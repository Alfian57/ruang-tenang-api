package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupJournalServiceForAnalytics(t *testing.T) *JournalService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE journals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			user_id INTEGER NOT NULL,
			title TEXT,
			content TEXT,
			summary TEXT,
			mood_id INTEGER,
			tags TEXT,
			is_private BOOLEAN,
			share_with_ai BOOLEAN,
			ai_accessed_at DATETIME,
			word_count INTEGER,
			sentiment_score REAL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO user_moods (id, user_id, mood, created_at, updated_at) VALUES (1, 1, 'happy', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed mood: %v", err)
	}
	if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, mood_id, word_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New().String(), 1, "Weekly", "Progress this week", 1, 120, now, now).Error; err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	return &JournalService{
		journalRepo:  repository.NewJournalRepository(db),
		userMoodRepo: repository.NewUserMoodRepository(db),
	}
}

func TestJournalService_AnalyticsPromptAndSummaryFallbacks(t *testing.T) {
	svc := setupJournalServiceForAnalytics(t)
	ctx := context.Background()

	analytics, err := svc.GetAnalytics(ctx, 1)
	if err != nil {
		t.Fatalf("get analytics failed: %v", err)
	}
	if analytics.TotalEntries < 1 || analytics.TotalWordCount < 1 {
		t.Fatalf("unexpected analytics payload: %+v", analytics)
	}

	prompt, err := svc.GetWritingPrompt(ctx, 1)
	if err != nil {
		t.Fatalf("get writing prompt failed: %v", err)
	}
	if prompt.Prompt == "" || prompt.Category == "" {
		t.Fatalf("unexpected prompt response: %+v", prompt)
	}

	summary, err := svc.GetWeeklySummary(ctx, 1)
	if err != nil {
		t.Fatalf("get weekly summary failed: %v", err)
	}
	if summary.EntriesCount < 1 || summary.Summary == "" {
		t.Fatalf("unexpected weekly summary response: %+v", summary)
	}

	summaryText, themes, insights, suggestions, moodTrend := svc.generateWeeklySummary(ctx, nil)
	if summaryText == "" || moodTrend != "stable" || len(themes) != 0 || len(insights) != 0 || len(suggestions) != 0 {
		t.Fatalf("unexpected empty-data weekly summary fallback: %q %q", summaryText, moodTrend)
	}

	parsedSummary, parsedThemes, parsedInsights, parsedSuggestions, parsedTrend := svc.parseWeeklySummaryResponse(ctx, strings.Join([]string{
		"SUMMARY: Ringkas",
		"THEMES: kerja, keluarga",
		"INSIGHTS: insight1 | insight2",
		"SUGGESTIONS: saran1 | saran2",
		"MOOD_TREND: improving",
	}, "\n"))
	if parsedSummary != "Ringkas" || parsedTrend != "improving" || len(parsedThemes) != 2 || len(parsedInsights) != 2 || len(parsedSuggestions) != 2 {
		t.Fatalf("unexpected parsed summary payload")
	}
}

func TestJournalService_GenerateWritingPromptDirect(t *testing.T) {
	svc := &JournalService{}
	resp := svc.generateWritingPrompt(context.Background(), "mood context", []string{"mindfulness", "sleep"})
	if resp == nil || resp.Prompt == "" || resp.Category == "" || len(resp.RelatedTags) != 2 {
		t.Fatalf("unexpected generated prompt: %+v", resp)
	}
}

func TestJournalService_GenerateSingleEntrySummaryFallback(t *testing.T) {
	svc := &JournalService{}
	summary, err := svc.generateSingleEntrySummary(context.Background(), "I feel okay")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "" {
		t.Fatalf("expected empty summary without genai client, got %q", summary)
	}
}

func TestJournalService_GenerateSingleEntrySummary_WithClientError(t *testing.T) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("dummy-test-key"))
	if err != nil {
		t.Skipf("unable to construct genai client in this environment: %v", err)
	}
	defer client.Close()

	svc := &JournalService{genaiClient: client}
	summary, err := svc.generateSingleEntrySummary(ctx, "hari ini cukup berat")
	if err == nil {
		t.Fatalf("expected error fallback from external AI call, got summary=%q", summary)
	}
}

func TestJournalService_GenerateWeeklySummary_WithClientErrorFallback(t *testing.T) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("dummy-test-key"))
	if err != nil {
		t.Skipf("unable to construct genai client in this environment: %v", err)
	}
	defer client.Close()

	svc := &JournalService{genaiClient: client}
	now := time.Now()
	moodHappy := model.UserMood{Mood: model.MoodHappy}
	journals := []model.Journal{
		{Content: "hari ini saya merasa sangat cemas", CreatedAt: now, Mood: &moodHappy},
		{Content: "saya mencoba teknik napas dan jadi lebih tenang", CreatedAt: now.Add(2 * time.Hour)},
	}

	summary, themes, insights, suggestions, moodTrend := svc.generateWeeklySummary(ctx, journals)
	if summary != "Gagal membuat ringkasan." {
		t.Fatalf("expected AI error fallback summary, got %q", summary)
	}
	if len(themes) != 0 || len(insights) != 0 || len(suggestions) != 0 || moodTrend != "stable" {
		t.Fatalf("expected empty fallback payload, got themes=%v insights=%v suggestions=%v trend=%s", themes, insights, suggestions, moodTrend)
	}
}

func TestJournalService_GenerateWeeklySummary_WithInjectedGeneratorSuccess(t *testing.T) {
	svc := &JournalService{}
	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{{
				Content: &genai.Content{Parts: []genai.Part{genai.Text(strings.Join([]string{
					"SUMMARY: Minggu ini kamu konsisten menulis.",
					"THEMES: refleksi, syukur, fokus",
					"INSIGHTS: kamu lebih sadar emosi | kamu punya rutinitas",
					"SUGGESTIONS: lanjutkan journaling pagi | evaluasi mingguan",
					"MOOD_TREND: improving",
				}, "\n"))}},
			}},
		}, nil
	}

	now := time.Now()
	moodHappy := model.UserMood{Mood: model.MoodHappy}
	journals := []model.Journal{
		{Content: "hari ini saya merasa lebih tenang", CreatedAt: now, Mood: &moodHappy},
		{Content: "saya berhasil menulis jurnal rutin", CreatedAt: now.Add(-time.Hour)},
	}

	summary, themes, insights, suggestions, moodTrend := svc.generateWeeklySummary(context.Background(), journals)
	if summary == "" || moodTrend != "improving" {
		t.Fatalf("expected parsed summary and improving trend, got summary=%q trend=%q", summary, moodTrend)
	}
	if len(themes) != 3 || len(insights) != 2 || len(suggestions) != 2 {
		t.Fatalf("unexpected parsed sections: themes=%v insights=%v suggestions=%v", themes, insights, suggestions)
	}
}

func TestJournalService_GetWritingPrompt_TopTagsCapped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE journals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			user_id INTEGER NOT NULL,
			title TEXT,
			content TEXT,
			summary TEXT,
			mood_id INTEGER,
			tags TEXT,
			is_private BOOLEAN,
			share_with_ai BOOLEAN,
			ai_accessed_at DATETIME,
			word_count INTEGER,
			sentiment_score REAL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	svc := &JournalService{
		journalRepo:  repository.NewJournalRepository(db),
		userMoodRepo: repository.NewUserMoodRepository(db),
	}
	ctx := context.Background()

	now := time.Now()
	if err := db.Exec(
		`INSERT INTO journals (uuid, user_id, title, content, tags, mood_id, word_count, created_at, updated_at) VALUES
		 (?, 1, 'T1', 'c1', '{"mindfulness","sleep","gratitude"}', 1, 50, ?, ?),
		 (?, 1, 'T2', 'c2', '{"focus","mindfulness"}', 1, 50, ?, ?)`,
		uuid.New().String(), now, now,
		uuid.New().String(), now, now,
	).Error; err != nil {
		t.Fatalf("seed tagged journals: %v", err)
	}

	prompt, err := svc.GetWritingPrompt(ctx, 1)
	if err != nil {
		t.Fatalf("GetWritingPrompt failed: %v", err)
	}
	if len(prompt.RelatedTags) == 0 || len(prompt.RelatedTags) > 3 {
		t.Fatalf("expected related tags count in range 1..3, got %+v", prompt.RelatedTags)
	}
}

func TestJournalService_AnalyticsAndPrompt_NoDataBranches(t *testing.T) {
	svc := setupJournalServiceForAnalytics(t)
	ctx := context.Background()

	analytics, err := svc.GetAnalytics(ctx, 999)
	if err != nil {
		t.Fatalf("get analytics no-data failed: %v", err)
	}
	if analytics.TotalEntries != 0 || analytics.AvgWordCount != 0 || analytics.EntriesThisMonth != 0 {
		t.Fatalf("unexpected no-data analytics payload: %+v", analytics)
	}

	prompt, err := svc.GetWritingPrompt(ctx, 999)
	if err != nil {
		t.Fatalf("get writing prompt no-data failed: %v", err)
	}
	if prompt.Prompt == "" || prompt.Category == "" {
		t.Fatalf("expected non-empty generated prompt on no-data user, got %+v", prompt)
	}
	if len(prompt.RelatedTags) != 0 {
		t.Fatalf("expected no related tags for no-data user, got %+v", prompt.RelatedTags)
	}

	weeklySummary, err := svc.GetWeeklySummary(ctx, 999)
	if err != nil {
		t.Fatalf("get weekly summary no-data failed: %v", err)
	}
	if weeklySummary.EntriesCount != 0 || weeklySummary.Summary == "" {
		t.Fatalf("unexpected no-data weekly summary: %+v", weeklySummary)
	}
}

func TestJournalService_GetAnalytics_NoCurrentMonthMatchBranch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE journals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			user_id INTEGER NOT NULL,
			title TEXT,
			content TEXT,
			summary TEXT,
			mood_id INTEGER,
			tags TEXT,
			is_private BOOLEAN,
			share_with_ai BOOLEAN,
			ai_accessed_at DATETIME,
			word_count INTEGER,
			sentiment_score REAL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	oldDate := time.Now().AddDate(0, -2, 0)
	if err := db.Exec(
		`INSERT INTO journals (uuid, user_id, title, content, mood_id, word_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), 42, "Old Month", "entry", 1, 80, oldDate, oldDate,
	).Error; err != nil {
		t.Fatalf("seed old journal: %v", err)
	}

	svc := &JournalService{
		journalRepo:  repository.NewJournalRepository(db),
		userMoodRepo: repository.NewUserMoodRepository(db),
	}

	analytics, err := svc.GetAnalytics(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetAnalytics failed: %v", err)
	}
	if analytics.TotalEntries != 1 || analytics.AvgWordCount != 80 {
		t.Fatalf("unexpected analytics for old-month entry: %+v", analytics)
	}
	if analytics.EntriesThisMonth != 0 {
		t.Fatalf("expected EntriesThisMonth=0 when no current month entries, got %d", analytics.EntriesThisMonth)
	}
}

func TestJournalService_GetAnalytics_AndPrompt_ErrorBranches(t *testing.T) {
	ctx := context.Background()

	createService := func(t *testing.T, schema []string) *JournalService {
		t.Helper()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		for _, stmt := range schema {
			if execErr := db.Exec(stmt).Error; execErr != nil {
				t.Fatalf("schema error: %v", execErr)
			}
		}
		return &JournalService{
			journalRepo:  repository.NewJournalRepository(db),
			userMoodRepo: repository.NewUserMoodRepository(db),
		}
	}

	t.Run("analytics fails when word_count column missing", func(t *testing.T) {
		svc := createService(t, []string{
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT UNIQUE,
				user_id INTEGER NOT NULL,
				title TEXT,
				content TEXT,
				mood_id INTEGER,
				tags TEXT,
				created_at DATETIME,
				updated_at DATETIME
			)`,
			`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
		})
		if _, err := svc.GetAnalytics(ctx, 1); err == nil {
			t.Fatal("expected GetAnalytics error when word_count is missing")
		}
	})

	t.Run("analytics fails when user_moods table missing", func(t *testing.T) {
		svc := createService(t, []string{
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT UNIQUE,
				user_id INTEGER NOT NULL,
				title TEXT,
				content TEXT,
				mood_id INTEGER,
				tags TEXT,
				word_count INTEGER,
				created_at DATETIME,
				updated_at DATETIME
			)`,
			`INSERT INTO journals (uuid, user_id, title, content, mood_id, tags, word_count, created_at, updated_at)
			 VALUES ('jid-1', 1, 't', 'c', 1, '{"tag1"}', 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		})
		if _, err := svc.GetAnalytics(ctx, 1); err == nil {
			t.Fatal("expected GetAnalytics error when user_moods table is missing")
		}
	})

	t.Run("analytics fails when tags column missing", func(t *testing.T) {
		svc := createService(t, []string{
			`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT UNIQUE,
				user_id INTEGER NOT NULL,
				title TEXT,
				content TEXT,
				mood_id INTEGER,
				word_count INTEGER,
				created_at DATETIME,
				updated_at DATETIME
			)`,
			`INSERT INTO journals (uuid, user_id, title, content, mood_id, word_count, created_at, updated_at)
			 VALUES ('jid-2', 1, 't', 'c', NULL, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		})
		if _, err := svc.GetAnalytics(ctx, 1); err == nil {
			t.Fatal("expected GetAnalytics error when tags column is missing")
		}
	})

	t.Run("analytics fails when created_at column missing", func(t *testing.T) {
		svc := createService(t, []string{
			`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT UNIQUE,
				user_id INTEGER NOT NULL,
				title TEXT,
				content TEXT,
				mood_id INTEGER,
				tags TEXT,
				word_count INTEGER,
				updated_at DATETIME
			)`,
			`INSERT INTO journals (uuid, user_id, title, content, mood_id, tags, word_count, updated_at)
			 VALUES ('jid-3', 1, 't', 'c', NULL, '{"tag1"}', 10, CURRENT_TIMESTAMP)`,
		})
		if _, err := svc.GetAnalytics(ctx, 1); err == nil {
			t.Fatal("expected GetAnalytics error when created_at column is missing")
		}
	})

	t.Run("writing prompt fails when user mood query errors", func(t *testing.T) {
		svc := createService(t, []string{
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT UNIQUE,
				user_id INTEGER NOT NULL,
				title TEXT,
				content TEXT,
				tags TEXT,
				created_at DATETIME,
				updated_at DATETIME
			)`,
		})
		if _, err := svc.GetWritingPrompt(ctx, 1); err == nil {
			t.Fatal("expected GetWritingPrompt error when user_moods table is missing")
		}
	})

	t.Run("writing prompt fails when tag query errors", func(t *testing.T) {
		svc := createService(t, []string{
			`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid TEXT UNIQUE,
				user_id INTEGER NOT NULL,
				title TEXT,
				content TEXT,
				created_at DATETIME,
				updated_at DATETIME
			)`,
			`INSERT INTO user_moods (user_id, mood, created_at, updated_at)
			 VALUES (1, 'happy', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		})
		if _, err := svc.GetWritingPrompt(ctx, 1); err == nil {
			t.Fatal("expected GetWritingPrompt error when tags column is missing")
		}
	})
}
