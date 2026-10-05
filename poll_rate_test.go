package fetchsms

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPollRetryAfter(t *testing.T) {
	n := 0
	start := time.Now()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			fmt.Fprint(w, `{"detail":"slow down"}`)
			return
		}
		if time.Since(start) < time.Second {
			t.Error("ignored Retry-After")
		}
		fmt.Fprint(w, `{"status":"received","code":"123"}`)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL), WithPollInterval(time.Millisecond))
	code, e := c.WaitForCode(nil, "v-1")
	if e != nil || code != "123" || n != 2 {
		t.Fatal(code, e, n)
	}
}
func TestPollRetryCancellable(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(429)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, e := c.WaitForCode(ctx, "v-1")
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
