package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type postgresNamedDialectorForJournal struct {
	gorm.Dialector
}

func (d postgresNamedDialectorForJournal) Name() string {
	return "postgres"
}

func setupJournalRepoDB(t *testing.T) (*gorm.DB, *JournalRepository) {
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
		`CREATE TABLE journal_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE,
			allow_ai_access BOOLEAN,
			ai_context_days INTEGER,
			ai_context_max_entries INTEGER,
			default_share_with_ai BOOLEAN,
			is_blocked BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE journal_ai_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			journal_id INTEGER,
			chat_session_id INTEGER,
			accessed_at DATETIME,
			context_type TEXT
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO user_moods (id, user_id, mood, created_at, updated_at) VALUES (1, 1, 'happy', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed user_moods: %v", err)
	}

	repo := NewJournalRepository(db)
	return db, repo
}

func TestJournalRepository_CRUDAndAnalyticsPaths(t *testing.T) {
	_, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	jUUID := uuid.New()
	j := &model.Journal{
		UUID:        jUUID,
		UserID:      1,
		Title:       "Today",
		Content:     "I feel better",
		Summary:     "good day",
		MoodID:      ptrUintJournal(1),
		ShareWithAI: true,
		WordCount:   3,
		CreatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, j); err != nil {
		t.Fatalf("create journal: %v", err)
	}

	if _, err := repo.FindByID(ctx, j.ID); err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if _, err := repo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected FindByID missing error")
	}
	if _, err := repo.FindByIDAndUserID(ctx, j.ID, 1); err != nil {
		t.Fatalf("find by id and user: %v", err)
	}
	if _, err := repo.FindByIDAndUserID(ctx, j.ID, 999); err == nil {
		t.Fatal("expected FindByIDAndUserID mismatched user error")
	}
	if _, err := repo.FindByUUID(ctx, jUUID); err != nil {
		t.Fatalf("find by uuid: %v", err)
	}
	if _, err := repo.FindByUUID(ctx, uuid.New()); err == nil {
		t.Fatal("expected FindByUUID missing error")
	}
	if _, err := repo.FindByUUIDAndUserID(ctx, jUUID, 1); err != nil {
		t.Fatalf("find by uuid and user: %v", err)
	}
	if _, err := repo.FindByUUIDAndUserID(ctx, jUUID, 999); err == nil {
		t.Fatal("expected FindByUUIDAndUserID mismatched user error")
	}

	j.Title = "Updated"
	if err := repo.Update(ctx, j); err != nil {
		t.Fatalf("update journal: %v", err)
	}
	if err := repo.UpdateSummary(ctx, j.ID, "new summary"); err != nil {
		t.Fatalf("update summary: %v", err)
	}

	list, total, err := repo.FindByUserID(ctx, 1, 1, 10, nil, nil, nil, nil)
	if err != nil || total < 1 || len(list) < 1 {
		t.Fatalf("find by user failed err=%v total=%d len=%d", err, total, len(list))
	}

	if _, err := repo.FindForAIContext(ctx, 1, 30, 5); err != nil {
		t.Fatalf("find for ai context: %v", err)
	}

	if err := repo.UpdateAIAccessedAt(ctx, j.ID); err != nil {
		t.Fatalf("update ai accessed at: %v", err)
	}

	count, err := repo.CountByUserID(ctx, 1)
	if err != nil || count < 1 {
		t.Fatalf("count by user failed: err=%v count=%d", err, count)
	}
	shared, err := repo.CountSharedWithAI(ctx, 1)
	if err != nil || shared < 1 {
		t.Fatalf("count shared failed: err=%v count=%d", err, shared)
	}

	distribution, err := repo.GetMoodDistribution(ctx, 1)
	if err != nil {
		t.Fatalf("mood distribution failed: %v", err)
	}
	if distribution["happy"] < 1 {
		t.Fatalf("expected happy distribution, got %+v", distribution)
	}

	wordCount, err := repo.GetTotalWordCount(ctx, 1)
	if err != nil || wordCount < 1 {
		t.Fatalf("word count failed: err=%v count=%d", err, wordCount)
	}

	if _, err := repo.GetWritingStreak(ctx, 1); err != nil {
		t.Fatalf("expected writing streak query to work, got %v", err)
	}

	if err := repo.Delete(ctx, j.ID, 1); err != nil {
		t.Fatalf("delete journal: %v", err)
	}
}

func TestJournalRepository_PostgresSpecificQueriesErrorOnSqlite(t *testing.T) {
	_, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	if _, err := repo.SearchByContent(ctx, 1, "today", 10); err == nil {
		t.Fatal("expected SearchByContent to fail on sqlite ILIKE")
	}
	if _, err := repo.FindRelevantForAIContext(ctx, 1, "today", 5); err == nil {
		t.Fatal("expected FindRelevantForAIContext to fail on sqlite ANY/ILIKE")
	}

	if err := repo.Create(ctx, &model.Journal{UUID: uuid.New(), UserID: 1, Title: "tags", Content: "c", Tags: []string{"calm", "focus", "calm"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("seed journal with tags: %v", err)
	}

	tagFrequency, err := repo.GetTagFrequency(ctx, 1)
	if err != nil {
		t.Fatalf("expected GetTagFrequency sqlite fallback success, got %v", err)
	}
	if tagFrequency["calm"] < 1 {
		t.Fatalf("expected calm tag to be counted, got %#v", tagFrequency)
	}

	if err := repo.Create(ctx, &model.Journal{UUID: uuid.New(), UserID: 1, Title: "empty-tags", Content: "c", Tags: []string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("seed journal with empty tags: %v", err)
	}
	if err := repo.Create(ctx, &model.Journal{UUID: uuid.New(), UserID: 1, Title: "quoted-tags", Content: "c", Tags: []string{"\"mindful\"", "focus"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("seed journal with quoted tags: %v", err)
	}

	tagFrequency2, err := repo.GetTagFrequency(ctx, 1)
	if err != nil {
		t.Fatalf("expected second GetTagFrequency sqlite fallback success, got %v", err)
	}
	if tagFrequency2["focus"] < 2 {
		t.Fatalf("expected focus tag aggregated count >= 2, got %#v", tagFrequency2)
	}
	if tagFrequency2[`\"mindful\`] < 1 {
		t.Fatalf("expected quoted mindful tag to be counted, got %#v", tagFrequency2)
	}
	if _, ok := tagFrequency2[""]; ok {
		t.Fatalf("expected empty tag to be ignored, got %#v", tagFrequency2)
	}
	entriesByMonth, err := repo.GetEntriesByMonth(ctx, 1, 3)
	if err != nil {
		t.Fatalf("expected GetEntriesByMonth sqlite fallback success, got %v", err)
	}
	if len(entriesByMonth) == 0 {
		t.Fatal("expected GetEntriesByMonth to return at least one bucket")
	}
}

func TestJournalRepository_GetEntriesByMonth_PostgresBranchErrorOnSqlite(t *testing.T) {
	dialector := postgresNamedDialectorForJournal{Dialector: sqlite.Open(":memory:")}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite with postgres name dialector: %v", err)
	}

	schema := []string{
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
		)`}
	for _, stmt := range schema {
		if execErr := db.Exec(stmt).Error; execErr != nil {
			t.Fatalf("schema error: %v", execErr)
		}
	}

	repo := NewJournalRepository(db)
	if _, err := repo.GetEntriesByMonth(context.Background(), 1, 3); err == nil {
		t.Fatal("expected GetEntriesByMonth postgres SQL to fail on sqlite backend")
	}
}

func TestJournalSettingsAndAccessLogRepository_BasicPaths(t *testing.T) {
	db, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	settingsRepo := NewJournalSettingsRepository(db)
	if settingsRepo == nil || repo == nil {
		t.Fatal("expected repositories to be created")
	}

	settings, err := settingsRepo.FindOrCreate(ctx, 42)
	if err != nil {
		t.Fatalf("find or create settings: %v", err)
	}
	settings2, err := settingsRepo.FindOrCreate(ctx, 42)
	if err != nil {
		t.Fatalf("find or create settings existing row: %v", err)
	}
	if settings2.UserID != settings.UserID {
		t.Fatalf("expected same user settings on second FindOrCreate, got %d vs %d", settings2.UserID, settings.UserID)
	}
	if settings.UserID != 42 {
		t.Fatalf("unexpected settings user id: %d", settings.UserID)
	}
	settings.AllowAIAccess = true
	if err := settingsRepo.Update(ctx, settings); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if _, err := settingsRepo.FindByUserID(ctx, 42); err != nil {
		t.Fatalf("find settings by user id: %v", err)
	}
	if _, err := settingsRepo.FindByUserID(ctx, 4042); err == nil {
		t.Fatal("expected find settings by unknown user to return error")
	}

	journal := &model.Journal{UUID: uuid.New(), UserID: 42, Title: "x", Content: "y", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(ctx, journal); err != nil {
		t.Fatalf("create journal for logs: %v", err)
	}

	logRepo := NewJournalAIAccessLogRepository(db)
	entry := &model.JournalAIAccessLog{UserID: 42, JournalID: journal.ID, ContextType: "full", AccessedAt: time.Now()}
	if err := logRepo.Create(ctx, entry); err != nil {
		t.Fatalf("create access log: %v", err)
	}
	if _, err := logRepo.FindByUserID(ctx, 42, 10); err != nil {
		t.Fatalf("find logs by user: %v", err)
	}
	if _, err := logRepo.FindByJournalID(ctx, journal.ID); err != nil {
		t.Fatalf("find logs by journal: %v", err)
	}

	errDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite errDB: %v", err)
	}
	errSettingsRepo := NewJournalSettingsRepository(errDB)
	if _, err := errSettingsRepo.FindOrCreate(ctx, 99); err == nil {
		t.Fatal("expected FindOrCreate error when journal_settings table missing")
	}
}

func TestJournalRepository_GetWritingStreak_EmptyData(t *testing.T) {
	db, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	if err := db.Exec(`DELETE FROM journals`).Error; err != nil {
		t.Fatalf("clear journals failed: %v", err)
	}

	streak, err := repo.GetWritingStreak(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error for empty streak data, got %v", err)
	}
	if streak != 0 {
		t.Fatalf("expected streak 0 for empty data, got %d", streak)
	}
}

func TestJournalRepository_GetWritingStreak_ConsecutiveDays(t *testing.T) {
	db, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	if err := db.Exec(`DELETE FROM journals`).Error; err != nil {
		t.Fatalf("clear journals failed: %v", err)
	}

	now := time.Now()
	for i := 0; i < 3; i++ {
		ts := now.AddDate(0, 0, -i)
		if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, share_with_ai, word_count, created_at, updated_at) VALUES (?, 1, ?, 'x', 0, 1, ?, ?)`, uuid.New().String(), "d", ts, ts).Error; err != nil {
			t.Fatalf("seed journal %d failed: %v", i, err)
		}
	}

	streak, err := repo.GetWritingStreak(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected streak error: %v", err)
	}
	if streak < 3 {
		t.Fatalf("expected streak at least 3, got %d", streak)
	}
}

func TestJournalRepository_FindByUserID_FilterBranches(t *testing.T) {
	_, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	now := time.Now()
	moodID := uint(1)
	if err := repo.Create(ctx, &model.Journal{
		UUID:      uuid.New(),
		UserID:    1,
		Title:     "filtered",
		Content:   "entry",
		MoodID:    &moodID,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed filtered journal failed: %v", err)
	}

	start := now.Add(-24 * time.Hour)
	end := now.Add(24 * time.Hour)
	list, total, err := repo.FindByUserID(ctx, 1, 1, 10, nil, &moodID, &start, &end)
	if err != nil {
		t.Fatalf("expected filtered FindByUserID success, got %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("expected non-empty filtered result, total=%d len=%d", total, len(list))
	}

	if _, _, err := repo.FindByUserID(ctx, 1, 1, 10, []string{"calm"}, nil, nil, nil); err == nil {
		t.Fatal("expected tags filter to fail on sqlite due postgres array operator")
	}
}

func TestJournalRepository_GetTagFrequency_ErrorBranch(t *testing.T) {
	db, repo := setupJournalRepoDB(t)
	ctx := context.Background()

	if err := db.Exec(`DROP TABLE journals`).Error; err != nil {
		t.Fatalf("drop journals failed: %v", err)
	}

	if _, err := repo.GetTagFrequency(ctx, 1); err == nil {
		t.Fatal("expected GetTagFrequency to fail when journals table is missing")
	}
}

func TestJournalRepository_GetTagFrequency_PostgresPathErrorOnSqlite(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(postgresNamedDialectorForJournal{Dialector: sqlite.Open(":memory:")}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite with postgres-named dialector failed: %v", err)
	}

	repo := NewJournalRepository(db)
	if _, err := repo.GetTagFrequency(ctx, 1); err == nil {
		t.Fatal("expected postgres query path to fail on sqlite backend")
	}
}

func ptrUintJournal(v uint) *uint {
	return &v
}
