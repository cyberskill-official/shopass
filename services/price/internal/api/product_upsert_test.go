package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/price/internal/price"
)

type stubProductUpserter struct {
	got    price.TrackedProduct
	called int
	out    price.TrackedProduct
	err    error
}

func (s *stubProductUpserter) Upsert(_ context.Context, p price.TrackedProduct) (price.TrackedProduct, error) {
	s.called++
	s.got = p
	if s.err != nil {
		return price.TrackedProduct{}, s.err
	}
	return s.out, nil
}

func productUpsertRequestForTest(token, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, productUpsertPath, io.NopCloser(bytes.NewBufferString(body)))
	if token != "" {
		req.Header.Set("X-Service-Token", token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestProductUpsertRequiresConfiguredMatchingServiceToken(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		supplied   string
		want       int
	}{
		{name: "missing configuration fails closed", supplied: "anything", want: http.StatusServiceUnavailable},
		{name: "missing header", configured: "secret", want: http.StatusUnauthorized},
		{name: "wrong header", configured: "secret", supplied: "wrong", want: http.StatusUnauthorized},
		{name: "matching header", configured: "secret", supplied: "secret", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubProductUpserter{out: price.TrackedProduct{ID: 41}}
			h := NewProductUpsertHandler(repo, tt.configured)
			rr := httptest.NewRecorder()
			h.HandleUpsert(rr, productUpsertRequestForTest(tt.supplied,
				`{"platform_id":1,"platform_item_id":"20114455667:88123","shop_id":"88123"}`))

			require.Equal(t, tt.want, rr.Code)
			if tt.want == http.StatusOK {
				require.Equal(t, 1, repo.called)
			} else {
				require.Zero(t, repo.called)
			}
		})
	}
}

func TestProductUpsertPassesValidatedProductToRepo(t *testing.T) {
	repo := &stubProductUpserter{out: price.TrackedProduct{ID: 42}}
	h := NewProductUpsertHandler(repo, "secret")
	rr := httptest.NewRecorder()
	h.HandleUpsert(rr, productUpsertRequestForTest("secret",
		`{"platform_id":1,"platform_item_id":"20114455667:88123","shop_id":"88123","title":"Tai nghe"}`))

	require.Equal(t, http.StatusOK, rr.Code)
	require.EqualValues(t, 1, repo.got.PlatformID)
	require.Equal(t, "20114455667:88123", repo.got.PlatformItemID)
	require.NotNil(t, repo.got.ShopID)
	require.Equal(t, "88123", *repo.got.ShopID)
	require.NotNil(t, repo.got.Title)
	require.Equal(t, "Tai nghe", *repo.got.Title)
	require.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))

	var got productUpsertResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.EqualValues(t, 42, got.ID)
}

func TestProductUpsertRejectsInvalidBodyBeforeRepo(t *testing.T) {
	repo := &stubProductUpserter{out: price.TrackedProduct{ID: 42}}
	h := NewProductUpsertHandler(repo, "secret")

	for _, body := range []string{
		`{}`,
		`{"platform_id":1,"platform_item_id":""}`,
		`{"platform_id":1,"platform_item_id":"x","unexpected":true}`,
		`{"platform_id":1,"platform_item_id":"x"}{}`,
	} {
		rr := httptest.NewRecorder()
		h.HandleUpsert(rr, productUpsertRequestForTest("secret", body))
		require.Equal(t, http.StatusBadRequest, rr.Code, body)
	}
	require.Zero(t, repo.called)
}

func TestProductUpsertHidesRepositoryFailure(t *testing.T) {
	repo := &stubProductUpserter{err: errors.New("database is unavailable")}
	h := NewProductUpsertHandler(repo, "secret")
	rr := httptest.NewRecorder()
	h.HandleUpsert(rr, productUpsertRequestForTest("secret", `{"platform_id":1,"platform_item_id":"x"}`))

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NotContains(t, rr.Body.String(), "database")
}
