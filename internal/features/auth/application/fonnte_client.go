package application

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WhatsAppSender interface {
	Send(ctx context.Context, number, message string) error
	IsConfigured() bool
}

type fonnteClient struct {
	token    string
	endpoint string
	client   *http.Client
}

func NewFonnteClient(token string) WhatsAppSender {
	return &fonnteClient{
		token:    strings.TrimSpace(token),
		endpoint: "https://api.fonnte.com/send",
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *fonnteClient) IsConfigured() bool { return c.token != "" }

func (c *fonnteClient) Send(ctx context.Context, number, message string) error {
	if !c.IsConfigured() {
		return errors.New("WhatsApp sender is not configured")
	}
	form := url.Values{"target": {number}, "message": {message}, "countryCode": {"62"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("WhatsApp provider rejected the request")
	}
	var result struct {
		Status bool `json:"status"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result); err != nil {
		return err
	}
	if !result.Status {
		return errors.New("WhatsApp provider did not accept the message")
	}
	return nil
}
