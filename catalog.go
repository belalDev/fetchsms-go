package fetchsms

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

// ServiceRef is a validated numeric service ID or compatibility slug. Its zero value is invalid.
type ServiceRef struct {
	id   int
	slug string
}

func ServiceID(id int) (ServiceRef, error) {
	if id <= 0 {
		return ServiceRef{}, errors.New("fetchsms: service ID must be positive")
	}
	return ServiceRef{id: id}, nil
}

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

func ServiceSlug(slug string) (ServiceRef, error) {
	if !slugPattern.MatchString(slug) {
		return ServiceRef{}, errors.New("fetchsms: invalid service slug")
	}
	return ServiceRef{slug: slug}, nil
}
func (s ServiceRef) String() string {
	if s.id > 0 {
		return strconv.Itoa(s.id)
	}
	return s.slug
}
func (s ServiceRef) valid() bool { return s.id > 0 || slugPattern.MatchString(s.slug) }
func (s ServiceRef) MarshalJSON() ([]byte, error) {
	if !s.valid() {
		return nil, errors.New("fetchsms: invalid service reference")
	}
	if s.id > 0 {
		return json.Marshal(s.id)
	}
	return json.Marshal(s.slug)
}

type Mode string

const (
	ModeShort Mode = "short"
	ModeLong  Mode = "long"
)

func validDays(n int) bool { return n == 1 || n == 7 || n == 30 || n == 90 || n == 365 }

// AreaCodesRequest defaults to short mode; a zero Service omits that filter.
type AreaCodesRequest struct {
	Mode    Mode
	Service ServiceRef
}

func (c *Client) AreaCodes(ctx context.Context, r AreaCodesRequest) ([]string, error) {
	if r.Mode == "" {
		r.Mode = ModeShort
	}
	if r.Mode != ModeShort && r.Mode != ModeLong {
		return nil, errors.New("fetchsms: invalid mode")
	}
	q := url.Values{"mode": {string(r.Mode)}}
	if r.Service.valid() {
		q.Set("service", r.Service.String())
	}
	var v []string
	err := c.do(ctx, http.MethodGet, "/services/area-codes", q, nil, &v, false)
	return v, err
}

type QuoteRequest struct {
	Service        ServiceRef
	Mode           Mode
	Days           int
	CustomAreaCode bool
}
type Quote struct {
	Service        string `json:"service"`
	Mode           Mode   `json:"mode"`
	Days           *int   `json:"days"`
	BaseDailyCents int64  `json:"base_daily_cents"`
	DiscountPct    int    `json:"discount_pct"`
	PerDayCents    *int64 `json:"per_day_cents"`
	SurchargeCents int64  `json:"surcharge_cents"`
	TotalCents     int64  `json:"total_cents"`
}

func (c *Client) Quote(ctx context.Context, r QuoteRequest) (Quote, error) {
	var v Quote
	if !r.Service.valid() || (r.Mode != ModeShort && r.Mode != ModeLong) || (r.Mode == ModeLong && (!validDays(r.Days) || (r.Service.String() != "1" && r.Service.String() != "unlimited-services"))) || (r.Mode == ModeShort && r.Days != 0) {
		return v, errors.New("fetchsms: invalid quote parameters")
	}
	q := url.Values{"service": {r.Service.String()}, "mode": {string(r.Mode)}, "custom_area_code": {strconv.FormatBool(r.CustomAreaCode)}}
	if r.Days != 0 {
		q.Set("days", strconv.Itoa(r.Days))
	}
	err := c.do(ctx, http.MethodGet, "/services/quote", q, nil, &v, false)
	return v, err
}
