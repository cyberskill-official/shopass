package apns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type staticTok struct{}

func (staticTok) Bearer(ctx context.Context) (string, error) { return "jwt", nil }

func TestAPNs_Send_ClassifiesStatus(t *testing.T) {
	cases := []struct {
		code int
		want SendResult
	}{
		{200, ResultSent},
		{410, ResultTokenDead},
		{429, ResultRetry},
		{500, ResultRetry},
		{403, ResultFailed},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "bearer jwt", r.Header.Get("authorization"))
			w.WriteHeader(tc.code)
			io.WriteString(w, "{}")
		}))
		c := NewClient("example.invalid", "bundle.id", staticTok{}, srv.Client())
		// Force host via custom transport by replacing Doer
		c.host = "127.0.0.1"
		c.http = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(req)
		})
		res, err := c.Send(context.Background(), "devtoken", []byte(`{"aps":{}}`))
		require.NoError(t, err)
		require.Equal(t, tc.want, res)
		srv.Close()
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
