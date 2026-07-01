package fcm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// --- test helpers ---

type staticToken struct{ tok string }

func (s *staticToken) Token(ctx context.Context) (string, error) { return s.tok, nil }

func stubFCM(t *testing.T, code int, body string, headers http.Header) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vs := range headers {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(code)
		io.WriteString(w, body)
	}))
}

func captureRequest(t *testing.T, gotPath *string, code int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotPath = r.URL.Path
		w.WriteHeader(code)
		io.WriteString(w, body)
	}))
}

func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c := NewClient("test-project", &staticToken{tok: "test-token"}, http.DefaultClient)
	c.baseURL = serverURL
	return c
}

// --- client_test ---

func TestSend_Success_MarksSent(t *testing.T) {
	srv := stubFCM(t, 200, `{"name":"projects/p/messages/123"}`, nil)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	out, err := c.Send(context.Background(), Message{Token: "tok-ok"})
	require.NoError(t, err)
	require.Equal(t, ResultSent, out.Result)
}

func TestSend_UsesHTTPv1Endpoint_NotLegacy(t *testing.T) {
	var gotPath string
	srv := captureRequest(t, &gotPath, 200, `{"name":"n"}`)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	c.Send(context.Background(), Message{Token: "tok"})
	require.True(t, strings.Contains(gotPath, "/v1/projects/"))
	require.True(t, strings.Contains(gotPath, "messages:send"))
	require.False(t, strings.Contains(gotPath, "/fcm/send")) // legacy not used (DEC-NOTIF-10)
}

func TestSend_429_TriggersRetry_NotDropped(t *testing.T) {
	srv := stubFCM(t, 429, `{"error":{"status":"RESOURCE_EXHAUSTED"}}`, nil)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	out, _ := c.Send(context.Background(), Message{Token: "tok"})
	require.Equal(t, ResultRetry, out.Result) // not dropped, will retry (DEC-NOTIF-12)
}

func TestSend_429_RespectsRetryAfter(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "30")
	srv := stubFCM(t, 429, `{}`, h)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	out, _ := c.Send(context.Background(), Message{Token: "tok"})
	require.Equal(t, 30*time.Second, out.RetryAfter) // respects FCM's Retry-After
	require.Equal(t, 30*time.Second, nextDelay(0, out.RetryAfter))
}

func TestSend_Unregistered_MarksTokenDead(t *testing.T) {
	srv := stubFCM(t, 404, `{"error":{"status":"UNREGISTERED"}}`, nil)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	out, _ := c.Send(context.Background(), Message{Token: "tok-dead"})
	require.Equal(t, ResultTokenDead, out.Result)
}

func TestSend_400_InvalidArgument_TokenDead(t *testing.T) {
	srv := stubFCM(t, 400, `{"error":{"status":"INVALID_ARGUMENT"}}`, nil)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	out, _ := c.Send(context.Background(), Message{Token: "bad-format"})
	require.Equal(t, ResultTokenDead, out.Result)
}

func TestSend_500_Retry(t *testing.T) {
	srv := stubFCM(t, 500, `{}`, nil)
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	out, _ := c.Send(context.Background(), Message{Token: "tok"})
	require.Equal(t, ResultRetry, out.Result)
}

func TestParseRetryAfter_Seconds(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "45")
	require.Equal(t, 45*time.Second, parseRetryAfter(h))
}

func TestParseRetryAfter_Empty(t *testing.T) {
	require.Equal(t, time.Duration(0), parseRetryAfter(http.Header{}))
}
