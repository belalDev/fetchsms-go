package fetchsms

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifyWebhookSignature checks X-FetchSMS-Signature against the untouched request body.
// This authenticates content, not freshness; callers must implement replay protection.
func VerifyWebhookSignature(rawBody []byte, signature, secret string) bool {
	if secret == "" {
		return false
	}
	given, e := hex.DecodeString(signature)
	if e != nil || len(given) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(rawBody)
	return hmac.Equal(given, mac.Sum(nil))
}
