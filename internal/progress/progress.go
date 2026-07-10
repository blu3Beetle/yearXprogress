// Package progress computes how far the current year has progressed and
// renders that as a text progress bar and post body.
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
	Bar        string
	Text       string
}

// Compute returns the year-progress snapshot for the given instant.
func Compute(now time.Time) Snapshot {
	year := now.Year()
	daysInYear := time.Date(year, time.December, 31, 0, 0, 0, 0, now.Location()).YearDay()

	pct := float64(now.YearDay()) / float64(daysInYear) * 100
	whole := int(pct)

	bar := renderBar(pct)
	text := fmt.Sprintf("YearProgress - %s %d%%", bar, whole)

	return Snapshot{
		Year:       year,
		Percentage: whole,
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
