package schemas

import (
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	portepg "github.com/FacileStudio/porte/pg"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var queryModels = []any{
	&User{}, &Workspace{}, &WorkspaceMember{},
	&Site{}, &Pageview{}, &Webhook{}, &VisitorSession{},
	&CustomEvent{}, &Goal{}, &APIKey{},
}

var timeType = reflect.TypeFor[time.Time]()

func compatibleTypes(field reflect.Type) []string {
	for field.Kind() == reflect.Pointer {
		field = field.Elem()
	}
	if field == timeType {
		return []string{"timestamp with time zone"}
	}
	switch field.Kind() {
	case reflect.String:
		return []string{"text", "character varying"}
	case reflect.Bool:
		return []string{"boolean"}
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		return []string{"bigint"}
	case reflect.Int32, reflect.Uint32:
		return []string{"integer", "bigint"}
	case reflect.Int16, reflect.Uint16:
		return []string{"smallint", "integer", "bigint"}
	default:
		return nil
	}
}

func migratedColumnTypes(t *testing.T, db *gorm.DB, table string) map[string]string {
	t.Helper()
	var rows []struct {
		ColumnName string
		DataType   string
	}
	err := db.Raw(`SELECT column_name, data_type FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ?`, table).Scan(&rows).Error
	if err != nil {
		t.Fatalf("read the columns of %s: %v", table, err)
	}
	columns := map[string]string{}
	for _, row := range rows {
		columns[row.ColumnName] = row.DataType
	}
	return columns
}

// TestGORMModelsMatchTheMigratedSchema checks every column a GORM model reads
// or writes exists in the schema the migrations build, with a type the field
// can hold. AutoMigrate no longer adds a missing column, so a field added to a
// struct without a migration fails here instead of at query time.
func TestGORMModelsMatchTheMigratedSchema(t *testing.T) {
	db := openTestDatabase(t)
	for _, model := range queryModels {
		parsed, err := schema.Parse(model, &sync.Map{}, db.NamingStrategy)
		if err != nil {
			t.Fatalf("parse %T: %v", model, err)
		}
		columns := migratedColumnTypes(t, db, parsed.Table)
		if len(columns) == 0 {
			t.Errorf("%T: table %s does not exist", model, parsed.Table)
			continue
		}
		for _, field := range parsed.Fields {
			if field.DBName == "" {
				continue
			}
			got, ok := columns[field.DBName]
			if !ok {
				t.Errorf("%s.%s: no such column", parsed.Table, field.DBName)
				continue
			}
			want := compatibleTypes(field.FieldType)
			if !slices.Contains(want, got) {
				t.Errorf("%s.%s: column is %s, field %s needs one of %v", parsed.Table, field.DBName, got, field.FieldType, want)
			}
		}
	}
}

func tableShape(t *testing.T, db *gorm.DB, table string) []string {
	t.Helper()
	var shape []string
	err := db.Raw(`SELECT column_name || ' ' || data_type || ' ' || is_nullable || ' ' || coalesce(column_default, '')
		FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ?
		UNION ALL SELECT regexp_replace(indexdef, ' ON \S+\.', ' ON ') FROM pg_indexes WHERE schemaname = current_schema() AND tablename = ?
		ORDER BY 1`, table, table).Scan(&shape).Error
	if err != nil {
		t.Fatalf("read the shape of %s: %v", table, err)
	}
	return shape
}

func columnShape(t *testing.T, db *gorm.DB, table string) []string {
	t.Helper()
	var shape []string
	err := db.Raw(`SELECT column_name || ' ' || data_type || ' ' || is_nullable || ' ' || coalesce(column_default, '')
		FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ?
		ORDER BY 1`, table).Scan(&shape).Error
	if err != nil {
		t.Fatalf("read the columns of %s: %v", table, err)
	}
	return shape
}

func referenceSchema(t *testing.T, prefix, ddl string) *gorm.DB {
	t.Helper()
	url, _ := testdb.URL()
	reference, err := testdb.Open(url, testdb.Config{Prefix: prefix, Migrate: func(db *gorm.DB) error {
		return db.Exec(ddl).Error
	}})
	if err != nil {
		t.Fatalf("build the reference schema %s: %v", prefix, err)
	}
	if handle, err := reference.DB(); err == nil {
		t.Cleanup(func() { _ = handle.Close() })
	}
	return reference
}

// TestLibrarySchemaIsInTheMigrations checks the migrations build mcp_actions
// exactly as tronc/mcp's PendingSchema does, columns and indexes, since Vision
// used to execute it at every boot. A tronc bump that changes it makes this
// fail, and the change then needs a numbered migration of its own.
func TestLibrarySchemaIsInTheMigrations(t *testing.T) {
	db := openTestDatabase(t)
	reference := referenceSchema(t, "vision_mcp_reference", mcpkit.PendingSchema)
	want := tableShape(t, reference, "mcp_actions")
	if len(want) == 0 {
		t.Fatal("PendingSchema no longer creates mcp_actions")
	}
	if got := tableShape(t, db, "mcp_actions"); !slices.Equal(got, want) {
		t.Fatalf("mcp_actions differs from PendingSchema:\nmigrations %v\nmcp        %v", got, want)
	}
}

// TestPorteTablesMatchPorteSchema checks the four porte tables in the
// migrations carry exactly the columns porte/pg's Schema gives them. Vision
// keeps its own copy, with porte_users left out and the foreign keys on its
// users, and used to run it at every boot; a porte bump that adds a column,
// as v0.10.0 did to porte_login_codes, now fails here and needs a numbered
// migration. Indexes are not compared: porte's magic link indexes on ip and
// expires_at were never in Vision's copy, and no Vision route issues a magic
// link.
func TestPorteTablesMatchPorteSchema(t *testing.T) {
	db := openTestDatabase(t)
	reference := referenceSchema(t, "vision_porte_reference", portepg.Schema)
	for _, table := range []string{"porte_identities", "porte_sessions", "porte_login_codes", "porte_magic_links"} {
		want := columnShape(t, reference, table)
		if len(want) == 0 {
			t.Fatalf("porte/pg's Schema no longer creates %s", table)
		}
		if got := columnShape(t, db, table); !slices.Equal(got, want) {
			t.Errorf("%s differs from porte/pg's Schema:\nmigrations %v\nporte      %v", table, got, want)
		}
	}
}
