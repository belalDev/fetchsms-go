package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRentals(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/rentals" || r.URL.Query().Get("tab") != "history" || r.URL.Query().Get("limit") != "200" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, "["+rentalJSON+"]")
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.ListRentals(nil, ListRentalsRequest{Tab: TabHistory, Limit: 200})
	if e != nil || len(v) != 1 {
		t.Fatal(v, e)
	}
	for _, r := range []ListRentalsRequest{{Tab: "bad"}, {Limit: -1}, {Limit: 201}} {
		if _, e := c.ListRentals(nil, r); e == nil {
			t.Fatal("bad params")
		}
	}
}
