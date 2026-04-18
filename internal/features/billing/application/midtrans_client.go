package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MidtransSnapRequest struct {
	TransactionDetails MidtransTransactionDetails `json:"transaction_details"`
	CustomerDetails    MidtransCustomerDetails    `json:"customer_details,omitempty"`
	ItemDetails        []MidtransItemDetails      `json:"item_details,omitempty"`
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

type midtransClient struct {
	baseURL    string
	serverKey  string
	httpClient *http.Client
}

type MidtransClient interface {
	CreateSnapTransaction(ctx context.Context, req MidtransSnapRequest) (*MidtransSnapResponse, error)
	IsConfigured() bool
}

func NewMidtransClient(baseURL string, serverKey string) MidtransClient {
	trimmedURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedURL == "" {
		trimmedURL = "https://app.sandbox.midtrans.com"
	}

	return &midtransClient{
		baseURL:   trimmedURL,
		serverKey: strings.TrimSpace(serverKey),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
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
