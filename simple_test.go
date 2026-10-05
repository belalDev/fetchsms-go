package fetchsms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type simpleTransport func(*http.Request) (*http.Response, error)

func (f simpleTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSimpleNumberDeadline(t *testing.T) {
	c := New("key")
	c.client.http.Timeout = 0 // Observe the wrapper's deadline, not net/http's.
	c.client.http.Transport = simpleTransport(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) < 29*time.Second || time.Until(deadline) > 30*time.Second {
			t.Error("number context must have a 30s deadline")
		}
		return nil, context.DeadlineExceeded
	})
	_, err := c.GetNumber("telegram")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestSimpleCodeDeadline(t *testing.T) {
	c := New("key")
	c.client.http.Timeout = 0
	c.client.http.Transport = simpleTransport(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) < 15*time.Minute-time.Second || time.Until(deadline) > 15*time.Minute {
			t.Error("code context must have a 15m deadline")
		}
		return nil, context.DeadlineExceeded
	})
	_, err := c.GetCode("v-1")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestSimpleTimeouts(t *testing.T) {
	for _, method := range []string{"number", "code"} {
		t.Run(method, func(t *testing.T) {
			c := New("key")
			c.numberTimeout = 10 * time.Millisecond
			c.codeTimeout = 10 * time.Millisecond
			c.client.http.Timeout = 0
			var requests int
			c.client.http.Transport = simpleTransport(func(r *http.Request) (*http.Response, error) {
				requests++
				<-r.Context().Done()
				return nil, r.Context().Err()
			})
			var err error
			if method == "number" {
				_, err = c.GetNumber("telegram")
			} else {
				_, err = c.GetCode("v-1")
			}
			if !errors.Is(err, context.DeadlineExceeded) || requests != 1 {
				t.Fatalf("error=%v requests=%d", err, requests)
			}
		})
	}
}

// These compile-time assertions preserve both public calling conventions.
var _ func(string) *SimpleClient = New
var _ func(string, ...Option) (*Client, error) = NewClient
var _ interface {
	GetNumber(string) (Verification, error)
	GetCode(string) (string, error)
} = (*SimpleClient)(nil)
var _ interface {
	GetNumber(context.Context, string) (Verification, error)
	WaitForCode(context.Context, string) (string, error)
} = (*Client)(nil)

func TestSimpleErrors(t *testing.T) {
	for _, status := range []int{402, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var requests atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.WriteHeader(status)
				fmt.Fprint(w, `{"detail":"rejected"}`)
			}))
			defer s.Close()
			c := New("key")
			c.client.baseURL = s.URL
			_, err := c.GetNumber("telegram")
			var api *APIError
			if !errors.As(err, &api) || api.StatusCode != status || string(api.Detail) != `"rejected"` || requests.Load() != 1 {
				t.Fatalf("error=%v requests=%d", err, requests.Load())
			}
			if status != 429 {
				_, err = c.GetCode("v-1")
				if !errors.As(err, &api) || api.StatusCode != status || requests.Load() != 2 {
					t.Fatalf("code error=%v requests=%d", err, requests.Load())
				}
			}
		})
	}
}

func TestSimpleInvalidInputs(t *testing.T) {
	for _, key := range []string{"key", "", " "} {
		c := New(key)
		c.client.http.Transport = simpleTransport(func(*http.Request) (*http.Response, error) {
			t.Error("invalid input made a request")
			return nil, fmt.Errorf("unexpected request")
		})
		for _, service := range []string{"", "Telegram/../../wallet"} {
			if v, err := c.GetNumber(service); err == nil || v != (Verification{}) {
				t.Fatalf("service=%q v=%+v err=%v", service, v, err)
			}
		}
		for _, id := range []string{"", "../wallet", "v?x=1"} {
			if code, err := c.GetCode(id); err == nil || code != "" {
				t.Fatalf("id=%q code=%q err=%v", id, code, err)
			}
		}
		if key != "key" {
			if _, err := c.GetNumber("telegram"); err == nil {
				t.Fatal("blank key accepted")
			}
			if _, err := c.GetCode("v-1"); err == nil {
				t.Fatal("blank key accepted")
			}
		}
	}
}

func TestSimplePollingDeadline(t *testing.T) {
	c := New("key")
	c.codeTimeout = 20 * time.Millisecond
	var requests atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		fmt.Fprint(w, `{"status":"waiting"}`)
	}))
	defer s.Close()
	c.client.baseURL = s.URL
	_, err := c.GetCode("v-1")
	if !errors.Is(err, context.DeadlineExceeded) || requests.Load() != 1 {
		t.Fatalf("error=%v requests=%d", err, requests.Load())
	}
}

func TestSimpleGetCode(t *testing.T) {
	var requests atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := requests.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/verifications/v-1" {
			t.Errorf("request: %s %s", r.Method, r.URL)
		}
		if n == 1 {
			fmt.Fprint(w, `{"status":"waiting","code":null}`)
		} else {
			fmt.Fprint(w, `{"status":"received","code":"0123"}`)
		}
	}))
	defer s.Close()
	c := New("key")
	c.client.baseURL = s.URL
	c.client.pollInterval = time.Millisecond
	code, err := c.GetCode("v-1")
	if err != nil || code != "0123" || requests.Load() != 2 {
		t.Fatalf("code=%q error=%v requests=%d", code, err, requests.Load())
	}
}

func TestSimpleGetNumber(t *testing.T) {
	var requests atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodPost || r.URL.Path != "/verifications" || r.URL.RawQuery != "" {
			t.Errorf("request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Error("missing authentication")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if len(body) != 1 || body["service"] != "telegram" {
			t.Errorf("body: %#v", body)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, verificationJSON)
	}))
	defer s.Close()
	c := New("key")
	if c.client.baseURL != DefaultBaseURL || c.client.http.Timeout != 30*time.Second || c.client.pollInterval != 3*time.Second {
		t.Fatal("unexpected defaults")
	}
	c.client.baseURL = s.URL
	if requests.Load() != 0 {
		t.Fatal("constructor made request")
	}
	v, err := c.GetNumber("telegram")
	if err != nil || v.ID != "v-1" || v.Number != "+1 332 555 0184" || v.Service != "telegram" || v.CostCents != 70 || v.Status != StatusWaiting {
		t.Fatalf("verification=%+v error=%v", v, err)
	}
	if requests.Load() != 1 {
		t.Fatal("purchase retried")
	}
}
