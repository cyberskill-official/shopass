package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// --- fakes ---

type fakeSocialRepo struct {
	socials map[string]int64 // "provider:subject" -> uid
	emails  map[string]int64 // email -> existing uid
	nextID  int64
	created []struct{ email, locale string }
	links   []struct {
		uid               int64
		provider, subject string
	}
}

func newFakeSocialRepo() *fakeSocialRepo {
	return &fakeSocialRepo{socials: map[string]int64{}, emails: map[string]int64{}, nextID: 100}
}

func (f *fakeSocialRepo) FindBySocial(_ context.Context, provider, subject string) (int64, bool, error) {
	uid, ok := f.socials[provider+":"+subject]
	return uid, ok, nil
}
func (f *fakeSocialRepo) FindByEmail(_ context.Context, email string) (AppUser, error) {
	if uid, ok := f.emails[email]; ok {
		return AppUser{ID: uid, Email: email}, nil
	}
	return AppUser{}, nil // not found -> ID 0, nil err (matches pgRepo)
}
func (f *fakeSocialRepo) LinkSocial(_ context.Context, uid int64, provider, subject string) error {
	f.socials[provider+":"+subject] = uid
	f.links = append(f.links, struct {
		uid               int64
		provider, subject string
	}{uid, provider, subject})
	return nil
}
func (f *fakeSocialRepo) CreateSocialUser(_ context.Context, email, locale string) (int64, error) {
	f.nextID++
	id := f.nextID
	f.created = append(f.created, struct{ email, locale string }{email, locale})
	if email != "" {
		f.emails[email] = id
	}
	return id, nil
}

type fakeProvider struct {
	claims                              OIDCClaims
	err                                 error
	lastState, lastChallenge, lastNonce string
}

func (p *fakeProvider) AuthCodeURL(state, challenge, nonce string) string {
	p.lastState, p.lastChallenge, p.lastNonce = state, challenge, nonce
	return "https://provider/authorize?state=" + state + "&code_challenge=" + challenge + "&nonce=" + nonce
}
func (p *fakeProvider) ExchangeAndVerify(_ context.Context, _, _, _ string) (OIDCClaims, error) {
	if p.err != nil {
		return OIDCClaims{}, p.err
	}
	return p.claims, nil
}

type fakeIssuer struct{ lastUID int64 }

func (i *fakeIssuer) IssueTokenPair(_ context.Context, uid int64) (TokenPair, error) {
	i.lastUID = uid
	return TokenPair{Access: "acc", Refresh: "ref", Type: "Bearer", Expires: 900}, nil
}

// --- resolveUser: the account-linking safety matrix (§1 #6, #7) ---

func TestResolveUser_ExistingSocialIdentityWins(t *testing.T) {
	repo := newFakeSocialRepo()
	repo.socials["google:sub-1"] = 42
	s := &OAuthService{repo: repo}
	uid, err := s.resolveUser(context.Background(), "google", OIDCClaims{Subject: "sub-1", Email: "x@e.com", EmailVerified: true})
	if err != nil || uid != 42 {
		t.Fatalf("want existing uid 42, got %d err %v", uid, err)
	}
	if len(repo.created) != 0 {
		t.Fatal("must not create a user when the social identity exists")
	}
}

func TestResolveUser_VerifiedEmailMergesIntoExistingAccount(t *testing.T) {
	repo := newFakeSocialRepo()
	repo.emails["chi@example.com"] = 7 // existing password account
	s := &OAuthService{repo: repo}
	uid, err := s.resolveUser(context.Background(), "google", OIDCClaims{Subject: "sub-9", Email: "chi@example.com", EmailVerified: true})
	if err != nil || uid != 7 {
		t.Fatalf("verified email should link to existing 7, got %d err %v", uid, err)
	}
	if len(repo.created) != 0 {
		t.Fatal("must link, not create, on a verified-email match")
	}
	if len(repo.links) != 1 || repo.links[0].uid != 7 {
		t.Fatalf("expected a link to user 7, got %+v", repo.links)
	}
}

func TestResolveUser_UnverifiedEmailDoesNotMerge(t *testing.T) {
	// The account-takeover case: a provider account claims the victim's email
	// but has NOT verified it. It must not merge into the victim, and must not
	// create a colliding email either.
	repo := newFakeSocialRepo()
	repo.emails["victim@example.com"] = 7
	s := &OAuthService{repo: repo}
	uid, err := s.resolveUser(context.Background(), "google", OIDCClaims{Subject: "attacker-sub", Email: "victim@example.com", EmailVerified: false})
	if err != nil {
		t.Fatal(err)
	}
	if uid == 7 {
		t.Fatal("must NOT link an unverified email into the victim account")
	}
	if len(repo.created) != 1 || repo.created[0].email != "" {
		t.Fatalf("new account must be created with a NULL/empty email, got %+v", repo.created)
	}
	for _, l := range repo.links {
		if l.uid == 7 {
			t.Fatal("victim account must receive no social link")
		}
	}
}

func TestResolveUser_NewVerifiedUser(t *testing.T) {
	repo := newFakeSocialRepo()
	s := &OAuthService{repo: repo}
	uid, err := s.resolveUser(context.Background(), "google", OIDCClaims{Subject: "sub-new", Email: "new@example.com", EmailVerified: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.created) != 1 || repo.created[0].email != "new@example.com" {
		t.Fatalf("expected a new account with the verified email, got %+v", repo.created)
	}
	if uid == 0 {
		t.Fatal("expected a real new uid")
	}
}

// --- StartOAuth / OAuthCallback flow (§1 #2, #3, #8, #11) ---

func newFlow(prov *fakeProvider, repo *fakeSocialRepo, iss *fakeIssuer) *OAuthService {
	return NewOAuthService(map[string]OAuthProvider{"google": prov}, NewMemTmpStore(), repo, iss)
}

func TestStartOAuth_UnknownProvider(t *testing.T) {
	s := newFlow(&fakeProvider{}, newFakeSocialRepo(), &fakeIssuer{})
	if _, err := s.StartOAuth(context.Background(), "myspace"); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("want ErrUnknownProvider, got %v", err)
	}
}

func TestOAuthFlow_HappyPathAndOneTimeState(t *testing.T) {
	prov := &fakeProvider{claims: OIDCClaims{Subject: "sub-1", Email: "chi@example.com", EmailVerified: true}}
	iss := &fakeIssuer{}
	s := newFlow(prov, newFakeSocialRepo(), iss)
	ctx := context.Background()

	url, err := s.StartOAuth(ctx, "google")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(url, "code_challenge=") || !strings.Contains(url, "nonce=") || prov.lastState == "" {
		t.Fatalf("authorize url missing pkce/state/nonce: %q", url)
	}

	pair, err := s.OAuthCallback(ctx, "google", "auth-code", prov.lastState)
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if pair.Access == "" || iss.lastUID == 0 {
		t.Fatalf("expected a token pair for a resolved user, got %+v uid %d", pair, iss.lastUID)
	}

	// Replaying the same state must fail: state is single-use (§1 #11).
	if _, err := s.OAuthCallback(ctx, "google", "auth-code", prov.lastState); !errors.Is(err, ErrBadState) {
		t.Fatalf("replayed state should be ErrBadState, got %v", err)
	}
}

func TestOAuthCallback_BadState(t *testing.T) {
	s := newFlow(&fakeProvider{}, newFakeSocialRepo(), &fakeIssuer{})
	if _, err := s.OAuthCallback(context.Background(), "google", "code", "never-issued"); !errors.Is(err, ErrBadState) {
		t.Fatalf("want ErrBadState, got %v", err)
	}
}

func TestOAuthCallback_VerifyErrorPropagates(t *testing.T) {
	prov := &fakeProvider{err: errors.New("id_token bad")}
	s := newFlow(prov, newFakeSocialRepo(), &fakeIssuer{})
	ctx := context.Background()
	if _, err := s.StartOAuth(ctx, "google"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OAuthCallback(ctx, "google", "code", prov.lastState); err == nil {
		t.Fatal("expected the provider verify error to propagate")
	}
}
