package mcp_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// pendingList is the confirm page's list as the client reads it.
type pendingList struct {
	Actions []struct {
		ID     string `json:"id"`
		Kind   string `json:"kind"`
		State  string `json:"state"`
		Target *struct {
			Type   string `json:"type"`
			Name   string `json:"name"`
			Domain string `json:"domain"`
		} `json:"target"`
	} `json:"actions"`
}

// page calls a confirm page route as user, optionally through a delegated token.
func (f *fixture) page(t *testing.T, method, path, user, actor string, out any) int {
	t.Helper()
	origin := strings.TrimSuffix(f.resource, "/mcp")
	req, _ := http.NewRequest(method, origin+"/api/mcp/pending"+path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", strconv.FormatInt(f.users[user], 10))
	if actor != "" {
		req.Header.Set("X-Test-Actor", actor)
	}
	resp, err := http.DefaultClient.Do(req)
	must(t, err)
	defer func() { _ = resp.Body.Close() }()
	if out != nil && resp.StatusCode == http.StatusOK {
		must(t, json.NewDecoder(resp.Body).Decode(out))
	}
	return resp.StatusCode
}

// TestPendingPageOpensAPublicLinkOnlyFromTheSession follows a client that cannot ask: the
// public link waits on /pending, an agent's token cannot confirm it, another person does
// not see it, and the owner's session opens it once.
func TestPendingPageOpensAPublicLinkOnlyFromTheSession(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner", mcptest.ClientOptions{})
	var asked siteResult
	call(t, session, "vision_share_site", map[string]any{"site_id": f.site.ID}, &asked)
	if asked.Status != "pending_confirmation" || !strings.Contains(asked.ConfirmURL, "/pending?action=") {
		t.Fatalf("want a confirm link, got %+v", asked)
	}
	if f.count("sites", "id = ? AND share_token IS NOT NULL", f.site.ID) != 0 {
		t.Fatal("nothing is shared before the person confirms")
	}
	var list pendingList
	if code := f.page(t, http.MethodGet, "/", "owner", "", &list); code != http.StatusOK || len(list.Actions) != 1 {
		t.Fatalf("the owner sees one action, got %d %+v", code, list)
	}
	action := list.Actions[0]
	if action.Kind != "share_site" || action.State != "ready" || action.Target == nil || action.Target.Domain != "acme.test" {
		t.Fatalf("the action must show the site as it is, got %+v", action)
	}
	if code := f.page(t, http.MethodPost, "/"+action.ID+"/confirm", "owner", "claude-code", nil); code != http.StatusForbidden {
		t.Fatalf("a delegated token must not confirm, got %d", code)
	}
	var other pendingList
	f.page(t, http.MethodGet, "/", "viewer", "", &other)
	if len(other.Actions) != 0 {
		t.Fatalf("another member must not see the owner's action, got %+v", other)
	}
	var done struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	}
	if code := f.page(t, http.MethodPost, "/"+action.ID+"/confirm", "owner", "", &done); code != http.StatusOK || done.Status != "shared" || !strings.Contains(done.URL, "/share/") {
		t.Fatalf("the confirm must open the link and return it, got %d %+v", code, done)
	}
	if code := f.page(t, http.MethodPost, "/"+action.ID+"/confirm", "owner", "", nil); code == http.StatusOK {
		t.Fatal("a confirmation runs once")
	}
}

// TestPendingPageRefusesAChangedSite closes a deletion when the site changed after the
// request, instead of deleting something the person did not see.
func TestPendingPageRefusesAChangedSite(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner", mcptest.ClientOptions{})
	var asked siteResult
	call(t, session, "vision_delete_site", map[string]any{"site_id": f.site.ID}, &asked)
	var renamed siteResult
	call(t, session, "vision_update_site", map[string]any{"site_id": f.site.ID, "name": "Renamed"}, &renamed)
	var list pendingList
	f.page(t, http.MethodGet, "/", "owner", "", &list)
	if len(list.Actions) != 1 || list.Actions[0].State != "changed" {
		t.Fatalf("the action must read changed, got %+v", list)
	}
	if code := f.page(t, http.MethodPost, "/"+list.Actions[0].ID+"/confirm", "owner", "", nil); code != http.StatusPreconditionFailed {
		t.Fatalf("a changed site must be refused with 412, got %d", code)
	}
	if f.count("sites", "id = ?", f.site.ID) != 1 {
		t.Fatal("the site must stay")
	}
}
