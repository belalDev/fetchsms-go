package main

import (
	fetchsms "github.com/belaldev/fetchsms-go"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Authentication-only demonstration. Add durable deduplication and processing
// before using this handler for business effects. Terminate TLS at a trusted proxy.
func handler(secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", 405)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid or oversized body", 413)
			return
		}
		if !fetchsms.VerifyWebhookSignature(raw, r.Header.Get("X-FetchSMS-Signature"), secret) {
			http.Error(w, "invalid signature", 401)
			return
		}
		event, err := fetchsms.DecodeWebhook(raw)
		if err != nil {
			http.Error(w, "invalid event", 400)
			return
		}
		if header := r.Header.Get("X-FetchSMS-Event"); header != "" && header != event.Event {
			http.Error(w, "event mismatch", 400)
			return
		}
		// Use the authenticated body event, not an unauthenticated header, for dispatch.
		// Persist/deduplicate/enqueue atomically BEFORE acknowledging in a real service.
		// Never log raw SMS bodies, codes, or the endpoint secret.
		_ = event
		w.WriteHeader(http.StatusNoContent)
	})
}
func main() {
	secret := os.Getenv("FETCHSMS_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("set FETCHSMS_WEBHOOK_SECRET")
	}
	mux := http.NewServeMux()
	mux.Handle("/webhook", handler(secret))
	s := &http.Server{Addr: "127.0.0.1:8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Fatal(s.ListenAndServe())
}
