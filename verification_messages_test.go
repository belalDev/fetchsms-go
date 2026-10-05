package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerificationMessages(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/verifications/v-1/messages" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, `[{"id":"m1","code":"0123","received_at":"2026-06-17T22:21:07"}]`)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.VerificationMessages(nil, "v-1")
	if e != nil || len(v) != 1 || v[0].Code != "0123" {
		t.Fatal(v, e)
	}
	if _, e := c.VerificationMessages(nil, ".."); e == nil {
		t.Fatal("bad id")
	}
}
