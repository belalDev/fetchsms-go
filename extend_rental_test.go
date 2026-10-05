package fetchsms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtendRental(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/rentals/r-1/extend" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		var b map[string]int
		json.NewDecoder(r.Body).Decode(&b)
		if b["days"] != 30 {
			t.Error(b)
		}
		fmt.Fprint(w, rentalJSON)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	v, e := c.ExtendRental(nil, "r-1", 30)
	if e != nil || v.ID != "r-1" {
		t.Fatal(v, e)
	}
	if _, e := c.ExtendRental(nil, "..", 30); e == nil {
		t.Fatal("bad id")
	}
	if _, e := c.ExtendRental(nil, "r-1", 2); e == nil {
		t.Fatal("bad days")
	}
}
