package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"github.com/google/uuid"
)

// ==========================================
// Recommendations
// ==========================================

func (s *breathingService) GetRecommendations(ctx context.Context, userID uint, mood string, timeOfDay string) (*dto.RecommendationsResponse, error) {
	techniques, err := s.repo.GetSystemTechniques(ctx)
	if err != nil {
		return nil, err
	}

	var recommended []dto.BreathingTechniqueResponse
	var reason string

	// Get favorites for marking
	favorites, _ := s.repo.GetFavorites(ctx, userID)
	favMap := make(map[uuid.UUID]bool)
	for _, f := range favorites {
		favMap[f.TechniqueID] = true
	}

	// Recommend based on mood
	for _, t := range techniques {
		match := false
		switch mood {
		case "stressed", "anxious":
			if t.Slug != nil && (*t.Slug == "box-breathing" || *t.Slug == "4-7-8-relaxing") {
				match = true
				reason = "Teknik ini sangat efektif untuk menenangkan pikiran dan mengurangi kecemasan"
			}
		case "tired", "low_energy":
			if t.Slug != nil && *t.Slug == "energizing-breath" {
				match = true
				reason = "Teknik ini dapat membantu meningkatkan energi dan fokus"
			}
		case "angry", "frustrated":
			if t.Slug != nil && *t.Slug == "deep-calm" {
				match = true
				reason = "Napas dalam dan panjang membantu menenangkan emosi yang kuat"
			}
		case "neutral", "calm":
			if t.Slug != nil && *t.Slug == "coherent-breathing" {
				match = true
				reason = "Teknik ini cocok untuk menjaga ketenangan dan keseimbangan"
			}
		default:
			if t.Slug != nil && *t.Slug == "box-breathing" {
				match = true
				reason = "Box Breathing adalah teknik dasar yang cocok untuk semua kondisi"
			}
		}

		if match {
			recommended = append(recommended, s.techniqueToResponse(ctx, &t, favMap[t.ID]))
		}
	}

	// Add time-based recommendations
	switch timeOfDay {
	case "morning":
		for _, t := range techniques {
			if t.Slug != nil && *t.Slug == "energizing-breath" {
				found := false
				for _, r := range recommended {
					if r.ID == t.ID {
						found = true
						break
					}
				}
				if !found {
					recommended = append(recommended, s.techniqueToResponse(ctx, &t, favMap[t.ID]))
				}
			}
		}
	case "night", "evening":
		for _, t := range techniques {
			if t.Slug != nil && *t.Slug == "4-7-8-relaxing" {
				found := false
				for _, r := range recommended {
					if r.ID == t.ID {
						found = true
						break
					}
				}
				if !found {
					recommended = append(recommended, s.techniqueToResponse(ctx, &t, favMap[t.ID]))
				}
			}
		}
	}

	// Default if no recommendations
	if len(recommended) == 0 && len(techniques) > 0 {
		recommended = append(recommended, s.techniqueToResponse(ctx, &techniques[0], favMap[techniques[0].ID]))
		reason = "Teknik dasar yang cocok untuk memulai latihan pernapasan"
	}

	// Build mood-based recommendations
	var basedOnMood []dto.TechniqueRecommendation
	for i, r := range recommended {
		basedOnMood = append(basedOnMood, dto.TechniqueRecommendation{
			Technique: r,
			Reason:    reason,
			Priority:  i + 1,
		})
	}

	// Build time-based recommendations
	var basedOnTime []dto.TechniqueRecommendation
	for _, t := range techniques {
		matched := false
		timeReason := ""
		switch timeOfDay {
		case "morning":
			if t.Slug != nil && *t.Slug == "energizing-breath" {
				matched = true
				timeReason = "Sempurna untuk memulai hari dengan energi"
			}
		case "afternoon":
			if t.Slug != nil && *t.Slug == "box-breathing" {
				matched = true
				timeReason = "Bantu fokus di tengah aktivitas"
			}
		case "night", "evening":
			if t.Slug != nil && (*t.Slug == "4-7-8-relaxing" || *t.Slug == "deep-calm") {
				matched = true
				timeReason = "Persiapkan tubuh untuk istirahat malam"
			}
		}
		if matched {
			basedOnTime = append(basedOnTime, dto.TechniqueRecommendation{
				Technique: s.techniqueToResponse(ctx, &t, favMap[t.ID]),
				Reason:    timeReason,
				Priority:  len(basedOnTime) + 1,
			})
		}
	}

	// Build default pick
	var defaultPick *dto.TechniqueRecommendation
	if len(basedOnMood) > 0 {
		defaultPick = &basedOnMood[0]
	} else if len(basedOnTime) > 0 {
		defaultPick = &basedOnTime[0]
	} else if len(techniques) > 0 {
		defaultPick = &dto.TechniqueRecommendation{
			Technique: s.techniqueToResponse(ctx, &techniques[0], favMap[techniques[0].ID]),
			Reason:    "Teknik dasar yang cocok untuk semua kondisi",
			Priority:  1,
		}
	}

	return &dto.RecommendationsResponse{
		BasedOnMood: basedOnMood,
		BasedOnTime: basedOnTime,
		DefaultPick: defaultPick,
	}, nil
}

// ==========================================
// Helper functions
// ==========================================

func (s *breathingService) techniqueToResponse(ctx context.Context, t *model.BreathingTechnique, isFavorite bool) dto.BreathingTechniqueResponse {
	return dto.BreathingTechniqueResponse{
		ID:                 t.ID,
		Name:               t.Name,
		Slug:               ptrToStr(t.Slug),
		Description:        ptrToStr(t.Description),
		Benefits:           ptrToStr(t.Benefits),
		BestFor:            ptrToStr(t.BestFor),
		InhaleDuration:     t.InhaleDuration,
		InhaleHoldDuration: t.InhaleHoldDuration,
		ExhaleDuration:     t.ExhaleDuration,
		ExhaleHoldDuration: t.ExhaleHoldDuration,
		TotalCycleDuration: t.GetTotalCycleDuration(),
		Icon:               t.Icon,
		Color:              t.Color,
		AnimationType:      t.AnimationType,
		Difficulty:         t.Difficulty,
		Category:           t.Category,
		Origin:             ptrToStr(t.Origin),
		IsSystem:           t.IsSystem,
		IsFavorite:         isFavorite,
		CreatedAt:          t.CreatedAt,
	}
}

func (s *breathingService) sessionToResponse(ctx context.Context, session *model.BreathingSession) dto.BreathingSessionResponse {
	resp := dto.BreathingSessionResponse{
		ID:                    session.ID,
		TechniqueID:           session.TechniqueID,
		DurationSeconds:       session.DurationSeconds,
		TargetDurationSeconds: session.TargetDurationSeconds,
		CyclesCompleted:       session.CyclesCompleted,
		VoiceGuidanceEnabled:  session.VoiceGuidanceEnabled,
		BackgroundSound:       ptrToStr(session.BackgroundSound),
		HapticFeedbackEnabled: session.HapticFeedbackEnabled,
		Completed:             session.Completed,
		CompletedPercentage:   session.CompletedPercentage,
		StartedAt:             session.StartedAt,
		EndedAt:               session.EndedAt,
		XPEarned:              session.XPEarned,
		MoodBefore:            ptrToStr(session.MoodBefore),
		MoodAfter:             ptrToStr(session.MoodAfter),
	}

	if session.Technique != nil {
		tech := s.techniqueToResponse(ctx, session.Technique, false)
		resp.Technique = &tech
	}

	return resp
}

func (s *breathingService) generateSlug(ctx context.Context, name string, userID uint) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	// Add user ID prefix for uniqueness
	return fmt.Sprintf("custom-%d-%s", userID, slug)
}

func (s *breathingService) defaultIfEmpty(ctx context.Context, value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func (s *breathingService) calculateXP(ctx context.Context, durationSeconds int) int {
	if durationSeconds < MinSessionSeconds {
		return 0
	}

	minutes := durationSeconds / 60
	if minutes >= 15 {
		return XPFor15MinPlus
	} else if minutes >= 10 {
		return XPFor10Min
	} else if minutes >= 5 {
		return XPFor5Min
	} else if minutes >= 2 {
		return XPFor2Min
	}
	return 0
}

func (s *breathingService) updateStreak(ctx context.Context, prefs *model.BreathingPreference) int {
	today := timeutil.Today()

	if prefs.LastPracticeDate == nil {
		// First practice
		prefs.CurrentStreak = 1
		prefs.LastPracticeDate = &today
	} else {
		lastPractice := timeutil.StartOfDay(*prefs.LastPracticeDate)
		diff := today.Sub(lastPractice).Hours() / 24

		if diff == 0 {
			// Same day, no change
		} else if diff == 1 {
			// Consecutive day
			prefs.CurrentStreak++
			prefs.LastPracticeDate = &today
		} else {
			// Streak broken
			prefs.CurrentStreak = 1
			prefs.LastPracticeDate = &today
		}
	}

	// Update longest streak
	if prefs.CurrentStreak > prefs.LongestStreak {
		prefs.LongestStreak = prefs.CurrentStreak
	}

	_ = s.repo.UpdatePreferences(ctx, prefs)

	return prefs.CurrentStreak
}

// Helper functions for pointer conversion
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
