package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListVerifications(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/verifications" || r.URL.Query().Get("tab") != "history" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, "["+verificationJSON+"]")
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.ListVerifications(nil, TabHistory)
	if e != nil || len(v) != 1 {
		t.Fatal(v, e)
	}
	if _, e := c.ListVerifications(nil, "bad"); e == nil {
		t.Fatal("bad tab")
	}
}
