package priceclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientUpsertSendsTokenAndProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != productUpsertPath {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-Service-Token"); got != "service-secret" {
			t.Fatalf("unexpected service token %q", got)
		}
		var got productUpsertRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got.PlatformID != 1 || got.PlatformItemID != "20114455667:88123" {
			t.Fatalf("unexpected product: %+v", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":77}`))
	}))
	defer server.Close()

	client, err := New(server.URL, "service-secret", time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := client.Upsert(context.Background(), TrackedProduct{
		PlatformID: 1, PlatformItemID: "20114455667:88123",
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if out.ID != 77 {
		t.Fatalf("product ID = %d, want 77", out.ID)
	}
}

func TestClientRejectsMissingTokenAndBadURL(t *testing.T) {
	if _, err := New("http://pricesvc:8081", "", time.Second); err == nil {
		t.Fatal("expected missing-token configuration error")
	}
	if _, err := New("not a url", "secret", time.Second); err == nil {
		t.Fatal("expected bad-url configuration error")
	}
}

func TestClientReportsRejectedUpsert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := New(server.URL, "service-secret", time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := client.Upsert(context.Background(), TrackedProduct{PlatformID: 1, PlatformItemID: "1:2"}); err == nil {
		t.Fatal("expected rejected-upsert error")
	}
}

func TestClientRecordsBrowserPriceThroughPrivatePriceService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != snapshotIngestPath {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-Service-Token"); got != "service-secret" {
			t.Fatalf("unexpected service token %q", got)
		}
		var got PriceSnapshot
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got.ProductID != 77 || got.Price != 199000 {
			t.Fatalf("unexpected snapshot: %+v", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"written":true}`))
	}))
	defer server.Close()

	client, err := New(server.URL, "service-secret", time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	written, err := client.RecordBrowserPrice(context.Background(), PriceSnapshot{ProductID: 77, Price: 199000})
	if err != nil || !written {
		t.Fatalf("RecordBrowserPrice = %v, %v", written, err)
	}
}
