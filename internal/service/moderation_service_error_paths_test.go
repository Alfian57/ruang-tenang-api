package service

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupModerationServiceErrorPaths(t *testing.T) *ModerationService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	moderationRepo := repository.NewModerationRepository(db)
	aiModeration := &AIModerationService{moderationRepo: moderationRepo}

	return NewModerationService(
		moderationRepo,
		repository.NewUserRepository(db),
		repository.NewArticleRepository(db),
		repository.NewForumRepository(db),
		aiModeration,
	)
}

func TestModerationService_ErrorHeavySafePaths(t *testing.T) {
	svc := setupModerationServiceErrorPaths(t)
	ctx := context.Background()

	modResult, err := svc.ModerateNewArticle(ctx, &model.Article{Title: "A", Content: "B"})
	if err != nil || modResult == nil {
		t.Fatalf("expected ModerateNewArticle fallback result, err=%v result=%v", err, modResult)
	}
	if modResult.Status == "" {
		t.Fatalf("unexpected moderation status: %+v", modResult)
	}

	crisis, err := svc.DetectCrisis(ctx, "saya merasa sangat tertekan")
	if err != nil || crisis == nil {
		t.Fatalf("expected DetectCrisis fallback result, err=%v result=%v", err, crisis)
	}

	if _, _, err := svc.GetModerationQueue(ctx, "pending", 1, 20); err == nil {
		t.Fatal("expected GetModerationQueue error")
	}
	if err := svc.ModerateArticle(ctx, 1, 1, &dto.ModerateArticleRequest{Action: "approve"}); err == nil {
		t.Fatal("expected ModerateArticle error")
	}
	if _, err := svc.CreateReport(ctx, 1, &dto.CreateReportRequest{ReportType: "article", Reason: "spam"}); err == nil {
		t.Fatal("expected CreateReport error")
	}
	if _, _, err := svc.GetReports(ctx, dto.ReportQueryParams{Page: 1, Limit: 20}); err == nil {
		t.Fatal("expected GetReports error")
	}
	if err := svc.HandleReport(ctx, 1, 1, &dto.HandleReportRequest{Action: "dismiss"}); err == nil {
		t.Fatal("expected HandleReport error")
	}

	if err := svc.BlockUser(ctx, 1, 1, "self"); err == nil {
		t.Fatal("expected self block error")
	}
	if err := svc.UnblockUser(ctx, 1, 2); err == nil {
		t.Fatal("expected UnblockUser error")
	}
	if _, _, err := svc.GetBlockedUsers(ctx, 1); err == nil {
		t.Fatal("expected GetBlockedUsers error")
	}
	if _, err := svc.IsUserBlocked(ctx, 1, 2); err == nil {
		t.Fatal("expected IsUserBlocked error")
	}
	if _, err := svc.GetBlockedUserIDs(ctx, 1); err == nil {
		t.Fatal("expected GetBlockedUserIDs error")
	}

	if _, err := svc.GetUserStrikes(ctx, 1, true); err == nil {
		t.Fatal("expected GetUserStrikes error")
	}
	if _, err := svc.CheckAutoSuspension(ctx, 1); err == nil {
		t.Fatal("expected CheckAutoSuspension error")
	}

	if err := svc.AddTriggerWarnings(ctx, "unknown", 1, 1, []string{"tw"}); err == nil {
		t.Fatal("expected AddTriggerWarnings invalid type error")
	}
	if err := svc.AcceptAIDisclaimer(ctx, 1); err == nil {
		t.Fatal("expected AcceptAIDisclaimer error")
	}
	if err := svc.UpdateContentWarningPreference(ctx, 1, "hide"); err == nil {
		t.Fatal("expected UpdateContentWarningPreference error")
	}

	stats, err := svc.GetModerationStats(ctx)
	if err != nil || stats == nil {
		t.Fatalf("expected stats response even on empty db, err=%v stats=%v", err, stats)
	}
	if _, _, err := svc.GetModeratorActions(ctx, dto.ModeratorActionQueryParams{Page: 1, Limit: 20}); err == nil {
		t.Fatal("expected GetModeratorActions error")
	}
	if _, err := svc.GetCrisisKeywords(ctx); err == nil {
		t.Fatal("expected GetCrisisKeywords error")
	}
	if _, err := svc.CreateCrisisKeyword(ctx, &dto.CreateCrisisKeywordRequest{Keyword: "k", Category: "suicide", Severity: "high"}); err == nil {
		t.Fatal("expected CreateCrisisKeyword error")
	}
	if err := svc.DeleteCrisisKeyword(ctx, 1); err == nil {
		t.Fatal("expected DeleteCrisisKeyword error")
	}

	if _, err := svc.CreateAppeal(ctx, 1, &dto.CreateAppealRequest{Reason: "please", Evidence: "x"}); err == nil {
		t.Fatal("expected CreateAppeal error")
	}
	if _, _, err := svc.GetAppeals(ctx, "pending", 1, 20); err == nil {
		t.Fatal("expected GetAppeals error")
	}
	if err := svc.ReviewAppeal(ctx, 1, 1, &dto.ReviewAppealRequest{Status: "approved"}); err == nil {
		t.Fatal("expected ReviewAppeal error")
	}
}

func setupModerationServiceForHandleReport(t *testing.T) *ModerationService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			email TEXT,
			avatar TEXT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE user_reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			reporter_id INTEGER,
			report_type TEXT,
			reported_content_id INTEGER,
			reported_user_id INTEGER,
			reason TEXT,
			description TEXT,
			status TEXT,
			handled_by_id INTEGER,
			handled_at DATETIME,
			action_taken TEXT,
			moderator_notes TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}
	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO users (id, name, email, avatar, deleted_at) VALUES (1, 'Reporter', 'r@example.com', '', NULL), (2, 'Moderator', 'm@example.com', '', NULL)`).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if err := db.Exec(`INSERT INTO user_reports (id, reporter_id, report_type, reason, status, created_at, updated_at) VALUES (1, 1, 'article', 'spam', 'pending', ?, ?), (2, 1, 'article', 'spam', 'resolved', ?, ?)`, now, now, now, now).Error; err != nil {
		t.Fatalf("seed reports: %v", err)
	}

	moderationRepo := repository.NewModerationRepository(db)
	return NewModerationService(moderationRepo, repository.NewUserRepository(db), repository.NewArticleRepository(db), repository.NewForumRepository(db), &AIModerationService{moderationRepo: moderationRepo})
}

func TestModerationService_HandleReportBranches(t *testing.T) {
	svc := setupModerationServiceForHandleReport(t)
	ctx := context.Background()

	if err := svc.HandleReport(ctx, 999, 2, &dto.HandleReportRequest{Action: "dismiss", Notes: "x"}); err == nil {
		t.Fatal("expected not found error")
	}
	if err := svc.HandleReport(ctx, 2, 2, &dto.HandleReportRequest{Action: "dismiss", Notes: "x"}); err == nil {
		t.Fatal("expected already handled error")
	}
	if err := svc.HandleReport(ctx, 1, 2, &dto.HandleReportRequest{Action: "invalid", Notes: "x"}); err == nil {
		t.Fatal("expected invalid action error")
	}
}
