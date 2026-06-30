package application

import (
	"context"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/prompts"
	"github.com/google/generative-ai-go/genai"
)

func (s *JournalService) generateSingleEntrySummary(ctx context.Context, content string) (string, error) {
	if s.genaiClient == nil {
		if s.generateContentFn == nil {
			return "", nil
		}
	}

	prompt := prompts.Format("journal", "single_summary", s.truncateContent(ctx, content, 2000))

	var (
		resp *genai.GenerateContentResponse
		err  error
	)
	if s.generateContentFn != nil {
		resp, err = s.generateContentFn(ctx, prompt)
	} else {
		model := s.genaiClient.GenerativeModel(s.aiModel)
		model.SetTemperature(0.5)
		resp, err = model.GenerateContent(ctx, genai.Text(prompt))
	}
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return strings.TrimSpace(string(text)), nil
		}
	}

	return "", nil
}
