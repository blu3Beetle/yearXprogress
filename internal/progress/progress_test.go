package progress

import (
	"strings"
	"testing"
	"time"
)

func TestComputeKnownDates(t *testing.T) {
	cases := []struct {
		name      string
		date      time.Time
		wantPct   int
		wantBarOn int // number of dark chars expected
	}{
		{"new year start", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 0, 0},
		{"halfway non-leap", time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC), 50, 10},
		{"leap day exists", time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 100, 20},
		{"year end", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), 100, 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := Compute(tc.date)
			if s.Percentage != tc.wantPct {
				t.Errorf("Percentage = %d, want %d", s.Percentage, tc.wantPct)
			}
			dark := strings.Count(s.Bar, darkBarChar)
			if dark != tc.wantBarOn {
				t.Errorf("dark bar chars = %d, want %d", dark, tc.wantBarOn)
			}
			if len([]rune(s.Bar)) != barSize {
				t.Errorf("bar length = %d, want %d", len([]rune(s.Bar)), barSize)
			}
		})
	}
}
