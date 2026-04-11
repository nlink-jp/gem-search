package gemini

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveRedirect(t *testing.T) {
	targetURL := "https://example.com/actual-page"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", targetURL)
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	c := &Client{http: srv.Client()}
	resolved, err := c.resolveRedirect(t.Context(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != targetURL {
		t.Errorf("resolved = %q, want %q", resolved, targetURL)
	}
}

func TestResolveRedirectNoLocation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{http: srv.Client()}
	resolved, err := c.resolveRedirect(t.Context(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return original URI when no redirect
	if resolved != srv.URL {
		t.Errorf("resolved = %q, want %q", resolved, srv.URL)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err       string
		retryable bool
	}{
		{"rpc error: code = Unavailable", true},
		{"status 429: rate limited", true},
		{"status 503: service unavailable", true},
		{"connection refused", true},
		{"status 400: bad request", false},
		{"invalid argument", false},
	}
	for _, tt := range tests {
		got := isRetryable(fmt.Errorf("%s", tt.err))
		if got != tt.retryable {
			t.Errorf("isRetryable(%q) = %v, want %v", tt.err, got, tt.retryable)
		}
	}
}
