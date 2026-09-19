package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func responseWithJSON(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestDeepSeekClientComplete(t *testing.T) {
	client := NewDeepSeekClient("https://deepseek.test", "test-key")
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s, want /chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("authorization = %q", got)
		}

		var req CompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "deepseek-flash" {
			t.Errorf("model = %q", req.Model)
		}
		if req.ReasoningEffort != "none" {
			t.Errorf("reasoning_effort = %q", req.ReasoningEffort)
		}
		if req.Stream {
			t.Error("stream must be disabled")
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "Halo" {
			t.Errorf("messages = %#v", req.Messages)
		}

		return responseWithJSON(`{"choices":[{"message":{"role":"assistant","content":"Halo juga"},"finish_reason":"stop"}]}`), nil
	})}

	response, err := client.Complete(context.Background(), CompletionRequest{
		Model:    "deepseek-flash",
		Messages: []Message{{Role: "user", Content: "Halo"}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if got := response.Choices[0].Message.Content; got != "Halo juga" {
		t.Errorf("content = %q", got)
	}
}

func TestDeepSeekClientParsesToolCallsAndJSONMode(t *testing.T) {
	client := NewDeepSeekClient("https://deepseek.test", "test-key")
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var req CompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.ResponseFormat == nil || req.ResponseFormat.Type != "json_object" {
			t.Fatalf("response format = %#v", req.ResponseFormat)
		}
		if len(req.Tools) != 1 || req.Tools[0].Function.Name != "search_articles" {
			t.Fatalf("tools = %#v", req.Tools)
		}
		return responseWithJSON(`{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"call-1","type":"function","function":{"name":"search_articles","arguments":"{\"query\":\"cemas\"}"}}]},"finish_reason":"tool_calls"}]}`), nil
	})}

	response, err := client.Complete(context.Background(), CompletionRequest{
		Model:          "deepseek-flash",
		Messages:       []Message{{Role: "user", Content: "Cari artikel"}},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
		Tools:          []Tool{{Type: "function", Function: FunctionDefinition{Name: "search_articles"}}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	call := response.Choices[0].Message.ToolCalls[0]
	if call.ID != "call-1" || call.Function.Name != "search_articles" {
		t.Fatalf("tool call = %#v", call)
	}
	if call.Function.Arguments != `{"query":"cemas"}` {
		t.Errorf("arguments = %q", call.Function.Arguments)
	}
}

func TestDeepSeekClientNotConfigured(t *testing.T) {
	client := NewDeepSeekClient("", "")
	if client.IsConfigured() {
		t.Fatal("client must not be configured without an API key")
	}
	_, err := client.Complete(context.Background(), CompletionRequest{Model: "deepseek-flash"})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("error = %v, want ErrNotConfigured", err)
	}
}
