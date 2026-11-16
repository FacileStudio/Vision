package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/schemas"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type insightFixture struct {
	db     *gorm.DB
	site   int64
	alice  int64
	target time.Time
}

func newInsightFixture(t *testing.T) *insightFixture {
	t.Helper()
	db := openSummaryDB(t)
	f := seedUsersAndSpaces(t, db)
	site := &schemas.Site{Name: "Acme", Domain: "acme.test", OwnerID: f.alice, WorkspaceID: f.acme}
	mustCreate(t, db, site)
	return &insightFixture{db: db, site: site.ID, alice: f.alice, target: utcDay(time.Now()).AddDate(0, 0, -1)}
}

type visits struct {
	day      time.Time
	count    int
	path     string
	referrer string
	country  string
	prefix   string
}

func (f *insightFixture) add(t *testing.T, v visits) {
	t.Helper()
	rows := make([]schemas.Pageview, v.count)
	for i := range rows {
		rows[i] = schemas.Pageview{
			SiteID: f.site, Path: v.path, Referrer: v.referrer, Country: v.country, Device: "Desktop",
			VisitorID: v.prefix + strconv.Itoa(i), CreatedAt: v.day.Add(time.Duration(i) * time.Second),
		}
	}
	mustCreate(t, f.db, &rows)
}

func (f *insightFixture) history(t *testing.T, weeks []int, v visits) {
	t.Helper()
	for _, week := range weeks {
		v.day = f.target.AddDate(0, 0, -7*week)
		v.prefix = fmt.Sprintf("w%d-", week)
		f.add(t, v)
	}
}

var allWeeks = []int{1, 2, 3, 4, 5, 6, 7, 8}

func byMetric(list []Anomaly, metric string) *Anomaly {
	for i := range list {
		if list[i].Metric == metric {
			return &list[i]
		}
	}
	return nil
}

func TestDetectFlagsASpikeAndNamesWhatMoved(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 40, path: "/", country: "FR"})
	f.add(t, visits{day: f.target, count: 40, path: "/", country: "FR", prefix: "usual-"})
	f.add(t, visits{day: f.target, count: 200, path: "/promo", referrer: "spam.example", country: "US", prefix: "burst-"})
	service := NewService(f.db)
	found, err := service.Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil {
		t.Fatal(err)
	}
	pageviews := byMetric(found, MetricPageviews)
	if pageviews == nil || pageviews.Direction != DirectionUp || pageviews.Value != 240 || pageviews.Baseline != 40 || pageviews.Weeks != 8 {
		t.Fatalf("pageviews spike: %+v", pageviews)
	}
	if pageviews.ChangePct == nil || *pageviews.ChangePct != 500 || pageviews.Day != f.target.Format(dayKeyLayout) {
		t.Fatalf("size and day: %+v", pageviews)
	}
	if visitors := byMetric(found, MetricVisitors); visitors == nil || visitors.Value != 240 {
		t.Fatalf("visitors spike: %+v", visitors)
	}
	share := byMetric(found, MetricTopReferrerShare)
	if share == nil || share.Direction != DirectionUp || share.ChangePct != nil {
		t.Fatalf("a share from nothing has no percentage: %+v", share)
	}
	top := pageviews.Breakdown
	if top.Referrers[0].Value != "spam.example" || top.Referrers[0].Count != 200 || top.Referrers[0].Baseline != 0 || top.Referrers[0].Delta != 200 {
		t.Fatalf("referrer breakdown: %+v", top.Referrers)
	}
	if top.Paths[0].Value != "/promo" || top.Countries[0].Value != "US" || top.Countries[0].Delta != 200 || top.Devices[0].Value != "Desktop" {
		t.Fatalf("breakdown: %+v", top)
	}
	again, err := service.Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil || len(again) != len(found) {
		t.Fatalf("a second run must update, got %d then %d, %v", len(found), len(again), err)
	}
	var stored int64
	f.db.Table("anomalies").Count(&stored)
	if int(stored) != len(found) {
		t.Fatalf("rows %d, want %d: a rerun must not duplicate", stored, len(found))
	}
}

func TestDetectStaysQuietOnAnOrdinaryDay(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 40, path: "/", referrer: "news.example"})
	f.add(t, visits{day: f.target, count: 42, path: "/", referrer: "news.example", prefix: "today-"})
	found, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil || len(found) != 0 {
		t.Fatalf("an ordinary day is not an anomaly: %+v, %v", found, err)
	}
}

func TestDetectNeedsFourWeeksOfHistory(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, []int{1, 2, 3}, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 400, path: "/", prefix: "burst-"})
	found, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil || len(found) != 0 {
		t.Fatalf("a site younger than four weeks has no baseline: %+v, %v", found, err)
	}
}

func TestDetectIgnoresWeeksBeforeTheFirstPageview(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, []int{1, 2, 3, 4}, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 400, path: "/", prefix: "burst-"})
	found, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil {
		t.Fatal(err)
	}
	if pageviews := byMetric(found, MetricPageviews); pageviews == nil || pageviews.Weeks != 4 {
		t.Fatalf("only the four weeks since the first pageview count: %+v", pageviews)
	}
}

func TestDetectTreatsAMissingWeekAsZero(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, []int{1, 3, 5, 7, 8}, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 40, path: "/", prefix: "today-"})
	found, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil || len(found) != 0 {
		t.Fatalf("a normal day after quiet weeks is not an anomaly: %+v, %v", found, err)
	}
}

func TestDetectReadsTheSameWeekdayNotTheDayBefore(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 200, path: "/"})
	for week := 1; week <= 8; week++ {
		f.add(t, visits{day: f.target.AddDate(0, 0, -7*week+1), count: 20, path: "/", prefix: fmt.Sprintf("off%d-", week)})
	}
	f.add(t, visits{day: f.target, count: 200, path: "/", prefix: "today-"})
	found, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil || len(found) != 0 {
		t.Fatalf("a busy weekday next to quiet days is its own baseline: %+v, %v", found, err)
	}
}

func TestDetectFlagsAGoalThatStoppedConverting(t *testing.T) {
	f := newInsightFixture(t)
	path := "/thanks"
	goal := &schemas.Goal{SiteID: f.site, OwnerID: f.alice, Name: "Thanks", GoalType: "pageview", PagePath: &path, MatchType: "exact"}
	mustCreate(t, f.db, goal)
	f.history(t, allWeeks, visits{count: 10, path: "/thanks"})
	f.history(t, allWeeks, visits{count: 10, path: "/", country: "FR"})
	f.add(t, visits{day: f.target, count: 20, path: "/", prefix: "today-"})
	found, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target})
	if err != nil {
		t.Fatal(err)
	}
	conversions := byMetric(found, MetricGoalConversions)
	if conversions == nil || conversions.Direction != DirectionDown || conversions.GoalID != goal.ID || conversions.GoalName != "Thanks" || conversions.Value != 0 {
		t.Fatalf("goal drop: %+v", conversions)
	}
}

func TestListAnomaliesFiltersAndOrders(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 400, path: "/", prefix: "burst-"})
	service := NewService(f.db)
	if _, err := service.Detect(context.Background(), f.site, []time.Time{f.target}); err != nil {
		t.Fatal(err)
	}
	all, err := service.ListAnomalies(context.Background(), f.site, AnomalyQuery{})
	if err != nil || len(all) < 2 {
		t.Fatalf("list: %+v, %v", all, err)
	}
	only, err := service.ListAnomalies(context.Background(), f.site, AnomalyQuery{Metric: MetricVisitors})
	if err != nil || len(only) != 1 || only[0].Metric != MetricVisitors {
		t.Fatalf("metric filter: %+v, %v", only, err)
	}
	after, err := service.ListAnomalies(context.Background(), f.site, AnomalyQuery{From: f.target.AddDate(0, 0, 1)})
	if err != nil || len(after) != 0 {
		t.Fatalf("from filter: %+v, %v", after, err)
	}
	if _, err := service.ListAnomalies(context.Background(), f.site, AnomalyQuery{Metric: "bounce"}); err == nil {
		t.Fatal("an unknown metric must be refused")
	}
	other, err := service.ListAnomalies(context.Background(), f.site+1, AnomalyQuery{})
	if err != nil || len(other) != 0 {
		t.Fatalf("another site has none: %+v, %v", other, err)
	}
}

func TestSweepChecksEverySiteWithRecentTraffic(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 400, path: "/", prefix: "burst-"})
	service := NewService(f.db)
	service.Sweep(context.Background(), []time.Time{f.target}, discardLogger())
	list, err := service.ListAnomalies(context.Background(), f.site, AnomalyQuery{})
	if err != nil || len(list) == 0 {
		t.Fatalf("the sweep must store the spike: %+v, %v", list, err)
	}
}

func TestDeletingASiteDeletesItsAnomalies(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 400, path: "/", prefix: "burst-"})
	if _, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target}); err != nil {
		t.Fatal(err)
	}
	mustCreate(t, f.db, &schemas.Goal{SiteID: f.site, OwnerID: f.alice, Name: "x", GoalType: "event", EventName: new(string)})
	if err := f.db.Delete(&schemas.Site{ID: f.site}).Error; err != nil {
		t.Fatal(err)
	}
	var left int64
	f.db.Table("anomalies").Count(&left)
	if left != 0 {
		t.Fatalf("%d anomalies outlived their site", left)
	}
}

func TestCompareRanksWhatMoved(t *testing.T) {
	f := newInsightFixture(t)
	path := "/thanks"
	mustCreate(t, f.db, &schemas.Goal{SiteID: f.site, OwnerID: f.alice, Name: "Thanks", GoalType: "pageview", PagePath: &path, MatchType: "exact"})
	prev := f.target.AddDate(0, 0, -7)
	f.add(t, visits{day: prev, count: 10, path: "/", referrer: "a.example", country: "FR", prefix: "p-"})
	f.add(t, visits{day: f.target, count: 30, path: "/thanks", referrer: "b.example", country: "US", prefix: "c-"})
	current := Period{From: f.target, To: f.target}
	got, err := NewService(f.db).Compare(context.Background(), f.site, current, Period{From: prev, To: prev}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Current.Pageviews != 30 || got.Current.Visitors != 30 || got.Previous.Pageviews != 10 {
		t.Fatalf("totals: %+v %+v", got.Current, got.Previous)
	}
	if len(got.Current.Goals) != 1 || got.Current.Goals[0].Conversions != 30 || got.Previous.Goals[0].Conversions != 0 {
		t.Fatalf("goals: %+v %+v", got.Current.Goals, got.Previous.Goals)
	}
	if len(got.Paths) != 2 || got.Paths[0].Value != "/thanks" || got.Paths[0].Delta != 30 || got.Paths[0].ChangePct != nil {
		t.Fatalf("paths: %+v", got.Paths)
	}
	if got.Referrers[1].Value != "a.example" || got.Referrers[1].Delta != -10 || *got.Referrers[1].ChangePct != -100 {
		t.Fatalf("referrers: %+v", got.Referrers)
	}
	if _, err := NewService(f.db).Compare(context.Background(), f.site, Period{From: f.target, To: prev}, Period{From: prev, To: prev}, 0); err == nil {
		t.Fatal("a period ending before it starts must be refused")
	}
}

func TestPreviousPeriodHasTheSameLength(t *testing.T) {
	from := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	got := PreviousPeriod(Period{From: from, To: from.AddDate(0, 0, 6)})
	if !got.From.Equal(from.AddDate(0, 0, -7)) || !got.To.Equal(from.AddDate(0, 0, -1)) {
		t.Fatalf("previous period: %+v", got)
	}
}

func insightRouter(f *insightFixture, user int64) http.Handler {
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: strconv.FormatInt(user, 10)})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	router.Route("/analytics", func(r chi.Router) { registerInsightRoutes(r, NewService(f.db), f.db) })
	return router
}

func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestInsightRoutes(t *testing.T) {
	f := newInsightFixture(t)
	f.history(t, allWeeks, visits{count: 40, path: "/"})
	f.add(t, visits{day: f.target, count: 400, path: "/", prefix: "burst-"})
	if _, err := NewService(f.db).Detect(context.Background(), f.site, []time.Time{f.target}); err != nil {
		t.Fatal(err)
	}
	var mallory schemas.User
	if err := f.db.Where("email = ?", "mallory@facile.test").First(&mallory).Error; err != nil {
		t.Fatal(err)
	}
	site := strconv.FormatInt(f.site, 10)
	day := f.target.Format(dayKeyLayout)
	cases := []struct {
		name string
		user int64
		path string
		code int
	}{
		{"list", f.alice, "/analytics/" + site + "/anomalies", http.StatusOK},
		{"list filtered", f.alice, "/analytics/" + site + "/anomalies?metric=pageviews&from=" + day + "&to=" + day + "&limit=1", http.StatusOK},
		{"bad metric", f.alice, "/analytics/" + site + "/anomalies?metric=bounce", http.StatusBadRequest},
		{"bad date", f.alice, "/analytics/" + site + "/anomalies?from=yesterday", http.StatusBadRequest},
		{"bad limit", f.alice, "/analytics/" + site + "/anomalies?limit=0", http.StatusBadRequest},
		{"stranger", mallory.ID, "/analytics/" + site + "/anomalies", http.StatusNotFound},
		{"compare", f.alice, "/analytics/" + site + "/compare?from=" + day + "&to=" + day, http.StatusOK},
		{"compare explicit", f.alice, "/analytics/" + site + "/compare?from=" + day + "&to=" + day + "&compare_from=" + day + "&compare_to=" + day + "&limit=3", http.StatusOK},
		{"compare reversed", f.alice, "/analytics/" + site + "/compare?from=" + day + "&to=2020-01-01", http.StatusBadRequest},
		{"compare bad date", f.alice, "/analytics/" + site + "/compare?compare_from=x", http.StatusBadRequest},
		{"compare stranger", mallory.ID, "/analytics/" + site + "/compare", http.StatusNotFound},
		{"bad site", f.alice, "/analytics/x/compare", http.StatusBadRequest},
	}
	for _, c := range cases {
		recorder := get(insightRouter(f, c.user), c.path)
		if recorder.Code != c.code {
			t.Errorf("%s: %d, want %d: %s", c.name, recorder.Code, c.code, recorder.Body)
		}
	}
	var listed AnomaliesResponse
	if err := json.Unmarshal(get(insightRouter(f, f.alice), "/analytics/"+site+"/anomalies").Body.Bytes(), &listed); err != nil || len(listed.Anomalies) < 2 {
		t.Fatalf("the list must decode: %+v, %v", listed, err)
	}
}
