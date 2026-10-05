package fetchsms

import (
	"testing"
)

func TestDecodeWebhook(t *testing.T) {
	for _, b := range []string{`{"event":"sms.received","data":{"verification_id":"v1","code":"0123","received_at":"2026-06-17T22:15:30"},"created_at":"2026-06-17T22:15:30"}`, `{"event":"sms.received","data":{"rental_id":"r1","body":"hello","code":"0123"}}`, `{"event":"rental.expired","data":{"rental_id":"r1","ref":"r-ref","number":"+1"}}`, `{"event":"future.event","data":{"unknown":[1,2,3]}}`} {
		e, err := DecodeWebhook([]byte(b))
		if err != nil || len(e.Data) == 0 {
			t.Fatal(e, err)
		}
		switch e.Event {
		case EventSMSReceived:
			if e.SMSReceived == nil || e.SMSReceived.Code != "0123" {
				t.Fatal(e)
			}
		case EventRentalExpired:
			if e.RentalExpired == nil || e.RentalExpired.RentalID != "r1" {
				t.Fatal(e)
			}
		default:
			if string(e.Data) != `{"unknown":[1,2,3]}` {
				t.Fatal(e)
			}
		}
	}
	for _, b := range []string{`bad`, `{}`, `{"event":"sms.received","data":1}`, `{"event":"rental.expired","data":null}`} {
		if _, e := DecodeWebhook([]byte(b)); e == nil {
			t.Fatal(b)
		}
	}
}
