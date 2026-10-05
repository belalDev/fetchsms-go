package fetchsms

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func signature(b []byte, key string) string {
	m := hmac.New(sha256.New, []byte(key))
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}
func TestWebhookSignature(t *testing.T) {
	b := []byte(`{"event":"sms.received"}`)
	sig := signature(b, "secret")
	if !VerifyWebhookSignature(b, sig, "secret") {
		t.Fatal("valid rejected")
	}
	for _, x := range []struct {
		b    []byte
		s, k string
	}{{append(b, ' '), sig, "secret"}, {b, sig, "wrong"}, {b, "xx", "secret"}, {b, sig, ""}, {b, "", "secret"}, {b, "sha256=" + sig, "secret"}} {
		if VerifyWebhookSignature(x.b, x.s, x.k) {
			t.Fatal("invalid accepted")
		}
	}
}
