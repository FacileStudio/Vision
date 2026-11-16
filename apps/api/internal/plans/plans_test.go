package plans

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

const issuer = "https://sso.test"

var testConfig = testdb.Config{Prefix: "vision_plans_test", Migrate: schemas.Migrate}

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	db, err := testdb.Open(url, testConfig)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	if err := testdb.Truncate(db, testConfig); err != nil {
		t.Fatalf("reset: %v", err)
	}
	return db
}

func snapshot(revision int64) directory.Snapshot {
	yes, no := json.RawMessage("true"), json.RawMessage("false")
	return directory.Snapshot{
		Version:              1,
		EntitlementsRevision: revision,
		Spaces: []directory.Space{
			{ID: "paid", Entitlements: map[string]json.RawMessage{AIAccess: yes}},
			{ID: "free", Entitlements: map[string]json.RawMessage{AIAccess: no}},
		},
		People: []directory.Person{{Subject: "sub-paid", Entitlements: map[string]json.RawMessage{AIAccess: yes}}},
	}
}

func newGate(db *gorm.DB) *Gate {
	return New(db, issuer, slog.New(slog.DiscardHandler))
}

func TestGateFailsOpenUntilRegistreAnswers(t *testing.T) {
	db := openDB(t)
	gate := newGate(db)
	if err := gate.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gate.Active() {
		t.Fatal("no snapshot yet, the gate must not be active")
	}
	if ok, err := gate.HasAI(context.Background(), 1, "free"); err != nil || !ok {
		t.Fatalf("before the first snapshot the gate fails open: %v, %v", ok, err)
	}
}

func TestGateEnforcesSpaceAndPersonPlans(t *testing.T) {
	db := openDB(t)
	paidUser := &schemas.User{Email: "paid@test", Name: "paid"}
	freeUser := &schemas.User{Email: "free@test", Name: "free"}
	for _, user := range []*schemas.User{paidUser, freeUser} {
		if err := db.Create(user).Error; err != nil {
			t.Fatal(err)
		}
	}
	for user, subject := range map[int64]string{paidUser.ID: "sub-paid", freeUser.ID: "sub-free"} {
		err := db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`, user, issuer, subject).Error
		if err != nil {
			t.Fatalf("identity: %v", err)
		}
	}
	gate := newGate(db)
	if err := gate.Apply(context.Background(), snapshot(1)); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		user  int64
		space string
		want  bool
	}{
		{"paid space", freeUser.ID, "paid", true},
		{"free space", paidUser.ID, "free", false},
		{"unknown space gets the default", paidUser.ID, "nowhere", false},
		{"person on a paid plan", paidUser.ID, "", true},
		{"person on the free plan", freeUser.ID, "", false},
		{"person with no identity", 999, "", false},
	}
	for _, c := range cases {
		got, err := gate.HasAI(context.Background(), c.user, c.space)
		if err != nil || got != c.want {
			t.Errorf("%s: %v, %v, want %v", c.name, got, err, c.want)
		}
	}
}

func TestGateKeepsTheLastPlansAcrossARestart(t *testing.T) {
	db := openDB(t)
	if err := newGate(db).Apply(context.Background(), snapshot(1)); err != nil {
		t.Fatal(err)
	}
	restarted := newGate(db)
	if err := restarted.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !restarted.Active() {
		t.Fatal("a saved snapshot must keep the gate enforcing after a restart")
	}
	if ok, _ := restarted.HasAI(context.Background(), 1, "free"); ok {
		t.Fatal("the free space must stay refused after a restart")
	}
	if ok, _ := restarted.HasAI(context.Background(), 1, "paid"); !ok {
		t.Fatal("the paid space must stay allowed after a restart")
	}
}
