// Package progress computes year-progress and renders it as a bar + post body.
package progress

import (
	"fmt"
	"strings"
	"time"
)

const (
	barSize      = 20
	darkBarChar  = "▓"
	lightBarChar = "░"
)

// Snapshot is a single point-in-time reading of year progress.
type Snapshot struct {
	Year       int
	Percentage int // floor'd whole-number percentage, 0-100
	DayOfYear  int
	DaysInYear int
	Bar        string
	Text       string
}

func Compute(now time.Time) Snapshot {
	year := now.Year()
	daysInYear := time.Date(year, time.December, 31, 0, 0, 0, 0, now.Location()).YearDay()
	dayOfYear := now.YearDay()

	pct := float64(dayOfYear) / float64(daysInYear) * 100
	whole := int(pct)

	bar := renderBar(pct)
	text := fmt.Sprintf("YearProgress - %s %d%%", bar, whole)

	return Snapshot{
		Year:       year,
		Percentage: whole,
		DayOfYear:  dayOfYear,
		DaysInYear: daysInYear,
		Bar:        bar,
		Text:       text,
	}
}

func renderBar(percentage float64) string {
	completed := int(float64(barSize) * (percentage / 100))
	if completed > barSize {
		completed = barSize
	}

	var b strings.Builder
	for i := 0; i < barSize; i++ {
		if i < completed {
			b.WriteString(darkBarChar)
		} else {
			b.WriteString(lightBarChar)
		}
	}
	return b.String()
}
