package migrate

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func freshDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; requires PostgreSQL 16 with citext")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		DROP TABLE IF EXISTS app_user;
		DROP TABLE IF EXISTS platform;
		DROP TABLE IF EXISTS schema_migrations;
		DROP EXTENSION IF EXISTS citext;
	`); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db, func() {
		db.Exec(`
			DROP TABLE IF EXISTS app_user;
			DROP TABLE IF EXISTS platform;
			DROP TABLE IF EXISTS schema_migrations;
			DROP EXTENSION IF EXISTS citext;
		`)
		db.Close()
	}
}

func newMigrator(t *testing.T, db *sql.DB) *Migrator {
	t.Helper()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}
	src, err := (&file.File{}).Open("file://" + migrationsDir(t))
	if err != nil {
		t.Fatal(err)
	}
	m, err := gomigrate.NewWithInstance("file", src, "postgres", driver)
	if err != nil {
		t.Fatal(err)
	}
	return New(m)
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", "migrations"))
}

func seedFile(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", "seed", "0001_platform_seed.sql"))
}

func TestUp_FromZero(t *testing.T) {
	db, cleanup := freshDB(t)
	defer cleanup()
	mg := newMigrator(t, db)
	if err := mg.Up(); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, db, "platform") || !tableExists(t, db, "app_user") {
		t.Fatal("expected platform and app_user tables")
	}
}

func TestUp_Idempotent(t *testing.T) {
	db, cleanup := freshDB(t)
	defer cleanup()
	mg := newMigrator(t, db)
	if err := mg.Up(); err != nil {
		t.Fatal(err)
	}
	v1, dirty1, err := mg.Version()
	if err != nil {
		t.Fatal(err)
	}
	if err := mg.Up(); err != nil {
		t.Fatal(err)
	}
	v2, dirty2, err := mg.Version()
	if err != nil {
		t.Fatal(err)
	}
	if v1 != v2 || dirty1 || dirty2 {
		t.Fatalf("expected stable clean version, got (%d,%v) -> (%d,%v)", v1, dirty1, v2, dirty2)
	}
}

func TestDown_OneStep(t *testing.T) {
	db, cleanup := freshDB(t)
	defer cleanup()
	mg := newMigrator(t, db)
	if err := mg.Up(); err != nil {
		t.Fatal(err)
	}
	if err := mg.Down(1); err != nil {
		t.Fatal(err)
	}
	if tableExists(t, db, "app_user") {
		t.Fatal("expected app_user rolled back")
	}
	if !tableExists(t, db, "platform") {
		t.Fatal("expected platform to remain after one-step rollback")
	}
}

func TestPlatformSeed_Idempotent(t *testing.T) {
	db := upWithSeed(t)
	seed(t, db)
	if got := countRows(t, db, "platform"); got != 3 {
		t.Fatalf("expected 3 platforms, got %d", got)
	}
}

func TestPlatform_CodeCheck(t *testing.T) {
	db := upWithSeed(t)
	if _, err := db.Exec(`INSERT INTO platform(id,code,country) VALUES (9,'amazon','VN')`); err == nil {
		t.Fatal("expected platform code CHECK error")
	}
}

func TestPlatform_CountryCheck(t *testing.T) {
	db := upWithSeed(t)
	if _, err := db.Exec(`INSERT INTO platform(id,code,country) VALUES (9,'shopee','Vietnam')`); err == nil {
		t.Fatal("expected country CHECK error")
	}
}

func TestAppUser_EmailCaseInsensitiveUnique(t *testing.T) {
	db := upWithSeed(t)
	if _, err := db.Exec(`INSERT INTO app_user(email) VALUES ('A@x.com')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO app_user(email) VALUES ('a@x.com')`); err == nil {
		t.Fatal("expected CITEXT unique violation")
	}
}

func TestAppUser_Defaults(t *testing.T) {
	db := upWithSeed(t)
	var locale, status string
	if err := db.QueryRow(`INSERT INTO app_user(email) VALUES ('d@x.com') RETURNING locale, status`).Scan(&locale, &status); err != nil {
		t.Fatal(err)
	}
	if locale != "vi-VN" || status != "active" {
		t.Fatalf("expected vi-VN/active defaults, got %s/%s", locale, status)
	}
}

func upWithSeed(t *testing.T) *sql.DB {
	t.Helper()
	db, cleanup := freshDB(t)
	t.Cleanup(cleanup)
	if err := newMigrator(t, db).Up(); err != nil {
		t.Fatal(err)
	}
	seed(t, db)
	return db
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var exists bool
	if err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", name).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	return exists
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func seed(t *testing.T, db *sql.DB) {
	t.Helper()
	sqlBytes, err := os.ReadFile(seedFile(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		t.Fatal(err)
	}
}
