package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"
	"gorm.io/gorm"
)

// GetAnalytics gets journal analytics for a user
func (s *JournalService) GetAnalytics(ctx context.Context, userID uint) (*dto.JournalAnalytics, error) {
	totalEntries, err := s.journalRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	totalWordCount, err := s.journalRepo.GetTotalWordCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	moodDistribution, err := s.journalRepo.GetMoodDistribution(ctx, userID)
	if err != nil {
		return nil, err
	}
	tagFrequency, err := s.journalRepo.GetTagFrequency(ctx, userID)
	if err != nil {
		return nil, err
	}
	entriesByMonth, err := s.journalRepo.GetEntriesByMonth(ctx, userID, 12)
	if err != nil {
		return nil, err
	}
	writingStreak, err := s.journalRepo.GetWritingStreak(ctx, userID)
	if err != nil {
		return nil, err
	}

	avgWordCount := 0
	if totalEntries > 0 {
		avgWordCount = totalWordCount / int(totalEntries)
	}

	entriesThisMonth := 0
	currentMonth := time.Now().Format("2006-01")
	for _, entry := range entriesByMonth {
		if entry.Month == currentMonth {
			entriesThisMonth = entry.Count
			break
		}
	}

	monthlyEntries := make([]dto.MonthlyEntryCount, len(entriesByMonth))
	for i, entry := range entriesByMonth {
		monthlyEntries[i] = dto.MonthlyEntryCount{Month: entry.Month, Count: entry.Count}
	}

	return &dto.JournalAnalytics{
		TotalEntries:     int(totalEntries),
		EntriesThisMonth: entriesThisMonth,
		TotalWordCount:   totalWordCount,
		AvgWordCount:     avgWordCount,
		MoodDistribution: moodDistribution,
		TagFrequency:     tagFrequency,
		EntriesByMonth:   monthlyEntries,
		WritingStreak:    writingStreak,
		LongestStreak:    writingStreak,
	}, nil
}

// GetWritingPrompt generates an AI writing prompt
func (s *JournalService) GetWritingPrompt(ctx context.Context, userID uint) (*dto.JournalPromptResponse, error) {
	latestMood, err := s.userMoodRepo.GetLatestByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	tagFrequency, err := s.journalRepo.GetTagFrequency(ctx, userID)
	if err != nil {
		return nil, err
	}

	moodContext := ""
	if latestMood != nil {
		moodContext = fmt.Sprintf("User's current mood is: %s (%s)", latestMood.Mood, latestMood.GetMoodEmoji())
	}

	topTags := make([]string, 0, 3)
	for tag := range tagFrequency {
		topTags = append(topTags, tag)
		if len(topTags) >= 3 {
			break
		}
	}

	return s.generateWritingPrompt(ctx, moodContext, topTags), nil
}

// GetWeeklySummary generates a weekly summary of journals
func (s *JournalService) GetWeeklySummary(ctx context.Context, userID uint) (*dto.JournalWeeklySummary, error) {
	weekStart := time.Now().AddDate(0, 0, -7)
	weekEnd := time.Now()

	journals, _, err := s.journalRepo.FindByUserID(ctx, userID, 1, 100, nil, nil, &weekStart, &weekEnd)
	if err != nil {
		return nil, err
	}

	if len(journals) == 0 {
		return &dto.JournalWeeklySummary{
			WeekStart:    weekStart,
			WeekEnd:      weekEnd,
			EntriesCount: 0,
			Summary:      "Tidak ada entri jurnal minggu ini.",
		}, nil
	}

	summary, themes, insights, suggestions, moodTrend := s.generateWeeklySummary(ctx, journals)
	return &dto.JournalWeeklySummary{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		EntriesCount: len(journals),
		Summary:      summary,
		KeyThemes:    themes,
		MoodTrend:    moodTrend,
		Insights:     insights,
		Suggestions:  suggestions,
	}, nil
}

func (s *JournalService) generateWritingPrompt(_ context.Context, _ string, topTags []string) *dto.JournalPromptResponse {
	categories := []string{"reflection", "gratitude", "goal", "emotion"}
	prompts := map[string][]string{
		"reflection": {
			"Apa satu hal yang kamu pelajari tentang dirimu hari ini?",
			"Bagaimana perasaanmu saat ini dan mengapa?",
			"Apa yang membuatmu bersyukur minggu ini?",
		},
		"gratitude": {
			"Sebutkan 3 hal kecil yang membuatmu tersenyum hari ini",
			"Siapa yang ingin kamu ucapkan terima kasih, dan mengapa?",
			"Apa momen sederhana yang membuat harimu lebih baik?",
		},
		"goal": {
			"Apa satu langkah kecil yang bisa kamu ambil besok untuk mencapai tujuanmu?",
			"Bagaimana kamu ingin perasaanmu di akhir minggu ini?",
			"Apa yang ingin kamu lepaskan untuk bisa maju?",
		},
		"emotion": {
			"Ceritakan tentang emosi yang paling kuat kamu rasakan hari ini",
			"Apa yang bisa kamu katakan pada dirimu sendiri untuk merasa lebih baik?",
			"Bagaimana tubuhmu merespons stres, dan apa yang bisa membantu?",
		},
	}

	category := categories[time.Now().Unix()%int64(len(categories))]
	promptList := prompts[category]
	prompt := promptList[time.Now().UnixNano()%int64(len(promptList))]

	return &dto.JournalPromptResponse{Prompt: prompt, Category: category, RelatedTags: topTags}
}

func (s *JournalService) generateWeeklySummary(ctx context.Context, journals []model.Journal) (string, []string, []string, []string, string) {
	if (s.genaiClient == nil && s.generateContentFn == nil) || len(journals) == 0 {
		return "Tidak cukup data untuk membuat ringkasan.", []string{}, []string{}, []string{}, "stable"
	}

	var contentBuilder strings.Builder
	for _, journal := range journals {
		mood := ""
		if journal.Mood != nil {
			mood = string(journal.Mood.Mood)
		}
		contentBuilder.WriteString(fmt.Sprintf(
			"[%s] Mood: %s | %s\n",
			journal.CreatedAt.Format("2006-01-02"),
			mood,
			s.truncateContent(ctx, journal.Content, 200),
		))
	}

	prompt := fmt.Sprintf(`Kamu adalah asisten kesehatan mental. Analisis entri jurnal minggu ini dan berikan:
1. Ringkasan singkat (2-3 kalimat)
2. 3 tema utama (pisahkan dengan koma)
3. 2 insight positif (pisahkan dengan |)
4. 2 saran untuk minggu depan (pisahkan dengan |)
5. Tren mood: "improving", "stable", atau "declining"

Format jawaban:
SUMMARY: [ringkasan]
THEMES: [tema1, tema2, tema3]
INSIGHTS: [insight1 | insight2]
SUGGESTIONS: [saran1 | saran2]
MOOD_TREND: [tren]

Entri Jurnal:
%s`, contentBuilder.String())

	var (
		resp *genai.GenerateContentResponse
		err  error
	)
	if s.generateContentFn != nil {
		resp, err = s.generateContentFn(context.Background(), prompt)
	} else {
		model := s.genaiClient.GenerativeModel("gemini-2.0-flash")
		model.SetTemperature(0.7)
		resp, err = model.GenerateContent(context.Background(), genai.Text(prompt))
	}
	if err != nil {
		return "Gagal membuat ringkasan.", []string{}, []string{}, []string{}, "stable"
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return s.parseWeeklySummaryResponse(ctx, string(text))
		}
	}

	return "Gagal membuat ringkasan.", []string{}, []string{}, []string{}, "stable"
}

func (s *JournalService) parseWeeklySummaryResponse(_ context.Context, response string) (string, []string, []string, []string, string) {
	summary := ""
	themes := []string{}
	insights := []string{}
	suggestions := []string{}
	moodTrend := "stable"

	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "SUMMARY:"):
			summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		case strings.HasPrefix(line, "THEMES:"):
			themesStr := strings.TrimSpace(strings.TrimPrefix(line, "THEMES:"))
			for _, theme := range strings.Split(themesStr, ",") {
				themes = append(themes, strings.TrimSpace(theme))
			}
		case strings.HasPrefix(line, "INSIGHTS:"):
			insightsStr := strings.TrimSpace(strings.TrimPrefix(line, "INSIGHTS:"))
			for _, insight := range strings.Split(insightsStr, "|") {
				insights = append(insights, strings.TrimSpace(insight))
			}
		case strings.HasPrefix(line, "SUGGESTIONS:"):
			suggestionsStr := strings.TrimSpace(strings.TrimPrefix(line, "SUGGESTIONS:"))
			for _, suggestion := range strings.Split(suggestionsStr, "|") {
				suggestions = append(suggestions, strings.TrimSpace(suggestion))
			}
		case strings.HasPrefix(line, "MOOD_TREND:"):
			moodTrend = strings.TrimSpace(strings.TrimPrefix(line, "MOOD_TREND:"))
		}
	}

	return summary, themes, insights, suggestions, moodTrend
}
