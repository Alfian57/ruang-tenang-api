package application

import (
	"context"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/shared/ai"
	"github.com/Alfian57/ruang-tenang-api/prompts"
)

func (s *JournalService) generateSingleEntrySummary(ctx context.Context, content string) (string, error) {
	if s.aiClient == nil || !s.aiClient.IsConfigured() {
		if s.generateContentFn == nil {
			return "", nil
		}
	}

	prompt := prompts.Format("journal", "single_summary", s.truncateContent(ctx, content, 2000))

	var resp *ai.CompletionResponse
	var err error
	if s.generateContentFn != nil {
		resp, err = s.generateContentFn(ctx, prompt)
	} else {
		temperature := 0.5
		resp, err = s.aiClient.Complete(ctx, ai.CompletionRequest{
			Model:       s.aiModel,
			Messages:    []ai.Message{{Role: "user", Content: prompt}},
			Temperature: &temperature,
		})
	}
	if err != nil {
		return "", err
	}

	if resp != nil && len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}

	return "", nil
}
