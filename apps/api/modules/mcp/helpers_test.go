package mcp_test

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/modules/mcp"
	"github.com/FacileStudio/Vision/apps/api/modules/sites"
	"github.com/FacileStudio/Vision/apps/api/modules/workspaces"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

var testConfig = testdb.Config{Prefix: "vision_mcp_test", Migrate: schemas.Migrate}

// identities resolves each test subject to its user.
type identities struct {
	porte.IdentityStore
	users map[string]int64
}

// Find resolves a known subject.
func (s identities) Find(_ context.Context, provider, sub string) (porte.StoredIdentity, error) {
	id, ok := s.users[sub]
	if !ok {
		return porte.StoredIdentity{}, porte.ErrNotFound
	}
	return porte.StoredIdentity{UserID: id, Provider: provider, Subject: sub}, nil
}

// fixture is the MCP endpoint over Vision's real services: an owner and a viewer of the
// Acme space with one tracked site, and a stranger outside it.
type fixture struct {
	db       *gorm.DB
	sqlDB    *sql.DB
	issuer   *mcptest.Issuer
	resource string
	users    map[string]int64
	acme     int64
	site     schemas.Site
	accounts *kit.Store
	front    func(http.Handler) http.Handler
}

// newFixture starts the endpoint and the confirm page routes on any free port.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureOn(t, "")
}

// newFixtureOn starts the endpoint on addr, any free port when empty.
func newFixtureOn(t *testing.T, addr string) *fixture {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	db, err := testdb.Open(url, testConfig)
	must(t, err)
	must(t, testdb.Truncate(db, testConfig))
	sqlDB, err := db.DB()
	must(t, err)
	accounts, err := auth.NewStore(context.Background(), sqlDB, workspaces.EraseData)
	must(t, err)
	f := &fixture{db: db, sqlDB: sqlDB, issuer: mcptest.NewIssuer(t), users: map[string]int64{}, accounts: accounts,
		front: func(next http.Handler) http.Handler { return next }}
	f.seed(t)
	var handler http.Handler
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f.front(handler).ServeHTTP(w, r) }))
	if addr != "" {
		listener, err := net.Listen("tcp", addr)
		must(t, err)
		srv.Listener = listener
	}
	srv.Start()
	t.Cleanup(srv.Close)
	f.resource = srv.URL + "/mcp"
	service := f.service(t)
	router := chi.NewRouter()
	mcp.RegisterRoutes(router, service)
	router.Route("/api", func(r chi.Router) { mcp.RegisterPendingRoutes(r, service, sessions{}) })
	handler = router
	return f
}

// service builds the MCP server over the fixture's database and issuer.
func (f *fixture) service(t *testing.T) *mcp.Service {
	t.Helper()
	verifier, err := mcpauth.Verifier(context.Background(), mcpauth.Config{
		Issuer: f.issuer.URL, Resource: f.resource, ScopePrefix: "vision", Identities: identities{users: f.users},
	})
	must(t, err)
	logger := slog.New(slog.DiscardHandler)
	live := events.NewActiveTracker()
	live.Touch(f.site.ID, "visitor-now")
	service, err := mcp.NewService(mcp.Deps{
		Workspaces: workspaces.NewService(f.db, ""), Sites: sites.NewService(f.db), Analytics: analytics.NewService(f.db),
		Goals: goals.NewService(f.db), Live: live, Accounts: f.accounts, DB: f.db,
	}, mcp.Config{
		Resource: f.resource, Issuer: f.issuer.URL, Version: "test", StateKey: []byte("state-key"),
		Verifier: verifier, Pending: mcpkit.PostgresPending{DB: f.sqlDB}, PerMinute: 100_000, Logger: logger,
	})
	must(t, err)
	return service
}

// seed creates the people, the Acme space with its site, and three pageviews today.
func (f *fixture) seed(t *testing.T) {
	t.Helper()
	for _, name := range []string{"owner", "viewer", "stranger"} {
		user := schemas.User{Email: name + "@example.test", Name: name}
		must(t, f.db.Create(&user).Error)
		f.users[name] = user.ID
	}
	acme := schemas.Workspace{Name: "Acme"}
	must(t, f.db.Create(&acme).Error)
	f.acme = acme.ID
	for name, role := range map[string]string{"owner": "owner", "viewer": "viewer"} {
		must(t, f.db.Create(&schemas.WorkspaceMember{WorkspaceID: acme.ID, UserID: f.users[name], Role: role}).Error)
	}
	f.site = schemas.Site{Name: "Acme site", Domain: "acme.test", OwnerID: f.users["owner"], WorkspaceID: acme.ID}
	must(t, f.db.Create(&f.site).Error)
	for i, path := range []string{"/", "/pricing", "/thanks"} {
		must(t, f.db.Create(&schemas.Pageview{
			SiteID: f.site.ID, Path: path, Referrer: "news.example", Country: "FR", Browser: "Firefox", OS: "Linux",
			Device: "Desktop", VisitorID: "v" + strconv.Itoa(i%2), CreatedAt: time.Now().Add(-time.Hour),
		}).Error)
	}
}

// token is a token for the owner with scopes, or both when none is given.
func (f *fixture) token(t testing.TB, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = []string{"vision.read", "vision.write"}
	}
	return f.issuer.Token(t, f.resource, "owner", scopes...)
}

// connect opens a session for subject with both scopes.
func (f *fixture) connect(t *testing.T, subject string, opts mcptest.ClientOptions) *sdk.ClientSession {
	t.Helper()
	return mcptest.Connect(t, f.resource, f.issuer.Token(t, f.resource, subject, "vision.read", "vision.write"), opts)
}

// newSite adds a site to Acme and returns it.
func (f *fixture) newSite(t testing.TB, domain string) schemas.Site {
	t.Helper()
	site := schemas.Site{Name: domain, Domain: domain, OwnerID: f.users["owner"], WorkspaceID: f.acme}
	must(t, f.db.Create(&site).Error)
	return site
}

// newGoal adds an owner's pageview goal to the site.
func (f *fixture) newGoal(t testing.TB, name string) schemas.Goal {
	t.Helper()
	path := "/thanks"
	goal := schemas.Goal{SiteID: f.site.ID, OwnerID: f.users["owner"], Name: name, GoalType: "pageview", PagePath: &path, MatchType: "exact"}
	must(t, f.db.Create(&goal).Error)
	return goal
}

// count is the number of rows of table matching where.
func (f *fixture) count(table, where string, args ...any) int {
	var n int64
	f.db.Table(table).Where(where, args...).Count(&n)
	return int(n)
}

// call runs a tool and decodes its result, failing on a tool error.
func call(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if result.IsError {
		t.Fatalf("%s: %s", tool, mcptest.Text(result))
	}
	mcptest.Decode(t, result, out)
}

// refused runs a tool and returns its error text, failing when it worked.
func refused(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if !result.IsError {
		t.Fatalf("%s must be refused, got %s", tool, mcptest.Text(result))
	}
	return mcptest.Text(result)
}

// must fails the test on err.
func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// sessions stands in for porte on the confirm page: X-Test-User is the signed-in user,
// X-Test-Actor makes the request a delegated token, and X-Test-Session: cli a session a
// CLI login minted, which carries no browser mark.
type sessions struct{}

// RequireAuth attaches the test identity, as porte/kit's middleware does.
func (sessions) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.Header.Get("X-Test-User"), 10, 64)
		if err != nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		identity := porte.Identity{UserID: id, SessionID: 1, Actor: r.Header.Get("X-Test-Actor"), BrowserSession: true}
		if identity.Actor != "" {
			identity.SessionID, identity.BrowserSession = 0, false
		}
		if r.Header.Get("X-Test-Session") == "cli" {
			identity.BrowserSession = false
		}
		ctx := kit.WithIdentity(porte.WithIdentity(r.Context(), identity), kit.Identity{UserID: id})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthenticateRequest is unused by the confirm page.
func (sessions) AuthenticateRequest(http.ResponseWriter, *http.Request) (int64, error) {
	return 0, porte.ErrNotFound
}

// AuthenticateToken is unused by the confirm page.
func (sessions) AuthenticateToken(http.ResponseWriter, *http.Request, string) (int64, error) {
	return 0, porte.ErrNotFound
}
