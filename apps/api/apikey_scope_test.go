package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/apikeys"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/Vision/apps/api/modules/webhooks"
	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var scopeTestConfig = testdb.Config{Prefix: "vision_scope_test", Migrate: schemas.Migrate}

type sessionStub struct{ userID int64 }

func (s sessionStub) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(w, request.WithContext(kit.WithIdentity(request.Context(), kit.Identity{UserID: s.userID, Email: "alice@facile.test"})))
	})
}

func (sessionStub) AuthenticateRequest(http.ResponseWriter, *http.Request) (int64, error) {
	return 0, nil
}

func (sessionStub) AuthenticateToken(http.ResponseWriter, *http.Request, string) (int64, error) {
	return 0, nil
}

type scopeFixture struct {
	router           http.Handler
	acme, beta       int64
	siteA1, siteA2   int64
	siteB            int64
	goalB            int64
	session, siteKey string
	roWorkspaceKey   string
	rwWorkspaceKey   string
	rwSiteKey        string
	roWideKey        string
	rwWideKey        string
}

func newScopeFixture(t *testing.T) *scopeFixture {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	db, err := testdb.Open(url, scopeTestConfig)
	if err != nil {
		t.Fatalf("open the test database: %v", err)
	}
	if err := testdb.Truncate(db, scopeTestConfig); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	alice := &schemas.User{Email: "alice@facile.test", Name: "alice"}
	acme := &schemas.Workspace{Name: "Acme"}
	beta := &schemas.Workspace{Name: "Beta"}
	for _, row := range []any{alice, acme, beta} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, ws := range []int64{acme.ID, beta.ID} {
		if err := db.Create(&schemas.WorkspaceMember{WorkspaceID: ws, UserID: alice.ID, Role: "owner"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	siteA1 := &schemas.Site{Name: "A1", Domain: "a1.test", OwnerID: alice.ID, WorkspaceID: acme.ID}
	siteA2 := &schemas.Site{Name: "A2", Domain: "a2.test", OwnerID: alice.ID, WorkspaceID: acme.ID}
	siteB := &schemas.Site{Name: "B", Domain: "b.test", OwnerID: alice.ID, WorkspaceID: beta.ID}
	for _, row := range []any{siteA1, siteA2, siteB} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	name := "signup"
	goalB := &schemas.Goal{SiteID: siteB.ID, OwnerID: alice.ID, Name: "B goal", GoalType: "event", EventName: &name}
	if err := db.Create(goalB).Error; err != nil {
		t.Fatal(err)
	}

	f := &scopeFixture{acme: acme.ID, beta: beta.ID, siteA1: siteA1.ID, siteA2: siteA2.ID, siteB: siteB.ID, goalB: goalB.ID}
	f.router = scopeRouter(db, alice.ID)
	f.roWorkspaceKey = f.mintKey(t, "read", nil, &acme.ID)
	f.rwWorkspaceKey = f.mintKey(t, "read,write", nil, &acme.ID)
	f.siteKey = f.mintKey(t, "read", &siteA1.ID, nil)
	f.rwSiteKey = f.mintKey(t, "read,write", &siteA1.ID, nil)
	f.roWideKey = f.mintKey(t, "read", nil, nil)
	f.rwWideKey = f.mintKey(t, "read,write", nil, nil)
	return f
}

func scopeRouter(db *gorm.DB, userID int64) http.Handler {
	authService := sessionStub{userID: userID}
	middleware.SetAPIKeyAuthenticator(apikeys.NewService(db))
	api := chi.NewRouter()
	tracker := events.NewActiveTracker()
	workspaces.RegisterRoutes(api, workspaces.NewService(db, ""), authService)
	sites.RegisterRoutes(api, sites.NewService(db), db, authService)
	analytics.RegisterRoutes(api, analytics.NewService(db), tracker, db, authService)
	goals.RegisterRoutes(api, goals.NewService(db), db, authService)
	apikeys.RegisterRoutes(api, apikeys.NewService(db), authService)
	webhooks.RegisterRoutes(api, webhooks.NewService(db), analytics.NewService(db), authService)
	return api
}

func (f *scopeFixture) do(method string, path string, key string, body any) *httptest.ResponseRecorder {
	var reader bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&reader).Encode(body)
	}
	request := httptest.NewRequest(method, path, &reader).WithContext(context.Background())
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("X-API-Key", key)
	}
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, request)
	return recorder
}

func (f *scopeFixture) mintKey(t *testing.T, scopes string, siteID *int64, workspaceID *int64) string {
	t.Helper()
	response := f.do(http.MethodPost, "/api-keys", "", map[string]any{"name": "k", "scopes": scopes, "site_id": siteID, "workspace_id": workspaceID})
	if response.Code != http.StatusCreated {
		t.Fatalf("mint key: %d %s", response.Code, response.Body.String())
	}
	var created struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	return created.Key
}

func ids(t *testing.T, response *httptest.ResponseRecorder, field string) []int64 {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var rows []map[string]any
	var wrapped struct {
		Sites []map[string]any `json:"sites"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &rows); err != nil {
		if err := json.Unmarshal(response.Body.Bytes(), &wrapped); err != nil {
			t.Fatal(err)
		}
		rows = wrapped.Sites
	}
	out := []int64{}
	for _, row := range rows {
		out = append(out, int64(row[field].(float64)))
	}
	return out
}

func expectIDs(t *testing.T, label string, got []int64, want ...int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", label, got, want)
	}
	seen := map[int64]bool{}
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			t.Fatalf("%s: got %v, want %v", label, got, want)
		}
	}
}

func expectStatus(t *testing.T, label string, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("%s: status %d, want %d: %s", label, response.Code, want, response.Body.String())
	}
}

func site(id int64) string { return "/sites/" + strconv.FormatInt(id, 10) }

func TestSiteKeyCannotReachAnotherSite(t *testing.T) {
	f := newScopeFixture(t)
	expectStatus(t, "own site", f.do(http.MethodGet, site(f.siteA1), f.siteKey, nil), http.StatusOK)
	expectStatus(t, "sibling site", f.do(http.MethodGet, site(f.siteA2), f.siteKey, nil), http.StatusForbidden)
	expectStatus(t, "other space site", f.do(http.MethodGet, site(f.siteB), f.siteKey, nil), http.StatusForbidden)
	expectStatus(t, "overview of site B", f.do(http.MethodGet, "/analytics/"+strconv.FormatInt(f.siteB, 10)+"/overview", f.siteKey, nil), http.StatusForbidden)
	expectStatus(t, "goals of site B", f.do(http.MethodGet, "/goals?site_id="+strconv.FormatInt(f.siteB, 10), f.siteKey, nil), http.StatusNotFound)
	expectStatus(t, "conversions of site B", f.do(http.MethodGet, "/goals/"+strconv.FormatInt(f.siteB, 10)+"/conversions", f.siteKey, nil), http.StatusForbidden)
	expectIDs(t, "site list", ids(t, f.do(http.MethodGet, "/sites", f.siteKey, nil), "id"), f.siteA1)
	expectIDs(t, "summary", ids(t, f.do(http.MethodGet, "/analytics/summary", f.siteKey, nil), "site_id"), f.siteA1)
	expectIDs(t, "spaces", ids(t, f.do(http.MethodGet, "/workspaces", f.siteKey, nil), "id"))
}

func TestWorkspaceKeySeesOnlyItsWorkspace(t *testing.T) {
	f := newScopeFixture(t)
	key := f.roWorkspaceKey
	expectStatus(t, "site in space", f.do(http.MethodGet, site(f.siteA2), key, nil), http.StatusOK)
	expectStatus(t, "site in other space", f.do(http.MethodGet, site(f.siteB), key, nil), http.StatusForbidden)
	expectStatus(t, "overview of other space", f.do(http.MethodGet, "/analytics/"+strconv.FormatInt(f.siteB, 10)+"/overview", key, nil), http.StatusForbidden)
	expectIDs(t, "site list", ids(t, f.do(http.MethodGet, "/sites", key, nil), "id"), f.siteA1, f.siteA2)
	expectIDs(t, "site list of other space", ids(t, f.do(http.MethodGet, "/sites?workspace_id="+strconv.FormatInt(f.beta, 10), key, nil), "id"))
	expectIDs(t, "summary", ids(t, f.do(http.MethodGet, "/analytics/summary", key, nil), "site_id"), f.siteA1, f.siteA2)
	expectIDs(t, "summary of other space", ids(t, f.do(http.MethodGet, "/analytics/summary?space="+strconv.FormatInt(f.beta, 10), key, nil), "site_id"))
	expectIDs(t, "spaces", ids(t, f.do(http.MethodGet, "/workspaces", key, nil), "id"), f.acme)
	expectStatus(t, "own space", f.do(http.MethodGet, "/workspaces/"+strconv.FormatInt(f.acme, 10), key, nil), http.StatusOK)
	expectStatus(t, "other space", f.do(http.MethodGet, "/workspaces/"+strconv.FormatInt(f.beta, 10), key, nil), http.StatusForbidden)
	expectStatus(t, "members", f.do(http.MethodGet, "/workspaces/"+strconv.FormatInt(f.acme, 10)+"/members", key, nil), http.StatusForbidden)
}

func TestReadKeysCannotWrite(t *testing.T) {
	f := newScopeFixture(t)
	body := map[string]string{"name": "renamed", "domain": "a1.test"}
	for label, key := range map[string]string{"site": f.siteKey, "workspace": f.roWorkspaceKey, "wide": f.roWideKey} {
		expectStatus(t, label+" key update", f.do(http.MethodPut, site(f.siteA1), key, body), http.StatusForbidden)
		expectStatus(t, label+" key create goal", f.do(http.MethodPost, "/goals", key, map[string]any{"site_id": f.siteA1, "name": "g", "goal_type": "event", "event_name": "x"}), http.StatusForbidden)
	}
}

func TestWriteKeysStayInsideTheirScope(t *testing.T) {
	f := newScopeFixture(t)
	body := map[string]string{"name": "renamed", "domain": "a1.test"}
	expectStatus(t, "site key writes its site", f.do(http.MethodPut, site(f.siteA1), f.rwSiteKey, body), http.StatusOK)
	expectStatus(t, "site key writes a sibling", f.do(http.MethodPut, site(f.siteA2), f.rwSiteKey, body), http.StatusForbidden)
	expectStatus(t, "workspace key writes its space", f.do(http.MethodPut, site(f.siteA2), f.rwWorkspaceKey, map[string]string{"name": "x", "domain": "a2.test"}), http.StatusOK)
	expectStatus(t, "workspace key writes another space", f.do(http.MethodPut, site(f.siteB), f.rwWorkspaceKey, map[string]string{"name": "x", "domain": "b.test"}), http.StatusForbidden)
	expectStatus(t, "workspace key goal on another space", f.do(http.MethodPost, "/goals", f.rwWorkspaceKey, map[string]any{"site_id": f.siteB, "name": "g", "goal_type": "event", "event_name": "x"}), http.StatusNotFound)
	expectStatus(t, "workspace key edits a goal of another space", f.do(http.MethodPut, "/goals/"+strconv.FormatInt(f.goalB, 10), f.rwWorkspaceKey, map[string]any{"name": "hijacked", "goal_type": "event", "event_name": "x"}), http.StatusNotFound)
	expectStatus(t, "workspace key deletes a goal of another space", f.do(http.MethodDelete, "/goals/"+strconv.FormatInt(f.goalB, 10), f.rwWorkspaceKey, nil), http.StatusNotFound)
	expectStatus(t, "workspace key creates a site in another space", f.do(http.MethodPost, "/sites", f.rwWorkspaceKey, map[string]any{"name": "n", "domain": "n.test", "workspace_id": f.beta}), http.StatusForbidden)
	expectStatus(t, "site key creates a site", f.do(http.MethodPost, "/sites", f.rwSiteKey, map[string]any{"name": "n", "domain": "n.test", "workspace_id": f.acme}), http.StatusForbidden)
}

func TestKeysCannotManageKeysWebhooksOrMembers(t *testing.T) {
	f := newScopeFixture(t)
	for label, key := range map[string]string{"site": f.rwSiteKey, "workspace": f.rwWorkspaceKey, "wide": f.rwWideKey} {
		expectStatus(t, label+" mints a key", f.do(http.MethodPost, "/api-keys", key, map[string]any{"name": "k"}), http.StatusForbidden)
		expectStatus(t, label+" lists keys", f.do(http.MethodGet, "/api-keys", key, nil), http.StatusForbidden)
		expectStatus(t, label+" lists webhooks", f.do(http.MethodGet, "/webhooks", key, nil), http.StatusForbidden)
		expectStatus(t, label+" adds a member", f.do(http.MethodPost, "/workspaces/"+strconv.FormatInt(f.acme, 10)+"/members", key, map[string]string{"email": "m@x.test", "role": "viewer"}), http.StatusForbidden)
	}
}

func TestUnscopedKeyReadsEverythingTheOwnerCan(t *testing.T) {
	f := newScopeFixture(t)
	expectIDs(t, "site list", ids(t, f.do(http.MethodGet, "/sites", f.roWideKey, nil), "id"), f.siteA1, f.siteA2, f.siteB)
	expectIDs(t, "summary", ids(t, f.do(http.MethodGet, "/analytics/summary", f.roWideKey, nil), "site_id"), f.siteA1, f.siteA2, f.siteB)
	expectStatus(t, "site", f.do(http.MethodGet, site(f.siteB), f.roWideKey, nil), http.StatusOK)
}

func TestSessionIsNotScoped(t *testing.T) {
	f := newScopeFixture(t)
	expectIDs(t, "site list", ids(t, f.do(http.MethodGet, "/sites", "", nil), "id"), f.siteA1, f.siteA2, f.siteB)
	expectIDs(t, "summary", ids(t, f.do(http.MethodGet, "/analytics/summary", "", nil), "site_id"), f.siteA1, f.siteA2, f.siteB)
	expectIDs(t, "spaces", ids(t, f.do(http.MethodGet, "/workspaces", "", nil), "id"), f.acme, f.beta)
	expectStatus(t, "update", f.do(http.MethodPut, site(f.siteB), "", map[string]string{"name": "x", "domain": "b.test"}), http.StatusOK)
	expectStatus(t, "mint a key", f.do(http.MethodGet, "/api-keys", "", nil), http.StatusOK)
	expectStatus(t, "webhooks", f.do(http.MethodGet, "/webhooks", "", nil), http.StatusOK)
}

func TestKeyCannotBeMintedForAForeignSpace(t *testing.T) {
	f := newScopeFixture(t)
	response := f.do(http.MethodPost, "/api-keys", "", map[string]any{"name": "k", "workspace_id": 999999})
	expectStatus(t, "foreign space", response, http.StatusNotFound)
}
