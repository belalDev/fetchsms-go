package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBalance(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/wallet/balance" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, `{"balance_cents":12345}`)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.Balance(nil)
	if e != nil || v.BalanceCents != 12345 {
		t.Fatal(v, e)
	}
	c, _ = NewClient("")
	if _, e := c.Balance(nil); e == nil {
		t.Fatal("missing key")
	}
}
