package fetchsms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.fetchsms.com/v1"

// MaxResponseBytes limits response bodies, including error bodies, to 4 MiB.
const MaxResponseBytes int64 = 4 << 20

var ErrResponseTooLarge = errors.New("fetchsms: response exceeds size limit")

// Client is safe for concurrent use. Configure it only through NewClient.
type Client struct {
	key, baseURL string
	http         *http.Client
	pollInterval time.Duration
}

// Option configures a client.
type Option func(*Client) error

// WithBaseURL overrides the API root, including /v1. HTTP is allowed for local tests.
func WithBaseURL(raw string) Option {
	return func(c *Client) error {
		u, err := url.Parse(raw)
		if err != nil || u == nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.ForceQuery {
			return errors.New("fetchsms: invalid base URL")
		}
		c.baseURL = strings.TrimRight(raw, "/")
		return nil
	}
}

// WithHTTPClient copies the client configuration; redirects are always rejected.
// Its Transport must not retry purchases or log credentials. Do not mutate it concurrently.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) error {
		if h == nil {
			return errors.New("fetchsms: nil HTTP client")
		}
		copy := *h
		c.http = &copy
		return nil
	}
}

// NewClient permits an empty key for public catalog calls.
func NewClient(key string, options ...Option) (*Client, error) {
	c := &Client{key: key, baseURL: DefaultBaseURL, pollInterval: 3 * time.Second, http: &http.Client{Timeout: 30 * time.Second}}
	for _, o := range options {
		if o == nil {
			return nil, errors.New("fetchsms: nil option")
		}
		if err := o(c); err != nil {
			return nil, err
		}
	}
	c.http.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return c, nil
}

// APIError preserves bounded server detail without including it in Error (it may contain secrets).
type APIError struct {
	StatusCode    int
	Detail        json.RawMessage
	Body          []byte
	RetryAfter    time.Duration
	RetryAfterRaw string
	Truncated     bool
}

func (e *APIError) Error() string { return fmt.Sprintf("fetchsms: HTTP %d", e.StatusCode) }
func retryAfter(raw string) time.Duration {
	if n, e := strconv.ParseInt(raw, 10, 64); e == nil && n >= 0 {
		if n > int64((1<<63-1)/time.Second) {
			return time.Duration(1<<63 - 1)
		}
		return time.Duration(n) * time.Second
	}
	if t, e := http.ParseTime(raw); e == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
func (c *Client) do(ctx context.Context, method, path string, q url.Values, body, out any, auth bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if auth && strings.TrimSpace(c.key) == "" {
		return errors.New("fetchsms: API key required")
	}
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("fetchsms: encode request: %w", err)
		}
	}
	endpoint := c.baseURL + path
	if len(q) > 0 {
		endpoint += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	// Prevent replay of POST bodies by automatic redirects or custom transports using GetBody.
	if method != http.MethodGet {
		req.GetBody = nil
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, readErr := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
	large := int64(len(b)) > MaxResponseBytes
	if large {
		b = b[:MaxResponseBytes]
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		e := &APIError{StatusCode: resp.StatusCode, Body: b, Truncated: large, RetryAfterRaw: resp.Header.Get("Retry-After")}
		e.RetryAfter = retryAfter(e.RetryAfterRaw)
		var v struct {
			Detail json.RawMessage `json:"detail"`
		}
		if json.Unmarshal(b, &v) == nil {
			e.Detail = v.Detail
		}
		return e
	}
	if readErr != nil {
		return fmt.Errorf("fetchsms: read response: %w", readErr)
	}
	if large {
		return ErrResponseTooLarge
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("fetchsms: decode response: %w", err)
	}
	return nil
}

type Service struct {
	ID             int              `json:"id"`
	Slug           string           `json:"slug"`
	Name           string           `json:"name"`
	PriceCents     int64            `json:"price_cents"`
	LongPrices     map[string]int64 `json:"long_prices"`
	ShortAvailable int              `json:"short_available"`
	LongAvailable  int              `json:"long_available"`
}

// Services returns the public catalog without sending an API key.
func (c *Client) Services(ctx context.Context) ([]Service, error) {
	var v []Service
	err := c.do(ctx, http.MethodGet, "/services", nil, nil, &v, false)
	return v, err
}
