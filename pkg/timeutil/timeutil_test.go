package timeutil

import (
	"sync"
	"testing"
	"time"
)

func resetTimezoneState() {
	appLocation = nil
	once = sync.Once{}
}

func TestLoadTimezoneFallbackAndHelpers(t *testing.T) {
	resetTimezoneState()
	t.Setenv("APP_TIMEZONE", "Invalid/Timezone")
	LoadTimezone()

	loc := GetLocation()
	if loc == nil {
		t.Fatal("location should not be nil")
	}
	if loc.String() != "Asia/Jakarta" {
		t.Fatalf("expected Asia/Jakarta fallback, got %s", loc.String())
	}

	base := time.Date(2026, 2, 20, 14, 30, 1, 123, time.UTC)
	start := StartOfDay(base)
	end := EndOfDay(base)

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Fatalf("unexpected start of day: %v", start)
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Fatalf("unexpected end of day: %v", end)
	}

	if !IsSameDay(start, end) {
		t.Fatal("start and end should be in same day")
	}

	formatted := Format(base, "2006-01-02")
	if formatted == "" {
		t.Fatal("formatted time should not be empty")
	}

	parsed, err := ParseInLocation("2006-01-02", "2026-02-20")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed.Location().String() != loc.String() {
		t.Fatalf("expected parsed location %s, got %s", loc.String(), parsed.Location().String())
	}
}

func TestNowAndTodayUseConfiguredLocation(t *testing.T) {
	resetTimezoneState()
	t.Setenv("APP_TIMEZONE", "Asia/Jakarta")
	LoadTimezone()

	now := Now()
	if now.Location().String() != "Asia/Jakarta" {
		t.Fatalf("expected Asia/Jakarta location for Now, got %s", now.Location().String())
	}

	today := Today()
	if today.Location().String() != "Asia/Jakarta" {
		t.Fatalf("expected Asia/Jakarta location for Today, got %s", today.Location().String())
	}
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Nanosecond() != 0 {
		t.Fatalf("expected Today at midnight, got %v", today)
	}
	if !IsSameDay(now, today) {
		t.Fatal("Now and Today should be on the same day")
	}
}

func TestGetLocationLazyLoadDefault(t *testing.T) {
	resetTimezoneState()
	t.Setenv("APP_TIMEZONE", "")

	loc := GetLocation()
	if loc == nil {
		t.Fatal("location should not be nil")
	}
	if loc.String() != "Asia/Jakarta" {
		t.Fatalf("expected default Asia/Jakarta, got %s", loc.String())
	}
}
