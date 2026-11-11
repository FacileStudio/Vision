package workspaces

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"github.com/go-chi/chi/v5"
)

func (f *fixture) count(table, where string, args ...any) int64 {
	var n int64
	f.must(f.db.Raw("SELECT count(*) FROM "+table+" WHERE "+where, args...).Scan(&n).Error)
	return n
}

func (f *fixture) deletions() *Deletions {
	sqlDB, err := f.db.DB()
	f.must(err)
	users, err := auth.NewStore(context.Background(), sqlDB, EraseData)
	f.must(err)
	return NewDeletions(f.db, directory.Client{Issuer: testIssuer}, users.EraseAccount, nil, slog.New(slog.DiscardHandler))
}

func (d *Deletions) eraseSubject(ctx context.Context, provider, subject string) error {
	return d.receiver.OnPersonDeleted(ctx, provider, subject)
}

func (f *fixture) site(workspaceID, ownerID int64, domain string) int64 {
	site := schemas.Site{Name: domain, Domain: domain, OwnerID: ownerID, WorkspaceID: workspaceID}
	f.must(f.db.Create(&site).Error)
	f.must(f.db.Create(&schemas.Pageview{SiteID: site.ID, Path: "/", VisitorID: "v"}).Error)
	f.must(f.db.Create(&schemas.Goal{SiteID: site.ID, OwnerID: ownerID, Name: "g", GoalType: "page"}).Error)
	return site.ID
}

func TestErasedPersonLosesTheirDataAndSharedDataStaysWithTheNextMember(t *testing.T) {
	f := newFixture(t)
	aliceID := f.users["alice"]
	mine := f.site(f.personal, aliceID, "mine.test")
	shared := f.site(f.acme, aliceID, "shared.test")
	f.must(f.db.Create(&schemas.Webhook{OwnerID: aliceID, WorkspaceID: &f.acme, URL: "https://h.test", Period: "daily"}).Error)
	f.must(f.db.Create(&schemas.APIKey{UserID: aliceID, Name: "k", KeyHash: "h1"}).Error)

	d := f.deletions()
	if err := d.eraseSubject(context.Background(), testIssuer, "sub-alice"); err != nil {
		t.Fatalf("erase: %v", err)
	}

	if n := f.count("users", "id = ?", aliceID); n != 0 {
		t.Errorf("the account survived")
	}
	if n := f.count("porte_identities", "user_id = ?", aliceID); n != 0 {
		t.Errorf("the identity survived")
	}
	for table, where := range map[string]string{
		"sites": "id = ?", "pageviews": "site_id = ?", "goals": "site_id = ?",
	} {
		if n := f.count(table, where, mine); n != 0 {
			t.Errorf("personal %s rows survived: %d", table, n)
		}
	}
	if n := f.count("workspaces", "id = ?", f.personal); n != 0 {
		t.Errorf("the personal workspace survived")
	}
	if n := f.count("webhooks", "owner_id = ?", aliceID) + f.count("api_keys", "user_id = ?", aliceID); n != 0 {
		t.Errorf("personal webhooks or keys survived")
	}
	if n := f.count("pageviews", "site_id = ?", shared); n != 1 {
		t.Errorf("the shared site lost its pageviews")
	}
	var owner int64
	f.must(f.db.Raw(`SELECT owner_id FROM sites WHERE id = ?`, shared).Scan(&owner).Error)
	if owner != f.users["carol"] && owner != f.users["dave"] {
		t.Errorf("the shared site went to %d, want an editor of Acme", owner)
	}
	if n := f.count("goals", "site_id = ? AND owner_id = ?", shared, owner); n != 1 {
		t.Errorf("the shared goal was not handed to the site's new owner")
	}
	if n := f.count("workspace_members", "user_id = ?", aliceID); n != 0 {
		t.Errorf("memberships survived")
	}
}

func TestErasingTwiceIsANoOp(t *testing.T) {
	f := newFixture(t)
	f.site(f.acme, f.users["alice"], "shared.test")
	d := f.deletions()
	if err := d.eraseSubject(context.Background(), testIssuer, "sub-alice"); err != nil {
		t.Fatal(err)
	}
	before := f.state()
	if err := d.eraseSubject(context.Background(), testIssuer, "sub-alice"); err != nil {
		t.Fatalf("second erase: %v", err)
	}
	if err := d.eraseSubject(context.Background(), testIssuer, "sub-nobody"); err != nil {
		t.Fatalf("unknown subject: %v", err)
	}
	if after := f.state(); after != before {
		t.Errorf("a repeated erase changed the database")
	}
}

func TestADeletedSoleMemberOfALinkedWorkspaceLeavesTheSitesOwnerless(t *testing.T) {
	f := newFixture(t)
	space := "s-solo"
	ws := schemas.Workspace{Name: "Solo", RegistreSpaceID: &space}
	f.must(f.db.Create(&ws).Error)
	f.must(f.db.Create(&schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: f.users["dave"], Role: "owner"}).Error)
	site := f.site(ws.ID, f.users["dave"], "solo.test")

	if err := f.deletions().eraseSubject(context.Background(), testIssuer, "sub-dave"); err != nil {
		t.Fatal(err)
	}
	if n := f.count("sites", "id = ? AND owner_id = 0", site); n != 1 {
		t.Errorf("the space's site was deleted or kept an owner")
	}
}

func TestDeletionPushRouteSitsOutsideTheSessionMiddleware(t *testing.T) {
	f := newFixture(t)
	router := chi.NewRouter()
	RegisterDeletions(router, f.deletions())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/registre/deletions", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("an unsigned push answered %d, want 400 from the verifier", recorder.Code)
	}
}

func TestDeletionStateRoundTrips(t *testing.T) {
	f := newFixture(t)
	state := deletionState{orm: f.db}
	if got, err := state.Load(context.Background()); err != nil || got != nil {
		t.Fatalf("empty load: %q %v", got, err)
	}
	for _, blob := range []string{`{"a":1}`, `{"a":2}`} {
		if err := state.Save(context.Background(), []byte(blob)); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := state.Load(context.Background()); string(got) != `{"a":2}` {
		t.Errorf("loaded %q", got)
	}
}
