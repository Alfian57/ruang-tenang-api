package repository

import (
	"testing"
	"time"
)

func TestCommunityProgressRepository_DateHelpers(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)

	monday := time.Date(2026, 2, 16, 15, 4, 5, 0, loc) // Monday
	gotMonday := getStartOfWeek(monday)
	if gotMonday.Day() != 16 || gotMonday.Weekday() != time.Monday {
		t.Fatalf("expected Monday start on 2026-02-16, got %v", gotMonday)
	}

	sunday := time.Date(2026, 2, 22, 9, 0, 0, 0, loc) // Sunday branch
	gotSunday := getStartOfWeek(sunday)
	if gotSunday.Day() != 16 || gotSunday.Weekday() != time.Monday {
		t.Fatalf("expected Sunday to map to prior Monday 2026-02-16, got %v", gotSunday)
	}

	monthMid := time.Date(2026, 2, 21, 8, 30, 0, 0, loc)
	gotMonth := getStartOfMonth(monthMid)
	if gotMonth.Day() != 1 || gotMonth.Month() != time.February {
		t.Fatalf("expected month start on 2026-02-01, got %v", gotMonth)
	}
}
