package fetchsms

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	for _, s := range []string{`"2026-06-17T22:14:05"`, `"2026-06-17T22:14:05.481203"`, `"2026-06-17T22:14:05Z"`, `"2026-06-18T00:14:05+02:00"`} {
		var v Timestamp
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Fatal(err)
		}
		if v.Location() != time.UTC || v.Hour() != 22 {
			t.Fatal(v)
		}
		b, err := json.Marshal(v)
		if err != nil || len(b) == 0 {
			t.Fatal(err)
		}
	}
	var v Timestamp
	if err := json.Unmarshal([]byte(`null`), &v); err != nil || !v.IsZero() {
		t.Fatal(v, err)
	}
	for _, s := range []string{`4`, `"garbage"`, `""`} {
		if json.Unmarshal([]byte(s), &v) == nil {
			t.Fatal(s)
		}
	}
}
