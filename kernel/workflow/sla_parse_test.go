package workflow

import (
	"testing"
	"time"
)

// sla_parse_test.go — QA G3 (parsing / edge cases): parseISODuration converts the
// ISO-8601 SLA durations in a workflow definition (due/remind_after) into
// time.Duration. A wrong parse silently mis-schedules SLA reminders/escalations,
// so its valid and invalid inputs are pinned here (white-box: the function is
// unexported meaningful logic, not a brittle internal detail).

func TestParseISODurationValid(t *testing.T) {
	cases := map[string]time.Duration{
		"":          0,
		"P1D":       24 * time.Hour,
		"P1W":       7 * 24 * time.Hour,
		"PT2H":      2 * time.Hour,
		"PT30M":     30 * time.Minute,
		"PT45S":     45 * time.Second,
		"P1DT12H":   36 * time.Hour,
		"P2DT3H30M": 2*24*time.Hour + 3*time.Hour + 30*time.Minute,
		"PT1H30M":   90 * time.Minute,
	}
	for in, want := range cases {
		got, err := parseISODuration(in)
		if err != nil {
			t.Errorf("parseISODuration(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseISODuration(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseISODurationInvalid(t *testing.T) {
	for _, in := range []string{
		"1D",    // missing leading P
		"P1X",   // unknown unit
		"PD",    // unit without a number
		"PT",    // T with no time components → tolerated? assert no panic below
		"PTH",   // H without a number
		"P1DTX", // bad time unit
	} {
		got, err := parseISODuration(in)
		// "PT" (empty date + empty time) is a boundary; accept either a clean 0
		// or an error, but it must never panic or return a bogus non-zero value.
		if in == "PT" {
			if err == nil && got != 0 {
				t.Errorf("parseISODuration(%q) = %v with no error; want 0 or an error", in, got)
			}
			continue
		}
		if err == nil {
			t.Errorf("parseISODuration(%q) = %v; want an error", in, got)
		}
	}
}
