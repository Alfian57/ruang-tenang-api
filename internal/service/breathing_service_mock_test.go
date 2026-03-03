package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
)

type mockBreathingRepo struct {
	sessionByID       *model.BreathingSession
	sessionByIDErr    error
	updateSessionErr  error
	updateSessionCall int
	updatedSession    *model.BreathingSession
	createSessionErr  error
	createSessionCall int
	createdSession    *model.BreathingSession

	prefs          *model.BreathingPreference
	updatePrefsErr error

	todayXP int

	techniqueByID       *model.BreathingTechnique
	techniqueByIDErr    error
	techniqueBySlug     *model.BreathingTechnique
	techniqueBySlugErr  error
	systemTechniques    []model.BreathingTechnique
	systemTechniquesErr error
	userTechniques      []model.BreathingTechnique
	userTechniquesErr   error
	favorites           []model.BreathingFavorite
	getFavoritesErr     error
	createTechniqueErr  error
	updateTechniqueErr  error
	deleteTechniqueErr  error
	removeFavoriteErr   error
	updateFavOrderErr   error
	getOrCreatePrefErr  error
	getUserSessionsErr  error
	monthlyCalendarErr  error
	usageStatsErr       error
	dailyStatsErr       error
	totalSessionsErr    error
	totalMinutesErr     error
	completionRateErr   error
	mostUsedErr         error
	countSinceErr       error
	userSessions        []model.BreathingSession
	userSessionsTotal   int64
	monthlyCalendar     []repository.DailyCalendarData
	usageStats          []repository.TechniqueUsageData
	dailyCount          int
	dailyMinutes        int
	totalSessions       int64
	totalMinutes        int
	completionRate      float64
	mostUsedTechnique   *model.BreathingTechnique
	mostUsedCount       int64
	countSince          int64
	createTechniqueCall int
	updateTechniqueCall int
	deleteTechniqueCall int
	createdTechnique    *model.BreathingTechnique
	updatedTechnique    *model.BreathingTechnique
	deletedTechniqueID  uuid.UUID
	isFavorite          bool
	addFavoriteCalls    int
}

func (m *mockBreathingRepo) GetAllTechniques(_ context.Context) ([]model.BreathingTechnique, error) {
	return nil, nil
}
func (m *mockBreathingRepo) GetSystemTechniques(_ context.Context) ([]model.BreathingTechnique, error) {
	return m.systemTechniques, m.systemTechniquesErr
}
func (m *mockBreathingRepo) GetUserTechniques(_ context.Context, _ uint) ([]model.BreathingTechnique, error) {
	return m.userTechniques, m.userTechniquesErr
}
func (m *mockBreathingRepo) GetTechniqueByID(_ context.Context, _ uuid.UUID) (*model.BreathingTechnique, error) {
	if m.techniqueByIDErr != nil {
		return nil, m.techniqueByIDErr
	}
	return m.techniqueByID, nil
}
func (m *mockBreathingRepo) GetTechniqueBySlug(_ context.Context, _ string) (*model.BreathingTechnique, error) {
	if m.techniqueBySlugErr != nil {
		return nil, m.techniqueBySlugErr
	}
	return m.techniqueBySlug, nil
}
func (m *mockBreathingRepo) CreateTechnique(_ context.Context, t *model.BreathingTechnique) error {
	m.createTechniqueCall++
	m.createdTechnique = t
	return m.createTechniqueErr
}
func (m *mockBreathingRepo) UpdateTechnique(_ context.Context, t *model.BreathingTechnique) error {
	m.updateTechniqueCall++
	m.updatedTechnique = t
	return m.updateTechniqueErr
}
func (m *mockBreathingRepo) DeleteTechnique(_ context.Context, id uuid.UUID) error {
	m.deleteTechniqueCall++
	m.deletedTechniqueID = id
	return m.deleteTechniqueErr
}

func (m *mockBreathingRepo) CreateSession(_ context.Context, s *model.BreathingSession) error {
	m.createSessionCall++
	m.createdSession = s
	return m.createSessionErr
}
func (m *mockBreathingRepo) UpdateSession(_ context.Context, session *model.BreathingSession) error {
	m.updateSessionCall++
	m.updatedSession = session
	return m.updateSessionErr
}
func (m *mockBreathingRepo) GetSessionByID(_ context.Context, _ uuid.UUID) (*model.BreathingSession, error) {
	if m.sessionByIDErr != nil {
		return nil, m.sessionByIDErr
	}
	return m.sessionByID, nil
}
func (m *mockBreathingRepo) GetUserSessions(_ context.Context, _ uint, _, _ *time.Time, _ *uuid.UUID, _, _ int) ([]model.BreathingSession, int64, error) {
	return m.userSessions, m.userSessionsTotal, m.getUserSessionsErr
}
func (m *mockBreathingRepo) GetUserSessionsToday(_ context.Context, _ uint) ([]model.BreathingSession, error) {
	return nil, nil
}
func (m *mockBreathingRepo) GetUserSessionsByDateRange(_ context.Context, _ uint, _, _ time.Time) ([]model.BreathingSession, error) {
	return nil, nil
}
func (m *mockBreathingRepo) GetUserTotalSessionsCount(_ context.Context, _ uint) (int64, error) {
	return m.totalSessions, m.totalSessionsErr
}
func (m *mockBreathingRepo) GetUserTotalMinutes(_ context.Context, _ uint) (int, error) {
	return m.totalMinutes, m.totalMinutesErr
}
func (m *mockBreathingRepo) GetUserTodayXP(_ context.Context, _ uint) (int, error) {
	return m.todayXP, nil
}
func (m *mockBreathingRepo) GetMostUsedTechnique(_ context.Context, _ uint) (*model.BreathingTechnique, int64, error) {
	return m.mostUsedTechnique, m.mostUsedCount, m.mostUsedErr
}
func (m *mockBreathingRepo) GetCompletionRate(_ context.Context, _ uint) (float64, error) {
	return m.completionRate, m.completionRateErr
}
func (m *mockBreathingRepo) CountSessionsSince(_ context.Context, _ uint, _ time.Time) (int64, error) {
	return m.countSince, m.countSinceErr
}

func (m *mockBreathingRepo) GetPreferences(_ context.Context, _ uint) (*model.BreathingPreference, error) {
	return m.prefs, nil
}
func (m *mockBreathingRepo) CreatePreferences(_ context.Context, _ *model.BreathingPreference) error {
	return nil
}
func (m *mockBreathingRepo) UpdatePreferences(_ context.Context, _ *model.BreathingPreference) error {
	return m.updatePrefsErr
}
func (m *mockBreathingRepo) GetOrCreatePreferences(_ context.Context, _ uint) (*model.BreathingPreference, error) {
	if m.getOrCreatePrefErr != nil {
		return nil, m.getOrCreatePrefErr
	}
	if m.prefs == nil {
		m.prefs = &model.BreathingPreference{}
	}
	return m.prefs, nil
}

func (m *mockBreathingRepo) GetFavorites(_ context.Context, _ uint) ([]model.BreathingFavorite, error) {
	return m.favorites, m.getFavoritesErr
}
func (m *mockBreathingRepo) AddFavorite(_ context.Context, _ *model.BreathingFavorite) error {
	m.addFavoriteCalls++
	return nil
}
func (m *mockBreathingRepo) RemoveFavorite(_ context.Context, _ uint, _ uuid.UUID) error {
	return m.removeFavoriteErr
}
func (m *mockBreathingRepo) IsFavorite(_ context.Context, _ uint, _ uuid.UUID) (bool, error) {
	return m.isFavorite, nil
}
func (m *mockBreathingRepo) UpdateFavoriteOrder(_ context.Context, _ uint, _ []uuid.UUID) error {
	return m.updateFavOrderErr
}

func (m *mockBreathingRepo) GetDailyStats(_ context.Context, _ uint, _ time.Time) (int, int, error) {
	return m.dailyCount, m.dailyMinutes, m.dailyStatsErr
}
func (m *mockBreathingRepo) GetMonthlyCalendar(_ context.Context, _ uint, _, _ int) ([]repository.DailyCalendarData, error) {
	return m.monthlyCalendar, m.monthlyCalendarErr
}
func (m *mockBreathingRepo) GetTechniqueUsageStats(_ context.Context, _ uint) ([]repository.TechniqueUsageData, error) {
	return m.usageStats, m.usageStatsErr
}

type mockDailyTaskSvc struct {
	calls int
	uid   uint
	typeV model.DailyTaskType
}

func (m *mockDailyTaskSvc) InitializeDailyTasks(_ context.Context, _ uint) error { return nil }
func (m *mockDailyTaskSvc) ProcessDailyLogin(_ context.Context, _ uint) (*DailyLoginResult, error) {
	return nil, nil
}
func (m *mockDailyTaskSvc) UpdateTaskProgress(_ context.Context, userID uint, taskType model.DailyTaskType) error {
	m.calls++
	m.uid = userID
	m.typeV = taskType
	return nil
}
func (m *mockDailyTaskSvc) GetTodayTasks(_ context.Context, _ uint) (*model.DailyTaskSummary, error) {
	return nil, nil
}
func (m *mockDailyTaskSvc) ClaimTaskReward(_ context.Context, _ uint, _ uint) (*ClaimResult, error) {
	return nil, nil
}
func (m *mockDailyTaskSvc) ClaimAllRewards(_ context.Context, _ uint) (*ClaimAllResult, error) {
	return nil, nil
}
func (m *mockDailyTaskSvc) GetTaskHistory(_ context.Context, _ uint, _, _ int) (*TaskHistoryResult, error) {
	return nil, nil
}

func TestBreathingService_CompleteSessionAppliesDailyCapAndUpdatesTask(t *testing.T) {
	sessionID := uuid.New()
	repo := &mockBreathingRepo{
		sessionByID: &model.BreathingSession{
			ID:        sessionID,
			UserID:    7,
			StartedAt: time.Now(),
		},
		prefs:   &model.BreathingPreference{},
		todayXP: 25,
	}
	dailyTask := &mockDailyTaskSvc{}
	svc := &breathingService{repo: repo, dailyTaskService: dailyTask}

	result, err := svc.CompleteSession(context.Background(), 7, sessionID, dto.CompleteBreathingSessionRequest{
		DurationSeconds:     300,
		CyclesCompleted:     8,
		Completed:           true,
		CompletedPercentage: 100,
		MoodAfter:           "calm",
	})
	if err != nil {
		t.Fatalf("unexpected complete session error: %v", err)
	}

	if result.TotalXP != 5 {
		t.Fatalf("expected capped total xp=5, got %d", result.TotalXP)
	}
	if repo.updateSessionCall != 1 || repo.updatedSession == nil {
		t.Fatalf("expected session update call, got calls=%d", repo.updateSessionCall)
	}
	if repo.updatedSession.XPEarned != 5 {
		t.Fatalf("expected session xp_earned=5, got %d", repo.updatedSession.XPEarned)
	}
	if dailyTask.calls != 1 || dailyTask.uid != 7 || dailyTask.typeV != model.TaskTypeBreathing {
		t.Fatalf("unexpected daily task update: calls=%d uid=%d type=%s", dailyTask.calls, dailyTask.uid, dailyTask.typeV)
	}
}

func TestBreathingService_CompleteSessionUnauthorized(t *testing.T) {
	repo := &mockBreathingRepo{
		sessionByID: &model.BreathingSession{ID: uuid.New(), UserID: 99, StartedAt: time.Now()},
	}
	svc := &breathingService{repo: repo}

	_, err := svc.CompleteSession(context.Background(), 7, uuid.New(), dto.CompleteBreathingSessionRequest{})
	if err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found, got %v", err)
	}
	if repo.updateSessionCall != 0 {
		t.Fatalf("update session must not be called on unauthorized, got %d", repo.updateSessionCall)
	}
}

func TestBreathingService_AddFavoriteFlow(t *testing.T) {
	tID := uuid.New()
	repo := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: tID}, isFavorite: false}
	svc := &breathingService{repo: repo}

	if err := svc.AddFavorite(context.Background(), 3, tID); err != nil {
		t.Fatalf("unexpected add favorite error: %v", err)
	}
	if repo.addFavoriteCalls != 1 {
		t.Fatalf("expected add favorite call, got %d", repo.addFavoriteCalls)
	}

	repo2 := &mockBreathingRepo{techniqueByIDErr: errors.New("not found")}
	svc2 := &breathingService{repo: repo2}
	if err := svc2.AddFavorite(context.Background(), 3, tID); err == nil || err.Error() != "technique not found" {
		t.Fatalf("expected technique not found, got %v", err)
	}
}

func TestBreathingService_CalculateXPBoundaries(t *testing.T) {
	svc := &breathingService{}
	cases := []struct {
		seconds int
		want    int
	}{
		{30, 0},
		{120, XPFor2Min},
		{300, XPFor5Min},
		{600, XPFor10Min},
		{900, XPFor15MinPlus},
	}

	for _, tc := range cases {
		got := svc.calculateXP(context.Background(), tc.seconds)
		if got != tc.want {
			t.Fatalf("for %d seconds expected %d xp, got %d", tc.seconds, tc.want, got)
		}
	}
}

func TestBreathingService_UpdateStreakBranches(t *testing.T) {
	repo := &mockBreathingRepo{prefs: &model.BreathingPreference{CurrentStreak: 0, LongestStreak: 0}}
	svc := &breathingService{repo: repo}

	prefs := &model.BreathingPreference{}
	streak := svc.updateStreak(context.Background(), prefs)
	if streak != 1 || prefs.LastPracticeDate == nil {
		t.Fatalf("expected first streak to initialize to 1, got %d", streak)
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	prefs2 := &model.BreathingPreference{CurrentStreak: 3, LongestStreak: 3, LastPracticeDate: &yesterday}
	streak = svc.updateStreak(context.Background(), prefs2)
	if streak != 4 || prefs2.LongestStreak != 4 {
		t.Fatalf("expected consecutive streak increment, got streak=%d longest=%d", streak, prefs2.LongestStreak)
	}

	lastWeek := time.Now().AddDate(0, 0, -7)
	prefs3 := &model.BreathingPreference{CurrentStreak: 5, LongestStreak: 9, LastPracticeDate: &lastWeek}
	streak = svc.updateStreak(context.Background(), prefs3)
	if streak != 1 || prefs3.LongestStreak != 9 {
		t.Fatalf("expected streak reset to 1 while keeping longest, got streak=%d longest=%d", streak, prefs3.LongestStreak)
	}
}

func TestBreathingService_GetAllTechniquesAndGetByIDSlug(t *testing.T) {
	systemID := uuid.New()
	customID := uuid.New()
	owner := 7
	repo := &mockBreathingRepo{
		systemTechniques: []model.BreathingTechnique{{ID: systemID, Name: "Box", IsSystem: true, IsActive: true}},
		userTechniques:   []model.BreathingTechnique{{ID: customID, Name: "Custom", IsSystem: false, IsActive: true, UserID: &owner}},
		favorites:        []model.BreathingFavorite{{TechniqueID: customID}},
		techniqueByID:    &model.BreathingTechnique{ID: customID, Name: "Custom", IsSystem: false, UserID: &owner},
		techniqueBySlug:  &model.BreathingTechnique{ID: systemID, Name: "Box", IsSystem: true},
		isFavorite:       true,
	}
	svc := &breathingService{repo: repo}

	all, err := svc.GetAllTechniques(context.Background(), 7)
	if err != nil {
		t.Fatalf("get all techniques failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 techniques, got %d", len(all))
	}

	byID, err := svc.GetTechniqueByID(context.Background(), 7, customID)
	if err != nil {
		t.Fatalf("get technique by id failed: %v", err)
	}
	if byID.ID != customID {
		t.Fatalf("expected technique %s, got %s", customID, byID.ID)
	}

	bySlug, err := svc.GetTechniqueBySlug(context.Background(), 7, "box")
	if err != nil {
		t.Fatalf("get technique by slug failed: %v", err)
	}
	if bySlug.ID != systemID {
		t.Fatalf("expected technique %s, got %s", systemID, bySlug.ID)
	}
}

func TestBreathingService_GetTechniqueAccessDenied(t *testing.T) {
	owner := 99
	techniqueID := uuid.New()
	repo := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &owner}}
	svc := &breathingService{repo: repo}

	if _, err := svc.GetTechniqueByID(context.Background(), 7, techniqueID); err == nil || err.Error() != "technique not found" {
		t.Fatalf("expected technique not found, got %v", err)
	}
}

func TestBreathingService_CustomTechniqueCRUD(t *testing.T) {
	owner := 7
	techniqueID := uuid.New()
	repo := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &owner}}
	svc := &breathingService{repo: repo}

	created, err := svc.CreateCustomTechnique(context.Background(), 7, dto.CreateBreathingTechniqueRequest{
		Name:               "Wind Down",
		Description:        "desc",
		InhaleDuration:     4,
		InhaleHoldDuration: 2,
		ExhaleDuration:     6,
		ExhaleHoldDuration: 2,
	})
	if err != nil {
		t.Fatalf("create custom technique failed: %v", err)
	}
	if created.Name != "Wind Down" || repo.createTechniqueCall != 1 {
		t.Fatalf("unexpected create result/calls: name=%s calls=%d", created.Name, repo.createTechniqueCall)
	}

	newName := "Wind Down Updated"
	newColor := "#000000"
	updated, err := svc.UpdateCustomTechnique(context.Background(), 7, techniqueID, dto.UpdateBreathingTechniqueRequest{
		Name:  &newName,
		Color: &newColor,
	})
	if err != nil {
		t.Fatalf("update custom technique failed: %v", err)
	}
	if updated.Name != newName || repo.updateTechniqueCall != 1 {
		t.Fatalf("unexpected update result/calls: name=%s calls=%d", updated.Name, repo.updateTechniqueCall)
	}

	if err := svc.DeleteCustomTechnique(context.Background(), 7, techniqueID); err != nil {
		t.Fatalf("delete custom technique failed: %v", err)
	}
	if repo.deleteTechniqueCall != 1 || repo.deletedTechniqueID != techniqueID {
		t.Fatalf("unexpected delete call: calls=%d id=%s", repo.deleteTechniqueCall, repo.deletedTechniqueID)
	}
}

func TestBreathingService_CustomTechniqueCRUDFailures(t *testing.T) {
	owner := 99
	techniqueID := uuid.New()
	repo := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &owner}}
	svc := &breathingService{repo: repo}

	if _, err := svc.UpdateCustomTechnique(context.Background(), 7, techniqueID, dto.UpdateBreathingTechniqueRequest{}); err == nil || err.Error() != "cannot update this technique" {
		t.Fatalf("expected cannot update this technique, got %v", err)
	}

	if err := svc.DeleteCustomTechnique(context.Background(), 7, techniqueID); err == nil || err.Error() != "cannot delete this technique" {
		t.Fatalf("expected cannot delete this technique, got %v", err)
	}
}

func TestBreathingService_StartSession(t *testing.T) {
	techniqueID := uuid.New()
	owner := 7
	repo := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &owner}}
	svc := &breathingService{repo: repo}

	resp, err := svc.StartSession(context.Background(), 7, dto.StartBreathingSessionRequest{
		TechniqueID:           techniqueID,
		TargetDurationSeconds: 300,
		VoiceGuidanceEnabled:  true,
		BackgroundSound:       "rain",
		HapticFeedbackEnabled: true,
		MoodBefore:            "anxious",
	})
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	if repo.createSessionCall != 1 || repo.createdSession == nil {
		t.Fatalf("expected create session called once, got %d", repo.createSessionCall)
	}
	if resp.TechniqueID != techniqueID {
		t.Fatalf("expected technique id %s, got %s", techniqueID, resp.TechniqueID)
	}

	repoErr := &mockBreathingRepo{techniqueByIDErr: errors.New("not found")}
	svcErr := &breathingService{repo: repoErr}
	if _, err := svcErr.StartSession(context.Background(), 7, dto.StartBreathingSessionRequest{TechniqueID: techniqueID}); err == nil || err.Error() != "technique not found" {
		t.Fatalf("expected technique not found, got %v", err)
	}
}

func TestBreathingService_HistoryAndPreferencesAndFavorites(t *testing.T) {
	techniqueID := uuid.New()
	now := time.Now()
	repo := &mockBreathingRepo{
		sessionByID: &model.BreathingSession{
			ID:          uuid.New(),
			UserID:      7,
			TechniqueID: techniqueID,
			Technique:   &model.BreathingTechnique{ID: techniqueID, Name: "Box", InhaleDuration: 4, ExhaleDuration: 4},
			StartedAt:   now,
		},
		userSessions: []model.BreathingSession{{
			ID:          uuid.New(),
			UserID:      7,
			TechniqueID: techniqueID,
			Technique:   &model.BreathingTechnique{ID: techniqueID, Name: "Box", InhaleDuration: 4, ExhaleDuration: 4},
			StartedAt:   now,
		}},
		userSessionsTotal: 1,
		prefs: &model.BreathingPreference{
			UserID:                 7,
			DefaultDurationSeconds: 300,
			VoiceGuidance:          "ask",
			BackgroundSound:        "ask",
			DefaultBackgroundSound: "none",
			HapticFeedback:         true,
			AnimationSpeed:         "normal",
			Theme:                  "default",
		},
		favorites: []model.BreathingFavorite{{TechniqueID: techniqueID, Technique: &model.BreathingTechnique{ID: techniqueID, Name: "Box", InhaleDuration: 4, ExhaleDuration: 4}}},
	}
	svc := &breathingService{repo: repo}

	history, err := svc.GetSessionHistory(context.Background(), 7, dto.SessionHistoryRequest{Page: 1, Limit: 5, StartDate: "bad", EndDate: "2026-02-20", TechniqueID: "bad-uuid"})
	if err != nil {
		t.Fatalf("get session history failed: %v", err)
	}
	if history.Total != 1 || len(history.Sessions) != 1 {
		t.Fatalf("unexpected history totals: total=%d len=%d", history.Total, len(history.Sessions))
	}

	sessionID := repo.sessionByID.ID
	byID, err := svc.GetSessionByID(context.Background(), 7, sessionID)
	if err != nil || byID.ID != sessionID {
		t.Fatalf("get session by id failed: err=%v", err)
	}
	if _, err := svc.GetSessionByID(context.Background(), 99, sessionID); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found for other user, got %v", err)
	}

	prefsResp, err := svc.GetPreferences(context.Background(), 7)
	if err != nil || prefsResp.Theme != "default" {
		t.Fatalf("get preferences failed: err=%v theme=%s", err, prefsResp.Theme)
	}

	theme := "dark"
	voice := "on"
	updatedPrefs, err := svc.UpdatePreferences(context.Background(), 7, dto.UpdateBreathingPreferencesRequest{Theme: &theme, VoiceGuidance: &voice})
	if err != nil || updatedPrefs.Theme != "dark" || updatedPrefs.VoiceGuidance != "on" {
		t.Fatalf("update preferences failed: err=%v prefs=%+v", err, updatedPrefs)
	}

	favorites, err := svc.GetFavorites(context.Background(), 7)
	if err != nil || len(favorites) != 1 {
		t.Fatalf("get favorites failed: err=%v len=%d", err, len(favorites))
	}

	if err := svc.RemoveFavorite(context.Background(), 7, techniqueID); err != nil {
		t.Fatalf("remove favorite failed: %v", err)
	}
	if err := svc.ReorderFavorites(context.Background(), 7, []uuid.UUID{techniqueID}); err != nil {
		t.Fatalf("reorder favorites failed: %v", err)
	}
}

func TestBreathingService_StatsCalendarUsageWidgetRecommendations(t *testing.T) {
	techBoxID := uuid.New()
	techEnergyID := uuid.New()
	box := "box-breathing"
	energizing := "energizing-breath"
	endedAt := time.Now().Add(-10 * time.Minute)
	yesterday := time.Now().AddDate(0, 0, -1)

	repo := &mockBreathingRepo{
		prefs:             &model.BreathingPreference{CurrentStreak: 3, LongestStreak: 5, LastPracticeDate: &yesterday},
		totalSessions:     12,
		totalMinutes:      60,
		completionRate:    75,
		mostUsedTechnique: &model.BreathingTechnique{ID: techBoxID, Name: "Box"},
		mostUsedCount:     8,
		dailyCount:        2,
		dailyMinutes:      12,
		favorites:         []model.BreathingFavorite{{TechniqueID: techBoxID, Technique: &model.BreathingTechnique{ID: techBoxID, Name: "Box", Slug: &box, InhaleDuration: 4, ExhaleDuration: 4}}},
		userSessions:      []model.BreathingSession{{EndedAt: &endedAt}},
		monthlyCalendar:   []repository.DailyCalendarData{{Date: time.Now(), SessionsCount: 1, TotalMinutes: 5}, {Date: time.Now(), SessionsCount: 2, TotalMinutes: 10}, {Date: time.Now(), SessionsCount: 4, TotalMinutes: 20}},
		usageStats:        []repository.TechniqueUsageData{{TechniqueID: techBoxID, TechniqueName: "Box", SessionsCount: 4, TotalMinutes: 20}},
		systemTechniques: []model.BreathingTechnique{
			{ID: techBoxID, Name: "Box", Slug: &box, InhaleDuration: 4, ExhaleDuration: 4},
			{ID: techEnergyID, Name: "Energizing", Slug: &energizing, InhaleDuration: 4, ExhaleDuration: 4},
		},
	}
	svc := &breathingService{repo: repo}

	stats, err := svc.GetStats(context.Background(), 7)
	if err != nil || stats.Overall.TotalSessions != 12 || stats.Today.SessionsCount != 2 {
		t.Fatalf("get stats failed: err=%v stats=%+v", err, stats)
	}

	calendar, err := svc.GetCalendar(context.Background(), 7, 2026, 2)
	if err != nil || len(calendar.Days) != 3 || calendar.Days[2].Intensity != 3 {
		t.Fatalf("get calendar failed: err=%v calendar=%+v", err, calendar)
	}

	usage, err := svc.GetTechniqueUsage(context.Background(), 7)
	if err != nil || len(usage) != 1 || usage[0].TechniqueID == "" {
		t.Fatalf("get technique usage failed: err=%v usage=%+v", err, usage)
	}

	widget, err := svc.GetWidgetData(context.Background(), 7)
	if err != nil || widget.DailyGoalProgress != 100 || widget.FavoriteTechnique == nil || widget.LastSessionAt == nil {
		t.Fatalf("get widget data failed: err=%v widget=%+v", err, widget)
	}

	reco, err := svc.GetRecommendations(context.Background(), 7, "stressed", "morning")
	if err != nil || reco.DefaultPick == nil || len(reco.BasedOnMood) == 0 {
		t.Fatalf("get recommendations failed: err=%v reco=%+v", err, reco)
	}
}

func TestNewBreathingService_Constructor(t *testing.T) {
	repo := &mockBreathingRepo{}
	dailyTask := &mockDailyTaskSvc{}
	if svc := NewBreathingService(repo, nil, dailyTask); svc == nil {
		t.Fatal("expected non-nil breathing service")
	}
}

func TestBreathingService_GetRecommendations_Branches(t *testing.T) {
	box := "box-breathing"
	relax := "4-7-8-relaxing"
	energy := "energizing-breath"
	deep := "deep-calm"
	coherent := "coherent-breathing"

	boxID := uuid.New()
	relaxID := uuid.New()
	energyID := uuid.New()
	deepID := uuid.New()
	coherentID := uuid.New()

	repo := &mockBreathingRepo{
		systemTechniques: []model.BreathingTechnique{
			{ID: boxID, Name: "Box", Slug: &box, InhaleDuration: 4, ExhaleDuration: 4},
			{ID: relaxID, Name: "Relax", Slug: &relax, InhaleDuration: 4, ExhaleDuration: 8},
			{ID: energyID, Name: "Energy", Slug: &energy, InhaleDuration: 4, ExhaleDuration: 4},
			{ID: deepID, Name: "Deep", Slug: &deep, InhaleDuration: 5, ExhaleDuration: 5},
			{ID: coherentID, Name: "Coherent", Slug: &coherent, InhaleDuration: 5, ExhaleDuration: 5},
		},
		favorites: []model.BreathingFavorite{{TechniqueID: boxID}},
	}
	svc := &breathingService{repo: repo}

	reco1, err := svc.GetRecommendations(context.Background(), 1, "anxious", "night")
	if err != nil {
		t.Fatalf("unexpected recommendations error: %v", err)
	}
	if reco1.DefaultPick == nil || len(reco1.BasedOnMood) == 0 || len(reco1.BasedOnTime) == 0 {
		t.Fatalf("expected mood+time recommendations, got %+v", reco1)
	}

	reco2, err := svc.GetRecommendations(context.Background(), 1, "tired", "morning")
	if err != nil {
		t.Fatalf("unexpected tired recommendations error: %v", err)
	}
	if reco2.DefaultPick == nil || reco2.DefaultPick.Technique.Slug != energy {
		t.Fatalf("expected energizing default for tired morning, got %+v", reco2.DefaultPick)
	}

	reco3, err := svc.GetRecommendations(context.Background(), 1, "angry", "evening")
	if err != nil {
		t.Fatalf("unexpected angry recommendations error: %v", err)
	}
	if reco3.DefaultPick == nil {
		t.Fatalf("expected default pick for angry evening")
	}

	reco4, err := svc.GetRecommendations(context.Background(), 1, "neutral", "afternoon")
	if err != nil {
		t.Fatalf("unexpected neutral recommendations error: %v", err)
	}
	if reco4.DefaultPick == nil {
		t.Fatalf("expected default pick for neutral afternoon")
	}

	reco5, err := svc.GetRecommendations(context.Background(), 1, "unknown-mood", "afternoon")
	if err != nil {
		t.Fatalf("unexpected unknown mood recommendations error: %v", err)
	}
	if reco5.DefaultPick == nil {
		t.Fatalf("expected default pick for unknown mood")
	}

	repo.systemTechniques = []model.BreathingTechnique{}
	reco6, err := svc.GetRecommendations(context.Background(), 1, "stressed", "morning")
	if err != nil {
		t.Fatalf("unexpected empty techniques recommendations error: %v", err)
	}
	if reco6.DefaultPick != nil {
		t.Fatalf("expected nil default pick when no techniques, got %+v", reco6.DefaultPick)
	}

	repo.systemTechniquesErr = errors.New("db fail")
	if _, err := svc.GetRecommendations(context.Background(), 1, "stressed", "morning"); err == nil {
		t.Fatal("expected error when GetSystemTechniques fails")
	}
}

func TestBreathingService_UpdatePreferences_AllFieldsAndErrors(t *testing.T) {
	tID := uuid.New()
	reminderTime := "21:30"
	defaultDuration := 420
	voice := "always"
	bg := "rain"
	defaultBg := "forest"
	haptic := true
	anim := "slow"
	theme := "night"
	reminderEnabled := true
	reminderDays := "135"

	repo := &mockBreathingRepo{
		prefs: &model.BreathingPreference{
			UserID:                 9,
			DefaultDurationSeconds: 300,
			VoiceGuidance:          "ask",
			BackgroundSound:        "none",
			DefaultBackgroundSound: "none",
			HapticFeedback:         false,
			AnimationSpeed:         "normal",
			Theme:                  "default",
			ReminderEnabled:        false,
			ReminderDays:           "2",
		},
	}
	svc := &breathingService{repo: repo}

	updated, err := svc.UpdatePreferences(context.Background(), 9, dto.UpdateBreathingPreferencesRequest{
		DefaultDurationSeconds: &defaultDuration,
		DefaultTechniqueID:     &tID,
		VoiceGuidance:          &voice,
		BackgroundSound:        &bg,
		DefaultBackgroundSound: &defaultBg,
		HapticFeedback:         &haptic,
		AnimationSpeed:         &anim,
		Theme:                  &theme,
		ReminderEnabled:        &reminderEnabled,
		ReminderTime:           &reminderTime,
		ReminderDays:           &reminderDays,
	})
	if err != nil {
		t.Fatalf("unexpected update preferences error: %v", err)
	}
	if updated.DefaultDurationSeconds != 420 || updated.Theme != "night" || updated.VoiceGuidance != "always" {
		t.Fatalf("unexpected updated preferences response: %+v", updated)
	}
	if updated.DefaultTechniqueID == nil || *updated.DefaultTechniqueID != tID {
		t.Fatalf("expected default technique id set, got %+v", updated.DefaultTechniqueID)
	}
	if updated.ReminderTime != reminderTime || updated.ReminderDays != reminderDays {
		t.Fatalf("expected reminder fields updated, got time=%s days=%v", updated.ReminderTime, updated.ReminderDays)
	}

	repo.getOrCreatePrefErr = errors.New("pref load fail")
	if _, err := svc.UpdatePreferences(context.Background(), 9, dto.UpdateBreathingPreferencesRequest{}); err == nil {
		t.Fatal("expected update preferences to fail when loading prefs fails")
	}

	repo.getOrCreatePrefErr = nil
	repo.updatePrefsErr = errors.New("pref save fail")
	if _, err := svc.UpdatePreferences(context.Background(), 9, dto.UpdateBreathingPreferencesRequest{Theme: &theme}); err == nil {
		t.Fatal("expected update preferences to fail when save fails")
	}
}

func TestBreathingService_UpdateCustomTechnique_AllFieldsAndErrors(t *testing.T) {
	techniqueID := uuid.New()
	owner := 7
	desc := "updated desc"
	name := "new name"
	inhale := 5
	inhaleHold := 2
	exhale := 7
	exhaleHold := 3
	icon := "🌀"
	color := "#123456"

	repo := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, Name: "Old", IsSystem: false, UserID: &owner}}
	svc := &breathingService{repo: repo}

	resp, err := svc.UpdateCustomTechnique(context.Background(), 7, techniqueID, dto.UpdateBreathingTechniqueRequest{
		Name:               &name,
		Description:        &desc,
		InhaleDuration:     &inhale,
		InhaleHoldDuration: &inhaleHold,
		ExhaleDuration:     &exhale,
		ExhaleHoldDuration: &exhaleHold,
		Icon:               &icon,
		Color:              &color,
	})
	if err != nil {
		t.Fatalf("unexpected update custom technique error: %v", err)
	}
	if resp.Name != name || resp.Icon != icon || resp.Color != color || resp.InhaleDuration != inhale || resp.ExhaleDuration != exhale {
		t.Fatalf("unexpected updated technique response: %+v", resp)
	}

	repo.techniqueByIDErr = errors.New("not found")
	if _, err := svc.UpdateCustomTechnique(context.Background(), 7, techniqueID, dto.UpdateBreathingTechniqueRequest{}); err == nil {
		t.Fatal("expected update custom technique to fail when technique not found")
	}

	repo.techniqueByIDErr = nil
	repo.updateTechniqueErr = errors.New("save failed")
	if _, err := svc.UpdateCustomTechnique(context.Background(), 7, techniqueID, dto.UpdateBreathingTechniqueRequest{Name: &name}); err == nil {
		t.Fatal("expected update custom technique to fail when save fails")
	}
}

func TestBreathingService_CompleteSession_MilestonesAndErrors(t *testing.T) {
	sessionID := uuid.New()
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	repo7 := &mockBreathingRepo{
		sessionByID: &model.BreathingSession{ID: sessionID, UserID: 11, StartedAt: now},
		prefs:       &model.BreathingPreference{CurrentStreak: 6, LastPracticeDate: &yesterday, LongestStreak: 6},
	}
	svc7 := &breathingService{repo: repo7}
	res7, err := svc7.CompleteSession(context.Background(), 11, sessionID, dto.CompleteBreathingSessionRequest{
		DurationSeconds:     600,
		CyclesCompleted:     10,
		Completed:           false,
		CompletedPercentage: 80,
		MoodAfter:           "calm",
	})
	if err != nil {
		t.Fatalf("unexpected complete session error for 7-day milestone: %v", err)
	}
	if !res7.StreakMilestone || res7.StreakMilestoneXP != StreakBonusWeek || res7.BonusReason == "" {
		t.Fatalf("expected 7-day milestone bonus, got %+v", res7)
	}

	repo30 := &mockBreathingRepo{
		sessionByID: &model.BreathingSession{ID: sessionID, UserID: 12, StartedAt: now},
		prefs:       &model.BreathingPreference{CurrentStreak: 29, LastPracticeDate: &yesterday, LongestStreak: 29},
	}
	svc30 := &breathingService{repo: repo30}
	res30, err := svc30.CompleteSession(context.Background(), 12, sessionID, dto.CompleteBreathingSessionRequest{DurationSeconds: 900, Completed: true, CompletedPercentage: 100})
	if err != nil {
		t.Fatalf("unexpected complete session error for 30-day milestone: %v", err)
	}
	if !res30.StreakMilestone || res30.StreakMilestoneXP != StreakBonusMonth || res30.BonusReason == "" {
		t.Fatalf("expected 30-day milestone bonus, got %+v", res30)
	}

	repoNoCap := &mockBreathingRepo{
		sessionByID: &model.BreathingSession{ID: sessionID, UserID: 13, StartedAt: now},
		prefs:       &model.BreathingPreference{},
		todayXP:     DailyXPCap,
	}
	svcNoCap := &breathingService{repo: repoNoCap}
	resNoCap, err := svcNoCap.CompleteSession(context.Background(), 13, sessionID, dto.CompleteBreathingSessionRequest{DurationSeconds: 600, Completed: true, CompletedPercentage: 100})
	if err != nil {
		t.Fatalf("unexpected complete session error for no-cap path: %v", err)
	}
	if resNoCap.TotalXP != 0 {
		t.Fatalf("expected total xp 0 when daily cap exhausted, got %d", resNoCap.TotalXP)
	}

	repoErrGet := &mockBreathingRepo{sessionByIDErr: errors.New("not found")}
	svcErrGet := &breathingService{repo: repoErrGet}
	if _, err := svcErrGet.CompleteSession(context.Background(), 7, sessionID, dto.CompleteBreathingSessionRequest{}); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found on load error, got %v", err)
	}

	repoErrUpdate := &mockBreathingRepo{
		sessionByID:      &model.BreathingSession{ID: sessionID, UserID: 14, StartedAt: now},
		prefs:            &model.BreathingPreference{},
		updateSessionErr: errors.New("update fail"),
	}
	svcErrUpdate := &breathingService{repo: repoErrUpdate}
	if _, err := svcErrUpdate.CompleteSession(context.Background(), 14, sessionID, dto.CompleteBreathingSessionRequest{DurationSeconds: 300}); err == nil {
		t.Fatal("expected complete session to fail when update session fails")
	}
}

func TestBreathingService_GetFavorites_Error(t *testing.T) {
	repo := &mockBreathingRepo{getFavoritesErr: errors.New("favorite list fail")}
	svc := &breathingService{repo: repo}
	if _, err := svc.GetFavorites(context.Background(), 5); err == nil {
		t.Fatal("expected get favorites to fail")
	}
}

func TestBreathingService_AdditionalErrorAndEdgeBranches(t *testing.T) {
	ctx := context.Background()
	techniqueID := uuid.New()
	otherUser := 99

	repoAllErr := &mockBreathingRepo{systemTechniquesErr: errors.New("system fail")}
	svcAllErr := &breathingService{repo: repoAllErr}
	if _, err := svcAllErr.GetAllTechniques(ctx, 1); err == nil {
		t.Fatal("expected get all techniques system error")
	}

	repoUserErr := &mockBreathingRepo{systemTechniques: []model.BreathingTechnique{{ID: techniqueID, IsSystem: true}}, userTechniquesErr: errors.New("user techniques fail")}
	svcUserErr := &breathingService{repo: repoUserErr}
	if _, err := svcUserErr.GetAllTechniques(ctx, 1); err == nil {
		t.Fatal("expected get all techniques user error")
	}

	repoSlugDenied := &mockBreathingRepo{techniqueBySlug: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &otherUser}}
	svcSlugDenied := &breathingService{repo: repoSlugDenied}
	if _, err := svcSlugDenied.GetTechniqueBySlug(ctx, 1, "custom"); err == nil || err.Error() != "technique not found" {
		t.Fatalf("expected technique not found from slug access check, got %v", err)
	}

	repoCreateErr := &mockBreathingRepo{createTechniqueErr: errors.New("create fail")}
	svcCreateErr := &breathingService{repo: repoCreateErr}
	if _, err := svcCreateErr.CreateCustomTechnique(ctx, 1, dto.CreateBreathingTechniqueRequest{Name: "N", InhaleDuration: 4, ExhaleDuration: 4}); err == nil {
		t.Fatal("expected create custom technique error")
	}

	repoDeleteGetErr := &mockBreathingRepo{techniqueByIDErr: errors.New("missing")}
	svcDeleteGetErr := &breathingService{repo: repoDeleteGetErr}
	if err := svcDeleteGetErr.DeleteCustomTechnique(ctx, 1, techniqueID); err == nil {
		t.Fatal("expected delete custom technique get error")
	}

	owner := 1
	repoDeleteErr := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &owner}, deleteTechniqueErr: errors.New("delete fail")}
	svcDeleteErr := &breathingService{repo: repoDeleteErr}
	if err := svcDeleteErr.DeleteCustomTechnique(ctx, 1, techniqueID); err == nil {
		t.Fatal("expected delete custom technique repo error")
	}

	repoStartDenied := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: false, UserID: &otherUser}}
	svcStartDenied := &breathingService{repo: repoStartDenied}
	if _, err := svcStartDenied.StartSession(ctx, 1, dto.StartBreathingSessionRequest{TechniqueID: techniqueID}); err == nil || err.Error() != "technique not found" {
		t.Fatalf("expected technique not found for start session access check, got %v", err)
	}

	repoStartCreateErr := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: true}, createSessionErr: errors.New("create session fail")}
	svcStartCreateErr := &breathingService{repo: repoStartCreateErr}
	if _, err := svcStartCreateErr.StartSession(ctx, 1, dto.StartBreathingSessionRequest{TechniqueID: techniqueID}); err == nil {
		t.Fatal("expected start session create error")
	}

	repoSessionHistoryErr := &mockBreathingRepo{getUserSessionsErr: errors.New("sessions fail")}
	svcSessionHistoryErr := &breathingService{repo: repoSessionHistoryErr}
	if _, err := svcSessionHistoryErr.GetSessionHistory(ctx, 1, dto.SessionHistoryRequest{Page: 1, Limit: 10}); err == nil {
		t.Fatal("expected session history error")
	}

	repoGetSessionErr := &mockBreathingRepo{sessionByIDErr: errors.New("session fail")}
	svcGetSessionErr := &breathingService{repo: repoGetSessionErr}
	if _, err := svcGetSessionErr.GetSessionByID(ctx, 1, techniqueID); err == nil {
		t.Fatal("expected get session by id error")
	}

	repoPrefErr := &mockBreathingRepo{getOrCreatePrefErr: errors.New("pref fail")}
	svcPrefErr := &breathingService{repo: repoPrefErr}
	if _, err := svcPrefErr.GetPreferences(ctx, 1); err == nil {
		t.Fatal("expected get preferences error")
	}

	repoFavExists := &mockBreathingRepo{techniqueByID: &model.BreathingTechnique{ID: techniqueID, IsSystem: true}, isFavorite: true}
	svcFavExists := &breathingService{repo: repoFavExists}
	if err := svcFavExists.AddFavorite(ctx, 1, techniqueID); err != nil {
		t.Fatalf("expected add favorite no-op success when already favorite, got %v", err)
	}
	if repoFavExists.addFavoriteCalls != 0 {
		t.Fatalf("expected no add favorite call when already favorite, got %d", repoFavExists.addFavoriteCalls)
	}

	repoUsageErr := &mockBreathingRepo{usageStatsErr: errors.New("usage fail")}
	svcUsageErr := &breathingService{repo: repoUsageErr}
	if _, err := svcUsageErr.GetTechniqueUsage(ctx, 1); err == nil {
		t.Fatal("expected get technique usage error")
	}

	repoCalendarErr := &mockBreathingRepo{monthlyCalendarErr: errors.New("calendar fail")}
	svcCalendarErr := &breathingService{repo: repoCalendarErr}
	if _, err := svcCalendarErr.GetCalendar(ctx, 1, 2026, 2); err == nil {
		t.Fatal("expected get calendar error")
	}
}
