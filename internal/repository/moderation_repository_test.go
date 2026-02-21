package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupModerationRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.ArticleCategory{},
		&model.Article{},
		&model.ContentFlag{},
		&model.UserReport{},
		&model.UserBlock{},
		&model.UserStrike{},
		&model.ModeratorAction{},
		&model.CrisisKeyword{},
		&model.ForumCategory{},
		&model.Forum{},
		&model.ForumPost{},
		&model.Appeal{},
	); err != nil {
		t.Fatalf("migrate moderation tables: %v", err)
	}

	return db
}

func seedModerationBase(t *testing.T, db *gorm.DB) (uint, uint, uint, uint, uint) {
	t.Helper()
	owner := model.User{Name: "Owner", Username: "owner_mod", Email: "owner_mod@test.local", Password: "x", Role: model.RoleMember}
	reporter := model.User{Name: "Reporter", Username: "reporter_mod", Email: "reporter_mod@test.local", Password: "x", Role: model.RoleMember}
	moderator := model.User{Name: "Moderator", Username: "moderator_mod", Email: "moderator_mod@test.local", Password: "x", Role: model.RoleModerator}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := db.Create(&reporter).Error; err != nil {
		t.Fatalf("create reporter: %v", err)
	}
	if err := db.Create(&moderator).Error; err != nil {
		t.Fatalf("create moderator: %v", err)
	}

	cat := model.ArticleCategory{Name: "Moderation", Slug: "moderation"}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create article category: %v", err)
	}

	article := model.Article{
		Title:             "Pending Article",
		Slug:              "pending-article",
		Content:           "content",
		ArticleCategoryID: cat.ID,
		UserID:            owner.ID,
		Status:            model.ArticleStatusDraft,
		ModerationStatus:  model.ArticleModerationPending,
		IsUserGenerated:   true,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("create article: %v", err)
	}

	forum := model.Forum{UserID: owner.ID, Title: "Forum A", Slug: "forum-a", Content: "forum content"}
	if err := db.Create(&forum).Error; err != nil {
		t.Fatalf("create forum: %v", err)
	}

	post := model.ForumPost{ForumID: forum.ID, UserID: owner.ID, Content: "reply"}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create forum post: %v", err)
	}

	return owner.ID, reporter.ID, moderator.ID, article.ID, post.ID
}

func TestModerationRepository_FullFlow(t *testing.T) {
	db := setupModerationRepoDB(t)
	ctx := context.Background()
	repo := NewModerationRepository(db)
	ownerID, reporterID, moderatorID, articleID, postID := seedModerationBase(t, db)

	// Content flag operations
	flag := &model.ContentFlag{ContentType: "article", ContentID: articleID, FlagType: model.ContentFlagTypeManual, FlagCategory: model.ContentFlagCategorySpam, Severity: model.FlagSeverityMedium, FlaggedByID: &reporterID}
	if err := repo.CreateContentFlag(ctx, flag); err != nil {
		t.Fatalf("create content flag: %v", err)
	}
	if _, err := repo.GetContentFlagByID(ctx, flag.ID); err != nil {
		t.Fatalf("get content flag by id: %v", err)
	}
	if flags, err := repo.GetContentFlags(ctx, "article", articleID); err != nil || len(flags) != 1 {
		t.Fatalf("get content flags failed: err=%v len=%d", err, len(flags))
	}
	if unresolved, total, err := repo.GetUnresolvedFlags(ctx, 1, 10); err != nil || total != 1 || len(unresolved) != 1 {
		t.Fatalf("get unresolved flags failed: err=%v total=%d len=%d", err, total, len(unresolved))
	}
	if err := repo.ResolveFlag(ctx, flag.ID, moderatorID, "handled"); err != nil {
		t.Fatalf("resolve flag failed: %v", err)
	}
	if err := repo.DeleteContentFlag(ctx, flag.ID); err != nil {
		t.Fatalf("delete content flag failed: %v", err)
	}

	// Report operations
	reportedContentID := articleID
	reportedUserID := ownerID
	report := &model.UserReport{ReporterID: reporterID, ReportType: model.ReportTypeArticle, ReportedContentID: &reportedContentID, ReportedUserID: &reportedUserID, Reason: model.ReportReasonSpam, Status: model.ReportStatusPending}
	if err := repo.CreateReport(ctx, report); err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	if _, err := repo.GetReportByID(ctx, report.ID); err != nil {
		t.Fatalf("get report by id failed: %v", err)
	}
	if reports, total, err := repo.GetReports(ctx, string(model.ReportStatusPending), string(model.ReportTypeArticle), string(model.ReportReasonSpam), 1, 10); err != nil || total < 1 || len(reports) < 1 {
		t.Fatalf("get reports failed: err=%v total=%d len=%d", err, total, len(reports))
	}
	if pendingCount, err := repo.GetPendingReportsCount(ctx); err != nil || pendingCount < 1 {
		t.Fatalf("pending reports count failed: err=%v count=%d", err, pendingCount)
	}
	if dup, err := repo.CheckDuplicateReport(ctx, reporterID, string(model.ReportTypeArticle), &reportedContentID, nil); err != nil || !dup {
		t.Fatalf("check duplicate report (content) failed: err=%v dup=%v", err, dup)
	}
	if dup, err := repo.CheckDuplicateReport(ctx, reporterID, string(model.ReportTypeArticle), nil, &reportedUserID); err != nil || !dup {
		t.Fatalf("check duplicate report (user) failed: err=%v dup=%v", err, dup)
	}
	report.Status = model.ReportStatusResolved
	now := time.Now()
	report.HandledAt = &now
	report.HandledByID = &moderatorID
	if err := repo.UpdateReport(ctx, report); err != nil {
		t.Fatalf("update report failed: %v", err)
	}
	if resolvedCount, err := repo.GetResolvedReportsTodayCount(ctx, nil); err != nil || resolvedCount < 1 {
		t.Fatalf("resolved reports today count failed: err=%v count=%d", err, resolvedCount)
	}
	if resolvedCount, err := repo.GetResolvedReportsTodayCount(ctx, &moderatorID); err != nil || resolvedCount < 1 {
		t.Fatalf("resolved reports by moderator count failed: err=%v count=%d", err, resolvedCount)
	}

	// User block operations
	block := &model.UserBlock{BlockerID: reporterID, BlockedID: ownerID, Reason: "spam"}
	if err := repo.CreateBlock(ctx, block); err != nil {
		t.Fatalf("create block failed: %v", err)
	}
	if _, err := repo.GetBlockByPair(ctx, reporterID, ownerID); err != nil {
		t.Fatalf("get block by pair failed: %v", err)
	}
	if blocks, err := repo.GetBlocksByBlocker(ctx, reporterID); err != nil || len(blocks) != 1 {
		t.Fatalf("get blocks by blocker failed: err=%v len=%d", err, len(blocks))
	}
	if ids, err := repo.GetBlockedUserIDs(ctx, reporterID); err != nil || len(ids) != 1 {
		t.Fatalf("get blocked user ids failed: err=%v len=%d", err, len(ids))
	}
	if isBlocked, err := repo.IsBlocked(ctx, reporterID, ownerID); err != nil || !isBlocked {
		t.Fatalf("is blocked failed: err=%v isBlocked=%v", err, isBlocked)
	}
	if blocksCount, err := repo.GetBlocksCount(ctx, reporterID); err != nil || blocksCount != 1 {
		t.Fatalf("get blocks count failed: err=%v count=%d", err, blocksCount)
	}
	if err := repo.DeleteBlock(ctx, reporterID, ownerID); err != nil {
		t.Fatalf("delete block failed: %v", err)
	}

	// Strike operations
	strike := &model.UserStrike{UserID: ownerID, Reason: "violation", Severity: model.StrikeSeverityMinor, IssuedByID: moderatorID, IsActive: true}
	if err := repo.CreateStrike(ctx, strike); err != nil {
		t.Fatalf("create strike failed: %v", err)
	}
	if _, err := repo.GetStrikeByID(ctx, strike.ID); err != nil {
		t.Fatalf("get strike by id failed: %v", err)
	}
	if strikes, err := repo.GetUserStrikes(ctx, ownerID, false); err != nil || len(strikes) < 1 {
		t.Fatalf("get user strikes failed: err=%v len=%d", err, len(strikes))
	}
	if activeStrikes, err := repo.GetActiveStrikesCount(ctx, ownerID); err != nil || activeStrikes < 1 {
		t.Fatalf("get active strikes count failed: err=%v count=%d", err, activeStrikes)
	}
	if allActive, err := repo.GetAllActiveStrikesCount(ctx); err != nil || allActive < 1 {
		t.Fatalf("get all active strikes count failed: err=%v count=%d", err, allActive)
	}
	if err := repo.DeactivateStrike(ctx, strike.ID); err != nil {
		t.Fatalf("deactivate strike failed: %v", err)
	}
	expiredAt := time.Now().Add(-1 * time.Hour)
	strike2 := &model.UserStrike{UserID: ownerID, Reason: "old", Severity: model.StrikeSeverityWarning, IssuedByID: moderatorID, IsActive: true, ExpiresAt: &expiredAt}
	if err := repo.CreateStrike(ctx, strike2); err != nil {
		t.Fatalf("create expired strike failed: %v", err)
	}
	if err := repo.DeactivateExpiredStrikes(ctx); err != nil {
		t.Fatalf("deactivate expired strikes failed: %v", err)
	}

	// Moderator action operations
	action := &model.ModeratorAction{ModeratorID: moderatorID, ActionType: model.ModeratorActionUserWarned, TargetType: "user", TargetID: ownerID, Reason: "warn"}
	if err := repo.CreateModeratorAction(ctx, action); err != nil {
		t.Fatalf("create moderator action failed: %v", err)
	}
	if actions, total, err := repo.GetModeratorActions(ctx, &moderatorID, string(model.ModeratorActionUserWarned), "user", 1, 10); err != nil || total < 1 || len(actions) < 1 {
		t.Fatalf("get moderator actions failed: err=%v total=%d len=%d", err, total, len(actions))
	}

	// Crisis keyword operations
	kw := &model.CrisisKeyword{Keyword: "bunuh diri", Category: model.CrisisCategorySuicide, Severity: model.CrisisSeverityCritical, Language: "id", IsActive: true}
	if err := repo.CreateCrisisKeyword(ctx, kw); err != nil {
		t.Fatalf("create crisis keyword failed: %v", err)
	}
	kw.Notes = "important"
	if err := repo.UpdateCrisisKeyword(ctx, kw); err != nil {
		t.Fatalf("update crisis keyword failed: %v", err)
	}
	if activeKeywords, err := repo.GetActiveCrisisKeywords(ctx, "id"); err != nil || len(activeKeywords) < 1 {
		t.Fatalf("get active crisis keywords failed: err=%v len=%d", err, len(activeKeywords))
	}
	if allKeywords, err := repo.GetAllCrisisKeywords(ctx); err != nil || len(allKeywords) < 1 {
		t.Fatalf("get all crisis keywords failed: err=%v len=%d", err, len(allKeywords))
	}
	if err := repo.DeleteCrisisKeyword(ctx, kw.ID); err != nil {
		t.Fatalf("delete crisis keyword failed: %v", err)
	}

	// Article moderation operations
	if pending, total, err := repo.GetArticlesPendingModeration(ctx, "", 1, 10); err != nil || total < 1 || len(pending) < 1 {
		t.Fatalf("get pending articles failed: err=%v total=%d len=%d", err, total, len(pending))
	}
	if err := repo.UpdateArticleModeration(ctx, articleID, model.ArticleModerationApproved, "ok", moderatorID, []string{"trauma"}); err != nil {
		t.Fatalf("update article moderation failed: %v", err)
	}
	if pendingCount, err := repo.GetPendingArticlesCount(ctx); err != nil {
		t.Fatalf("get pending articles count failed: %v", err)
	} else if pendingCount != 0 {
		t.Fatalf("expected 0 pending articles, got %d", pendingCount)
	}
	if flaggedCount, err := repo.GetFlaggedArticlesCount(ctx); err != nil {
		t.Fatalf("get flagged articles count failed: %v", err)
	} else if flaggedCount != 0 {
		t.Fatalf("expected 0 flagged articles, got %d", flaggedCount)
	}

	// User suspension/ban/disclaimer operations
	end := time.Now().Add(24 * time.Hour)
	if err := repo.SuspendUser(ctx, ownerID, end, "cooldown"); err != nil {
		t.Fatalf("suspend user failed: %v", err)
	}
	if suspendedCount, err := repo.GetSuspendedUsersCount(ctx); err != nil || suspendedCount < 1 {
		t.Fatalf("get suspended users count failed: err=%v count=%d", err, suspendedCount)
	}
	if err := repo.UnsuspendUser(ctx, ownerID); err != nil {
		t.Fatalf("unsuspend user failed: %v", err)
	}
	if err := repo.BanUser(ctx, ownerID, "ban reason"); err != nil {
		t.Fatalf("ban user failed: %v", err)
	}
	if bannedCount, err := repo.GetBannedUsersCount(ctx); err != nil || bannedCount < 1 {
		t.Fatalf("get banned users count failed: err=%v count=%d", err, bannedCount)
	}
	if err := repo.UnbanUser(ctx, ownerID); err != nil {
		t.Fatalf("unban user failed: %v", err)
	}
	if err := repo.SetAIDisclaimerAccepted(ctx, ownerID, true); err != nil {
		t.Fatalf("set ai disclaimer accepted failed: %v", err)
	}
	if err := repo.SetContentWarningPreference(ctx, ownerID, model.ContentWarningHideAll); err != nil {
		t.Fatalf("set content warning preference failed: %v", err)
	}

	// Forum moderation operations
	forum := model.Forum{}
	if err := db.Where("user_id = ?", ownerID).First(&forum).Error; err != nil {
		t.Fatalf("find forum failed: %v", err)
	}
	if err := repo.UpdateForumTriggerWarnings(ctx, forum.ID, []string{"trauma"}); err != nil {
		t.Fatalf("update forum trigger warnings failed: %v", err)
	}
	if err := repo.FlagForum(ctx, forum.ID, "flagged"); err != nil {
		t.Fatalf("flag forum failed: %v", err)
	}
	if err := repo.UnflagForum(ctx, forum.ID); err != nil {
		t.Fatalf("unflag forum failed: %v", err)
	}
	if err := repo.FlagForumPost(ctx, postID, "flag post"); err != nil {
		t.Fatalf("flag forum post failed: %v", err)
	}
	if err := repo.UnflagForumPost(ctx, postID); err != nil {
		t.Fatalf("unflag forum post failed: %v", err)
	}

	// Appeal operations
	appeal := &model.Appeal{UserID: ownerID, Reason: "please review", Status: model.AppealStatusPending}
	if err := repo.CreateAppeal(ctx, appeal); err != nil {
		t.Fatalf("create appeal failed: %v", err)
	}
	if _, err := repo.GetAppealByID(ctx, appeal.ID); err != nil {
		t.Fatalf("get appeal by id failed: %v", err)
	}
	if appeals, total, err := repo.GetAppeals(ctx, model.AppealStatusPending, ownerID, 1, 10); err != nil || total < 1 || len(appeals) < 1 {
		t.Fatalf("get appeals failed: err=%v total=%d len=%d", err, total, len(appeals))
	}
	appeal.Status = model.AppealStatusApproved
	appeal.ReviewerID = &moderatorID
	if err := repo.UpdateAppeal(ctx, appeal); err != nil {
		t.Fatalf("update appeal failed: %v", err)
	}
	if pendingAppeals, err := repo.GetPendingAppealsCount(ctx); err != nil {
		t.Fatalf("get pending appeals count failed: %v", err)
	} else if pendingAppeals != 0 {
		t.Fatalf("expected 0 pending appeals, got %d", pendingAppeals)
	}
}
