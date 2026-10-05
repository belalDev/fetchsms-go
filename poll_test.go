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

func TestWaitForCode(t *testing.T) {
	n := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n == 1 {
			fmt.Fprint(w, `{"status":"waiting","code":null}`)
		} else {
			fmt.Fprint(w, `{"status":"received","code":"0123"}`)
		}
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL), WithPollInterval(time.Millisecond))
	code, e := c.WaitForCode(nil, "v-1")
	if e != nil || code != "0123" || n != 2 {
		t.Fatal(code, e, n)
	}
	if _, e := NewClient("", WithPollInterval(0)); e == nil {
		t.Fatal("zero interval")
	}
}
func TestWaitTerminal(t *testing.T) {
	for _, status := range []string{"expired", "cancelled", "new-status"} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, `{"status":%q,"code":"old"}`, status) }))
		c, _ := NewClient("key", WithBaseURL(s.URL))
		_, e := c.WaitForCode(nil, "v-1")
		var terminal *TerminalStatusError
		if !errors.As(e, &terminal) || string(terminal.Status) != status {
			t.Fatal(e)
		}
		s.Close()
	}
}
func TestWaitDeadline(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"status":"waiting"}`) }))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, e := c.WaitForCode(ctx, "v-1")
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
