package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaitForRentalCode(t *testing.T) {
	n := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if r.URL.Path != "/rentals/r1" {
			t.Error(r.URL)
		}
		if n == 1 {
			w.WriteHeader(429)
			return
		}
		if n == 2 {
			fmt.Fprint(w, `{"status":"waiting"}`)
			return
		}
		fmt.Fprint(w, `{"status":"received","last_code":"321"}`)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL), WithPollInterval(time.Millisecond))
	code, e := c.WaitForRentalCode(nil, "r1")
	if e != nil || code != "321" || n != 3 {
		t.Fatal(code, e, n)
	}
}
