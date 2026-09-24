package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
)

type MidtransSnapRequest struct {
	TransactionDetails MidtransTransactionDetails `json:"transaction_details"`
	CustomerDetails    MidtransCustomerDetails    `json:"customer_details,omitempty"`
	ItemDetails        []MidtransItemDetails      `json:"item_details,omitempty"`
	Callbacks          *MidtransCallbacks         `json:"callbacks,omitempty"`
}

type MidtransCallbacks struct {
	Finish string `json:"finish,omitempty"`
}

type MidtransTransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int    `json:"gross_amount"`
}

type MidtransCustomerDetails struct {
	FirstName string `json:"first_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

type MidtransItemDetails struct {
	ID       string `json:"id"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type MidtransRefundRequest struct {
	RefundKey string `json:"refund_key"`
	Amount    int64  `json:"amount"`
	Reason    string `json:"reason"`
}

type MidtransRefundResponse struct {
	StatusCode         string          `json:"status_code"`
	StatusMessage      string          `json:"status_message"`
	TransactionStatus  string          `json:"transaction_status"`
	TransactionID      string          `json:"transaction_id"`
	RefundChargebackID json.RawMessage `json:"refund_chargeback_id"`
	RefundAmount       string          `json:"refund_amount"`
	RefundKey          string          `json:"refund_key"`
}

type MidtransAPIError struct {
	StatusCode int
	Message    string
}

func (e *MidtransAPIError) Error() string {
	return fmt.Sprintf("midtrans request failed: status=%d message=%s", e.StatusCode, e.Message)
}

type midtransClient struct {
	baseURL    string
	apiBaseURL string
	serverKey  string
	httpClient *http.Client
}

type MidtransClient interface {
	CreateSnapTransaction(ctx context.Context, req MidtransSnapRequest) (*MidtransSnapResponse, error)
	RefundTransaction(ctx context.Context, orderID string, req MidtransRefundRequest) (*MidtransRefundResponse, error)
	GetTransactionStatus(ctx context.Context, orderID string) (*dto.MidtransWebhookRequest, error)
	IsConfigured() bool
}

func NewMidtransClient(baseURL string, serverKey string) MidtransClient {
	trimmedURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedURL == "" {
		trimmedURL = "https://app.sandbox.midtrans.com"
	}
	apiBaseURL := midtransAPIBaseURL(trimmedURL)

	return &midtransClient{
		baseURL:    trimmedURL,
		apiBaseURL: apiBaseURL,
		serverKey:  strings.TrimSpace(serverKey),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func midtransAPIBaseURL(snapBaseURL string) string {
	parsed, err := url.Parse(snapBaseURL)
	if err != nil {
		return snapBaseURL
	}
	if parsed.Host == "app.sandbox.midtrans.com" {
		parsed.Host = "api.sandbox.midtrans.com"
	} else if parsed.Host == "app.midtrans.com" {
		parsed.Host = "api.midtrans.com"
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return strings.TrimRight(parsed.String(), "/")
}

func (c *midtransClient) IsConfigured() bool {
	return c.serverKey != ""
}

func (c *midtransClient) CreateSnapTransaction(ctx context.Context, req MidtransSnapRequest) (*MidtransSnapResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/snap/v1/transactions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("midtrans create transaction failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var snapResp MidtransSnapResponse
	if err := json.Unmarshal(body, &snapResp); err != nil {
		return nil, err
	}

	if snapResp.Token == "" || snapResp.RedirectURL == "" {
		return nil, fmt.Errorf("midtrans response incomplete")
	}

	return &snapResp, nil
}

func (c *midtransClient) RefundTransaction(ctx context.Context, orderID string, req MidtransRefundRequest) (*MidtransRefundResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	endpoint := c.apiBaseURL + "/v2/" + url.PathEscape(orderID) + "/refund"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var refundResp MidtransRefundResponse
	var decodeErr error
	if len(body) > 0 {
		decodeErr = json.Unmarshal(body, &refundResp)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := refundResp.StatusMessage
		if message == "" && len(body) > 0 {
			message = string(body)
		}
		return &refundResp, &MidtransAPIError{StatusCode: resp.StatusCode, Message: message}
	}
	if decodeErr != nil {
		return nil, decodeErr
	}
	if refundResp.RefundKey == "" {
		refundResp.RefundKey = req.RefundKey
	}
	if refundResp.StatusCode != "200" || refundResp.RefundKey == "" {
		return &refundResp, fmt.Errorf("midtrans refund response incomplete")
	}
	return &refundResp, nil
}

func (c *midtransClient) GetTransactionStatus(ctx context.Context, orderID string) (*dto.MidtransWebhookRequest, error) {
	endpoint := c.apiBaseURL + "/v2/" + url.PathEscape(orderID) + "/status"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var status dto.MidtransWebhookRequest
	decodeErr := json.Unmarshal(body, &status)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := string(body)
		if decodeErr == nil && status.TransactionStatus != "" {
			message = status.TransactionStatus
		}
		return nil, &MidtransAPIError{StatusCode: resp.StatusCode, Message: message}
	}
	if decodeErr != nil {
		return nil, decodeErr
	}
	if status.OrderID == "" || status.TransactionID == "" || status.TransactionStatus == "" || status.SignatureKey == "" {
		return nil, fmt.Errorf("midtrans transaction status response incomplete")
	}
	return &status, nil
}
