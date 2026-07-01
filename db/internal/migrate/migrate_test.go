package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newMigratorFreshDB(t *testing.T) (*Migrator, *sql.DB) {
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase("shopass"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, postgresContainer.Terminate(ctx))
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	// find the migrations folder from the module root
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// path relative to internal/migrate
	sourceURL := fmt.Sprintf("file://%s/../../migrations", cwd)

	mg, err := NewMigrator(db, sourceURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		mg.Close()
	})

	return mg, db
}

func upWithSeed(t *testing.T) (*Migrator, *sql.DB) {
	mg, db := newMigratorFreshDB(t)
	require.NoError(t, mg.Up())
	seed(t, db)
	return mg, db
}

func seed(t *testing.T, db *sql.DB) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	seedPath := fmt.Sprintf("%s/../../seed/0001_platform_seed.sql", cwd)
	seedQuery, err := os.ReadFile(seedPath)
	require.NoError(t, err)

	_, err = db.Exec(string(seedQuery))
	require.NoError(t, err)
}

func tableExists(t *testing.T, db *sql.DB, tableName string) bool {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`, tableName).Scan(&exists)
	require.NoError(t, err)
	return exists
}

func countRows(t *testing.T, db *sql.DB, tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT count(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&count)
	require.NoError(t, err)
	return count
}

func TestUp_FromZero(t *testing.T) {
	mg, db := newMigratorFreshDB(t)
	require.NoError(t, mg.Up())
	require.True(t, tableExists(t, db, "platform"))
	require.True(t, tableExists(t, db, "app_user"))
}

func TestUp_Idempotent(t *testing.T) {
	mg, _ := newMigratorFreshDB(t)
	require.NoError(t, mg.Up())
	require.NoError(t, mg.Up()) // second time should not error
}

func TestDown_OneStep(t *testing.T) {
	mg, db := newMigratorFreshDB(t)
	require.NoError(t, mg.Up())
	require.NoError(t, mg.Down(1))
	require.False(t, tableExists(t, db, "app_user")) // last migration rolled back
	require.True(t, tableExists(t, db, "platform"))  // prev still exists
}

func TestPlatformSeed_Idempotent(t *testing.T) {
	_, db := upWithSeed(t)
	seed(t, db) // seed a second time
	require.Equal(t, 3, countRows(t, db, "platform"))
}

func TestPlatform_CodeCheck(t *testing.T) {
	_, db := upWithSeed(t)
	_, err := db.Exec(`INSERT INTO platform(id,code,country) VALUES (9,'amazon','VN')`)
	require.Error(t, err) // CHECK code IN (...)
}

func TestPlatform_CountryCheck(t *testing.T) {
	_, db := upWithSeed(t)
	_, err := db.Exec(`INSERT INTO platform(id,code,country) VALUES (9,'shopee','Vietnam')`)
	require.Error(t, err) // CHECK country ~ '^[A-Z]{2}$'
}

func TestAppUser_EmailCaseInsensitiveUnique(t *testing.T) {
	_, db := upWithSeed(t)
	_, err1 := db.Exec(`INSERT INTO app_user(email) VALUES ('A@x.com')`)
	require.NoError(t, err1)
	_, err2 := db.Exec(`INSERT INTO app_user(email) VALUES ('a@x.com')`)
	require.Error(t, err2) // CITEXT unique
}

func TestAppUser_Defaults(t *testing.T) {
	_, db := upWithSeed(t)
	var locale, status string
	err := db.QueryRow(`INSERT INTO app_user(email) VALUES ('d@x.com') RETURNING locale, status`).Scan(&locale, &status)
	require.NoError(t, err)
	require.Equal(t, "vi-VN", locale)
	require.Equal(t, "active", status)
}
