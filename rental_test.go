package fetchsms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const rentalJSON = `{"id":"r-1","ref":"r-ref","service":"unlimited-services","service_name":"Unlimited Services","number":"+1 415 555 0142","duration_days":30,"extended_days":0,"custom_area_code":true,"last_code":null,"status":"waiting","cost_cents":2800,"refunded":false,"created_at":"2026-06-17T22:14:05","expires_at":"2026-07-17T22:14:05","cancel_deadline":"2026-06-17T22:29:05"}`

func TestCreateRental(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/rentals" || r.Header.Get("Authorization") != "Bearer key" {
			t.Error(r.URL)
		}
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		if b["service"] != float64(1) || b["days"] != float64(30) || b["area_code"] != "415" {
			t.Error(b)
		}
		w.WriteHeader(201)
		fmt.Fprint(w, rentalJSON)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	ref, _ := ServiceID(1)
	v, e := c.CreateRental(nil, CreateRentalRequest{Service: ref, Days: 30, AreaCode: "415"})
	if e != nil || v.ID != "r-1" || v.DurationDays != 30 || v.LastCode != nil || v.CostCents != 2800 || !v.CustomAreaCode || v.CancelDeadline.IsZero() {
		t.Fatal(v, e)
	}
	if _, e := c.CreateRental(nil, CreateRentalRequest{Service: ref, Days: 2}); e == nil {
		t.Fatal("bad days")
	}
}
