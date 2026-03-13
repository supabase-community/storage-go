package storage_go

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewClientHeaders verifies that NewClient sets the correct HTTP headers
// for both legacy JWT tokens and new-style Supabase API keys.
func TestNewClientHeaders(t *testing.T) {
	tests := []struct {
		name              string
		token             string
		wantApikey        string
		wantAuthorization string
	}{
		{
			name:              "legacy JWT token gets Bearer prefix",
			token:             "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantApikey:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantAuthorization: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
		},
		{
			name:              "new publishable key does not get Bearer prefix",
			token:             "sb_publishable_abc123",
			wantApikey:        "sb_publishable_abc123",
			wantAuthorization: "sb_publishable_abc123",
		},
		{
			name:              "new secret key does not get Bearer prefix",
			token:             "sb_secret_xyz789",
			wantApikey:        "sb_secret_xyz789",
			wantAuthorization: "sb_secret_xyz789",
		},
		{
			name:              "empty token falls back to Bearer scheme",
			token:             "",
			wantApikey:        "",
			wantAuthorization: "Bearer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedReq *http.Request
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("[]"))
			}))
			defer srv.Close()

			c := NewClient(srv.URL, tc.token, map[string]string{})
			req, err := c.NewRequest(http.MethodGet, "/bucket")
			if err != nil {
				t.Fatalf("NewRequest failed: %v", err)
			}
			if _, err = c.Do(req, nil); err != nil {
				t.Fatalf("Do failed: %v", err)
			}

			gotApikey := capturedReq.Header.Get("Apikey")
			gotAuthorization := capturedReq.Header.Get("Authorization")

			if gotApikey != tc.wantApikey {
				t.Errorf("apikey header: got %q, want %q", gotApikey, tc.wantApikey)
			}
			if gotAuthorization != tc.wantAuthorization {
				t.Errorf("Authorization header: got %q, want %q", gotAuthorization, tc.wantAuthorization)
			}
		})
	}
}
