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
)

var testConfig = testdb.Config{Prefix: "vision_events_test", Migrate: schemas.Migrate}

func ingestRouter(t *testing.T) http.Handler {
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
	return router
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
