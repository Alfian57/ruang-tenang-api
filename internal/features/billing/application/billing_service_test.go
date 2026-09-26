package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func TestDuitkuWebhookValidation(t *testing.T) {
	s := &Service{apiKey: "test-key", merchantCode: "DTEST"}
	payload := &dto.DuitkuWebhookRequest{
		MerchantCode: "DTEST", Amount: "10000", MerchantOrderID: "RT-1", PaymentCode: "SP",
		ResultCode: "00", Reference: "DTEST-REF",
	}
	mac := hmac.New(sha256.New, []byte(s.apiKey))
	_, _ = mac.Write([]byte(payload.MerchantCode + payload.Amount + payload.MerchantOrderID))
	payload.Signature = hex.EncodeToString(mac.Sum(nil))
	if !s.verifyWebhookSignature(payload) {
		t.Fatal("valid signature rejected")
	}
	if !webhookAmountMatches(payload.Amount, 10000) {
		t.Fatal("valid amount rejected")
	}
	if webhookAmountMatches(payload.Amount, 9000) {
		t.Fatal("mismatched amount accepted")
	}
	if webhookAmountMatches("10000.00", 10000) {
		t.Fatal("fractional amount accepted")
	}
	if got, ok := mapDuitkuResultCode("00"); !ok || got != model.PaymentStatusPaid {
		t.Fatalf("success result = %s, valid=%t", got, ok)
	}
	if got, ok := mapDuitkuResultCode("01"); !ok || got != model.PaymentStatusFailed {
		t.Fatalf("failure result = %s, valid=%t", got, ok)
	}
	if _, ok := mapDuitkuResultCode("99"); ok {
		t.Fatal("unknown result code accepted")
	}
	if s.verifyWebhookSignature(&dto.DuitkuWebhookRequest{MerchantCode: "OTHER", Amount: payload.Amount, MerchantOrderID: payload.MerchantOrderID, Signature: payload.Signature}) {
		t.Fatal("wrong merchant accepted")
	}
	payload.Amount = "10001"
	if s.verifyWebhookSignature(payload) {
		t.Fatal("tampered amount accepted")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDuitkuCreateInvoice(t *testing.T) {
	client := &duitkuClient{
		baseURL:      "https://duitku.test",
		merchantCode: "DTEST",
		apiKey:       "test-key",
		httpClient: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost || r.URL.String() != "https://duitku.test/api/merchant/createInvoice" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			}
			if got := r.Header.Get("x-duitku-merchantcode"); got != "DTEST" {
				t.Errorf("merchant header = %q", got)
			}
			timestamp := r.Header.Get("x-duitku-timestamp")
			if timestamp == "" {
				t.Error("timestamp header is missing")
			}
			mac := hmac.New(sha256.New, []byte("test-key"))
			_, _ = mac.Write([]byte("DTEST" + timestamp))
			if got, want := r.Header.Get("x-duitku-signature"), hex.EncodeToString(mac.Sum(nil)); got != want {
				t.Errorf("request signature = %q, want %q", got, want)
			}
			var request DuitkuInvoiceRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode request: %v", err)
			}
			if request.MerchantOrderID != "RT-1" || request.PaymentAmount != 10000 || request.CallbackURL == "" || request.ReturnURL == "" {
				t.Errorf("unexpected invoice request: %+v", request)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"merchantCode":"DTEST","reference":"DTEST-REF","paymentUrl":"https://app-sandbox.duitku.com/redirect_checkout?reference=DTEST-REF","statusCode":"00","statusMessage":"SUCCESS"}`)),
				Request:    r,
			}, nil
		})},
	}
	result, err := client.CreateInvoice(t.Context(), DuitkuInvoiceRequest{
		PaymentAmount: 10000, MerchantOrderID: "RT-1", ProductDetails: "Premium", Email: "member@example.com",
		CallbackURL: "https://api.example.com/api/v1/billing/webhooks/duitku", ReturnURL: "https://app.example.com/payment/success",
	})
	if err != nil {
		t.Fatalf("CreateInvoice returned error: %v", err)
	}
	if result.Reference != "DTEST-REF" || !strings.Contains(result.PaymentURL, "redirect_checkout") {
		t.Fatalf("unexpected invoice response: %+v", result)
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
