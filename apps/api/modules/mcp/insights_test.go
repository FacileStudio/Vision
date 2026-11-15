package mcp_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// seedSpike gives the site eight quiet weeks of 40 pageviews, then a burst of 400 from one
// referrer yesterday, and stores the anomalies the detector finds. It returns that day.
func (f *fixture) seedSpike(t *testing.T) time.Time {
	t.Helper()
	y, m, d := time.Now().UTC().Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	add := func(day time.Time, n int, path, referrer, prefix string) {
		rows := make([]schemas.Pageview, n)
		for i := range rows {
			rows[i] = schemas.Pageview{SiteID: f.site.ID, Path: path, Referrer: referrer, Country: "FR", Device: "Desktop",
				VisitorID: prefix + strconv.Itoa(i), CreatedAt: day.Add(time.Duration(i) * time.Second)}
		}
		must(t, f.db.Create(&rows).Error)
	}
	must(t, f.db.Exec(`DELETE FROM pageviews WHERE site_id = ?`, f.site.ID).Error)
	for week := 1; week <= 8; week++ {
		add(yesterday.AddDate(0, 0, -7*week), 40, "/", "", "h"+strconv.Itoa(week)+"-")
	}
	add(yesterday, 40, "/", "", "usual-")
	add(yesterday, 400, "/promo", "spam.example", "burst-")
	_, err := analytics.NewService(f.db).Detect(context.Background(), f.site.ID, []time.Time{yesterday})
	must(t, err)
	return yesterday
}

// TestMCPListsAnomaliesAndComparesPeriods reads the detector's findings and the period
// comparison through the tools a model would use to explain a spike.
func TestMCPListsAnomaliesAndComparesPeriods(t *testing.T) {
	f := newFixture(t)
	yesterday := f.seedSpike(t)
	session := f.connect(t, "viewer", mcptest.ClientOptions{})
	var found struct {
		Site struct {
			ID int64 `json:"id"`
		} `json:"site"`
		Anomalies []struct {
			Metric    string  `json:"metric"`
			Day       string  `json:"day"`
			Direction string  `json:"direction"`
			Value     float64 `json:"value"`
			Baseline  float64 `json:"baseline"`
			Breakdown struct {
				Referrers []struct {
					Value string  `json:"value"`
					Delta float64 `json:"delta"`
				} `json:"referrers"`
			} `json:"breakdown"`
		} `json:"anomalies"`
	}
	call(t, session, "vision_list_anomalies", map[string]any{"site_id": f.site.ID}, &found)
	if found.Site.ID != f.site.ID || len(found.Anomalies) < 2 {
		t.Fatalf("want the spike on pageviews and visitors, got %+v", found)
	}
	top := found.Anomalies[0]
	if top.Day != yesterday.Format("2006-01-02") || top.Direction != "up" || len(top.Breakdown.Referrers) == 0 || top.Breakdown.Referrers[0].Value != "spam.example" || top.Breakdown.Referrers[0].Delta != 400 {
		t.Fatalf("anomaly: %+v", top)
	}
	call(t, session, "vision_list_anomalies", map[string]any{"site_id": f.site.ID, "metric": "visitors", "limit": 1}, &found)
	if len(found.Anomalies) != 1 || found.Anomalies[0].Metric != "visitors" {
		t.Fatalf("metric and limit filters: %+v", found)
	}
	day := yesterday.Format("2006-01-02")
	var compared struct {
		Current  struct{ Pageviews, Visitors int64 }
		Previous struct{ Pageviews int64 }
		Paths    []struct {
			Value string `json:"value"`
			Delta int64  `json:"delta"`
		} `json:"paths"`
	}
	call(t, session, "vision_compare_periods", map[string]any{"site_id": f.site.ID, "from": day, "to": day}, &compared)
	if compared.Current.Pageviews != 440 || compared.Previous.Pageviews != 0 || len(compared.Paths) == 0 || compared.Paths[0].Value != "/promo" {
		t.Fatalf("a day against the day before: %+v", compared)
	}
	week := yesterday.AddDate(0, 0, -7).Format("2006-01-02")
	call(t, session, "vision_compare_periods", map[string]any{"site_id": f.site.ID, "from": day, "to": day, "compare_from": week, "compare_to": week}, &compared)
	if compared.Previous.Pageviews != 40 || compared.Paths[0].Value != "/promo" || compared.Paths[0].Delta != 400 {
		t.Fatalf("a day against the same weekday: %+v", compared)
	}
}

// TestMCPInsightToolsRefuseBadInputAndStrangers keeps the new reads behind the same site
// access as the others and explains every refusal.
func TestMCPInsightToolsRefuseBadInputAndStrangers(t *testing.T) {
	f := newFixture(t)
	owner := f.connect(t, "owner", mcptest.ClientOptions{})
	cases := []struct {
		tool string
		args map[string]any
		want string
	}{
		{"vision_list_anomalies", map[string]any{"site_id": f.site.ID, "metric": "bounce"}, "visitors, pageviews, goal_conversions"},
		{"vision_list_anomalies", map[string]any{"site_id": f.site.ID, "from": "last week"}, "2026-11-01"},
		{"vision_compare_periods", map[string]any{"site_id": f.site.ID, "compare_from": "2026-11-01"}, "go together"},
		{"vision_compare_periods", map[string]any{"site_id": f.site.ID, "from": "2026-11-05", "to": "2026-11-01"}, "from must not be after to"},
	}
	for _, c := range cases {
		if text := refused(t, owner, c.tool, c.args); !strings.Contains(text, c.want) {
			t.Errorf("%s %v: want %q in %q", c.tool, c.args, c.want, text)
		}
	}
	stranger := f.connect(t, "stranger", mcptest.ClientOptions{})
	for _, tool := range []string{"vision_list_anomalies", "vision_compare_periods"} {
		if text := refused(t, stranger, tool, map[string]any{"site_id": f.site.ID}); !strings.Contains(text, "vision_list_sites") {
			t.Errorf("%s: a stranger must be pointed at the list, got %q", tool, text)
		}
	}
}

// TestMCPInsightToolsAreReadOnly checks both tools are annotated read-only and idempotent,
// open world because visitors write the values they return, and typed: no free-form query.
func TestMCPInsightToolsAreReadOnly(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "viewer", mcptest.ClientOptions{})
	listed, err := session.ListTools(context.Background(), nil)
	must(t, err)
	seen := 0
	for _, tool := range listed.Tools {
		if tool.Name != "vision_list_anomalies" && tool.Name != "vision_compare_periods" {
			continue
		}
		seen++
		a := tool.Annotations
		if a == nil || !a.ReadOnlyHint || !a.IdempotentHint || a.OpenWorldHint == nil || !*a.OpenWorldHint {
			t.Errorf("%s annotations: %+v", tool.Name, a)
		}
		schema, _ := tool.InputSchema.(map[string]any)
		for name := range schema["properties"].(map[string]any) {
			if name == "query" || name == "sql" || name == "filter" {
				t.Errorf("%s has a free-form %s parameter", tool.Name, name)
			}
		}
	}
	if seen != 2 {
		t.Fatalf("both insight tools must be listed, got %d", seen)
	}
}
