package analytics

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/schemas"

	"github.com/go-chi/chi/v5"
)

var botDay = time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)

// seedBots gives Acme hits on two days and the foreign site hits that must not count.
func seedBots(t *testing.T) (summaryFixture, *Service) {
	t.Helper()
	db := openSummaryDB(t)
	f := seedSummary(t, db, botDay)
	var foreign schemas.Site
	if err := db.Where("name = ?", "Foreign").First(&foreign).Error; err != nil {
		t.Fatal(err)
	}
	next := botDay.AddDate(0, 0, 1)
	rows := []schemas.BotHit{
		{SiteID: f.acmeSite, Day: botDay, Category: "search", Name: "Googlebot", Path: "/", Hits: 10},
		{SiteID: f.acmeSite, Day: next, Category: "search", Name: "Googlebot", Path: "/docs", Hits: 5},
		{SiteID: f.acmeSite, Day: next, Category: "ai", Name: "GPTBot", Path: "/", Hits: 7},
		{SiteID: f.acmeSite, Day: next, Category: "ai", Name: "ClaudeBot", Path: "/docs", Hits: 7},
		{SiteID: f.acmeSite, Day: botDay, Category: "monitor", Name: "UptimeRobot", Path: "/health", Hits: 1},
		{SiteID: foreign.ID, Day: botDay, Category: "spoofed", Name: "Fake", Path: "/", Hits: 1000},
	}
	mustCreate(t, db, &rows)
	return f, NewService(db)
}

func TestBotsAggregatesOneSite(t *testing.T) {
	f, service := seedBots(t)
	got, err := service.Bots(context.Background(), f.acmeSite, botDay.AddDate(0, 0, -1), botDay.AddDate(0, 0, 2).Add(-time.Nanosecond))
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 30 {
		t.Fatalf("total = %d, want 30", got.Total)
	}
	if want := []BotCategoryStat{{"search", 15}, {"ai", 14}, {"monitor", 1}}; !slices.Equal(got.Categories, want) {
		t.Fatalf("categories = %+v, want %+v", got.Categories, want)
	}
	if want := []BotStat{{"Googlebot", "search", 15}, {"ClaudeBot", "ai", 7}, {"GPTBot", "ai", 7}, {"UptimeRobot", "monitor", 1}}; !slices.Equal(got.Bots, want) {
		t.Fatalf("bots = %+v, want %+v", got.Bots, want)
	}
	if want := []BotPathStat{{"/", 17}, {"/docs", 12}, {"/health", 1}}; !slices.Equal(got.Paths, want) {
		t.Fatalf("paths = %+v, want %+v", got.Paths, want)
	}
	if want := []BotDayStat{{"2026-03-01", 0}, {"2026-03-02", 11}, {"2026-03-03", 19}}; !slices.Equal(got.Series, want) {
		t.Fatalf("series = %+v, want %+v", got.Series, want)
	}
}

func TestBotsRoute(t *testing.T) {
	f, service := seedBots(t)
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: strconv.FormatInt(f.alice, 10)})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	router.Route("/analytics", func(r chi.Router) { registerBotRoutes(r, service, service.orm) })
	site := "/analytics/" + strconv.FormatInt(f.acmeSite, 10) + "/bots"
	empty := get(router, site+"?from=2020-01-01&to=2020-01-01")
	if empty.Code != http.StatusOK || !strings.Contains(empty.Body.String(), `"categories":[],"bots":[],"paths":[],"series":[{"day":"2020-01-01","hits":0}]`) {
		t.Fatalf("empty period: %d %s", empty.Code, empty.Body)
	}
	if long := get(router, site+"?from=2020-01-01&to=2022-01-01"); long.Code != http.StatusBadRequest {
		t.Fatalf("two years: %d, want 400", long.Code)
	}
	if stranger := get(router, "/analytics/"+strconv.FormatInt(f.acmeSite+1000, 10)+"/bots"); stranger.Code != http.StatusNotFound {
		t.Fatalf("missing site: %d, want 404", stranger.Code)
	}
}

func TestOverviewCountsInteractedVisitors(t *testing.T) {
	f, service := seedBots(t)
	start := botDay.Add(time.Hour)
	sessions := []schemas.VisitorSession{
		{SiteID: f.acmeSite, VisitorID: "human", StartedAt: start, EndedAt: start, Interacted: true},
		{SiteID: f.acmeSite, VisitorID: "human", StartedAt: start.Add(time.Hour), EndedAt: start, Interacted: true},
		{SiteID: f.acmeSite, VisitorID: "idle", StartedAt: start, EndedAt: start},
		{SiteID: f.acmeSite + 1, VisitorID: "elsewhere", StartedAt: start, EndedAt: start, Interacted: true},
	}
	mustCreate(t, service.orm, &sessions)
	got, err := service.Overview(context.Background(), f.acmeSite, botDay, botDay.Add(24*time.Hour-time.Nanosecond), "day", Filters{})
	if err != nil || got.InteractedVisitors != 1 {
		t.Fatalf("interacted visitors = %+v, %v; want 1", got, err)
	}
}
