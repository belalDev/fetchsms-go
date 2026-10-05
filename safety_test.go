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

func TestMutationsNeverRetry(t *testing.T) {
	ref, _ := ServiceID(4)
	unlimited, _ := ServiceID(1)
	for _, status := range []int{429, 500} {
		for _, call := range []func(*Client) error{func(c *Client) error {
			_, e := c.CreateVerification(nil, CreateVerificationRequest{Service: ref})
			return e
		}, func(c *Client) error { _, e := c.ReuseVerification(nil, "v1"); return e }, func(c *Client) error { return c.CancelVerification(nil, "v1") }, func(c *Client) error {
			_, e := c.CreateRental(nil, CreateRentalRequest{Service: unlimited, Days: 1})
			return e
		}, func(c *Client) error { _, e := c.ExtendRental(nil, "r1", 1); return e }, func(c *Client) error { _, e := c.CancelRental(nil, "r1"); return e }} {
			n := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n++
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(status)
			}))
			c, _ := NewClient("key", WithBaseURL(s.URL))
			if e := call(c); e == nil || n != 1 {
				t.Fatal(e, n)
			}
			s.Close()
		}
	}
}
func TestBoundedSuccessAndMalformedJSON(t *testing.T) {
	for _, body := range []string{strings.Repeat("x", int(MaxResponseBytes)+1), "not-json", `[] []`} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
		c, _ := NewClient("", WithBaseURL(s.URL))
		_, e := c.Services(nil)
		if e == nil {
			t.Fatal("accepted invalid response")
		}
		if len(body) > int(MaxResponseBytes) && !errors.Is(e, ErrResponseTooLarge) {
			t.Fatal(e)
		}
		s.Close()
	}
}
func TestErrorDetail(t *testing.T) {
	for _, detail := range []string{`"secret detail"`, `[{"msg":"bad"}]`, `{"code":5}`} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(422)
			fmt.Fprintf(w, `{"detail":%s}`, detail)
		}))
		c, _ := NewClient("", WithBaseURL(s.URL))
		_, e := c.Services(nil)
		var api *APIError
		if !errors.As(e, &api) || string(api.Detail) != detail || strings.Contains(e.Error(), "secret") {
			t.Fatal(e)
		}
		s.Close()
	}
}
func TestPollNonRetryErrorAndCancelled(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) }))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	_, e := c.WaitForCode(nil, "v1")
	var api *APIError
	if !errors.As(e, &api) || api.StatusCode != 401 {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e = c.WaitForCode(ctx, "v1")
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}
func TestRetryAfterFormats(t *testing.T) {
	for _, s := range []string{"garbage", "-2", ""} {
		if retryAfter(s) != 0 {
			t.Fatal(s)
		}
	}
	if retryAfter(time.Now().Add(3*time.Second).UTC().Format(http.TimeFormat)) <= 0 {
		t.Fatal("HTTP date")
	}
}
