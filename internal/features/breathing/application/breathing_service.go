package application

import (
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	gamificationpkg "github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/breathing/infrastructure")

// XP configuration constants
const (
	XPFor2Min         = 5
	XPFor5Min         = 10
	XPFor10Min        = 15
	XPFor15MinPlus    = 20
	DailyXPCap        = 30
	StreakBonusWeek   = 50  // 7 day streak bonus
	StreakBonusMonth  = 200 // 30 day streak bonus
	MinSessionSeconds = 60  // Minimum 1 minute for XP
)

type BreathingService interface {
	// Techniques
	GetAllTechniques(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error)
	GetTechniqueByID(ctx context.Context, userID uint, techniqueID uuid.UUID) (*dto.BreathingTechniqueResponse, error)
	GetTechniqueBySlug(ctx context.Context, userID uint, slug string) (*dto.BreathingTechniqueResponse, error)
	CreateCustomTechnique(ctx context.Context, userID uint, req dto.CreateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error)
	UpdateCustomTechnique(ctx context.Context, userID uint, techniqueID uuid.UUID, req dto.UpdateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error)
	DeleteCustomTechnique(ctx context.Context, userID uint, techniqueID uuid.UUID) error

	// Sessions
	StartSession(ctx context.Context, userID uint, req dto.StartBreathingSessionRequest) (*dto.BreathingSessionResponse, error)
	CompleteSession(ctx context.Context, userID uint, sessionID uuid.UUID, req dto.CompleteBreathingSessionRequest) (*dto.SessionCompletionResult, error)
	GetSessionHistory(ctx context.Context, userID uint, req dto.SessionHistoryRequest) (*dto.SessionHistoryResponse, error)
	GetSessionByID(ctx context.Context, userID uint, sessionID uuid.UUID) (*dto.BreathingSessionResponse, error)

	// Preferences
	GetPreferences(ctx context.Context, userID uint) (*dto.BreathingPreferencesResponse, error)
	UpdatePreferences(ctx context.Context, userID uint, req dto.UpdateBreathingPreferencesRequest) (*dto.BreathingPreferencesResponse, error)

	// Favorites
	GetFavorites(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error)
	AddFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) error
	RemoveFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) error
	ReorderFavorites(ctx context.Context, userID uint, techniqueIDs []uuid.UUID) error

	// Stats
	GetStats(ctx context.Context, userID uint) (*dto.BreathingStatsResponse, error)
	GetCalendar(ctx context.Context, userID uint, year, month int) (*dto.BreathingCalendarResponse, error)
	GetTechniqueUsage(ctx context.Context, userID uint) ([]dto.TechniqueUsageStats, error)
	GetWidgetData(ctx context.Context, userID uint) (*dto.BreathingWidgetData, error)

	// Recommendations
	GetRecommendations(ctx context.Context, userID uint, mood string, timeOfDay string) (*dto.RecommendationsResponse, error)
}

type breathingService struct {
	repo             infrastructure.BreathingRepository
	gamificationSvc  *gamificationapp.GamificationService
	dailyTaskService dailytaskapp.DailyTaskService
}

func NewBreathingService(repo infrastructure.BreathingRepository, gamificationSvc *gamificationapp.GamificationService, dailyTaskService dailytaskapp.DailyTaskService) BreathingService {
	return &breathingService{
		repo:             repo,
		gamificationSvc:  gamificationSvc,
		dailyTaskService: dailyTaskService,
	}
}

// ==========================================
// Techniques
// ==========================================

func (s *breathingService) GetAllTechniques(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error) {
	// Get system techniques
	systemTechniques, err := s.repo.GetSystemTechniques(ctx)
	if err != nil {
		return nil, err
	}

	// Get user's custom techniques
	userTechniques, err := s.repo.GetUserTechniques(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Combine all techniques
	allTechniques := append(systemTechniques, userTechniques...)

	// Get favorites to mark them
	favorites, _ := s.repo.GetFavorites(ctx, userID)
	favMap := make(map[uuid.UUID]bool)
	for _, f := range favorites {
		favMap[f.TechniqueID] = true
	}

	result := make([]dto.BreathingTechniqueResponse, len(allTechniques))
	for i, t := range allTechniques {
		result[i] = s.techniqueToResponse(ctx, &t, favMap[t.ID])
	}

	return result, nil
}

func (s *breathingService) GetTechniqueByID(ctx context.Context, userID uint, techniqueID uuid.UUID) (*dto.BreathingTechniqueResponse, error) {
	technique, err := s.repo.GetTechniqueByID(ctx, techniqueID)
	if err != nil {
		return nil, err
	}

	// Check access: system techniques are accessible by all, custom only by owner
	if !technique.IsSystem && technique.UserID != nil && *technique.UserID != int(userID) {
		return nil, errors.New("technique not found")
	}

	isFav, _ := s.repo.IsFavorite(ctx, userID, techniqueID)
	resp := s.techniqueToResponse(ctx, technique, isFav)
	return &resp, nil
}

func (s *breathingService) GetTechniqueBySlug(ctx context.Context, userID uint, slug string) (*dto.BreathingTechniqueResponse, error) {
	technique, err := s.repo.GetTechniqueBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if !technique.IsSystem && technique.UserID != nil && *technique.UserID != int(userID) {
		return nil, errors.New("technique not found")
	}

	isFav, _ := s.repo.IsFavorite(ctx, userID, technique.ID)
	resp := s.techniqueToResponse(ctx, technique, isFav)
	return &resp, nil
}

func (s *breathingService) CreateCustomTechnique(ctx context.Context, userID uint, req dto.CreateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
	slug := s.generateSlug(ctx, req.Name, userID)
	userIDInt := int(userID)

	technique := &model.BreathingTechnique{
		Name:               req.Name,
		Slug:               &slug,
		Description:        strPtr(req.Description),
		InhaleDuration:     req.InhaleDuration,
		InhaleHoldDuration: req.InhaleHoldDuration,
		ExhaleDuration:     req.ExhaleDuration,
		ExhaleHoldDuration: req.ExhaleHoldDuration,
		Icon:               s.defaultIfEmpty(ctx, req.Icon, "🌬️"),
		Color:              s.defaultIfEmpty(ctx, req.Color, "#6366F1"),
		AnimationType:      "circle",
		Difficulty:         "custom",
		Category:           "custom",
		IsSystem:           false,
		IsActive:           true,
		UserID:             &userIDInt,
	}

	if err := s.repo.CreateTechnique(ctx, technique); err != nil {
		return nil, err
	}

	resp := s.techniqueToResponse(ctx, technique, false)
	return &resp, nil
}

func (s *breathingService) UpdateCustomTechnique(ctx context.Context, userID uint, techniqueID uuid.UUID, req dto.UpdateBreathingTechniqueRequest) (*dto.BreathingTechniqueResponse, error) {
	technique, err := s.repo.GetTechniqueByID(ctx, techniqueID)
	if err != nil {
		return nil, err
	}

	// Only allow updating custom techniques owned by user
	if technique.IsSystem || technique.UserID == nil || *technique.UserID != int(userID) {
		return nil, errors.New("cannot update this technique")
	}

	// Update fields if provided
	if req.Name != nil {
		technique.Name = *req.Name
	}
	if req.Description != nil {
		technique.Description = req.Description
	}
	if req.InhaleDuration != nil {
		technique.InhaleDuration = *req.InhaleDuration
	}
	if req.InhaleHoldDuration != nil {
		technique.InhaleHoldDuration = *req.InhaleHoldDuration
	}
	if req.ExhaleDuration != nil {
		technique.ExhaleDuration = *req.ExhaleDuration
	}
	if req.ExhaleHoldDuration != nil {
		technique.ExhaleHoldDuration = *req.ExhaleHoldDuration
	}
	if req.Icon != nil {
		technique.Icon = *req.Icon
	}
	if req.Color != nil {
		technique.Color = *req.Color
	}

	if err := s.repo.UpdateTechnique(ctx, technique); err != nil {
		return nil, err
	}

	isFav, _ := s.repo.IsFavorite(ctx, userID, techniqueID)
	resp := s.techniqueToResponse(ctx, technique, isFav)
	return &resp, nil
}

func (s *breathingService) DeleteCustomTechnique(ctx context.Context, userID uint, techniqueID uuid.UUID) error {
	technique, err := s.repo.GetTechniqueByID(ctx, techniqueID)
	if err != nil {
		return err
	}

	if technique.IsSystem || technique.UserID == nil || *technique.UserID != int(userID) {
		return errors.New("cannot delete this technique")
	}

	return s.repo.DeleteTechnique(ctx, techniqueID)
}

// ==========================================
// Sessions
// ==========================================

func (s *breathingService) StartSession(ctx context.Context, userID uint, req dto.StartBreathingSessionRequest) (*dto.BreathingSessionResponse, error) {
	// Verify technique exists and is accessible
	technique, err := s.repo.GetTechniqueByID(ctx, req.TechniqueID)
	if err != nil {
		return nil, errors.New("technique not found")
	}

	if !technique.IsSystem && technique.UserID != nil && *technique.UserID != int(userID) {
		return nil, errors.New("technique not found")
	}

	session := &model.BreathingSession{
		UserID:                int(userID),
		TechniqueID:           req.TechniqueID,
		Technique:             technique,
		TargetDurationSeconds: req.TargetDurationSeconds,
		DurationSeconds:       0,
		VoiceGuidanceEnabled:  req.VoiceGuidanceEnabled,
		BackgroundSound:       strPtr(req.BackgroundSound),
		HapticFeedbackEnabled: req.HapticFeedbackEnabled,
		MoodBefore:            strPtr(req.MoodBefore),
		StartedAt:             time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	resp := s.sessionToResponse(ctx, session)
	return &resp, nil
}

func (s *breathingService) CompleteSession(ctx context.Context, userID uint, sessionID uuid.UUID, req dto.CompleteBreathingSessionRequest) (*dto.SessionCompletionResult, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.UserID != int(userID) {
		return nil, errors.New("session not found")
	}

	// Update session
	now := time.Now()
	session.DurationSeconds = req.DurationSeconds
	session.CyclesCompleted = req.CyclesCompleted
	session.Completed = req.Completed
	session.CompletedPercentage = req.CompletedPercentage
	session.EndedAt = &now
	session.MoodAfter = strPtr(req.MoodAfter)

	// Calculate XP
	baseXP := s.calculateXP(ctx, req.DurationSeconds)
	bonusXP := 0
	bonusReason := ""

	// Check for streak bonus
	prefs, _ := s.repo.GetOrCreatePreferences(ctx, userID)
	newStreak := s.updateStreak(ctx, prefs)
	streakMilestone := false
	streakMilestoneXP := 0

	// Award streak milestones
	if newStreak == 7 {
		streakMilestone = true
		streakMilestoneXP = StreakBonusWeek
		bonusXP += StreakBonusWeek
		bonusReason = "7 day streak bonus!"
	} else if newStreak == 30 {
		streakMilestone = true
		streakMilestoneXP = StreakBonusMonth
		bonusXP += StreakBonusMonth
		bonusReason = "30 day streak bonus!"
	}

	// Apply daily cap
	todayXP, _ := s.repo.GetUserTodayXP(ctx, userID)
	remainingCap := DailyXPCap - todayXP
	totalXP := baseXP + bonusXP
	if totalXP > remainingCap && remainingCap > 0 {
		totalXP = remainingCap
	} else if remainingCap <= 0 {
		totalXP = 0
	}

	session.XPEarned = totalXP

	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	// Update daily task progress (ONLY if completed)
	if session.Completed && s.dailyTaskService != nil {
		// Use fire-and-forget or handle error? Handling error is safer but shouldn't block session completion
		_ = s.dailyTaskService.UpdateTaskProgress(ctx, userID, model.TaskTypeBreathing)
	}

	// Award XP to user (Gamification Service)
	if totalXP > 0 && s.gamificationSvc != nil {
		_ = s.gamificationSvc.AwardExp(ctx, userID, gamificationpkg.ActivityBreathing, int64(totalXP))
	}

	return &dto.SessionCompletionResult{
		Session:           s.sessionToResponse(ctx, session),
		XPEarned:          baseXP,
		BonusXP:           bonusXP,
		BonusReason:       bonusReason,
		TotalXP:           totalXP,
		NewStreak:         newStreak,
		StreakMilestone:   streakMilestone,
		StreakMilestoneXP: streakMilestoneXP,
		DailyXPRemaining:  DailyXPCap - todayXP - totalXP,
	}, nil
}

func (s *breathingService) GetSessionHistory(ctx context.Context, userID uint, req dto.SessionHistoryRequest) (*dto.SessionHistoryResponse, error) {
	limit := 20
	page := 1
	if req.Limit > 0 {
		limit = req.Limit
	}
	if req.Page > 0 {
		page = req.Page
	}
	offset := (page - 1) * limit

	// Parse dates if provided
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}

	// Parse technique ID if provided
	var techniqueID *uuid.UUID
	if req.TechniqueID != "" {
		if id, err := uuid.Parse(req.TechniqueID); err == nil {
			techniqueID = &id
		}
	}

	sessions, total, err := s.repo.GetUserSessions(ctx, userID, startDate, endDate, techniqueID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.BreathingSessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = s.sessionToResponse(ctx, &session)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.SessionHistoryResponse{
		Sessions:   responses,
		Total:      int(total),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *breathingService) GetSessionByID(ctx context.Context, userID uint, sessionID uuid.UUID) (*dto.BreathingSessionResponse, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != int(userID) {
		return nil, errors.New("session not found")
	}

	resp := s.sessionToResponse(ctx, session)
	return &resp, nil
}

// ==========================================
// Preferences
// ==========================================

func (s *breathingService) GetPreferences(ctx context.Context, userID uint) (*dto.BreathingPreferencesResponse, error) {
	prefs, err := s.repo.GetOrCreatePreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.BreathingPreferencesResponse{
		DefaultDurationSeconds: prefs.DefaultDurationSeconds,
		DefaultTechniqueID:     prefs.DefaultTechniqueID,
		VoiceGuidance:          prefs.VoiceGuidance,
		BackgroundSound:        prefs.BackgroundSound,
		DefaultBackgroundSound: prefs.DefaultBackgroundSound,
		HapticFeedback:         prefs.HapticFeedback,
		AnimationSpeed:         prefs.AnimationSpeed,
		Theme:                  prefs.Theme,
		ReminderEnabled:        prefs.ReminderEnabled,
		ReminderTime:           ptrToStr(prefs.ReminderTime),
		ReminderDays:           prefs.ReminderDays,
		TutorialCompleted:      prefs.TutorialCompleted,
	}, nil
}

func (s *breathingService) UpdatePreferences(ctx context.Context, userID uint, req dto.UpdateBreathingPreferencesRequest) (*dto.BreathingPreferencesResponse, error) {
	prefs, err := s.repo.GetOrCreatePreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.DefaultDurationSeconds != nil {
		prefs.DefaultDurationSeconds = *req.DefaultDurationSeconds
	}
	if req.DefaultTechniqueID != nil {
		prefs.DefaultTechniqueID = req.DefaultTechniqueID
	}
	if req.VoiceGuidance != nil {
		prefs.VoiceGuidance = *req.VoiceGuidance
	}
	if req.BackgroundSound != nil {
		prefs.BackgroundSound = *req.BackgroundSound
	}
	if req.DefaultBackgroundSound != nil {
		prefs.DefaultBackgroundSound = *req.DefaultBackgroundSound
	}
	if req.HapticFeedback != nil {
		prefs.HapticFeedback = *req.HapticFeedback
	}
	if req.AnimationSpeed != nil {
		prefs.AnimationSpeed = *req.AnimationSpeed
	}
	if req.Theme != nil {
		prefs.Theme = *req.Theme
	}
	if req.ReminderEnabled != nil {
		prefs.ReminderEnabled = *req.ReminderEnabled
	}
	if req.ReminderTime != nil {
		prefs.ReminderTime = req.ReminderTime
	}
	if req.ReminderDays != nil {
		prefs.ReminderDays = *req.ReminderDays
	}

	if err := s.repo.UpdatePreferences(ctx, prefs); err != nil {
		return nil, err
	}

	return s.GetPreferences(ctx, userID)
}

// ==========================================
// Favorites
// ==========================================

func (s *breathingService) GetFavorites(ctx context.Context, userID uint) ([]dto.BreathingTechniqueResponse, error) {
	favorites, err := s.repo.GetFavorites(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BreathingTechniqueResponse, 0, len(favorites))
	for _, f := range favorites {
		if f.Technique != nil {
			result = append(result, s.techniqueToResponse(ctx, f.Technique, true))
		}
	}

	return result, nil
}

func (s *breathingService) AddFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) error {
	// Verify technique exists
	_, err := s.repo.GetTechniqueByID(ctx, techniqueID)
	if err != nil {
		return errors.New("technique not found")
	}

	// Check if already favorite
	isFav, _ := s.repo.IsFavorite(ctx, userID, techniqueID)
	if isFav {
		return nil // Already a favorite
	}

	favorite := &model.BreathingFavorite{
		UserID:      int(userID),
		TechniqueID: techniqueID,
	}

	return s.repo.AddFavorite(ctx, favorite)
}

func (s *breathingService) RemoveFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) error {
	return s.repo.RemoveFavorite(ctx, userID, techniqueID)
}

func (s *breathingService) ReorderFavorites(ctx context.Context, userID uint, techniqueIDs []uuid.UUID) error {
	return s.repo.UpdateFavoriteOrder(ctx, userID, techniqueIDs)
}

// ==========================================
// Stats
// ==========================================

func (s *breathingService) GetStats(ctx context.Context, userID uint) (*dto.BreathingStatsResponse, error) {
	prefs, _ := s.repo.GetOrCreatePreferences(ctx, userID)

	totalSessions, _ := s.repo.GetUserTotalSessionsCount(ctx, userID)
	totalMinutes, _ := s.repo.GetUserTotalMinutes(ctx, userID)
	completionRate, _ := s.repo.GetCompletionRate(ctx, userID)
	mostUsedTechnique, _, _ := s.repo.GetMostUsedTechnique(ctx, userID)

	todaySessions, todayMinutes, _ := s.repo.GetDailyStats(ctx, userID, time.Now())

	var mostUsedName, mostUsedID string
	if mostUsedTechnique != nil {
		mostUsedName = mostUsedTechnique.Name
		mostUsedID = mostUsedTechnique.ID.String()
	}

	// Calculate average sessions per week (last 30 days)
	avgSessionsPerWeek := float64(totalSessions) / 4.0 // Simple approximation

	// Get favorite technique name for today
	favTechniqueName := ""
	favorites, _ := s.repo.GetFavorites(ctx, userID)
	if len(favorites) > 0 && favorites[0].Technique != nil {
		favTechniqueName = favorites[0].Technique.Name
	}

	// Determine last practice date and streak info
	lastPracticeDate := ""
	daysUntilBreak := 1
	if prefs.LastPracticeDate != nil {
		lastPracticeDate = prefs.LastPracticeDate.Format("2006-01-02")
		today := time.Now().Truncate(24 * time.Hour)
		lastPractice := prefs.LastPracticeDate.Truncate(24 * time.Hour)
		if today.Sub(lastPractice).Hours()/24 >= 1 {
			daysUntilBreak = 0 // Must practice today to keep streak
		}
	}

	return &dto.BreathingStatsResponse{
		Today: dto.BreathingDailyStats{
			Date:              time.Now().Format("2006-01-02"),
			SessionsCount:     todaySessions,
			TotalMinutes:      todayMinutes,
			FavoriteTechnique: favTechniqueName,
		},
		Overall: dto.BreathingOverallStats{
			TotalSessions:          int(totalSessions),
			TotalMinutes:           totalMinutes,
			CurrentStreak:          prefs.CurrentStreak,
			LongestStreak:          prefs.LongestStreak,
			MostUsedTechnique:      mostUsedName,
			MostUsedTechniqueID:    mostUsedID,
			AverageSessionsPerWeek: avgSessionsPerWeek,
			CompletionRate:         completionRate,
		},
		StreakInfo: dto.StreakInfo{
			CurrentStreak:         prefs.CurrentStreak,
			LongestStreak:         prefs.LongestStreak,
			LastPracticeDate:      lastPracticeDate,
			StreakFreezeAvailable: false, // Not implemented yet
			DaysUntilStreakBreak:  daysUntilBreak,
		},
	}, nil
}

func (s *breathingService) GetCalendar(ctx context.Context, userID uint, year, month int) (*dto.BreathingCalendarResponse, error) {
	calendarData, err := s.repo.GetMonthlyCalendar(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	days := make([]dto.BreathingCalendarDay, len(calendarData))
	for i, d := range calendarData {
		// Calculate intensity (0-3)
		intensity := 0
		if d.SessionsCount > 0 {
			intensity = 1
		}
		if d.SessionsCount >= 2 {
			intensity = 2
		}
		if d.SessionsCount >= 4 {
			intensity = 3
		}

		days[i] = dto.BreathingCalendarDay{
			Date:           d.Date.Format("2006-01-02"),
			SessionsCount:  d.SessionsCount,
			TotalMinutes:   d.TotalMinutes,
			TechniquesUsed: d.TechniquesUsed,
			Intensity:      intensity,
		}
	}

	return &dto.BreathingCalendarResponse{
		Year:  year,
		Month: month,
		Days:  days,
	}, nil
}

func (s *breathingService) GetTechniqueUsage(ctx context.Context, userID uint) ([]dto.TechniqueUsageStats, error) {
	usage, err := s.repo.GetTechniqueUsageStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.TechniqueUsageStats, len(usage))
	for i, u := range usage {
		result[i] = dto.TechniqueUsageStats{
			TechniqueID:   u.TechniqueID.String(),
			TechniqueName: u.TechniqueName,
			SessionsCount: int(u.SessionsCount),
			TotalMinutes:  u.TotalMinutes,
		}
	}

	return result, nil
}

func (s *breathingService) GetWidgetData(ctx context.Context, userID uint) (*dto.BreathingWidgetData, error) {
	prefs, _ := s.repo.GetOrCreatePreferences(ctx, userID)
	todaySessions, todayMinutes, _ := s.repo.GetDailyStats(ctx, userID, time.Now())

	var favTechnique *dto.BreathingTechniqueResponse
	favorites, _ := s.repo.GetFavorites(ctx, userID)
	if len(favorites) > 0 && favorites[0].Technique != nil {
		resp := s.techniqueToResponse(ctx, favorites[0].Technique, true)
		favTechnique = &resp
	}

	// Default daily goal: 5 minutes
	dailyGoalMinutes := 5
	dailyGoalProgress := 0
	if dailyGoalMinutes > 0 {
		dailyGoalProgress = (todayMinutes * 100) / dailyGoalMinutes
		if dailyGoalProgress > 100 {
			dailyGoalProgress = 100
		}
	}

	// Determine if needs practice today (to maintain streak)
	needsPracticeToday := true
	streakAtRisk := false
	if prefs.LastPracticeDate != nil {
		today := time.Now().Truncate(24 * time.Hour)
		lastPractice := prefs.LastPracticeDate.Truncate(24 * time.Hour)
		daysDiff := today.Sub(lastPractice).Hours() / 24
		if daysDiff == 0 {
			needsPracticeToday = false // Already practiced today
		} else if daysDiff == 1 && prefs.CurrentStreak > 0 {
			streakAtRisk = true // Streak at risk, needs to practice today
		}
	}

	// Get last session time
	var lastSessionAt *time.Time
	sessions, _, _ := s.repo.GetUserSessions(ctx, userID, nil, nil, nil, 1, 0)
	if len(sessions) > 0 && sessions[0].EndedAt != nil {
		lastSessionAt = sessions[0].EndedAt
	}

	return &dto.BreathingWidgetData{
		CurrentStreak:      prefs.CurrentStreak,
		TodaySessions:      todaySessions,
		TodayMinutes:       todayMinutes,
		DailyGoalMinutes:   dailyGoalMinutes,
		DailyGoalProgress:  dailyGoalProgress,
		FavoriteTechnique:  favTechnique,
		LastSessionAt:      lastSessionAt,
		NeedsPracticeToday: needsPracticeToday,
		StreakAtRisk:       streakAtRisk,
	}, nil
}

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
	today := time.Now().Truncate(24 * time.Hour)

	if prefs.LastPracticeDate == nil {
		// First practice
		prefs.CurrentStreak = 1
		prefs.LastPracticeDate = &today
	} else {
		lastPractice := prefs.LastPracticeDate.Truncate(24 * time.Hour)
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

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
