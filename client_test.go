package fetchsms

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTransportSafety(t *testing.T) {
	for _, o := range []Option{WithBaseURL("relative"), WithBaseURL("https://user:pass@example.com"), WithBaseURL("https://example.com?q=x"), WithHTTPClient(nil), nil} {
		if _, err := NewClient("key", o); err == nil {
			t.Error("accepted invalid option")
		}
	}
	for _, body := range []string{`{"detail":[{"msg":"invalid"}]}`, `not json`, strings.Repeat("x", int(MaxResponseBytes)+1)} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "3")
			w.WriteHeader(429)
			fmt.Fprint(w, body)
		}))
		c, _ := NewClient("", WithBaseURL(s.URL))
		_, err := c.Services(nil)
		var api *APIError
		if !errors.As(err, &api) || api.StatusCode != 429 || api.RetryAfter != 3*time.Second || api.RetryAfterRaw != "3" {
			t.Fatalf("%#v", err)
		}
		if len(api.Body) > int(MaxResponseBytes) {
			t.Fatal("unbounded body")
		}
		s.Close()
	}
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("followed redirect") }))
	defer target.Close()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
	defer s.Close()
	h := s.Client()
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return nil }
	c, _ := NewClient("", WithBaseURL(s.URL), WithHTTPClient(h))
	if _, err := c.Services(context.Background()); err == nil {
		t.Fatal("accepted redirect")
	}
}

func TestServices(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/services" || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected request: %s %s auth=%q", r.Method, r.URL, r.Header.Get("Authorization"))
		}
		fmt.Fprint(w, `[{"id":4,"slug":"telegram","name":"Telegram","price_cents":70,"long_prices":{},"short_available":3,"long_available":0}]`)
	}))
	defer s.Close()
	c, err := NewClient("secret", WithBaseURL(s.URL+"/v1"), WithHTTPClient(s.Client()))
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Services(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != 4 || got[0].PriceCents != 70 {
		t.Fatalf("%+v", got)
	}
}
