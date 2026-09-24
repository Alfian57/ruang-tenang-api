package application

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func TestMidtransWebhookValidation(t *testing.T) {
	s := &Service{serverKey: "test-key"}
	payload := &dto.MidtransWebhookRequest{OrderID: "RT-1", StatusCode: "200", GrossAmount: "10000.00", TransactionStatus: "settlement", FraudStatus: "accept"}
	hash := sha512.Sum512([]byte(payload.OrderID + payload.StatusCode + payload.GrossAmount + s.serverKey))
	payload.SignatureKey = hex.EncodeToString(hash[:])
	if !s.verifyWebhookSignature(payload) {
		t.Fatal("valid signature rejected")
	}
	if !webhookAmountMatches(payload.GrossAmount, 10000) {
		t.Fatal("valid amount rejected")
	}
	if webhookAmountMatches(payload.GrossAmount, 9000) {
		t.Fatal("mismatched amount accepted")
	}
	if got := s.mapTransactionStatus("capture", "challenge"); got != model.PaymentStatusPending {
		t.Fatalf("challenge = %s", got)
	}
	if got := s.mapTransactionStatus("settlement", "deny"); got != model.PaymentStatusFailed {
		t.Fatalf("fraud deny = %s", got)
	}
	if got := s.mapTransactionStatus("settlement", "accept"); got != model.PaymentStatusPaid {
		t.Fatalf("settlement = %s", got)
	}
	if got := s.mapTransactionStatus("refund", ""); got != model.PaymentStatusRefunded {
		t.Fatalf("refund = %s", got)
	}
	if got := s.mapTransactionStatus("partial_refund", ""); got != model.PaymentStatusPaid {
		t.Fatalf("partial refund = %s", got)
	}
	payload.GrossAmount = "1.00"
	if s.verifyWebhookSignature(payload) {
		t.Fatal("tampered amount accepted")
	}
}

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
