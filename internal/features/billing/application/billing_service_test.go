package application

import (
	"testing"
	"time"
)

func TestParseChatQuotaResetInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "empty uses default", value: "", want: 24 * time.Hour},
		{name: "minutes", value: "30m", want: 30 * time.Minute},
		{name: "hours", value: "6h", want: 6 * time.Hour},
		{name: "days shorthand", value: "2d", want: 48 * time.Hour},
		{name: "invalid uses default", value: "soon", want: 24 * time.Hour},
		{name: "too small clamps to minimum", value: "10s", want: time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseChatQuotaResetInterval(tt.value); got != tt.want {
				t.Fatalf("parseChatQuotaResetInterval(%q) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}

func TestChatQuotaWindowFor(t *testing.T) {
	loc := time.FixedZone("WIB", 7*60*60)

	t.Run("hourly window is anchored to local day", func(t *testing.T) {
		now := time.Date(2026, 5, 1, 10, 17, 30, 0, loc)
		window := chatQuotaWindowFor(now, 30*time.Minute)

		wantStart := time.Date(2026, 5, 1, 10, 0, 0, 0, loc)
		wantEnd := time.Date(2026, 5, 1, 10, 30, 0, 0, loc)
		if !window.Start.Equal(wantStart) || !window.End.Equal(wantEnd) {
			t.Fatalf("window = %s - %s, want %s - %s", window.Start, window.End, wantStart, wantEnd)
		}
	})

	t.Run("daily window resets at local midnight", func(t *testing.T) {
		now := time.Date(2026, 5, 1, 10, 17, 30, 0, loc)
		window := chatQuotaWindowFor(now, 24*time.Hour)

		wantStart := time.Date(2026, 5, 1, 0, 0, 0, 0, loc)
		wantEnd := time.Date(2026, 5, 2, 0, 0, 0, 0, loc)
		if !window.Start.Equal(wantStart) || !window.End.Equal(wantEnd) {
			t.Fatalf("window = %s - %s, want %s - %s", window.Start, window.End, wantStart, wantEnd)
		}
	})
}
