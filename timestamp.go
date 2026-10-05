package fetchsms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp interprets offset-free API timestamps as UTC. Null becomes the zero value.
type Timestamp struct{ time.Time }

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		t.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		v, err = time.ParseInLocation("2006-01-02T15:04:05", s, time.UTC)
	}
	if err != nil {
		return fmt.Errorf("fetchsms: invalid timestamp: %w", err)
	}
	t.Time = v.UTC()
	return nil
}
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return t.UTC().MarshalJSON()
}
