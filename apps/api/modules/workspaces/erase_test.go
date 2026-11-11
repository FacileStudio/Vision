package workspaces

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/modules/auth"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte"
)

const (
	erasedEmail   = "zoe.erasable@facile.test"
	erasedName    = "Zoe Erasable"
	erasedSubject = "sub-zoe-erasable"
)

func (f *fixture) signInErasable() int64 {
	sqlDB, err := f.db.DB()
	f.must(err)
	users, err := auth.NewStore(context.Background(), sqlDB, EraseData)
	f.must(err)
	id, err := users.UpsertFromOIDC(context.Background(), porte.Claims{
		Provider: testIssuer, Subject: erasedSubject, Email: erasedEmail, EmailVerified: true,
		Name: erasedName, Picture: testIssuer + "/avatars/" + erasedSubject,
	})
	f.must(err)
	f.must(f.db.Exec(`INSERT INTO porte_identities (user_id, provider, subject, access_token, refresh_token)
		VALUES (?, ?, ?, 'at-'||?, 'rt-'||?)`, id, testIssuer, erasedSubject, erasedSubject, erasedSubject).Error)
	f.must(f.db.Exec(`INSERT INTO porte_sessions (token_hash, user_id, label) VALUES ('h-zoe', ?, ?)`, id, erasedName).Error)
	f.must(f.db.Exec(`INSERT INTO porte_login_codes (code_hash, user_id, expires_at, provider, subject)
		VALUES ('c-zoe', ?, now(), ?, ?)`, id, testIssuer, erasedSubject).Error)
	f.must(f.db.Exec(`INSERT INTO porte_magic_links (token_hash, user_id, created_at, expires_at)
		VALUES ('m-zoe', ?, now(), now())`, id).Error)
	return id
}

func (f *fixture) seedErasableData(zoe int64) {
	f.users["zoe"] = zoe
	own := f.workspace(erasedName, map[string]string{"zoe": "owner"})
	namedAfter := f.workspace(erasedEmail, map[string]string{"zoe": "owner", "carol": "editor"})
	f.must(f.db.Create(&schemas.WorkspaceMember{WorkspaceID: f.acme, UserID: zoe, Role: "editor"}).Error)
	space := "s-linked"
	linked := schemas.Workspace{Name: "Linked", RegistreSpaceID: &space}
	f.must(f.db.Create(&linked).Error)
	f.must(f.db.Create(&schemas.WorkspaceMember{WorkspaceID: linked.ID, UserID: zoe, Role: "editor"}).Error)

	mine := f.site(own, zoe, "zoe-own.test")
	f.must(f.db.Create(&schemas.CustomEvent{SiteID: mine, VisitorID: "v", Path: "/", Name: "signup"}).Error)
	f.must(f.db.Create(&schemas.VisitorSession{SiteID: mine, VisitorID: "v", EntryPath: "/"}).Error)
	f.must(f.db.Create(&schemas.APIKey{UserID: zoe, SiteID: &mine, Name: "own", KeyHash: "k-own"}).Error)
	f.site(namedAfter, zoe, "zoe-shared.test")
	f.site(f.acme, zoe, "zoe-acme.test")
	f.must(f.db.Create(&schemas.Webhook{OwnerID: zoe, WorkspaceID: &f.acme, URL: "https://hooks.test/zoe", Secret: "s", Period: "daily"}).Error)
	f.must(f.db.Create(&schemas.APIKey{UserID: zoe, WorkspaceID: &f.acme, Name: "acme", KeyHash: "k-acme"}).Error)
	f.must(f.db.Exec(`INSERT INTO mcp_actions (id, user_id, kind, target, digest, summary, expires_at)
		VALUES ('a-zoe', ?, 'delete_site', ?, 'd', ?, now() + interval '1 day')`,
		fmt.Sprint(zoe), fmt.Sprint(mine), "Delete the site of "+erasedName+"?").Error)
}

// traces counts, column by column, the rows that still name the person: any text, jsonb
// or uuid value containing their address, name or subject, and any user_id or owner_id
// equal to their account id.
func (f *fixture) traces(userID int64) map[string]int64 {
	var columns []struct{ Table, Column, Type string }
	f.must(f.db.Raw(`SELECT table_name AS "table", column_name AS "column", data_type AS "type"
		FROM information_schema.columns WHERE table_schema = current_schema()
		  AND (data_type IN ('text', 'character varying', 'character', 'jsonb', 'uuid')
		       OR column_name IN ('user_id', 'owner_id'))`).Scan(&columns).Error)
	out := map[string]int64{}
	for _, c := range columns {
		where := fmt.Sprintf(`%q::text ILIKE ANY (ARRAY[?, ?, ?])`, c.Column)
		args := []any{"%" + erasedEmail + "%", "%" + erasedName + "%", "%" + erasedSubject + "%"}
		if !strings.Contains("text character varying character jsonb uuid", c.Type) {
			where, args = fmt.Sprintf(`%q = ?`, c.Column), []any{userID}
		}
		var n int64
		f.must(f.db.Raw(fmt.Sprintf(`SELECT count(*) FROM %q WHERE `, c.Table)+where, args...).Scan(&n).Error)
		if n > 0 {
			out[c.Table+"."+c.Column] = n
		}
	}
	return out
}

func TestErasureLeavesNoTraceOfThePersonInAnyColumn(t *testing.T) {
	f := newFixture(t)
	zoe := f.signInErasable()
	f.seedErasableData(zoe)

	before := f.traces(zoe)
	for _, column := range []string{"users.email", "users.name", "porte_identities.subject", "workspaces.name",
		"porte_sessions.user_id", "workspace_members.user_id", "sites.owner_id", "goals.owner_id",
		"webhooks.owner_id", "api_keys.user_id", "porte_login_codes.subject", "porte_magic_links.user_id",
		"mcp_actions.summary"} {
		if before[column] == 0 {
			t.Errorf("the seed left nothing in %s, so the scan proves nothing there", column)
		}
	}

	if err := f.deletions().eraseSubject(context.Background(), testIssuer+"/", erasedSubject); err != nil {
		t.Fatalf("erase: %v", err)
	}

	if after := f.traces(zoe); len(after) > 0 {
		t.Errorf("rows still name the erased person: %v", after)
	}
	if n := f.count("sites", "domain = ?", "zoe-acme.test"); n != 1 {
		t.Errorf("the shared site went with the person")
	}
	if n := f.count("workspaces", "id = ?", f.acme); n != 1 {
		t.Errorf("the shared workspace went with the person")
	}
}
