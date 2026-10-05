package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRentalMessages(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/rentals/r-1/messages" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, `[{"id":"m1","sender":"WhatsApp","body":"code 821-193","code":"821-193","received_at":"2026-06-17T22:31:10"}]`)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.RentalMessages(nil, "r-1")
	if e != nil || len(v) != 1 || v[0].Code != "821-193" || v[0].Sender != "WhatsApp" || v[0].Body != "code 821-193" {
		t.Fatal(v, e)
	}
	if _, e := c.RentalMessages(nil, ".."); e == nil {
		t.Fatal("bad id")
	}
}
