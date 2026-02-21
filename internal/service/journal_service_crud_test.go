package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupJournalServiceForCRUD(t *testing.T) (*JournalService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
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
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO journal_settings (user_id, allow_ai_access, ai_context_days, ai_context_max_entries, default_share_with_ai, is_blocked, created_at, updated_at) VALUES (1, 1, 7, 5, 0, 0, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, share_with_ai, word_count, created_at, updated_at) VALUES (?, 1, 'Existing', 'existing content', 0, 2, ?, ?)`, uuid.New().String(), now, now).Error; err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	svc := &JournalService{
		journalRepo:  repository.NewJournalRepository(db),
		settingsRepo: repository.NewJournalSettingsRepository(db),
	}
	return svc, db
}

func TestJournalService_CRUDBasicFlow(t *testing.T) {
	svc, db := setupJournalServiceForCRUD(t)
	ctx := context.Background()

	created, err := svc.CreateJournal(ctx, 1, dto.CreateJournalRequest{
		Title:   "New Entry",
		Content: "today was better and calm",
		Tags:    []string{"calm"},
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}
	if created == nil || created.Title != "New Entry" {
		t.Fatalf("unexpected create response: %+v", created)
	}

	fetched, err := svc.GetJournal(ctx, 1, created.ID)
	if err != nil {
		t.Fatalf("get journal failed: %v", err)
	}
	if fetched.ID != created.ID {
		t.Fatalf("unexpected get journal response: %+v", fetched)
	}

	seedUUID := uuid.New().String()
	if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, share_with_ai, word_count, created_at, updated_at) VALUES (?, 1, 'UUID Target', 'uuid content', 0, 2, ?, ?)`, seedUUID, time.Now(), time.Now()).Error; err != nil {
		t.Fatalf("seed uuid journal failed: %v", err)
	}

	byUUID, err := svc.GetJournalByUUID(ctx, 1, seedUUID)
	if err != nil {
		t.Fatalf("get journal by uuid failed: %v", err)
	}
	if byUUID.UUID != seedUUID {
		t.Fatalf("unexpected get by uuid response: %+v", byUUID)
	}

	newTitle := "Updated Entry"
	newContent := "updated content with more words"
	share := true
	updated, err := svc.UpdateJournal(ctx, 1, created.ID, dto.UpdateJournalRequest{
		Title:       &newTitle,
		Content:     &newContent,
		ShareWithAI: &share,
		Tags:        []string{"tag1", "tag2"},
	})
	if err != nil {
		t.Fatalf("update journal failed: %v", err)
	}
	if updated.Title != newTitle || !updated.ShareWithAI {
		t.Fatalf("unexpected update response: %+v", updated)
	}

	newTitle2 := "Updated by UUID"
	updatedByUUID, err := svc.UpdateJournalByUUID(ctx, 1, seedUUID, dto.UpdateJournalRequest{Title: &newTitle2})
	if err != nil {
		t.Fatalf("update journal by uuid failed: %v", err)
	}
	if updatedByUUID.Title != newTitle2 {
		t.Fatalf("unexpected update by uuid response: %+v", updatedByUUID)
	}

	list, total, err := svc.ListJournals(ctx, 1, 1, 20, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("list journals failed: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("unexpected list result total=%d len=%d", total, len(list))
	}

	if err := svc.DeleteJournal(ctx, 1, created.ID); err != nil {
		t.Fatalf("delete journal failed: %v", err)
	}

	deleteUUID := uuid.New().String()
	if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, share_with_ai, word_count, created_at, updated_at) VALUES (?, 1, 'Delete by UUID', 'content', 0, 1, ?, ?)`, deleteUUID, time.Now(), time.Now()).Error; err != nil {
		t.Fatalf("seed delete uuid journal failed: %v", err)
	}
	if err := svc.DeleteJournalByUUID(ctx, 1, deleteUUID); err != nil {
		t.Fatalf("delete journal by uuid failed: %v", err)
	}

	if _, err := svc.GetJournalByUUID(ctx, 1, "bad-uuid"); err == nil {
		t.Fatal("expected invalid uuid error")
	}
	if _, err := svc.UpdateJournalByUUID(ctx, 1, "bad-uuid", dto.UpdateJournalRequest{}); err == nil {
		t.Fatal("expected invalid uuid error on update")
	}
	if err := svc.DeleteJournalByUUID(ctx, 1, "bad-uuid"); err == nil {
		t.Fatal("expected invalid uuid error on delete")
	}

	if err := db.Exec(`UPDATE journal_settings SET is_blocked = 1 WHERE user_id = 1`).Error; err != nil {
		t.Fatalf("set blocked failed: %v", err)
	}
	if _, err := svc.CreateJournal(ctx, 1, dto.CreateJournalRequest{Title: "blocked", Content: "x"}); err == nil {
		t.Fatal("expected blocked create error")
	}
	blockedTitle := "Blocked Edit"
	if _, err := svc.UpdateJournal(ctx, 1, 1, dto.UpdateJournalRequest{Title: &blockedTitle}); err == nil {
		t.Fatal("expected blocked update error")
	}
}

func TestJournalService_SearchAndHelpers(t *testing.T) {
	svc, _ := setupJournalServiceForCRUD(t)
	ctx := context.Background()

	if _, err := svc.SearchJournals(ctx, 1, "existing", 10); err == nil {
		t.Fatal("expected sqlite search error due ILIKE")
	}

	journal := &model.Journal{Title: "A", Content: "old content", WordCount: 2}
	newTitle := "B"
	newContent := "new content now"
	share := true
	moodID := uint(99)
	svc.applyUpdateRequest(journal, dto.UpdateJournalRequest{
		Title:       &newTitle,
		Content:     &newContent,
		MoodID:      &moodID,
		Tags:        []string{"x", "y"},
		ShareWithAI: &share,
	})
	if journal.Title != "B" || journal.Content != newContent || journal.WordCount == 0 || !journal.ShareWithAI || journal.MoodID == nil || *journal.MoodID != 99 {
		t.Fatalf("unexpected applyUpdateRequest result: %+v", journal)
	}

	contentShort := "too short"
	svc.scheduleSummaryRegeneration(1, &contentShort)
	var nilContent *string
	svc.scheduleSummaryRegeneration(1, nilContent)
	longContent := ""
	for i := 0; i < 120; i++ {
		longContent += "word "
	}
	svc.scheduleSummaryRegeneration(1, &longContent)

	t.Run("search journals success via injected search", func(t *testing.T) {
		svc.searchByContentFn = func(_ context.Context, _ uint, _ string, _ int) ([]model.Journal, error) {
			return []model.Journal{{ID: 77, UUID: uuid.New(), UserID: 1, Title: "Found", Content: "match"}}, nil
		}
		res, err := svc.SearchJournals(ctx, 1, "match", 5)
		if err != nil {
			t.Fatalf("expected search success, got %v", err)
		}
		if len(res) != 1 || res[0].ID != 77 {
			t.Fatalf("unexpected search response: %+v", res)
		}
	})
}

func TestJournalService_CreateAndDeleteJournal_ErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("create fails when settings table missing", func(t *testing.T) {
		svc, db := setupJournalServiceForCRUD(t)
		if err := db.Exec(`DROP TABLE journal_settings`).Error; err != nil {
			t.Fatalf("drop settings table failed: %v", err)
		}

		_, err := svc.CreateJournal(ctx, 1, dto.CreateJournalRequest{Title: "x", Content: "y"})
		if err == nil || !strings.Contains(err.Error(), "failed to get settings") {
			t.Fatalf("expected failed to get settings error, got %v", err)
		}
	})

	t.Run("create fails when insert fails", func(t *testing.T) {
		svc, db := setupJournalServiceForCRUD(t)
		if err := db.Exec(`DROP TABLE journals`).Error; err != nil {
			t.Fatalf("drop journals table failed: %v", err)
		}

		_, err := svc.CreateJournal(ctx, 1, dto.CreateJournalRequest{Title: "x", Content: "y"})
		if err == nil || !strings.Contains(err.Error(), "failed to create journal") {
			t.Fatalf("expected failed to create journal error, got %v", err)
		}
	})

	t.Run("create succeeds with share override", func(t *testing.T) {
		svc, _ := setupJournalServiceForCRUD(t)
		share := true
		created, err := svc.CreateJournal(ctx, 1, dto.CreateJournalRequest{
			Title:       "Override Share",
			Content:     "short content",
			ShareWithAI: &share,
		})
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}
		if !created.ShareWithAI {
			t.Fatalf("expected share_with_ai override true, got %+v", created)
		}
	})

	t.Run("delete not found branch", func(t *testing.T) {
		svc, _ := setupJournalServiceForCRUD(t)
		err := svc.DeleteJournal(ctx, 1, 999999)
		if err == nil {
			t.Fatal("expected delete to fail for missing journal")
		}
	})

	t.Run("delete fails when delete query fails", func(t *testing.T) {
		svc, db := setupJournalServiceForCRUD(t)
		created, err := svc.CreateJournal(ctx, 1, dto.CreateJournalRequest{Title: "to-delete", Content: "ok"})
		if err != nil {
			t.Fatalf("seed create failed: %v", err)
		}

		if err := db.Exec(`ALTER TABLE journals RENAME TO journals_old`).Error; err != nil {
			t.Fatalf("rename journals table failed: %v", err)
		}

		err = svc.DeleteJournal(ctx, 1, created.ID)
		if err == nil {
			t.Fatal("expected delete to fail when journals table missing")
		}
	})
}
