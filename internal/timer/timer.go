package timer

import (
	"fmt"
	"time"
)

// Duration represents a selectable blocking duration
type Duration struct {
	Label    string
	Duration time.Duration
}

// PredefinedDurations returns the list of available durations
func PredefinedDurations() []Duration {
	return []Duration{
		{Label: "30 seconds", Duration: 30 * time.Second},
		{Label: "5 minutes", Duration: 5 * time.Minute},
		{Label: "15 minutes", Duration: 15 * time.Minute},
		{Label: "1 hour", Duration: 1 * time.Hour},
		{Label: "4 hours", Duration: 4 * time.Hour},
		{Label: "6 hours", Duration: 6 * time.Hour},
		{Label: "8 hours", Duration: 8 * time.Hour},
	}
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
