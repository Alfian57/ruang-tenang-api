package timeutil

import (
	"os"
	"sync"
	"time"
)

var (
	appLocation *time.Location
	once        sync.Once
)

// LoadTimezone loads the timezone from APP_TIMEZONE environment variable.
// Should be called once at application startup.
// Defaults to "Asia/Jakarta" if not set or invalid.
func LoadTimezone() {
	once.Do(func() {
		tz := os.Getenv("APP_TIMEZONE")
		if tz == "" {
			tz = "Asia/Jakarta"
		}
		
		loc, err := time.LoadLocation(tz)
		if err != nil {
			// Fallback to Asia/Jakarta if timezone is invalid
			loc, _ = time.LoadLocation("Asia/Jakarta")
		}
		appLocation = loc
	})
}

// GetLocation returns the application timezone location.
// If LoadTimezone hasn't been called, it will load with defaults.
func GetLocation() *time.Location {
	if appLocation == nil {
		LoadTimezone()
	}
	return appLocation
}

// Now returns the current time in the application timezone.
func Now() time.Time {
	return time.Now().In(GetLocation())
}

// Today returns the start of today in the application timezone.
func Today() time.Time {
	now := Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, GetLocation())
}

// StartOfDay returns the start of the day for the given time in the application timezone.
func StartOfDay(t time.Time) time.Time {
	t = t.In(GetLocation())
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, GetLocation())
}

// EndOfDay returns the end of the day (23:59:59.999999999) for the given time in the application timezone.
func EndOfDay(t time.Time) time.Time {
	t = t.In(GetLocation())
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, GetLocation())
}

// IsSameDay checks if two times are on the same calendar day in the application timezone.
func IsSameDay(t1, t2 time.Time) bool {
	t1 = t1.In(GetLocation())
	t2 = t2.In(GetLocation())
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

// Format formats the time in the application timezone using the given layout.
func Format(t time.Time, layout string) string {
	return t.In(GetLocation()).Format(layout)
}

// ParseInLocation parses a time string in the application timezone.
func ParseInLocation(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, GetLocation())
}
