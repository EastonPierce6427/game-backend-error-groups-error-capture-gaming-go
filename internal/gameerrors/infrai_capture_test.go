package gameerrors

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCaptureRetriesWithStableIdentity(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.Path != "/v1/errors/capture" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("Idempotency-Key") != "run-17" {
			t.Fatal("request identity headers are incorrect")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"exception":"round close transaction failed"`) {
			t.Fatal("exception payload is missing")
		}
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"message":"retry later"},"metadata":{}}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data":{"event_id":"event-17"},"error":null,"metadata":{}}`))
	}))
	defer server.Close()

	client := NewCaptureClient("test-key")
	client.endpoint = server.URL + "/v1/errors/capture"
	client.sleep = func(context.Context, time.Duration) error { return nil }
	err := client.Capture(context.Background(), "run-17", Capture{Exception: "round close transaction failed"})
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}
