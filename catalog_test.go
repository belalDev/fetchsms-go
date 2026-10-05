package fetchsms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQuote(t *testing.T) {
	ref, err := ServiceID(4)
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/quote" || r.URL.Query().Get("service") != "4" || r.URL.Query().Get("mode") != "short" || r.URL.Query().Get("custom_area_code") != "true" || r.Header.Get("Authorization") != "" {
			t.Error(r.URL)
		}
		fmt.Fprint(w, `{"service":"telegram","mode":"short","days":null,"base_daily_cents":70,"discount_pct":0,"per_day_cents":null,"surcharge_cents":10,"total_cents":80}`)
	}))
	defer s.Close()
	c, _ := NewClient("secret", WithBaseURL(s.URL))
	v, err := c.Quote(context.Background(), QuoteRequest{Service: ref, Mode: ModeShort, CustomAreaCode: true})
	if err != nil || v.TotalCents != 80 || v.Days != nil {
		t.Fatalf("%+v %v", v, err)
	}
	b, _ := json.Marshal(ref)
	if string(b) != "4" {
		t.Fatal(string(b))
	}
	slug, _ := ServiceSlug("telegram")
	b, _ = json.Marshal(slug)
	if string(b) != `"telegram"` {
		t.Fatal(string(b))
	}
	for _, id := range []int{0, -1} {
		if _, err := ServiceID(id); err == nil {
			t.Fatal("invalid id")
		}
	}
	for _, s := range []string{"", "4", "a/b", "a?b", " telegram"} {
		if _, err := ServiceSlug(s); err == nil {
			t.Fatal("invalid slug", s)
		}
	}
	if _, err := json.Marshal(ServiceRef{}); err == nil {
		t.Fatal("zero reference")
	}
	if _, err := c.Quote(nil, QuoteRequest{}); err == nil {
		t.Fatal("invalid quote")
	}
}
