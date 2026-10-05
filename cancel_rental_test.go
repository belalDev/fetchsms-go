package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCancelRental(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/rentals/r-1/cancel" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, rentalJSON)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.CancelRental(nil, "r-1")
	if e != nil || v.ID != "r-1" {
		t.Fatal(v, e)
	}
	if _, e := c.CancelRental(nil, ".."); e == nil {
		t.Fatal("bad id")
	}
}
