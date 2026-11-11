package mcp_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// siteResult is the part of a site tool result the tests read.
type siteResult struct {
	Status string `json:"status"`
	Site   struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Domain   string `json:"domain"`
		ShareURL string `json:"share_url"`
	} `json:"site"`
	TrackingSnippet string `json:"tracking_snippet"`
	ConfirmURL      string `json:"confirm_url"`
}

// TestMCPConformance runs tronc's conformance kit with the three confirmed tools.
func TestMCPConformance(t *testing.T) {
	f := newFixture(t)
	n := 0
	fresh := func(testing.TB) string { n++; return "fresh-" + strconv.Itoa(n) + ".test" }
	mcptest.Run(t, mcptest.Config{
		URL:   f.resource,
		Token: f.token,
		Confirm: []mcptest.ConfirmCase{
			{Tool: "vision_delete_site", Setup: func(tb testing.TB) (map[string]any, func() int) {
				site := f.newSite(tb, fresh(tb))
				return map[string]any{"site_id": site.ID}, func() int { return 1 - f.count("sites", "id = ?", site.ID) }
			}},
			{Tool: "vision_share_site", Setup: func(tb testing.TB) (map[string]any, func() int) {
				site := f.newSite(tb, fresh(tb))
				return map[string]any{"site_id": site.ID}, func() int { return f.count("sites", "id = ? AND share_token IS NOT NULL", site.ID) }
			}},
			{Tool: "vision_delete_goal", Setup: func(tb testing.TB) (map[string]any, func() int) {
				goal := f.newGoal(tb, "Fresh goal")
				return map[string]any{"goal_id": goal.ID}, func() int { return 1 - f.count("goals", "id = ?", goal.ID) }
			}},
		},
	})
}

// TestMCPReadsWhatTheDashboardShows walks a model's usual read path: spaces, sites, the
// site's numbers in both formats, and its goals with conversions.
func TestMCPReadsWhatTheDashboardShows(t *testing.T) {
	f := newFixture(t)
	f.newGoal(t, "Thanks page")
	session := f.connect(t, "viewer", mcptest.ClientOptions{})
	var spaces struct {
		Spaces []struct {
			ID    int64  `json:"id"`
			Role  string `json:"role"`
			Sites int64  `json:"sites"`
		} `json:"spaces"`
	}
	call(t, session, "vision_list_spaces", nil, &spaces)
	if len(spaces.Spaces) != 1 || spaces.Spaces[0].Role != "viewer" || spaces.Spaces[0].Sites != 1 {
		t.Fatalf("the viewer sees Acme only, got %+v", spaces)
	}
	var list struct {
		Sites []struct {
			ID     int64  `json:"id"`
			Domain string `json:"domain"`
			Space  string `json:"space"`
		} `json:"sites"`
	}
	call(t, session, "vision_list_sites", map[string]any{"query": "ACME"}, &list)
	if len(list.Sites) != 1 || list.Sites[0].ID != f.site.ID || list.Sites[0].Space != "Acme" {
		t.Fatalf("the query must find the Acme site, got %+v", list)
	}
	checkStats(t, f, session)
	var goals struct {
		TotalVisitors int64 `json:"total_visitors"`
		Goals         []struct {
			Name        string `json:"name"`
			PagePath    string `json:"page_path"`
			Conversions int64  `json:"conversions"`
		} `json:"goals"`
	}
	call(t, session, "vision_list_goals", map[string]any{"site_id": f.site.ID}, &goals)
	if goals.TotalVisitors != 2 || len(goals.Goals) != 1 || goals.Goals[0].Conversions != 1 || goals.Goals[0].PagePath != "/thanks" {
		t.Fatalf("one of two visitors reached /thanks, got %+v", goals)
	}
}

// checkStats reads the seeded traffic, concise then detailed, and refuses a bad date.
func checkStats(t *testing.T, f *fixture, s *sdk.ClientSession) {
	t.Helper()
	var stats struct {
		Pageviews    int64 `json:"pageviews"`
		Visitors     int64 `json:"visitors"`
		LiveVisitors int64 `json:"live_visitors"`
		TopPages     []struct {
			Value string `json:"value"`
		} `json:"top_pages"`
		Detailed *struct {
			PageviewsPerDay []struct{ Count int64 } `json:"pageviews_per_day"`
		} `json:"detailed"`
	}
	call(t, s, "vision_get_stats", map[string]any{"site_id": f.site.ID}, &stats)
	if stats.Pageviews != 3 || stats.Visitors != 2 || stats.LiveVisitors != 1 || len(stats.TopPages) != 3 || stats.Detailed != nil {
		t.Fatalf("want 3 pageviews by 2 visitors and one live, concise, got %+v", stats)
	}
	call(t, s, "vision_get_stats", map[string]any{"site_id": f.site.ID, "response_format": "detailed", "path": "/pricing"}, &stats)
	if stats.Pageviews != 1 || stats.Detailed == nil || len(stats.Detailed.PageviewsPerDay) == 0 {
		t.Fatalf("the path filter and the detailed series must apply, got %+v", stats)
	}
	if text := refused(t, s, "vision_get_stats", map[string]any{"site_id": f.site.ID, "from": "last week"}); !strings.Contains(text, "2026-11-01") {
		t.Fatalf("a bad date must show the expected form, got %q", text)
	}
}

// TestMCPKeepsRolesAsTheWebAppDoes refuses a viewer's writes and hides the site from a
// stranger, with errors that say what to do.
func TestMCPKeepsRolesAsTheWebAppDoes(t *testing.T) {
	f := newFixture(t)
	viewer := f.connect(t, "viewer", mcptest.ClientOptions{})
	if text := refused(t, viewer, "vision_update_site", map[string]any{"site_id": f.site.ID, "name": "Mine"}); !strings.Contains(text, "owners, admins and editors") {
		t.Fatalf("a viewer cannot edit, got %q", text)
	}
	if text := refused(t, viewer, "vision_delete_site", map[string]any{"site_id": f.site.ID}); !strings.Contains(text, "owners, admins and editors") {
		t.Fatalf("a viewer is refused before anyone is asked, got %q", text)
	}
	if text := refused(t, viewer, "vision_create_site", map[string]any{"space_id": f.acme, "name": "X", "domain": "x.test"}); !strings.Contains(text, "write access") {
		t.Fatalf("a viewer cannot add a site, got %q", text)
	}
	stranger := f.connect(t, "stranger", mcptest.ClientOptions{})
	if text := refused(t, stranger, "vision_get_stats", map[string]any{"site_id": f.site.ID}); !strings.Contains(text, "vision_list_sites") {
		t.Fatalf("a stranger must not see the site and is pointed at the list, got %q", text)
	}
	if f.count("mcp_actions", "1 = 1") != 0 {
		t.Fatal("a refused request must leave no pending action")
	}
}

// TestMCPWritesSitesAndGoals adds a site, edits it field by field, adds and edits a goal,
// and closes a public link without a confirmation.
func TestMCPWritesSitesAndGoals(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner", mcptest.ClientOptions{})
	var site siteResult
	call(t, session, "vision_create_site", map[string]any{"space_id": f.acme, "name": "Blog", "domain": "Blog.Example"}, &site)
	if site.Status != "created" || site.Site.Domain != "blog.example" || !strings.Contains(site.TrackingSnippet, "/s.js") {
		t.Fatalf("want the site and its script tag, got %+v", site)
	}
	call(t, session, "vision_update_site", map[string]any{"site_id": site.Site.ID, "name": "Engineering blog"}, &site)
	if site.Site.Name != "Engineering blog" || site.Site.Domain != "blog.example" {
		t.Fatalf("a rename must keep the domain, got %+v", site)
	}
	if text := refused(t, session, "vision_create_site", map[string]any{"space_id": f.acme, "name": "Again", "domain": "blog.example"}); !strings.Contains(text, "already registered") {
		t.Fatalf("a tracked domain must be refused, got %q", text)
	}
	type goalResult struct {
		Goal struct {
			ID        int64  `json:"id"`
			Name      string `json:"name"`
			GoalType  string `json:"goal_type"`
			EventName string `json:"event_name"`
			PagePath  string `json:"page_path"`
		} `json:"goal"`
	}
	var goal, edited goalResult
	call(t, session, "vision_create_goal", map[string]any{"site_id": f.site.ID, "name": "Pricing", "goal_type": "pageview", "page_path": "/pricing"}, &goal)
	call(t, session, "vision_update_goal", map[string]any{"goal_id": goal.Goal.ID, "goal_type": "event", "event_name": "signup"}, &edited)
	if edited.Goal.Name != "Pricing" || edited.Goal.GoalType != "event" || edited.Goal.EventName != "signup" || edited.Goal.PagePath != "" {
		t.Fatalf("switching the type keeps the name and clears the path, got %+v", edited)
	}
	must(t, f.db.Exec(`UPDATE sites SET share_token = 'old' WHERE id = ?`, f.site.ID).Error)
	call(t, session, "vision_unshare_site", map[string]any{"site_id": f.site.ID}, &site)
	if site.Status != "unshared" || site.Site.ShareURL != "" || f.count("sites", "id = ? AND share_token IS NULL", f.site.ID) != 1 {
		t.Fatalf("the public link must close, got %+v", site)
	}
}
