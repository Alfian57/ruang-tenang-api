package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/features/wellness/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
)

var (
	ErrWellnessOnboardingRequired = errors.New("wellness onboarding required")
	ErrUnsupportedNeedCondition   = errors.New("unsupported need condition")
)

type WellnessService struct {
	repo        *infrastructure.WellnessRepository
	genaiClient *genai.Client
}

func NewWellnessService(repo *infrastructure.WellnessRepository, genaiClient *genai.Client) *WellnessService {
	return &WellnessService{repo: repo, genaiClient: genaiClient}
}

func (s *WellnessService) GetOnboarding(ctx context.Context, userID uint) (*dto.WellnessOnboardingResponse, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrWellnessProfileNotFound) {
			return &dto.WellnessOnboardingResponse{NeedsOnboarding: true}, nil
		}
		return nil, err
	}

	plan, planErr := s.getOrCreatePlan(ctx, userID, profile)
	if planErr != nil {
		return nil, planErr
	}

	return &dto.WellnessOnboardingResponse{
		NeedsOnboarding: profile.OnboardingCompletedAt == nil,
		Profile:         toProfileDTO(profile),
		Plan:            toPlanDTO(plan),
	}, nil
}

func (s *WellnessService) CompleteOnboarding(ctx context.Context, userID uint, req *dto.WellnessOnboardingRequest) (*dto.WellnessOnboardingResponse, error) {
	now := time.Now()
	profile := &model.UserWellnessProfile{
		UserID:                userID,
		InitialMood:           strings.TrimSpace(strings.ToLower(req.InitialMood)),
		GoalsJSON:             marshalStringSlice(normalizeStringSlice(req.Goals, 5)),
		HabitsJSON:            marshalStringSlice(normalizeStringSlice(req.Habits, 8)),
		OnboardingCompletedAt: &now,
	}
	if err := s.repo.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	stored, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ArchiveActivePlans(ctx, userID); err != nil {
		return nil, err
	}
	plan, err := s.createSevenDayPlan(ctx, userID, stored)
	if err != nil {
		return nil, err
	}

	return &dto.WellnessOnboardingResponse{
		NeedsOnboarding: false,
		Profile:         toProfileDTO(stored),
		Plan:            toPlanDTO(plan),
	}, nil
}

func (s *WellnessService) GetCurrentPlan(ctx context.Context, userID uint) (*dto.WellnessOnboardingResponse, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrWellnessProfileNotFound) {
			return &dto.WellnessOnboardingResponse{NeedsOnboarding: true}, nil
		}
		return nil, err
	}
	plan, err := s.getOrCreatePlan(ctx, userID, profile)
	if err != nil {
		return nil, err
	}
	return &dto.WellnessOnboardingResponse{
		NeedsOnboarding: profile.OnboardingCompletedAt == nil,
		Profile:         toProfileDTO(profile),
		Plan:            toPlanDTO(plan),
	}, nil
}

func (s *WellnessService) CompletePlanItem(ctx context.Context, userID uint, itemID uuid.UUID) (*dto.WellnessPlanItemDTO, error) {
	item, err := s.repo.GetPlanItemForUser(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	item.Status = model.WellnessPlanItemStatusCompleted
	item.CompletedAt = &now
	if err := s.repo.SavePlanItem(ctx, item); err != nil {
		return nil, err
	}
	dtoItem := toPlanItemDTO(*item)
	return &dtoItem, nil
}

func (s *WellnessService) NeedNow(ctx context.Context, userID uint, condition string) (*dto.WellnessNeedNowResponse, error) {
	isPremium, err := s.repo.HasPremiumAccess(ctx, userID, time.Now())
	if err != nil {
		return nil, err
	}
	response, ok := buildNeedNowResponse(strings.TrimSpace(strings.ToLower(condition)), isPremium)
	if !ok {
		return nil, ErrUnsupportedNeedCondition
	}
	event := &model.WellnessNeedEvent{
		UserID:              userID,
		Condition:           response.Condition,
		RecommendationsJSON: marshalRecommendations(response.Recommendations),
	}
	if err := s.repo.CreateNeedEvent(ctx, event); err != nil {
		return nil, err
	}
	return response, nil
}

func (s *WellnessService) GetWeeklyInsight(ctx context.Context, userID uint, weekStartRaw string) (*dto.WeeklyInsightResponse, error) {
	weekStart, weekEnd := resolveWeekRange(weekStartRaw)
	isPremium, err := s.repo.HasPremiumAccess(ctx, userID, time.Now())
	if err != nil {
		return nil, err
	}

	snapshot, err := s.repo.GetWeeklyInsight(ctx, userID, weekStart)
	if err != nil && !errors.Is(err, infrastructure.ErrWeeklyInsightNotFound) {
		return nil, err
	}
	if snapshot == nil {
		snapshot, err = s.generateWeeklyInsight(ctx, userID, weekStart, weekEnd)
		if err != nil {
			return nil, err
		}
	}

	return s.weeklySnapshotToResponse(snapshot, isPremium), nil
}

func (s *WellnessService) CompleteTour(ctx context.Context, userID uint) (*dto.WellnessTourCompleteResponse, error) {
	completedAt := time.Now()
	if err := s.repo.MarkTourCompleted(ctx, userID, completedAt); err != nil {
		return nil, err
	}
	return &dto.WellnessTourCompleteResponse{TourCompletedAt: completedAt}, nil
}

func (s *WellnessService) GetJourneyMap(ctx context.Context, userID uint) (*dto.WellnessJourneyMapResponse, error) {
	since := time.Now().AddDate(0, 0, -14)
	signals, err := s.repo.CountJourneySignals(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	streak, _ := s.repo.GetUserStreak(ctx, userID)

	nodes := []dto.WellnessJourneyNodeDTO{
		buildJourneyNode("mood", "Mengenali Rasa", "Mood check-in membentuk titik awal perjalananmu.", signals["mood"], 7, "/dashboard/mood-tracker", "rose"),
		buildJourneyNode("journal", "Merapikan Pikiran", "Jurnal menangkap konteks dan pemicu yang sering muncul.", signals["journal"], 4, "/dashboard/journal", "sky"),
		buildJourneyNode("breathing", "Menata Napas", "Sesi napas memberi jeda saat tubuh mulai tegang.", signals["breathing"], 4, "/dashboard/breathing", "emerald"),
		buildJourneyNode("chat", "Mencari Arah", "Chat AI membantu mengubah cerita menjadi langkah kecil.", signals["chat"], 3, "/dashboard/chat", "violet"),
		buildJourneyNode("reward", "Merayakan Progres", "Reward dan landmark menjaga perjalanan terasa hidup.", signals["reward"]+signals["landmarks"], 3, "/dashboard/rewards", "amber"),
	}

	total := 0.0
	weakest := nodes[0]
	for _, node := range nodes {
		total += node.Progress
		if node.Progress < weakest.Progress {
			weakest = node
		}
	}
	overall := int(total / float64(len(nodes)))
	narrative := "Perjalanan tenangmu mulai terbentuk dari kebiasaan kecil yang saling terhubung."
	if overall >= 80 {
		narrative = "Perjalananmu tampak stabil: refleksi, regulasi napas, dan progres mulai bergerak bersama."
	} else if overall >= 45 {
		narrative = "Ada momentum yang sudah terlihat. Menguatkan satu area lemah akan membuat perjalanan terasa lebih utuh."
	}

	return &dto.WellnessJourneyMapResponse{
		Title:           "Peta Perjalanan Tenang",
		Narrative:       narrative,
		OverallProgress: overall,
		Streak:          streak,
		Nodes:           nodes,
		NextRecommendation: dto.WellnessRecommendationDTO{
			Type:        weakest.Key,
			Title:       "Lanjutkan titik berikutnya",
			Description: fmt.Sprintf("Fokus kecil hari ini: %s.", strings.ToLower(weakest.Label)),
			Route:       weakest.Route,
			Locked:      false,
		},
	}, nil
}

func (s *WellnessService) getOrCreatePlan(ctx context.Context, userID uint, profile *model.UserWellnessProfile) (*model.WellnessPlan, error) {
	plan, err := s.repo.GetActivePlan(ctx, userID)
	if err == nil {
		return plan, nil
	}
	if !errors.Is(err, infrastructure.ErrWellnessPlanNotFound) {
		return nil, err
	}
	return s.createSevenDayPlan(ctx, userID, profile)
}

func (s *WellnessService) createSevenDayPlan(ctx context.Context, userID uint, profile *model.UserWellnessProfile) (*model.WellnessPlan, error) {
	start := dayStart(time.Now())
	goals := parseStringSlice(profile.GoalsJSON)
	mainGoal := "membangun ritme tenang"
	if len(goals) > 0 {
		mainGoal = goals[0]
	}

	plan := &model.WellnessPlan{
		UserID:            userID,
		ProfileID:         &profile.ID,
		Title:             "Rencana Tenang 7 Hari",
		Summary:           fmt.Sprintf("Rencana ringan berdasarkan mood %s dan tujuan: %s.", displayFallback(profile.InitialMood, "awal"), mainGoal),
		Status:            model.WellnessPlanStatusActive,
		StartsOn:          start,
		EndsOn:            start.AddDate(0, 0, 6),
		GeneratedFromMood: profile.InitialMood,
	}

	templates := []struct {
		title       string
		description string
		actionType  string
		route       string
	}{
		{"Hari 1: Kenali Kondisi", "Mulai dari mood check-in agar sistem punya titik awal yang jujur.", "mood", "/dashboard/mood-tracker"},
		{"Hari 2: Tenangkan Tubuh", "Ambil satu sesi pernapasan pendek untuk menurunkan intensitas tubuh.", "breathing", "/dashboard/breathing"},
		{"Hari 3: Tulis Pola", "Catat satu kejadian, satu rasa, dan satu kebutuhan dalam jurnal.", "journal", "/dashboard/journal/create?mode=structured-reflection"},
		{"Hari 4: Rapikan Pikiran", "Gunakan Teman Cerita AI untuk mengubah cerita menjadi langkah kecil.", "chat", "/dashboard/chat"},
		{"Hari 5: Pulihkan Energi", "Dengarkan musik sesuai mood atau pilih latihan fokus singkat.", "music", "/dashboard/music"},
		{"Hari 6: Lihat Progres", "Buka reward dan progress map untuk melihat titik perjalananmu.", "progress", "/dashboard/progress-map"},
		{"Hari 7: Refleksi Mingguan", "Baca insight mingguan dan pilih satu fokus untuk minggu depan.", "weekly_insight", "/dashboard"},
	}

	items := make([]model.WellnessPlanItem, 0, len(templates))
	for i, template := range templates {
		items = append(items, model.WellnessPlanItem{
			UserID:      userID,
			DayNumber:   i + 1,
			ItemDate:    start.AddDate(0, 0, i),
			Title:       template.title,
			Description: template.description,
			ActionType:  template.actionType,
			Route:       template.route,
			Status:      model.WellnessPlanItemStatusPending,
			MetadataJSON: marshalJSONMap(map[string]any{
				"initial_mood": profile.InitialMood,
				"primary_goal": mainGoal,
			}),
		})
	}

	if err := s.repo.CreatePlanWithItems(ctx, plan, items); err != nil {
		return nil, err
	}
	return s.repo.GetActivePlan(ctx, userID)
}

func (s *WellnessService) generateWeeklyInsight(ctx context.Context, userID uint, weekStart, weekEnd time.Time) (*model.WeeklyInsightSnapshot, error) {
	aggregate, err := s.repo.AggregateWeekly(ctx, userID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	moodSummary := buildMoodSummary(aggregate)
	activitySummary := buildActivitySummary(aggregate)
	insight := buildInsight(aggregate)
	premiumSections := buildPremiumSections(aggregate)
	recommendations := buildWeeklyRecommendations(aggregate)
	narrative := buildWeeklyNarrative(aggregate, insight)
	isAIEnhanced := false
	if aiNarrative, ok := s.tryEnhanceWeeklyNarrative(ctx, moodSummary, activitySummary, insight); ok {
		narrative = aiNarrative
		isAIEnhanced = true
	}

	snapshot := &model.WeeklyInsightSnapshot{
		UserID:              userID,
		WeekStart:           weekStart,
		WeekEnd:             weekEnd,
		MoodSummaryJSON:     marshalJSONMap(moodSummary),
		ActivitySummaryJSON: marshalJSONMap(activitySummary),
		InsightJSON:         marshalJSONMap(insight),
		PremiumSectionsJSON: marshalJSONMap(premiumSections),
		Narrative:           narrative,
		RecommendationsJSON: marshalRecommendations(recommendations),
		IsAIEnhanced:        isAIEnhanced,
	}
	if err := s.repo.UpsertWeeklyInsight(ctx, snapshot); err != nil {
		return nil, err
	}
	return s.repo.GetWeeklyInsight(ctx, userID, weekStart)
}

func (s *WellnessService) tryEnhanceWeeklyNarrative(ctx context.Context, moodSummary, activitySummary, insight map[string]any) (string, bool) {
	if s.genaiClient == nil {
		return "", false
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	payload := marshalJSONMap(map[string]any{
		"mood_summary":     moodSummary,
		"activity_summary": activitySummary,
		"insight":          insight,
	})
	prompt := fmt.Sprintf("Buat ringkasan wellbeing mingguan bahasa Indonesia, 2 kalimat, aman dan tidak mengklaim diagnosis. Data agregat: %s", payload)
	model := s.genaiClient.GenerativeModel("gemini-flash-latest")
	model.SetTemperature(0.4)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", false
	}
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", false
	}
	narrative := strings.TrimSpace(string(text))
	if narrative == "" || len(narrative) > 800 {
		return "", false
	}
	return narrative, true
}

func (s *WellnessService) weeklySnapshotToResponse(snapshot *model.WeeklyInsightSnapshot, isPremium bool) *dto.WeeklyInsightResponse {
	moodSummary := parseJSONMap(snapshot.MoodSummaryJSON)
	activitySummary := parseJSONMap(snapshot.ActivitySummaryJSON)
	insight := parseJSONMap(snapshot.InsightJSON)
	premiumSections := parseJSONMap(snapshot.PremiumSectionsJSON)
	recommendations := parseRecommendations(snapshot.RecommendationsJSON)
	premiumPreview := map[string]any{"locked": false}

	if isPremium {
		insight["premium_sections"] = premiumSections
	} else {
		premiumPreview = map[string]any{
			"locked": true,
			"title":  "Insight lanjutan tersedia di Premium",
			"sections": []string{
				"Pola pemicu lintas fitur",
				"Rekomendasi AI lanjutan",
				"Fokus minggu depan yang lebih personal",
			},
			"teaser": "Kami menemukan beberapa sinyal yang bisa dibaca lebih dalam setelah Premium aktif.",
		}
		insight = map[string]any{
			"pattern":        insight["pattern"],
			"progress_label": insight["progress_label"],
			"locked_count":   len(premiumSections),
		}
		for i := range recommendations {
			if i >= 2 {
				recommendations[i].Locked = true
				recommendations[i].Prompt = ""
			}
		}
	}

	return &dto.WeeklyInsightResponse{
		ID:              snapshot.ID.String(),
		WeekStart:       snapshot.WeekStart.Format("2006-01-02"),
		WeekEnd:         snapshot.WeekEnd.AddDate(0, 0, -1).Format("2006-01-02"),
		MoodSummary:     moodSummary,
		ActivitySummary: activitySummary,
		Insight:         insight,
		Narrative:       snapshot.Narrative,
		Recommendations: recommendations,
		PremiumPreview:  premiumPreview,
		IsPremium:       isPremium,
		IsAIEnhanced:    snapshot.IsAIEnhanced,
		GeneratedAt:     snapshot.UpdatedAt,
	}
}

func buildNeedNowResponse(condition string, isPremium bool) (*dto.WellnessNeedNowResponse, bool) {
	base := map[string]dto.WellnessNeedNowResponse{
		"cemas": {
			Condition:   "cemas",
			Title:       "Turunkan intensitas dulu",
			Description: "Mulai dari tubuh, lalu rapikan pikiran setelah napas lebih stabil.",
			Recommendations: []dto.WellnessRecommendationDTO{
				{Type: "breathing", Title: "Box breathing 3 menit", Description: "Latihan pendek untuk memberi sinyal aman ke tubuh.", Route: "/dashboard/breathing", Locked: false},
				{Type: "music", Title: "Musik grounding", Description: "Pilih suara lembut sebelum masuk ke aktivitas berat.", Route: "/dashboard/music", Locked: false},
				{Type: "chat", Title: "Chat AI grounding", Description: "Minta AI memandu satu pertanyaan grounding.", Route: "/dashboard/chat", Prompt: "Bantu aku grounding saat cemas dengan 3 langkah singkat.", Locked: !isPremium},
			},
		},
		"capek": {
			Condition:   "capek",
			Title:       "Pulihkan energi kecil",
			Description: "Pilih aktivitas ringan yang tidak menambah beban keputusan.",
			Recommendations: []dto.WellnessRecommendationDTO{
				{Type: "music", Title: "Playlist pemulihan", Description: "Mulai dari musik pelan selama beberapa menit.", Route: "/dashboard/music", Locked: false},
				{Type: "journal", Title: "Jurnal satu kalimat", Description: "Tulis apa yang paling menguras energi hari ini.", Route: "/dashboard/journal/create?mode=brain-dump", Locked: false},
				{Type: "chat", Title: "Rencana 10 menit", Description: "AI bantu pilih satu langkah paling ringan.", Route: "/dashboard/chat", Prompt: "Aku capek. Bantu pilih satu langkah 10 menit yang realistis.", Locked: !isPremium},
			},
		},
		"sedih": {
			Condition:   "sedih",
			Title:       "Beri ruang untuk rasa",
			Description: "Validasi rasa lebih dulu, lalu pilih satu dukungan kecil.",
			Recommendations: []dto.WellnessRecommendationDTO{
				{Type: "journal", Title: "Jurnal validasi rasa", Description: "Tulis rasa tanpa harus langsung menyelesaikannya.", Route: "/dashboard/journal/create?mode=structured-reflection", Locked: false},
				{Type: "music", Title: "Mood lift lembut", Description: "Dengarkan musik yang mendukung, bukan memaksa ceria.", Route: "/dashboard/music", Locked: false},
				{Type: "chat", Title: "Teman cerita", Description: "AI membantu menyusun kalimat yang menenangkan.", Route: "/dashboard/chat", Prompt: "Aku sedang sedih. Tolong bantu aku validasi perasaan dan pilih satu dukungan kecil.", Locked: !isPremium},
			},
		},
		"marah": {
			Condition:   "marah",
			Title:       "Buat jeda sebelum bereaksi",
			Description: "Regulasi tubuh dulu agar respons berikutnya lebih aman.",
			Recommendations: []dto.WellnessRecommendationDTO{
				{Type: "breathing", Title: "Exhale panjang", Description: "Fokus pada embusan napas lebih panjang.", Route: "/dashboard/breathing", Locked: false},
				{Type: "journal", Title: "Fakta vs emosi", Description: "Pisahkan kejadian, tafsir, dan kebutuhan.", Route: "/dashboard/journal/create?mode=structured-reflection", Locked: false},
				{Type: "chat", Title: "Respons aman", Description: "AI bantu membuat respons yang tidak reaktif.", Route: "/dashboard/chat", Prompt: "Aku sedang marah. Bantu pisahkan fakta, emosi, dan respons yang aman.", Locked: !isPremium},
			},
		},
		"bingung": {
			Condition:   "bingung",
			Title:       "Ubah kabut jadi daftar",
			Description: "Tujuannya bukan selesai semua, tapi menemukan langkah pertama.",
			Recommendations: []dto.WellnessRecommendationDTO{
				{Type: "journal", Title: "Brain dump", Description: "Keluarkan semua hal yang berputar di kepala.", Route: "/dashboard/journal/create?mode=brain-dump", Locked: false},
				{Type: "chat", Title: "Urutkan prioritas", Description: "AI bantu membuat 3 prioritas paling dekat.", Route: "/dashboard/chat", Prompt: "Aku bingung. Bantu urutkan masalah jadi 3 prioritas kecil.", Locked: false},
				{Type: "progress", Title: "Lihat peta progres", Description: "Kembali ke perjalanan agar arah besarnya terlihat.", Route: "/dashboard/progress-map", Locked: !isPremium},
			},
		},
		"fokus": {
			Condition:   "fokus",
			Title:       "Siapkan mode fokus",
			Description: "Bangun ritme singkat: napas, musik, lalu satu target.",
			Recommendations: []dto.WellnessRecommendationDTO{
				{Type: "breathing", Title: "Napas fokus", Description: "Ambil satu sesi pendek sebelum mulai.", Route: "/dashboard/breathing", Locked: false},
				{Type: "music", Title: "Playlist fokus", Description: "Pilih musik latar yang tidak mengganggu.", Route: "/dashboard/music", Locked: false},
				{Type: "chat", Title: "Target 25 menit", Description: "AI bantu menyusun target fokus yang jelas.", Route: "/dashboard/chat", Prompt: "Bantu aku membuat target fokus 25 menit dengan langkah pembuka.", Locked: !isPremium},
			},
		},
	}
	response, ok := base[condition]
	return &response, ok
}

func buildMoodSummary(aggregate *infrastructure.WeeklyAggregate) map[string]any {
	dominant := ""
	maxCount := 0
	total := 0
	keys := make([]string, 0, len(aggregate.MoodCounts))
	for mood := range aggregate.MoodCounts {
		keys = append(keys, mood)
	}
	sort.Strings(keys)
	for _, mood := range keys {
		count := aggregate.MoodCounts[mood]
		total += count
		if count > maxCount {
			dominant = mood
			maxCount = count
		}
	}
	return map[string]any{
		"counts":        aggregate.MoodCounts,
		"checkins":      total,
		"dominant_mood": dominant,
		"latest_mood":   aggregate.LatestMood,
	}
}

func buildActivitySummary(aggregate *infrastructure.WeeklyAggregate) map[string]any {
	return map[string]any{
		"journals":           aggregate.JournalCount,
		"journal_words":      aggregate.JournalWords,
		"breathing_sessions": aggregate.BreathingCount,
		"breathing_minutes":  aggregate.BreathingMinutes,
		"chat_sessions":      aggregate.ChatSessionCount,
		"chat_messages":      aggregate.ChatMessageCount,
		"tasks_completed":    aggregate.TaskCompleted,
		"tasks_claimed":      aggregate.TaskClaimed,
		"reward_claims":      aggregate.RewardClaims,
		"landmarks_unlocked": aggregate.UnlockedLandmarks,
		"streak":             aggregate.Streak,
	}
}

func buildInsight(aggregate *infrastructure.WeeklyAggregate) map[string]any {
	score := weeklyProgressScore(aggregate)
	label := "mulai terbentuk"
	if score >= 80 {
		label = "stabil"
	} else if score >= 50 {
		label = "bertumbuh"
	}
	pattern := "Aktivitas minggu ini masih sedikit, jadi insight akan makin akurat setelah beberapa check-in."
	if aggregate.JournalCount+aggregate.BreathingCount+aggregate.ChatSessionCount >= 5 {
		pattern = "Ada ritme refleksi dan regulasi yang mulai saling mendukung sepanjang minggu."
	}
	return map[string]any{
		"progress_score": score,
		"progress_label": label,
		"pattern":        pattern,
		"trigger_guess":  inferTriggerGuess(aggregate),
	}
}

func buildPremiumSections(aggregate *infrastructure.WeeklyAggregate) map[string]any {
	return map[string]any{
		"cross_feature_pattern":   fmt.Sprintf("Jurnal %d, napas %d, chat %d: pola terbaik muncul saat refleksi diikuti regulasi tubuh.", aggregate.JournalCount, aggregate.BreathingCount, aggregate.ChatSessionCount),
		"next_week_focus":         inferNextWeekFocus(aggregate),
		"advanced_recommendation": "Pilih satu rutinitas tetap: mood check-in pagi, napas siang, jurnal singkat malam selama 3 hari.",
	}
}

func buildWeeklyRecommendations(aggregate *infrastructure.WeeklyAggregate) []dto.WellnessRecommendationDTO {
	recommendations := []dto.WellnessRecommendationDTO{
		{Type: "mood", Title: "Pertahankan check-in", Description: "Catat mood minimal 4 kali agar pola mingguan lebih jelas.", Route: "/dashboard/mood-tracker", Locked: false},
		{Type: "breathing", Title: "Tambahkan jeda napas", Description: "Satu sesi pendek bisa menjadi transisi sebelum jurnal atau chat.", Route: "/dashboard/breathing", Locked: false},
		{Type: "journal", Title: "Prompt pemicu", Description: "Tulis satu pemicu yang muncul berulang minggu ini.", Route: "/dashboard/journal/create?mode=structured-reflection", Locked: false},
		{Type: "chat", Title: "Rencana minggu depan", Description: "Minta AI menyusun 3 langkah kecil berdasarkan insight ini.", Route: "/dashboard/chat", Prompt: "Bantu aku membuat rencana minggu depan dari pola wellbeing minggu ini.", Locked: false},
	}
	if aggregate.JournalCount == 0 {
		recommendations[0] = dto.WellnessRecommendationDTO{Type: "journal", Title: "Mulai dari jurnal 3 menit", Description: "Satu catatan akan membuat insight minggu depan lebih personal.", Route: "/dashboard/journal", Locked: false}
	}
	return recommendations
}

func buildWeeklyNarrative(aggregate *infrastructure.WeeklyAggregate, insight map[string]any) string {
	if aggregate.JournalCount+aggregate.BreathingCount+aggregate.ChatSessionCount+len(aggregate.MoodCounts) == 0 {
		return "Belum banyak data minggu ini. Mulai dari satu check-in mood dan satu aktivitas ringan agar pola pertamamu terbentuk."
	}
	return fmt.Sprintf("Minggu ini progresmu %s. Sinyal paling kuat terlihat dari %d mood check-in, %d jurnal, %d sesi napas, dan %d sesi chat.",
		insight["progress_label"],
		totalMoodCount(aggregate.MoodCounts),
		aggregate.JournalCount,
		aggregate.BreathingCount,
		aggregate.ChatSessionCount,
	)
}

func weeklyProgressScore(aggregate *infrastructure.WeeklyAggregate) int {
	parts := []float64{
		minRatio(totalMoodCount(aggregate.MoodCounts), 4),
		minRatio(aggregate.JournalCount, 3),
		minRatio(aggregate.BreathingCount, 3),
		minRatio(aggregate.ChatSessionCount, 2),
		minRatio(aggregate.TaskCompleted, 8),
	}
	total := 0.0
	for _, part := range parts {
		total += part
	}
	return int((total / float64(len(parts))) * 100)
}

func inferTriggerGuess(aggregate *infrastructure.WeeklyAggregate) string {
	if aggregate.LatestMood == "anxious" || aggregate.LatestMood == "sad" {
		return "Mood akhir minggu menunjukkan perlunya dukungan lembut dan aktivitas rendah tekanan."
	}
	if aggregate.JournalCount > 0 && aggregate.BreathingCount == 0 {
		return "Refleksi sudah muncul, tetapi regulasi tubuh belum banyak tercatat."
	}
	return "Belum ada pemicu dominan yang cukup kuat dari data agregat minggu ini."
}

func inferNextWeekFocus(aggregate *infrastructure.WeeklyAggregate) string {
	switch {
	case aggregate.BreathingCount < 2:
		return "Tambahkan 2 sesi napas singkat sebagai jangkar regulasi."
	case aggregate.JournalCount < 2:
		return "Tulis 2 jurnal pendek untuk menangkap konteks emosi."
	case aggregate.ChatSessionCount < 1:
		return "Gunakan chat AI sekali untuk merapikan rencana mingguan."
	default:
		return "Pertahankan ritme dan pilih satu reward kecil sebagai penutup minggu."
	}
}

func buildJourneyNode(key, label, description string, value, target int, route, tone string) dto.WellnessJourneyNodeDTO {
	progress := minRatio(value, target) * 100
	return dto.WellnessJourneyNodeDTO{
		Key:         key,
		Label:       label,
		Description: description,
		Value:       value,
		Target:      target,
		Progress:    progress,
		Route:       route,
		Tone:        tone,
	}
}

func toProfileDTO(profile *model.UserWellnessProfile) *dto.WellnessProfileDTO {
	if profile == nil {
		return nil
	}
	return &dto.WellnessProfileDTO{
		ID:                    profile.ID,
		UserID:                profile.UserID,
		InitialMood:           profile.InitialMood,
		Goals:                 parseStringSlice(profile.GoalsJSON),
		Habits:                parseStringSlice(profile.HabitsJSON),
		TourCompletedAt:       profile.TourCompletedAt,
		OnboardingCompletedAt: profile.OnboardingCompletedAt,
	}
}

func toPlanDTO(plan *model.WellnessPlan) *dto.WellnessPlanDTO {
	if plan == nil {
		return nil
	}
	items := make([]dto.WellnessPlanItemDTO, 0, len(plan.Items))
	completed := 0
	for _, item := range plan.Items {
		if item.Status == model.WellnessPlanItemStatusCompleted {
			completed++
		}
		items = append(items, toPlanItemDTO(item))
	}
	completion := 0
	if len(items) > 0 {
		completion = int(float64(completed) / float64(len(items)) * 100)
	}
	return &dto.WellnessPlanDTO{
		ID:                plan.ID.String(),
		Title:             plan.Title,
		Summary:           plan.Summary,
		Status:            string(plan.Status),
		StartsOn:          plan.StartsOn.Format("2006-01-02"),
		EndsOn:            plan.EndsOn.Format("2006-01-02"),
		GeneratedFromMood: plan.GeneratedFromMood,
		CompletionPercent: completion,
		Items:             items,
	}
}

func toPlanItemDTO(item model.WellnessPlanItem) dto.WellnessPlanItemDTO {
	return dto.WellnessPlanItemDTO{
		ID:          item.ID.String(),
		DayNumber:   item.DayNumber,
		ItemDate:    item.ItemDate.Format("2006-01-02"),
		Title:       item.Title,
		Description: item.Description,
		ActionType:  item.ActionType,
		Route:       item.Route,
		Status:      string(item.Status),
		CompletedAt: item.CompletedAt,
		Metadata:    parseJSONMap(item.MetadataJSON),
	}
}

func normalizeStringSlice(input []string, limit int) []string {
	result := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func marshalStringSlice(input []string) string {
	encoded, err := json.Marshal(input)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func parseStringSlice(raw string) []string {
	var output []string
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return []string{}
	}
	return output
}

func marshalJSONMap(input map[string]any) string {
	if input == nil {
		return "{}"
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func parseJSONMap(raw string) map[string]any {
	output := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return map[string]any{}
	}
	return output
}

func marshalRecommendations(input []dto.WellnessRecommendationDTO) string {
	encoded, err := json.Marshal(input)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func parseRecommendations(raw string) []dto.WellnessRecommendationDTO {
	var output []dto.WellnessRecommendationDTO
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return []dto.WellnessRecommendationDTO{}
	}
	return output
}

func resolveWeekRange(raw string) (time.Time, time.Time) {
	loc := appLocation()
	now := time.Now().In(loc)
	if strings.TrimSpace(raw) != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), loc); err == nil {
			start := dayStart(parsed)
			return start, start.AddDate(0, 0, 7)
		}
	}
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := dayStart(now.AddDate(0, 0, -(weekday - 1)))
	return start, start.AddDate(0, 0, 7)
}

func appLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

func dayStart(value time.Time) time.Time {
	loc := appLocation()
	v := value.In(loc)
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, loc)
}

func displayFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func minRatio(value, target int) float64 {
	if target <= 0 {
		return 0
	}
	ratio := float64(value) / float64(target)
	if ratio > 1 {
		return 1
	}
	if ratio < 0 {
		return 0
	}
	return ratio
}

func totalMoodCount(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}
