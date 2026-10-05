package fetchsms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestGetNumber(t *testing.T) {
	for _, tt := range []struct {
		name         string
		slug         string
		status       int
		wantRequests int32
	}{
		{"valid slug", "telegram", http.StatusCreated, 1},
		{"blank slug", "", http.StatusCreated, 0},
		{"invalid slug", "Telegram/../../wallet", http.StatusCreated, 0},
		{"API failure", "telegram", http.StatusPaymentRequired, 1},
		{"rate limited without retry", "telegram", http.StatusTooManyRequests, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method != http.MethodPost || r.URL.Path != "/verifications" || r.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				if r.Header.Get("Authorization") != "Bearer key" || r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing authentication or JSON content type")
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if len(body) != 1 || body["service"] != tt.slug {
					t.Errorf("unexpected purchase body: %#v", body)
				}
				w.WriteHeader(tt.status)
				if tt.status == http.StatusCreated {
					fmt.Fprint(w, verificationJSON)
				} else {
					fmt.Fprint(w, `{"detail":"purchase rejected"}`)
				}
			}))
			defer s.Close()
			c, err := NewClient("key", WithBaseURL(s.URL))
			if err != nil {
				t.Fatal(err)
			}
			v, err := c.GetNumber(context.Background(), tt.slug)
			if got := requests.Load(); got != tt.wantRequests {
				t.Fatalf("requests = %d, want %d", got, tt.wantRequests)
			}
			if tt.wantRequests == 0 {
				_, wantErr := ServiceSlug(tt.slug)
				if err == nil || wantErr == nil || err.Error() != wantErr.Error() || v != (Verification{}) {
					t.Fatalf("invalid slug: verification=%+v error=%v; want ServiceSlug error %v", v, err, wantErr)
				}
			} else if tt.status != http.StatusCreated {
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != tt.status || string(apiErr.Detail) != `"purchase rejected"` {
					t.Fatalf("API failure not propagated: %v", err)
				}
			} else if err != nil || v.ID != "v-1" || v.Number != "+1 332 555 0184" || v.Service != tt.slug || v.CostCents != 70 || v.Status != StatusWaiting {
				t.Fatalf("verification=%+v error=%v", v, err)
			}
		})
	}
}
