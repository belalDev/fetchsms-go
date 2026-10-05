package fetchsms

import (
	"bytes"
	"encoding/json"
	"errors"
)

const (
	EventSMSReceived   = "sms.received"
	EventRentalExpired = "rental.expired"
)

// WebhookEnvelope always preserves Data, including unknown future event payloads.
// Typed pointers are populated by DecodeWebhook, not by json.Unmarshal.
type WebhookEnvelope struct {
	Event         string              `json:"event"`
	Data          json.RawMessage     `json:"data"`
	CreatedAt     Timestamp           `json:"created_at"`
	SMSReceived   *SMSReceivedEvent   `json:"-"`
	RentalExpired *RentalExpiredEvent `json:"-"`
}
type SMSReceivedEvent struct {
	VerificationID  string    `json:"verification_id"`
	RentalID        string    `json:"rental_id"`
	Ref             string    `json:"ref"`
	Number          string    `json:"number"`
	Sender          string    `json:"sender"`
	Code            string    `json:"code"`
	DetectedService string    `json:"detected_service"`
	ReceivedAt      Timestamp `json:"received_at"`
	Body            *string   `json:"body"`
}
type RentalExpiredEvent struct {
	RentalID string `json:"rental_id"`
	Ref      string `json:"ref"`
	Number   string `json:"number"`
}

// DecodeWebhook does not authenticate; verify the raw bytes before calling it.
func DecodeWebhook(raw []byte) (WebhookEnvelope, error) {
	var v WebhookEnvelope
	if int64(len(raw)) > MaxResponseBytes {
		return v, ErrResponseTooLarge
	}
	if e := json.Unmarshal(raw, &v); e != nil {
		return v, e
	}
	if v.Event == "" || len(v.Data) == 0 || bytes.Equal(bytes.TrimSpace(v.Data), []byte("null")) {
		return v, errors.New("fetchsms: invalid webhook envelope")
	}
	switch v.Event {
	case EventSMSReceived:
		v.SMSReceived = &SMSReceivedEvent{}
		if e := json.Unmarshal(v.Data, v.SMSReceived); e != nil {
			return v, e
		}
	case EventRentalExpired:
		v.RentalExpired = &RentalExpiredEvent{}
		if e := json.Unmarshal(v.Data, v.RentalExpired); e != nil {
			return v, e
		}
	}
	return v, nil
}
