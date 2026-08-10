package schemas

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openTestDatabase returns a connection to a throwaway database, or skips.
//
// PostgreSQL and nothing else: AdoptPorte is a DO block, to_regclass and
// ON CONFLICT, so a SQLite run would build a different schema from the struct
// tags and then agree with itself about it. It is also the one piece of this
// migration that can sign every user out, which makes an untested version
// worse than no version.
//
// The schema is reset per test and only User is migrated, so each case starts
// from the shape production is in *before* the deploy and drives the migration
// itself rather than inheriting whatever a previous case left behind.
func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is unset — createdb vision_test, or point at any scratch database")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: logger.Discard, TranslateError: true})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public").Error; err != nil {
		t.Fatalf("reset the schema: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	return db
}
