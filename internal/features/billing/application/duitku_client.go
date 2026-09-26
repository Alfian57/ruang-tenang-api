package application

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DuitkuItemDetail struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type DuitkuCustomerDetail struct {
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type DuitkuInvoiceRequest struct {
	PaymentAmount    int                  `json:"paymentAmount"`
	MerchantOrderID  string               `json:"merchantOrderId"`
	ProductDetails   string               `json:"productDetails"`
	Email            string               `json:"email"`
	MerchantUserInfo string               `json:"merchantUserInfo,omitempty"`
	CustomerVaName   string               `json:"customerVaName,omitempty"`
	ItemDetails      []DuitkuItemDetail   `json:"itemDetails,omitempty"`
	CustomerDetail   DuitkuCustomerDetail `json:"customerDetail,omitempty"`
	CallbackURL      string               `json:"callbackUrl"`
	ReturnURL        string               `json:"returnUrl"`
}

type DuitkuInvoiceResponse struct {
	MerchantCode  string `json:"merchantCode"`
	Reference     string `json:"reference"`
	PaymentURL    string `json:"paymentUrl"`
	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
}

type DuitkuClient interface {
	CreateInvoice(ctx context.Context, req DuitkuInvoiceRequest) (*DuitkuInvoiceResponse, error)
	IsConfigured() bool
}

type duitkuClient struct {
	baseURL      string
	merchantCode string
	apiKey       string
	httpClient   *http.Client
}

func NewDuitkuClient(baseURL, merchantCode, apiKey string) DuitkuClient {
	return &duitkuClient{
		baseURL:      strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		merchantCode: strings.TrimSpace(merchantCode),
		apiKey:       strings.TrimSpace(apiKey),
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *duitkuClient) IsConfigured() bool {
	if c == nil || c.baseURL == "" || c.merchantCode == "" || c.apiKey == "" {
		return false
	}
	parsed, err := url.ParseRequestURI(c.baseURL)
	return err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host != ""
}

func (c *duitkuClient) CreateInvoice(ctx context.Context, req DuitkuInvoiceRequest) (*DuitkuInvoiceResponse, error) {
	if !c.IsConfigured() {
		return nil, ErrDuitkuNotConfigured
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode Duitku invoice request: %w", err)
	}
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	mac := hmac.New(sha256.New, []byte(c.apiKey))
	_, _ = mac.Write([]byte(c.merchantCode + timestamp))
	signature := hex.EncodeToString(mac.Sum(nil))

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/merchant/createInvoice", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Duitku invoice request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("x-duitku-signature", signature)
	httpRequest.Header.Set("x-duitku-timestamp", timestamp)
	httpRequest.Header.Set("x-duitku-merchantcode", c.merchantCode)

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("Duitku invoice request failed: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Duitku invoice response: %w", err)
	}
	var result DuitkuInvoiceResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("Duitku invoice request failed with HTTP %d", response.StatusCode)
		}
		return nil, fmt.Errorf("decode Duitku invoice response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(result.StatusMessage)
		if message == "" {
			message = "request rejected"
		}
		return nil, fmt.Errorf("Duitku invoice request failed with HTTP %d: %s", response.StatusCode, message)
	}
	if result.StatusCode != "00" || strings.TrimSpace(result.Reference) == "" || strings.TrimSpace(result.PaymentURL) == "" {
		message := strings.TrimSpace(result.StatusMessage)
		if message == "" {
			message = "invalid invoice response"
		}
		return nil, fmt.Errorf("Duitku invoice request failed: %s", message)
	}
	return &result, nil
}
