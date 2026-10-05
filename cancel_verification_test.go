package fetchsms

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCancelVerification(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/verifications/v-1/cancel" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		w.WriteHeader(204)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	if e := c.CancelVerification(nil, "v-1"); e != nil {
		t.Fatal(e)
	}
	if e := c.CancelVerification(nil, ".."); e == nil {
		t.Fatal("bad id")
	}
}
