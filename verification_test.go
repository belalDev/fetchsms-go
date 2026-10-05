package fetchsms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const verificationJSON = `{"id":"v-1","ref":"v-ref","service":"telegram","service_name":"Telegram","number":"+1 332 555 0184","code":null,"code_count":2,"status":"waiting","cost_cents":70,"created_at":"2026-06-17T22:14:05.481203","expires_at":"2026-06-17T22:29:05","reusable":true,"reuse_cost_cents":80,"reuse_note":null}`

func TestCreateVerification(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/verifications" || r.Header.Get("Authorization") != "Bearer key" || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("bad request %v", r)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["service"] != float64(4) || body["area_code"] != "332" {
			t.Error(body)
		}
		w.WriteHeader(201)
		fmt.Fprint(w, verificationJSON)
	}))
	defer s.Close()
	c, _ := NewClient("key", WithBaseURL(s.URL))
	ref, _ := ServiceID(4)
	v, err := c.CreateVerification(context.Background(), CreateVerificationRequest{Service: ref, AreaCode: "332"})
	if err != nil || v.ID != "v-1" || v.Code != nil || v.CodeCount != 2 || !v.Reusable || v.ReuseCostCents == nil || *v.ReuseCostCents != 80 || v.CreatedAt.Nanosecond() != 481203000 {
		t.Fatalf("%+v %v", v, err)
	}
	for _, r := range []CreateVerificationRequest{{}, {Service: ref, AreaCode: "12/"}} {
		if _, err := c.CreateVerification(nil, r); err == nil {
			t.Fatal("invalid request")
		}
	}
}
