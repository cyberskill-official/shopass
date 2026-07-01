package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockNotifier struct {
	sentTokens map[int64]string
}

func (m *mockNotifier) SendReset(ctx context.Context, u AppUser, token string) error {
	if m.sentTokens == nil {
		m.sentTokens = make(map[int64]string)
	}
	m.sentTokens[u.ID] = token
	return nil
}

type testLifecycleSuite struct {
	repo   *pgRepo
	notif  *mockNotifier
	ls     *LifecycleService
	ts     *TokenService
	params Argon2Params
}

func newTestLifecycleSuite(t *testing.T) *testLifecycleSuite {
	db := setupTestDB(t)
	// create password_reset table for tests
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS app_user (
			id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			phone TEXT,
			locale TEXT,
			status TEXT DEFAULT 'active',
			pwd_hash TEXT NOT NULL,
			referral_code_id BIGINT,
			tier TEXT DEFAULT 'silver'
		);
		CREATE TABLE IF NOT EXISTS password_reset (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token_hash TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS platform_account (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			platform_id SMALLINT NOT NULL,
			ext_user_ref TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Exec(`DROP TABLE IF EXISTS password_reset CASCADE;`)
		db.Exec(`DROP TABLE IF EXISTS platform_account CASCADE;`)
		db.Exec(`DROP TABLE IF EXISTS app_user CASCADE;`)
	})

	repo := NewRepo(db).(*pgRepo)
	notif := &mockNotifier{}
	params := defaultParams
	ls := NewLifecycleService(repo, notif, params)
	
	ts := NewTokenService(repo, "test-auth", "test-gw", time.Minute)
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	ts.AddSigningKey("key1", priv)

	return &testLifecycleSuite{repo, notif, ls, ts, params}
}

func (suite *testLifecycleSuite) seedUser(t *testing.T, email string) int64 {
	return suite.seedActiveUser(t, email, "p@ssword123")
}

func (suite *testLifecycleSuite) seedActiveUser(t *testing.T, email, pwd string) int64 {
	hash, err := Hash(pwd, suite.params)
	require.NoError(t, err)
	u := AppUser{
		Email:   email,
		Status:  "active",
		PwdHash: hash,
	}
	id, err := suite.repo.InsertUser(context.Background(), u)
	require.NoError(t, err)
	return id
}

func (suite *testLifecycleSuite) seedUserWithLinks(t *testing.T, email string) int64 {
	id := suite.seedUser(t, email)
	err := suite.repo.UpsertPlatformAccount(context.Background(), PlatformAccount{
		UserID:     id,
		PlatformID: 1,
		ExtUserRef: "google_123",
	})
	require.NoError(t, err)
	return id
}

func (suite *testLifecycleSuite) issueResetFor(t *testing.T, email string) string {
	u, ok := suite.repo.FindByIdentifier(context.Background(), email)
	require.True(t, ok)
	err := suite.ls.RequestReset(context.Background(), email)
	require.NoError(t, err)
	token := suite.notif.sentTokens[u.ID]
	require.NotEmpty(t, token)
	return token
}
