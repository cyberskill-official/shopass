package auth

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// setupSocialDB creates a minimal app_user plus the social_identity table from
// migration 0007 (email as TEXT here to avoid the citext extension dependency).
func setupSocialDB(t *testing.T) (*sql.DB, SocialRepo) {
	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Skip("TEST_DB_URL not set; skipping social_identity integration test")
	}
	db, err := sql.Open("postgres", url)
	require.NoError(t, err)
	ctx := context.Background()
	for _, ddl := range []string{
		`DROP TABLE IF EXISTS social_identity`,
		`DROP TABLE IF EXISTS app_user`,
		// Columns match what pgRepo.FindByEmail selects (id, email, phone, locale,
		// status, pwd_hash, referral_code_id).
		`CREATE TABLE app_user (
			id               BIGSERIAL PRIMARY KEY,
			email            TEXT UNIQUE,
			phone            TEXT,
			locale           TEXT NOT NULL DEFAULT 'vi-VN',
			status           TEXT NOT NULL DEFAULT 'active',
			pwd_hash         TEXT,
			referral_code_id BIGINT
		)`,
		`CREATE TABLE social_identity (
			id         BIGSERIAL PRIMARY KEY,
			user_id    BIGINT NOT NULL REFERENCES app_user(id),
			provider   TEXT NOT NULL CHECK (provider IN ('google','facebook','zalo')),
			subject    TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			UNIQUE (provider, subject)
		)`,
	} {
		_, err := db.ExecContext(ctx, ddl)
		require.NoError(t, err, ddl)
	}
	return db, NewRepo(db).(SocialRepo)
}

func TestSocialRepo_CreateFindLink(t *testing.T) {
	db, repo := setupSocialDB(t)
	defer db.Close()
	ctx := context.Background()

	// Create a social-only user with a verified email, find it back.
	id1, err := repo.CreateSocialUser(ctx, "chi@example.com", "vi-VN")
	require.NoError(t, err)
	require.NotZero(t, id1)
	u, err := repo.FindByEmail(ctx, "chi@example.com")
	require.NoError(t, err)
	require.Equal(t, id1, u.ID)

	// Link a provider subject, then find it.
	require.NoError(t, repo.LinkSocial(ctx, id1, "google", "sub-a"))
	uid, found, err := repo.FindBySocial(ctx, "google", "sub-a")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, id1, uid)

	// Unknown subject is not found.
	_, found, err = repo.FindBySocial(ctx, "google", "nope")
	require.NoError(t, err)
	require.False(t, found)

	// Re-linking the same subject is idempotent (ON CONFLICT DO NOTHING).
	require.NoError(t, repo.LinkSocial(ctx, id1, "google", "sub-a"))
	var n int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM social_identity WHERE provider='google' AND subject='sub-a'`).Scan(&n))
	require.Equal(t, 1, n)
}

func TestSocialRepo_NullEmailForSocialOnly(t *testing.T) {
	db, repo := setupSocialDB(t)
	defer db.Close()
	ctx := context.Background()

	// Two social-only users with no email must both insert (NULL emails do not
	// collide on the UNIQUE(email) constraint).
	a, err := repo.CreateSocialUser(ctx, "", "vi-VN")
	require.NoError(t, err)
	b, err := repo.CreateSocialUser(ctx, "", "vi-VN")
	require.NoError(t, err)
	require.NotEqual(t, a, b)

	var emailIsNull bool
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT email IS NULL FROM app_user WHERE id = $1`, a).Scan(&emailIsNull))
	require.True(t, emailIsNull, "social-only account must have NULL email")
}
