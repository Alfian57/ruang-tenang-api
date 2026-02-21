package service

import (
	"strings"
	"testing"
	"time"
)

func TestCommunityProgressService_HelperDatesAndLabels(t *testing.T) {
	loc := time.UTC
	monday := time.Date(2026, 2, 16, 15, 0, 0, 0, loc)
	sunday := time.Date(2026, 2, 22, 9, 0, 0, 0, loc)

	startMon := getStartOfWeek(monday)
	startSun := getStartOfWeek(sunday)
	if startMon.Day() != 16 || startMon.Weekday() != time.Monday {
		t.Fatalf("unexpected monday start: %v", startMon)
	}
	if startSun.Day() != 16 || startSun.Weekday() != time.Monday {
		t.Fatalf("unexpected sunday start: %v", startSun)
	}

	month := getStartOfMonth(time.Date(2026, 2, 20, 7, 0, 0, 0, loc))
	if month.Day() != 1 || month.Hour() != 0 {
		t.Fatalf("unexpected month start: %v", month)
	}

	if getActivityLabel("chat_ai") != "Chat dengan AI" {
		t.Fatal("expected mapped activity label")
	}
	if getActivityLabel("unknown_event") != "unknown_event" {
		t.Fatal("expected fallback activity label")
	}
}

func TestCommunityProgressService_GenerateLevelUpMessage(t *testing.T) {
	msg1 := generateLevelUpMessage(1, "Newcomer")
	msg10 := generateLevelUpMessage(10, "Guardian")
	msgOther := generateLevelUpMessage(99, "Other")

	if !strings.Contains(msg1, "Selamat datang") {
		t.Fatalf("unexpected level 1 message: %s", msg1)
	}
	if !strings.Contains(msg10, "Legendaris") {
		t.Fatalf("unexpected level 10 message: %s", msg10)
	}
	if !strings.Contains(msgOther, "Selamat naik level") {
		t.Fatalf("unexpected fallback message: %s", msgOther)
	}
}
