package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.deepseek.com"

var ErrNotConfigured = errors.New("AI service is not configured")

// Client is the provider-neutral interface used by application services.
// The current implementation speaks the DeepSeek OpenAI-compatible API.
type Client interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	IsConfigured() bool
}

type CompletionRequest struct {
	Model           string          `json:"model"`
	Messages        []Message       `json:"messages"`
	Temperature     *float64        `json:"temperature,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
	ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
	Tools           []Tool          `json:"tools,omitempty"`
	ToolChoice      string          `json:"tool_choice,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	Stream          bool            `json:"stream"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments string         `json:"arguments"`
	Args      map[string]any `json:"-"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type CompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type DeepSeekClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewDeepSeekClient(baseURL, apiKey string) *DeepSeekClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &DeepSeekClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *DeepSeekClient) IsConfigured() bool {
	return c != nil && c.apiKey != ""
}

func (c *DeepSeekClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, errors.New("AI model is not configured")
	}

	// The application does not use streaming. Explicitly disable it so the
	// response remains a normal JSON completion even if provider defaults change.
	req.Stream = false
	if req.ReasoningEffort == "" {
		// Keep responses focused and avoid exposing chain-of-thought content to
		// the application or persisting it in chat history.
		req.ReasoningEffort = "none"
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal AI request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create AI request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read AI response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Do not include the provider body: it can contain request details or
		// other sensitive data and is not useful to end users.
		return nil, fmt.Errorf("AI request returned status %d", resp.StatusCode)
	}

	var result CompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode AI response: %w", err)
	}
	return &result, nil
}
