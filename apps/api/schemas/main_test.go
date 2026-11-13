package schemas

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/migrations"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

const testBootstrap = "createdb vision_test"

var (
	testDBOnce sync.Once
	testDB     *gorm.DB
	testDBErr  error
	testDBCfg  = testdb.Config{Prefix: "vision_schemas_test", Migrate: applyMigrations}
)

// applyMigrations builds the schema from the SQL migrations alone, without
// Reconcile, so each case drives the boot data steps itself.
func applyMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return migrations.Up(context.Background(), migrations.Config(sqlDB, slog.New(slog.DiscardHandler)))
}

// openTestDatabase hands the test a migrated Postgres schema, emptied per case,
// or skips.
//
// PostgreSQL and nothing else: AdoptPorte is a DO block, to_regclass and
// ON CONFLICT, so a SQLite run would build a different schema from the struct
// tags and then agree with itself about it. It is also the one piece of the
// boot that can sign every user out, which makes an untested version worse
// than no version.
func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce(testBootstrap)
		t.Skip(testdb.SkipReason(testBootstrap))
	}
	testDBOnce.Do(func() { testDB, testDBErr = testdb.Open(url, testDBCfg) })
	if testDBErr != nil {
		t.Fatalf("test database: %v", testDBErr)
	}
	if err := testdb.Truncate(testDB, testDBCfg); err != nil {
		t.Fatalf("reset the test database: %v", err)
	}
	return testDB
}
