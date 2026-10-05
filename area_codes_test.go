package fetchsms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAreaCodes(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/services/area-codes" || r.URL.Query().Get("mode") != "short" || r.URL.Query().Get("service") != "2" || r.Header.Get("Authorization") != "" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, `["212","415"]`)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	ref, _ := ServiceID(2)
	v, err := c.AreaCodes(nil, AreaCodesRequest{Mode: ModeShort, Service: ref})
	if err != nil || len(v) != 2 {
		t.Fatalf("%v %v", v, err)
	}
	if _, err := c.AreaCodes(nil, AreaCodesRequest{Mode: "bad"}); err == nil {
		t.Fatal("invalid mode")
	}
}
