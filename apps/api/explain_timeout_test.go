package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/ideas"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
)

const slowBulleDelay = 20 * time.Second

type porteSession struct{ userID int64 }

func (s porteSession) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: s.userID})
		next.ServeHTTP(w, r.WithContext(kit.WithIdentity(ctx, kit.Identity{UserID: s.userID})))
	})
}

// TestExplainSurvivesASlowBulle runs the explain route through the production router and
// the production server timeouts, against a Bulle that answers after 20 seconds. The
// server's write timeout is 15 seconds, so the answer arrives only if the handler
// extended its own deadline through every response wrapper in between.
func TestExplainSurvivesASlowBulle(t *testing.T) {
	if testing.Short() {
		t.Skip("waits 20 seconds on purpose")
	}
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	cfg := testdb.Config{Prefix: "vision_timeout_test", Migrate: schemas.Migrate}
	db, err := testdb.Open(url, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.Truncate(db, cfg); err != nil {
		t.Fatal(err)
	}
	user := schemas.User{Email: "alice@facile.test", Name: "alice"}
	workspace := schemas.Workspace{Name: "Solo"}
	for _, row := range []any{&user, &workspace} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	site := schemas.Site{Name: "S", Domain: "s.test", OwnerID: user.ID, WorkspaceID: workspace.ID}
	if err := db.Create(&schemas.WorkspaceMember{WorkspaceID: workspace.ID, UserID: user.ID, Role: "owner"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&site).Error; err != nil {
		t.Fatal(err)
	}
	var anomaly int64
	err = db.Raw(`INSERT INTO anomalies (site_id, metric, goal_id, day, direction, value, baseline, mad, score, weeks, breakdown)
		VALUES (?, 'pageviews', 0, '2026-11-15', 'up', 240, 40, 2, 20, 8, '{}') RETURNING id`, site.ID).Scan(&anomaly).Error
	if err != nil {
		t.Fatal(err)
	}

	bulle := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(slowBulleDelay)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"json":{"summary":"Slow but here.","likely_causes":[],"what_to_check":[]},"model":"m"}`)
	}))
	defer bulle.Close()

	service := ideas.NewService(db, ideas.Config{
		URL: bulle.URL, UpgradeURL: "https://sso.test/spaces",
		Token:     func(context.Context) (string, error) { return "t", nil },
		Analytics: analytics.NewService(db),
	})
	api := chi.NewRouter()
	if err := ideas.RegisterRoutes(api, service, porteSession{userID: user.ID}); err != nil {
		t.Fatal(err)
	}
	router := httpx.NewRouter(httpx.Config{})
	router.Mount("/api", api)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := newServer("", router)
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close() }()

	client := &http.Client{Timeout: 60 * time.Second}
	started := time.Now()
	response, err := client.Post("http://"+listener.Addr().String()+"/api/anomalies/"+strconv.FormatInt(anomaly, 10)+"/explain", "application/json", nil)
	if err != nil {
		t.Fatalf("the answer was cut after %s: %v", time.Since(started), err)
	}
	defer func() { _ = response.Body.Close() }()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK || time.Since(started) < slowBulleDelay {
		t.Fatalf("status %d after %s: %s", response.StatusCode, time.Since(started), body)
	}
}
