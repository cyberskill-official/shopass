package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchema_NoCredentialColumns(t *testing.T) {
	// This test asserts the struct has no tokens. We can also verify against the DB in integration test.
	db := setupTestDB(t) // Reuse setupTestDB from token_test.go
	
	// Create platform table mock just for this test so platform_account can be created
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS platform (
			id SMALLSERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);
		INSERT INTO platform (id, name) VALUES (1, 'shopee') ON CONFLICT DO NOTHING;
		INSERT INTO platform (id, name) VALUES (2, 'lazada') ON CONFLICT DO NOTHING;

		CREATE TABLE IF NOT EXISTS app_user (
			id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE
		);
		INSERT INTO app_user (id, email) VALUES (1, 'user1@test.com') ON CONFLICT DO NOTHING;
		INSERT INTO app_user (id, email) VALUES (2, 'user2@test.com') ON CONFLICT DO NOTHING;
		
		CREATE TABLE IF NOT EXISTS platform_account (
		  id           BIGSERIAL    PRIMARY KEY,
		  user_id      BIGINT       NOT NULL REFERENCES app_user(id),
		  platform_id  SMALLINT     NOT NULL REFERENCES platform(id),
		  ext_user_ref TEXT         NOT NULL
		                 CHECK (length(ext_user_ref) BETWEEN 1 AND 128),
		  linked_at    TIMESTAMPTZ  DEFAULT now(),
		  UNIQUE (user_id, platform_id)
		);
	`)
	require.NoError(t, err)

	rows, err := db.Query(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'platform_account'
	`)
	require.NoError(t, err)
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		require.NoError(t, err)
		cols = append(cols, name)
	}
	
	for _, c := range cols {
		lower := strings.ToLower(c)
		require.NotContains(t, lower, "token")
		require.NotContains(t, lower, "cookie")
		require.NotContains(t, lower, "session")
		require.NotContains(t, lower, "password")
	}
}

func setupServiceWithDB(t *testing.T) *Service {
	db := setupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS platform (
			id SMALLSERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);
		INSERT INTO platform (id, name) VALUES (1, 'shopee') ON CONFLICT DO NOTHING;
		INSERT INTO platform (id, name) VALUES (2, 'lazada') ON CONFLICT DO NOTHING;

		CREATE TABLE IF NOT EXISTS app_user (
			id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE
		);
		INSERT INTO app_user (id, email) VALUES (1, 'user1@test.com') ON CONFLICT DO NOTHING;
		INSERT INTO app_user (id, email) VALUES (2, 'user2@test.com') ON CONFLICT DO NOTHING;
		
		CREATE TABLE IF NOT EXISTS platform_account (
		  id           BIGSERIAL    PRIMARY KEY,
		  user_id      BIGINT       NOT NULL REFERENCES app_user(id),
		  platform_id  SMALLINT     NOT NULL REFERENCES platform(id),
		  ext_user_ref TEXT         NOT NULL
		                 CHECK (length(ext_user_ref) BETWEEN 1 AND 128),
		  linked_at    TIMESTAMPTZ  DEFAULT now(),
		  UNIQUE (user_id, platform_id)
		);
	`)
	require.NoError(t, err)
	repo := NewRepo(db)
	return NewService(repo, defaultParams)
}

func TestLink_NewAndUpsert(t *testing.T) {
	s := setupServiceWithDB(t)
	ctx := context.Background()
	var u1 int64 = 1
	var shopee int16 = 1

	require.NoError(t, s.LinkAccount(ctx, u1, shopee, "ref-abc"))
	require.NoError(t, s.LinkAccount(ctx, u1, shopee, "ref-xyz")) // cùng sàn → upsert

	links, _ := s.ListLinks(ctx, u1)
	require.Len(t, links, 1)
	require.Equal(t, "ref-xyz", links[0].ExtUserRef)
}

func TestLink_MultiPlatform(t *testing.T) {
	s := setupServiceWithDB(t)
	ctx := context.Background()
	var u1 int64 = 1
	var shopee int16 = 1
	var lazada int16 = 2

	s.LinkAccount(ctx, u1, shopee, "ref-1")
	s.LinkAccount(ctx, u1, lazada, "ref-2")

	links, _ := s.ListLinks(ctx, u1)
	require.Len(t, links, 2)
}

func TestLink_RejectsRawCredential(t *testing.T) {
	s := setupServiceWithDB(t)
	ctx := context.Background()
	var u1 int64 = 1
	var shopee int16 = 1

	require.ErrorIs(t, s.LinkAccount(ctx, u1, shopee, ""), ErrInvalidExtRef)
	require.ErrorIs(t, s.LinkAccount(ctx, u1, shopee, "chi@gmail.com"), ErrExtRefNotAnonymized)
	require.ErrorIs(t, s.LinkAccount(ctx, u1, shopee, "session_id=123"), ErrExtRefNotAnonymized)
	require.ErrorIs(t, s.LinkAccount(ctx, u1, shopee, "eyJhbGciOiJIUzI1NiI"), ErrExtRefNotAnonymized)
}

func TestList_IsolatedPerUser(t *testing.T) {
	s := setupServiceWithDB(t)
	ctx := context.Background()
	var u1 int64 = 1
	var u2 int64 = 2
	var shopee int16 = 1

	s.LinkAccount(ctx, u1, shopee, "ref-1")

	out, _ := s.ListLinks(ctx, u2)
	require.Empty(t, out) // u2 không thấy liên kết của u1
}

func TestUnlink_Idempotent(t *testing.T) {
	s := setupServiceWithDB(t)
	ctx := context.Background()
	var u1 int64 = 1
	var shopee int16 = 1

	s.LinkAccount(ctx, u1, shopee, "ref-1")
	require.NoError(t, s.UnlinkAccount(ctx, u1, shopee))
	require.NoError(t, s.UnlinkAccount(ctx, u1, shopee)) // lần hai không lỗi
}
