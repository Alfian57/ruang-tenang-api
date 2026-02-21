package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
)

func TestBreathingService_HelperFunctions(t *testing.T) {
	s := &breathingService{}

	slug := s.generateSlug(context.Background(), "4-7-8 Breathing! Pro", 42)
	if slug != "custom-42-4-7-8-breathing-pro" {
		t.Fatalf("unexpected slug: %s", slug)
	}

	if got := s.defaultIfEmpty(context.Background(), "", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
	if got := s.defaultIfEmpty(context.Background(), "value", "fallback"); got != "value" {
		t.Fatalf("expected value, got %s", got)
	}

	if s.calculateXP(context.Background(), 30) != 0 {
		t.Fatal("expected 0 XP for short session")
	}
	if s.calculateXP(context.Background(), 120) != XPFor2Min {
		t.Fatal("expected XPFor2Min")
	}
	if s.calculateXP(context.Background(), 300) != XPFor5Min {
		t.Fatal("expected XPFor5Min")
	}
	if s.calculateXP(context.Background(), 600) != XPFor10Min {
		t.Fatal("expected XPFor10Min")
	}
	if s.calculateXP(context.Background(), 900) != XPFor15MinPlus {
		t.Fatal("expected XPFor15MinPlus")
	}
}

func TestBreathingService_ResponseMappers(t *testing.T) {
	s := &breathingService{}
	now := time.Now()

	slug := "box-breathing"
	desc := "desc"
	benefits := "benefits"
	bestFor := "best"
	origin := "origin"
	bg := "rain"
	moodBefore := "anxious"
	moodAfter := "calm"

	tech := &model.BreathingTechnique{
		ID:                 uuid.New(),
		Name:               "Box",
		Slug:               &slug,
		Description:        &desc,
		Benefits:           &benefits,
		BestFor:            &bestFor,
		InhaleDuration:     4,
		InhaleHoldDuration: 4,
		ExhaleDuration:     4,
		ExhaleHoldDuration: 4,
		Icon:               "⬜",
		Color:              "#fff",
		AnimationType:      "square",
		Difficulty:         "easy",
		Category:           "stress",
		Origin:             &origin,
		IsSystem:           true,
		CreatedAt:          now,
	}

	resp := s.techniqueToResponse(context.Background(), tech, true)
	if !resp.IsFavorite || resp.TotalCycleDuration != 16 || resp.Slug != slug {
		t.Fatalf("unexpected technique response: %+v", resp)
	}

	session := &model.BreathingSession{
		ID:                    uuid.New(),
		TechniqueID:           tech.ID,
		DurationSeconds:       120,
		TargetDurationSeconds: 300,
		CyclesCompleted:       8,
		VoiceGuidanceEnabled:  true,
		BackgroundSound:       &bg,
		HapticFeedbackEnabled: true,
		Completed:             true,
		CompletedPercentage:   100,
		StartedAt:             now,
		EndedAt:               &now,
		XPEarned:              10,
		MoodBefore:            &moodBefore,
		MoodAfter:             &moodAfter,
		Technique:             tech,
	}

	sessionResp := s.sessionToResponse(context.Background(), session)
	if sessionResp.Technique == nil || sessionResp.BackgroundSound != "rain" || sessionResp.MoodAfter != "calm" {
		t.Fatalf("unexpected session response: %+v", sessionResp)
	}
}

func TestPointerAndFilenameAndCrisisHelpers(t *testing.T) {
	if strPtr("") != nil {
		t.Fatal("expected nil for empty strPtr")
	}
	v := strPtr("x")
	if v == nil || *v != "x" {
		t.Fatal("expected pointer to x")
	}
	if ptrToStr(nil) != "" || ptrToStr(v) != "x" {
		t.Fatal("ptrToStr mismatch")
	}

	clean := sanitizeFilename("A/B:C*D?E\"F<G>H| I")
	if strings.ContainsAny(clean, "/\\:*?\"<>| ") {
		t.Fatalf("sanitizeFilename still contains invalid chars: %s", clean)
	}
	long := sanitizeFilename(strings.Repeat("x", 80))
	if len(long) != 50 {
		t.Fatalf("expected max filename length 50, got %d", len(long))
	}

	ai := &AIModerationService{}
	resp1 := ai.generateCrisisResponse(context.Background(), model.CrisisCategorySuicide, model.CrisisSeverityCritical)
	resp2 := ai.generateCrisisResponse(context.Background(), model.CrisisCategorySevereDepression, model.CrisisSeverityHigh)
	resp3 := ai.generateCrisisResponse(context.Background(), model.CrisisCategoryEmergency, model.CrisisSeverityMedium)
	if !strings.Contains(resp1, "119 ext 8") || !strings.Contains(resp2, "Kamu tidak sendirian") || !strings.Contains(resp3, "Bantuan tersedia") {
		t.Fatal("unexpected crisis response content")
	}
}
