package timex

import "time"

const (
	LayoutDateTime = "2006-01-02 15:04:05"
	LayoutDate     = "2006-01-02"
	LayoutTime     = "15:04:05"
	LayoutRFC3339  = time.RFC3339Nano
)

// NowUTC returns the current UTC time.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// Format returns an empty string for zero time, otherwise formats with layout.
func Format(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

// Parse parses a time value with layout.
func Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// FormatDateTime formats with the standard date-time layout.
func FormatDateTime(t time.Time) string {
	return Format(t, LayoutDateTime)
}

// FormatDate formats with the standard date layout.
func FormatDate(t time.Time) string {
	return Format(t, LayoutDate)
}

// SinceSeconds returns whole seconds since t.
func SinceSeconds(t time.Time) int64 {
	return int64(time.Since(t).Seconds())
}
