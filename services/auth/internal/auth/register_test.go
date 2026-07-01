package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	users     map[int64]AppUser
	byEmail   map[string]AppUser
	nextID    int64
	emailCase func(string) string
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:     make(map[int64]AppUser),
		byEmail:   make(map[string]AppUser),
		nextID:    1,
		emailCase: strings.ToLower,
	}
}

func (m *mockRepo) InsertUser(ctx context.Context, u AppUser) (int64, error) {
	lowerEmail := m.emailCase(u.Email)
	for _, ex := range m.byEmail {
		if m.emailCase(ex.Email) == lowerEmail && lowerEmail != "" {
			return 0, &pq.Error{Code: "23505", Constraint: "app_user_email_key"}
		}
	}
	u.ID = m.nextID
	m.nextID++
	m.users[u.ID] = u
	if lowerEmail != "" {
		m.byEmail[lowerEmail] = u
	}
	return u.ID, nil
}

func (m *mockRepo) FindByEmail(ctx context.Context, email string) (AppUser, error) {
	lowerEmail := m.emailCase(email)
	for _, ex := range m.byEmail {
		if m.emailCase(ex.Email) == lowerEmail {
			return ex, nil
		}
	}
	return AppUser{}, nil
}

func (m *mockRepo) FindRefreshByHash(ctx context.Context, hash string) (RefreshTokenRow, error) {
	return RefreshTokenRow{}, nil
}

func (m *mockRepo) RevokeFamily(ctx context.Context, familyID string) error {
	return nil
}

func (m *mockRepo) MarkUsed(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRepo) InsertRefreshToken(ctx context.Context, userID int64, hash, familyID string, expiresAt time.Time) error {
	return nil
}

func (m *mockRepo) UpsertPlatformAccount(ctx context.Context, pa PlatformAccount) error {
	return nil
}

func (m *mockRepo) ListPlatformAccountsByUser(ctx context.Context, userID int64) ([]PlatformAccount, error) {
	return nil, nil
}

func (m *mockRepo) DeletePlatformAccount(ctx context.Context, userID int64, platformID int16) error {
	return nil
}

func TestRegister_NoIdentifier(t *testing.T) {
	ctx := context.Background()
	s := NewService(newMockRepo(), defaultParams)
	_, err := s.Register(ctx, RegisterInput{Password: "p@ss12345"})
	require.ErrorIs(t, err, ErrNoIdentifier)
}

func TestRegister_WeakPassword(t *testing.T) {
	ctx := context.Background()
	s := NewService(newMockRepo(), defaultParams)
	_, err := s.Register(ctx, RegisterInput{Email: "a@x.com", Password: "123"})
	require.Error(t, err)
}

func TestRegister_DuplicateEmail_CaseInsensitive(t *testing.T) {
	ctx := context.Background()
	s := NewService(newMockRepo(), defaultParams)
	_, err1 := s.Register(ctx, RegisterInput{Email: "Chi@Mail.com", Password: "p@ss12345"})
	require.NoError(t, err1)
	_, err2 := s.Register(ctx, RegisterInput{Email: "chi@mail.com", Password: "p@ss12345"})
	require.ErrorIs(t, err2, ErrEmailTaken)
}

func TestRegister_TrimsEmail(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepo()
	s := NewService(repo, defaultParams)
	id, err := s.Register(ctx, RegisterInput{Email: " a@x.com ", Password: "p@ss12345"})
	require.NoError(t, err)

	u, err := repo.FindByEmail(ctx, "a@x.com")
	require.NoError(t, err)
	require.Equal(t, id, u.ID)
}
