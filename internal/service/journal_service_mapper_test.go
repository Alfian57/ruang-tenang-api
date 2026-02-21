package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func TestToJournalResponse(t *testing.T) {
	service := &JournalService{}
	now := time.Now()
	moodID := uint(2)

	entry := &model.Journal{
		ID:          1,
		UUID:        uuid.New(),
		Title:       "Judul",
		Content:     "Konten",
		Summary:     "Ringkasan",
		MoodID:      &moodID,
		Tags:        nil,
		IsPrivate:   true,
		ShareWithAI: true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Mood: &model.UserMood{
			Mood: model.MoodHappy,
		},
	}

	resp := service.toJournalResponse(context.Background(), entry)
	if resp.ID != entry.ID || resp.UUID == "" {
		t.Fatalf("unexpected mapped response: %+v", resp)
	}
	if resp.MoodLabel != "happy" || resp.MoodEmoji != "😊" {
		t.Fatalf("unexpected mood mapping: %s %s", resp.MoodLabel, resp.MoodEmoji)
	}
	if resp.Tags == nil || len(resp.Tags) != 0 {
		t.Fatalf("expected empty tags slice, got %+v", resp.Tags)
	}
}

func TestToJournalListResponse(t *testing.T) {
	service := &JournalService{}
	content := strings.Repeat("a", 180)
	moodID := uint(3)

	entry := &model.Journal{
		ID:          10,
		UUID:        uuid.New(),
		Title:       "Catatan",
		Content:     content,
		MoodID:      &moodID,
		Tags:        pq.StringArray{"tag1", "tag2"},
		ShareWithAI: false,
		Mood: &model.UserMood{
			Mood: model.MoodSad,
		},
	}

	resp := service.toJournalListResponse(context.Background(), entry)
	if len(resp.Preview) != 153 {
		t.Fatalf("expected truncated preview with ellipsis, got length %d", len(resp.Preview))
	}
	if !strings.HasSuffix(resp.Preview, "...") {
		t.Fatalf("expected preview to end with ellipsis: %s", resp.Preview)
	}
	if resp.MoodEmoji != "😢" {
		t.Fatalf("unexpected mood emoji: %s", resp.MoodEmoji)
	}
	if len(resp.Tags) != 2 {
		t.Fatalf("unexpected tags: %+v", resp.Tags)
	}
}
