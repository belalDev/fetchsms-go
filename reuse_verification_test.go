package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReuseVerification(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/verifications/v-1/reuse" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		w.WriteHeader(201)
		fmt.Fprint(w, verificationJSON)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.ReuseVerification(nil, "v-1")
	if e != nil || v.ID != "v-1" {
		t.Fatal(v, e)
	}
	if _, e := c.ReuseVerification(nil, ".."); e == nil {
		t.Fatal("bad id")
	}
}
