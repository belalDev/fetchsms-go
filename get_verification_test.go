package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetVerification(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/verifications/v-1" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, verificationJSON)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.GetVerification(nil, "v-1")
	if e != nil || v.ID != "v-1" {
		t.Fatal(v, e)
	}
	for _, id := range []string{"", ".", "..", "a/b", "a?x", "a#x", "%2f", " x", "a\\b"} {
		if _, e := c.GetVerification(nil, id); e == nil {
			t.Fatal(id)
		}
	}
}
