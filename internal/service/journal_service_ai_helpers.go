package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

func (s *JournalService) generateSingleEntrySummary(ctx context.Context, content string) (string, error) {
	if s.genaiClient == nil {
		return "", nil
	}

	model := s.genaiClient.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.5)

	prompt := fmt.Sprintf(`Buatlah ringkasan singkat (1 kalimat) dari entri jurnal berikut. Fokus pada inti kejadian atau perasaan utama.
Jurnal: "%s"`, s.truncateContent(ctx, content, 2000))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
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
