package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func (s *JournalService) toJournalResponse(_ context.Context, journal *model.Journal) *dto.JournalResponse {
	response := &dto.JournalResponse{
		ID:             journal.ID,
		UUID:           journal.UUID.String(),
		Title:          journal.Title,
		Content:        journal.Content,
		Summary:        journal.Summary,
		MoodID:         journal.MoodID,
		Tags:           []string(journal.Tags),
		IsPrivate:      journal.IsPrivate,
		ShareWithAI:    journal.ShareWithAI,
		AIAccessedAt:   journal.AIAccessedAt,
		WordCount:      journal.WordCount,
		SentimentScore: journal.SentimentScore,
		CreatedAt:      journal.CreatedAt,
		UpdatedAt:      journal.UpdatedAt,
	}

	if journal.Mood != nil {
		response.MoodLabel = string(journal.Mood.Mood)
		response.MoodEmoji = journal.Mood.GetMoodEmoji()
	}

	if response.Tags == nil {
		response.Tags = []string{}
	}

	return response
}

func (s *JournalService) toJournalListResponse(_ context.Context, journal *model.Journal) dto.JournalListResponse {
	preview := journal.Content
	if len(preview) > 150 {
		preview = preview[:150] + "..."
	}

	response := dto.JournalListResponse{
		ID:           journal.ID,
		UUID:         journal.UUID.String(),
		Title:        journal.Title,
		Preview:      preview,
		MoodID:       journal.MoodID,
		Tags:         []string(journal.Tags),
		ShareWithAI:  journal.ShareWithAI,
		AIAccessedAt: journal.AIAccessedAt,
		WordCount:    journal.WordCount,
		CreatedAt:    journal.CreatedAt,
	}

	if journal.Mood != nil {
		response.MoodLabel = string(journal.Mood.Mood)
		response.MoodEmoji = journal.Mood.GetMoodEmoji()
	}

	if response.Tags == nil {
		response.Tags = []string{}
	}

	return response
}
