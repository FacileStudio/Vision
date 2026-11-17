package events

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/httpx"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var testConfig = testdb.Config{Prefix: "vision_events_test", Migrate: schemas.Migrate}

func ingestRouter(t *testing.T) http.Handler {
	t.Helper()
	handler, _ := ingestRouterWithDB(t)
	return handler
}

func ingestRouterWithDB(t *testing.T) (http.Handler, *gorm.DB) {
	t.Helper()
	dbURL, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	db, err := testdb.Open(dbURL, testConfig)
	if err != nil {
		t.Fatalf("open the test database: %v", err)
	}
	if err := testdb.Truncate(db, testConfig); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	if err := db.Create(&schemas.Site{Name: "Blog", Domain: "blog.example", WorkspaceID: 1}).Error; err != nil {
		t.Fatalf("seed the site: %v", err)
	}
	router := httpx.NewRouter(httpx.Config{
		APIPrefix: httpx.RootAPI,
		CORS:      troncmiddleware.CORSConfig{AllowedOrigins: []string{"https://vision.test"}},
	})
	RegisterRoutes(router, NewService(db), NewHub(), NewActiveTracker(), nil, db)
	return router, db
}

func ingest(handler http.Handler, method, target, origin, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Origin", origin)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertCrossOriginSafe(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	header := recorder.Header()
	if header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("a tracked site cannot read the answer: ACAO %q", header.Get("Access-Control-Allow-Origin"))
	}
	if header.Get("Cross-Origin-Resource-Policy") != "" {
		t.Errorf("CORP %q would block the pixel on a third-party page", header.Get("Cross-Origin-Resource-Policy"))
	}
	if header.Get("Content-Security-Policy") == "" {
		t.Error("the chassis security headers are missing")
	}
}

func TestPixelIngestStaysOpenToTrackedSitesUnderTroncSecurityHeaders(t *testing.T) {
	handler := ingestRouter(t)
	data := url.QueryEscape(`{"hostname":"blog.example","path":"/a","visitor_id":"v1"}`)
	recorder := ingest(handler, http.MethodGet, "/e/p?data="+data, "https://blog.example", "")
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "image/gif" {
		t.Fatalf("pixel: %d %s", recorder.Code, recorder.Header().Get("Content-Type"))
	}
	assertCrossOriginSafe(t, recorder)
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("nosniff is missing, so the content type is not the contract")
	}
}

func TestJSONIngestStaysOpenToTrackedSitesUnderTroncSecurityHeaders(t *testing.T) {
	handler := ingestRouter(t)
	body := `{"hostname":"blog.example","path":"/a","visitor_id":"v1"}`
	recorder := ingest(handler, http.MethodPost, "/e/p", "https://blog.example", body)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("ingest: %d %s", recorder.Code, recorder.Body.String())
	}
	assertCrossOriginSafe(t, recorder)
}

func TestIngestRefusesAnUnregisteredOrigin(t *testing.T) {
	handler := ingestRouter(t)
	body := `{"hostname":"evil.example","path":"/a","visitor_id":"v1"}`
	if recorder := ingest(handler, http.MethodPost, "/e/p", "https://evil.example", body); recorder.Code == http.StatusNoContent {
		t.Error("an unregistered domain inflated somebody's numbers")
	}
}

func TestPreflightFromATrackedSiteIsAnsweredBeforeTheRouterCORS(t *testing.T) {
	router := httpx.NewRouter(httpx.Config{
		CORS: troncmiddleware.CORSConfig{AllowedOrigins: []string{"https://vision.test"}},
	})
	router.Route("/api", func(api chi.Router) {
		RegisterRoutes(api, nil, nil, nil, nil, nil)
	})

	for _, target := range []string{"/api/e/p", "/api/e/t", "/api/e/h"} {
		if bare := ingest(router, http.MethodOptions, target, "https://blog.example", ""); bare.Code != http.StatusForbidden {
			t.Fatalf("the router CORS answered %s with %d; the wrapper is no longer needed", target, bare.Code)
		}
		recorder := ingest(Preflight(router, "/api/e/"), http.MethodOptions, target, "https://blog.example", "")
		if recorder.Code != http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("preflight %s: %d, ACAO %q", target, recorder.Code, recorder.Header().Get("Access-Control-Allow-Origin"))
		}
	}

	if recorder := ingest(Preflight(router, "/api/e/"), http.MethodOptions, "/api/sites", "https://blog.example", ""); recorder.Code != http.StatusForbidden {
		t.Errorf("the wrapper answered a preflight outside the tracker: %d", recorder.Code)
	}
}

func TestTrackerAnswersCarryTheHeadersTheOldProxyAdded(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, nil, nil, nil, nil, nil)

	recorder := ingest(router, http.MethodGet, "/e/p", "https://blog.example", "")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("GET /e/p without data: %d", recorder.Code)
	}
	header := recorder.Header()
	if header.Get("Content-Type") != "application/json" || header.Get("Cache-Control") != "no-cache, no-store" || header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("headers: %v", header)
	}
}

func ingestAs(handler http.Handler, userAgent, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/e/p", strings.NewReader(body))
	request.Header.Set("Origin", "https://blog.example")
	request.Header.Set("User-Agent", userAgent)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestBotPageviewsCountAsBotHitsNotPageviews(t *testing.T) {
	handler, db := ingestRouterWithDB(t)
	gptBot := "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)"
	body := `{"hostname":"blog.example","path":"/a","visitor_id":"v1"}`
	for range 2 {
		if recorder := ingestAs(handler, gptBot, body); recorder.Code != http.StatusNoContent {
			t.Fatalf("bot ingest: %d %s", recorder.Code, recorder.Body.String())
		}
	}
	var pageviews int64
	db.Model(&schemas.Pageview{}).Count(&pageviews)
	var hits []schemas.BotHit
	db.Find(&hits)
	if pageviews != 0 {
		t.Errorf("a bot made %d pageviews", pageviews)
	}
	if len(hits) != 1 || hits[0].Hits != 2 || hits[0].Category != "ai" || hits[0].Name != "GPTBot" || hits[0].Path != "/a" {
		t.Fatalf("bot hits: %+v", hits)
	}

	chrome := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	if recorder := ingestAs(handler, chrome, body); recorder.Code != http.StatusNoContent {
		t.Fatalf("human ingest: %d %s", recorder.Code, recorder.Body.String())
	}
	db.Model(&schemas.Pageview{}).Count(&pageviews)
	if pageviews != 1 {
		t.Errorf("a human made %d pageviews", pageviews)
	}
}

func TestHeartbeatMarksTheSessionInteracted(t *testing.T) {
	handler, db := ingestRouterWithDB(t)
	chrome := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	ingestAs(handler, chrome, `{"hostname":"blog.example","path":"/a","visitor_id":"v1"}`)
	data := url.QueryEscape(`{"hostname":"blog.example","visitor_id":"v1","interacted":true}`)
	request := httptest.NewRequest(http.MethodGet, "/e/h?data="+data, nil)
	request.Header.Set("User-Agent", chrome)
	handler.ServeHTTP(httptest.NewRecorder(), request)
	var session schemas.VisitorSession
	if err := db.Where("visitor_id = ?", "v1").First(&session).Error; err != nil {
		t.Fatalf("session: %v", err)
	}
	if !session.Interacted {
		t.Error("the heartbeat did not mark the session interacted")
	}
}
