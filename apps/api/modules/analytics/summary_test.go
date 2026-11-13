package analytics

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

var testConfig = testdb.Config{Prefix: "vision_analytics_test", Migrate: schemas.Migrate}

// openSummaryDB hands the test a migrated, emptied Postgres schema, or skips.
func openSummaryDB(t *testing.T) *gorm.DB {
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
	return db
}

// mustCreate inserts value or fails the test.
func mustCreate(t *testing.T, db *gorm.DB, value any) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatal(err)
	}
}

// summaryFixture holds the ids seedSummary created.
type summaryFixture struct {
	alice, mallory, acme, other, acmeSite int64
}

// seedUsersAndSpaces creates alice in Acme and mallory in Other.
func seedUsersAndSpaces(t *testing.T, db *gorm.DB) summaryFixture {
	t.Helper()
	alice := &schemas.User{Email: "alice@facile.test", Name: "alice"}
	mallory := &schemas.User{Email: "mallory@facile.test", Name: "mallory"}
	acme := &schemas.Workspace{Name: "Acme"}
	other := &schemas.Workspace{Name: "Other"}
	for _, row := range []any{alice, mallory, acme, other} {
		mustCreate(t, db, row)
	}
	mustCreate(t, db, &schemas.WorkspaceMember{WorkspaceID: acme.ID, UserID: alice.ID, Role: "owner"})
	mustCreate(t, db, &schemas.WorkspaceMember{WorkspaceID: other.ID, UserID: mallory.ID, Role: "owner"})
	return summaryFixture{alice: alice.ID, mallory: mallory.ID, acme: acme.ID, other: other.ID}
}

// seedSummary adds Acme's busy and quiet sites and Other's foreign site, with traffic in
// the window, in the previous window and before both.
func seedSummary(t *testing.T, db *gorm.DB, from time.Time) summaryFixture {
	t.Helper()
	f := seedUsersAndSpaces(t, db)
	site := &schemas.Site{Name: "Acme", Domain: "acme.test", OwnerID: f.alice, WorkspaceID: f.acme}
	quiet := &schemas.Site{Name: "Quiet", Domain: "quiet.test", OwnerID: f.alice, WorkspaceID: f.acme}
	foreign := &schemas.Site{Name: "Foreign", Domain: "foreign.test", OwnerID: f.mallory, WorkspaceID: f.other}
	for _, row := range []any{site, quiet, foreign} {
		mustCreate(t, db, row)
	}
	f.acmeSite = site.ID
	visitors := []string{"v1", "v1", "v2", "", "v3", "v9"}
	offsets := []time.Duration{time.Hour, 2 * time.Hour, 26 * time.Hour, 27 * time.Hour, -time.Hour, -240 * time.Hour}
	for i, visitor := range visitors {
		mustCreate(t, db, &schemas.Pageview{SiteID: site.ID, VisitorID: visitor, Path: "/", CreatedAt: from.Add(offsets[i])})
	}
	mustCreate(t, db, &schemas.Pageview{SiteID: foreign.ID, VisitorID: "f1", Path: "/", CreatedAt: from.Add(time.Hour)})
	return f
}

func TestSummaryCountsTheCallersSites(t *testing.T) {
	db := openSummaryDB(t)
	from := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	f := seedSummary(t, db, from)
	got, err := NewService(db).Summary(context.Background(), f.alice, 0, from, from.Add(3*24*time.Hour-time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sites) != 2 || got.Sites[0].SiteID != f.acmeSite || got.Sites[1].Name != "Quiet" {
		t.Fatalf("alice's sites = %+v, want Acme then Quiet", got.Sites)
	}
	top := got.Sites[0]
	if top.Pageviews != 4 || top.UniqueVisitors != 2 || top.PrevPageviews != 1 || top.PrevUniqueVisitors != 1 {
		t.Fatalf("counts = %+v, want 4 pageviews, 2 visitors, previous 1 and 1", top)
	}
	if top.WorkspaceID != f.acme || top.WorkspaceName != "Acme" || got.Granularity != "day" {
		t.Fatalf("workspace %d %q, granularity %q", top.WorkspaceID, top.WorkspaceName, got.Granularity)
	}
	want := []DayStat{{"2026-01-10", 2}, {"2026-01-11", 2}, {"2026-01-12", 0}}
	if !slices.Equal(top.Series, want) {
		t.Fatalf("series = %+v, want %+v", top.Series, want)
	}
}

func TestSummaryShowsANonMemberNothing(t *testing.T) {
	db := openSummaryDB(t)
	from := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	to := from.Add(72 * time.Hour)
	f := seedSummary(t, db, from)
	service := NewService(db)
	cases := map[string][2]int64{
		"alice filtering on Other":  {f.alice, f.other},
		"alice on a missing space":  {f.alice, f.other + 1000},
		"mallory filtering on Acme": {f.mallory, f.acme},
	}
	for name, c := range cases {
		got, err := service.Summary(context.Background(), c[0], c[1], from, to)
		if err != nil || len(got.Sites) != 0 {
			t.Fatalf("%s: %+v, %v; want no sites", name, got, err)
		}
	}
	own, err := service.Summary(context.Background(), f.mallory, 0, from, to)
	if err != nil || len(own.Sites) != 1 || own.Sites[0].Name != "Foreign" {
		t.Fatalf("mallory's own sites: %+v, %v", own, err)
	}
}

func TestSummaryBucketsShortWindowsByHour(t *testing.T) {
	from := time.Date(2026, 1, 10, 10, 30, 0, 0, time.UTC)
	starts := bucketStarts(from, from.Add(24*time.Hour), summaryUnit(from, from.Add(24*time.Hour)))
	if len(starts) != 25 || bucketKey(starts[0], "hour") != "2026-01-10T10:00" {
		t.Fatalf("hour buckets: %d starting %v", len(starts), starts[0])
	}
}

func TestParseSpace(t *testing.T) {
	for raw, want := range map[string]int64{"": 0, "all": 0, "42": 42} {
		if got, err := parseSpace(raw); err != nil || got != want {
			t.Fatalf("parseSpace(%q) = %d, %v", raw, got, err)
		}
	}
	for _, raw := range []string{"0", "-1", "abc"} {
		if _, err := parseSpace(raw); err == nil {
			t.Fatalf("parseSpace(%q) accepted", raw)
		}
	}
}
