package workspaces

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

const testIssuer = "https://sso.test"

var testConfig = testdb.Config{Prefix: "vision_workspaces_test", Migrate: schemas.Migrate}

// fixture is a database holding one shared workspace and one personal one, before
// Registre has ever been heard from.
type fixture struct {
	t        *testing.T
	db       *gorm.DB
	dir      *Directory
	acme     int64
	personal int64
	users    map[string]int64
}

// newFixture seeds Acme with an SSO owner, a local password editor, an SSO editor who
// will leave in Registre and an SSO viewer, plus Alice's personal workspace.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	db, err := testdb.Open(url, testConfig)
	if err != nil {
		t.Fatalf("open the test database: %v", err)
	}
	if err := testdb.Truncate(db, testConfig); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	f := &fixture{t: t, db: db, users: map[string]int64{}}
	f.dir = NewDirectory(db, testIssuer, "vision", "secret", slog.New(slog.DiscardHandler))
	for _, name := range []string{"alice", "bob", "carol", "dave", "vic"} {
		user := schemas.User{Email: name + "@facile.test", Name: name}
		f.must(db.Create(&user).Error)
		f.users[name] = user.ID
	}
	for _, name := range []string{"alice", "dave", "vic"} {
		f.must(db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`,
			f.users[name], testIssuer, "sub-"+name).Error)
	}
	f.acme = f.workspace("Acme", map[string]string{"alice": "owner", "carol": "editor", "dave": "editor", "vic": "viewer"})
	f.personal = f.workspace("Alice", map[string]string{"alice": "owner"})
	return f
}

// must fails the test on err.
func (f *fixture) must(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatal(err)
	}
}

// workspace creates an unlinked workspace with members at roles.
func (f *fixture) workspace(name string, roles map[string]string) int64 {
	ws := schemas.Workspace{Name: name}
	f.must(f.db.Create(&ws).Error)
	for user, role := range roles {
		f.must(f.db.Create(&schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: f.users[user], Role: role}).Error)
	}
	return ws.ID
}

// roles is every member of a workspace by email, at their role.
func (f *fixture) roles(workspaceID int64) map[string]string {
	var rows []struct{ Email, Role string }
	f.must(f.db.Raw(`SELECT u.email, wm.role FROM workspace_members wm JOIN users u ON u.id = wm.user_id
		WHERE wm.workspace_id = ?`, workspaceID).Scan(&rows).Error)
	out := map[string]string{}
	for _, row := range rows {
		out[row.Email] = row.Role
	}
	return out
}

// state is everything Apply may write, so two applies can be compared.
func (f *fixture) state() string {
	var rows []string
	f.must(f.db.Raw(`SELECT concat_ws('|', 'w', id, name, registre_space_id, updated_at) FROM workspaces
		UNION ALL SELECT concat_ws('|', 'm', id, workspace_id, user_id, role) FROM workspace_members
		UNION ALL SELECT concat_ws('|', 'u', id, email, oidc_picture_url) FROM users
		UNION ALL SELECT concat_ws('|', 'i', user_id, provider, subject) FROM porte_identities
		ORDER BY 1`).Scan(&rows).Error)
	b, _ := json.Marshal(rows)
	return string(b)
}

// registreSnapshot is Registre after the import: Acme renamed in case, Dave gone, Bob
// added by email, Erin new to Vision, and a space sharing Alice's personal workspace name.
func registreSnapshot() directory.Snapshot {
	erin := directory.Member{Subject: "sub-erin", Email: "erin@facile.test", Name: "Erin",
		Picture: testIssuer + "/avatars/sub-erin", Role: directory.RoleAdmin}
	acme := []directory.Member{
		{Subject: "sub-alice", Email: "alice@facile.test", Role: directory.RoleOwner},
		{Subject: "sub-bob", Email: "Bob@facile.test", Role: directory.RoleMember},
		erin,
	}
	alice := []directory.Member{{Subject: "sub-alice", Email: "alice@facile.test", Role: directory.RoleMember}}
	return directory.Snapshot{Version: 3, Spaces: []directory.Space{
		{ID: "s-acme", Name: "ACME", Members: acme},
		{ID: "s-alice", Name: "alice", Members: alice},
	}}
}

// Apply links Acme by name, keeps local and viewer members, removes the departed SSO
// member, pre-provisions the newcomer and never touches the personal workspace. The
// second apply writes nothing.
func TestApplyMirrorsRegistre(t *testing.T) {
	f := newFixture(t)
	snapshot := registreSnapshot()
	f.must(f.dir.Apply(context.Background(), snapshot))

	var acme schemas.Workspace
	f.must(f.db.First(&acme, f.acme).Error)
	if acme.Name != "ACME" || acme.RegistreSpaceID == nil || *acme.RegistreSpaceID != "s-acme" {
		t.Fatalf("acme = %q linked to %v, want ACME linked to s-acme", acme.Name, acme.RegistreSpaceID)
	}
	want := map[string]string{
		"alice@facile.test": "owner", "bob@facile.test": "editor", "erin@facile.test": "admin",
		"carol@facile.test": "editor", "vic@facile.test": "viewer",
	}
	if got := f.roles(f.acme); !maps.Equal(got, want) {
		t.Errorf("acme members = %v, want %v", got, want)
	}

	var erin schemas.User
	f.must(f.db.Where("email = ?", "erin@facile.test").First(&erin).Error)
	if erin.Name != "Erin" || erin.OIDCPictureURL != testIssuer+"/avatars/sub-erin" {
		t.Errorf("erin pre-provisioned as %q with picture %q", erin.Name, erin.OIDCPictureURL)
	}
	var identities int64
	f.must(f.db.Raw(`SELECT count(*) FROM porte_identities WHERE provider = ? AND subject IN ('sub-bob', 'sub-erin')`,
		testIssuer).Scan(&identities).Error)
	if identities != 2 {
		t.Errorf("bob and erin have %d porte identities, want 2", identities)
	}

	var personal schemas.Workspace
	f.must(f.db.First(&personal, f.personal).Error)
	if personal.RegistreSpaceID != nil || personal.Name != "Alice" {
		t.Errorf("the personal workspace was linked to %v as %q", personal.RegistreSpaceID, personal.Name)
	}
	var mirror schemas.Workspace
	f.must(f.db.Where("registre_space_id = ?", "s-alice").First(&mirror).Error)
	if got := f.roles(mirror.ID); !maps.Equal(got, map[string]string{"alice@facile.test": "editor"}) {
		t.Errorf("s-alice members = %v", got)
	}

	before := f.state()
	f.must(f.dir.Apply(context.Background(), snapshot))
	if after := f.state(); after != before {
		t.Errorf("a second apply changed the database:\nbefore %s\nafter  %s", before, after)
	}
}

// Run keeps retrying the import while Registre answers 404, and the first snapshot is
// pulled only after an import was accepted. The import carries the shared workspace
// without its viewer and never the personal one.
func TestImportLandsBeforeTheFirstSync(t *testing.T) {
	f := newFixture(t)
	fake := &fakeRegistre{synced: make(chan struct{}, 1)}
	server := httptest.NewServer(fake)
	defer server.Close()

	f.dir = NewDirectory(f.db, server.URL, "vision", "secret", slog.New(slog.DiscardHandler))
	f.dir.issuer = testIssuer
	f.dir.minBackoff, f.dir.maxBackoff = time.Millisecond, 2*time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go f.dir.Run(ctx)
	select {
	case <-fake.synced:
	case <-time.After(10 * time.Second):
		t.Fatal("no snapshot was ever pulled")
	}
	cancel()

	fake.mu.Lock()
	defer fake.mu.Unlock()
	calls, imported := fake.calls, fake.imported
	wantCalls := []string{"/api/directory/import", "/api/directory/import", "/api/directory/import",
		"/api/directory/snapshot"}
	if !slices.Equal(calls[:4], wantCalls) {
		t.Errorf("calls = %v, want %v first", calls, wantCalls)
	}
	if len(imported) != 1 || imported[0].Name != "Acme" || len(imported[0].Members) != 3 {
		t.Fatalf("imported = %+v, want Acme with its three non-viewer members", imported)
	}
	for _, member := range imported[0].Members {
		if member.Role == directory.RoleOwner && member.Subject != "sub-alice" {
			t.Errorf("the SSO owner was imported as %+v, want her subject", member)
		}
		if member.Subject == "" && member.Email != "carol@facile.test" {
			t.Errorf("a member without an identity was imported as %+v", member)
		}
	}
}

// fakeRegistre answers the directory API: 404 to the first two imports, as a Registre
// that predates it does, then the import and every snapshot.
type fakeRegistre struct {
	mu       sync.Mutex
	calls    []string
	imported []directory.ImportSpace
	synced   chan struct{}
}

// ServeHTTP records the call, then answers it.
func (f *fakeRegistre) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id, secret, _ := r.BasicAuth(); id != "vision" || secret != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	f.calls = append(f.calls, r.URL.Path)
	if r.URL.Path == "/api/directory/import" && len(f.calls) < 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.URL.Path == "/api/directory/import" {
		var payload struct{ Spaces []directory.ImportSpace }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f.imported = payload.Spaces
	} else {
		select {
		case f.synced <- struct{}{}:
		default:
		}
	}
	if err := json.NewEncoder(w).Encode(directory.Snapshot{Version: 1}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Once Registre manages spaces, a linked workspace refuses a rename and every workspace
// refuses a membership write, each with a 409 naming the Registre page. A personal
// workspace can still be renamed, and a newcomer in no workspace is given one.
func TestRegistreOwnsSharedWrites(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.must(f.dir.Apply(ctx, registreSnapshot()))
	service := NewService(f.db, testIssuer+"/")
	alice := strconv.FormatInt(f.users["alice"], 10)

	_, err := service.updateWorkspace(ctx, alice, strconv.FormatInt(f.acme, 10), "Renamed")
	wantConflict(t, err, testIssuer+"/spaces/s-acme")
	_, err = service.addMember(ctx, alice, strconv.FormatInt(f.personal, 10), "bob@facile.test", "editor")
	wantConflict(t, err, testIssuer+"/spaces")
	if _, err := service.updateWorkspace(ctx, alice, strconv.FormatInt(f.personal, 10), "Mine"); err != nil {
		t.Errorf("renaming a personal workspace: %v", err)
	}

	nina := schemas.User{Email: "nina@facile.test", Name: "Nina"}
	f.must(f.db.Create(&nina).Error)
	for range 2 {
		list, err := service.listWorkspaces(ctx, strconv.FormatInt(nina.ID, 10))
		if err != nil || len(list) != 1 || list[0].Name != "Nina" || list[0].ManageURL != "" {
			t.Fatalf("newcomer workspaces = %+v, %v; want one personal workspace", list, err)
		}
	}
}

// wantConflict fails unless err is a 409 whose message names url.
func wantConflict(t *testing.T, err error, url string) {
	t.Helper()
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) || apiErr.Code != "already_exists" || !strings.Contains(apiErr.Message, url) {
		t.Errorf("err = %v, want a conflict naming %s", err, url)
	}
}
