package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupModerationServiceSuccess(t *testing.T) (*ModerationService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.ArticleCategory{},
		&model.Article{},
		&model.ContentFlag{},
		&model.UserReport{},
		&model.UserBlock{},
		&model.UserStrike{},
		&model.ModeratorAction{},
		&model.CrisisKeyword{},
		&model.Appeal{},
		&model.Forum{},
		&model.ForumPost{},
	)
	if err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	users := []model.User{
		{ID: 1, Name: "Reporter", Username: "reporter", Email: "reporter@example.com", Password: "x", Role: model.RoleMember},
		{ID: 2, Name: "Moderator", Username: "moderator", Email: "moderator@example.com", Password: "x", Role: model.RoleModerator},
		{ID: 3, Name: "Author", Username: "author", Email: "author@example.com", Password: "x", Role: model.RoleMember, IsBanned: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	suspendedUntil := time.Now().Add(48 * time.Hour)
	if err := db.Model(&model.User{}).Where("id = ?", 3).Updates(map[string]interface{}{
		"suspension_end": suspendedUntil,
	}).Error; err != nil {
		t.Fatalf("seed suspension: %v", err)
	}

	category := model.ArticleCategory{ID: 1, Name: "Mind", Slug: "mind"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}

	article := model.Article{
		ID:                1,
		Title:             "Article One",
		Slug:              "article-one",
		Content:           strings.Repeat("a", 240),
		ArticleCategoryID: 1,
		UserID:            3,
		Status:            model.ArticleStatusDraft,
		ModerationStatus:  model.ArticleModerationPending,
		IsUserGenerated:   true,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("seed article: %v", err)
	}

	flag := model.ContentFlag{
		ID:           1,
		ContentType:  "article",
		ContentID:    1,
		FlagType:     model.ContentFlagTypeAIModeration,
		FlagCategory: model.ContentFlagCategorySpam,
		Severity:     model.FlagSeverityHigh,
		AIReason:     "possible spam",
		IsResolved:   false,
	}
	if err := db.Create(&flag).Error; err != nil {
		t.Fatalf("seed flag: %v", err)
	}

	reportedUserID := uint(3)
	reportedContentID := uint(1)
	report := model.UserReport{
		ID:                1,
		ReporterID:        1,
		ReportType:        model.ReportTypeArticle,
		ReportedContentID: &reportedContentID,
		ReportedUserID:    &reportedUserID,
		Reason:            model.ReportReasonSpam,
		Description:       "contains spam links",
		Status:            model.ReportStatusPending,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatalf("seed report: %v", err)
	}

	block := model.UserBlock{ID: 1, BlockerID: 1, BlockedID: 3, Reason: "toxic"}
	if err := db.Create(&block).Error; err != nil {
		t.Fatalf("seed block: %v", err)
	}

	strike := model.UserStrike{ID: 1, UserID: 3, ReportID: &report.ID, Reason: "warning", Severity: model.StrikeSeverityWarning, IssuedByID: 2, IsActive: true}
	if err := db.Create(&strike).Error; err != nil {
		t.Fatalf("seed strike: %v", err)
	}

	keyword := model.CrisisKeyword{ID: 1, Keyword: "bunuh diri", Category: model.CrisisCategorySuicide, Severity: model.CrisisSeverityCritical, Language: "id", IsActive: true}
	if err := db.Create(&keyword).Error; err != nil {
		t.Fatalf("seed crisis keyword: %v", err)
	}

	appeal := model.Appeal{ID: 1, UserID: 3, Reason: "please reconsider", Evidence: "new evidence", Status: model.AppealStatusPending}
	if err := db.Create(&appeal).Error; err != nil {
		t.Fatalf("seed appeal: %v", err)
	}

	action := model.ModeratorAction{ID: 1, ModeratorID: 2, ActionType: model.ModeratorActionArticleApproved, TargetType: "article", TargetID: 1, Reason: "initial action"}
	if err := db.Create(&action).Error; err != nil {
		t.Fatalf("seed moderator action: %v", err)
	}

	forum := model.Forum{ID: 1, UserID: 1, Title: "Forum One", Slug: "forum-one", Content: "forum content"}
	if err := db.Create(&forum).Error; err != nil {
		t.Fatalf("seed forum: %v", err)
	}

	moderationRepo := repository.NewModerationRepository(db)
	svc := NewModerationService(
		moderationRepo,
		repository.NewUserRepository(db),
		repository.NewArticleRepository(db),
		repository.NewForumRepository(db),
		&AIModerationService{moderationRepo: moderationRepo},
	)

	return svc, db
}

func TestModerationService_SuccessMappings(t *testing.T) {
	svc, _ := setupModerationServiceSuccess(t)
	ctx := context.Background()

	queueItems, totalQueue, err := svc.GetModerationQueue(ctx, "pending", 1, 10)
	if err != nil {
		t.Fatalf("GetModerationQueue error: %v", err)
	}
	if totalQueue != 1 || len(queueItems) != 1 {
		t.Fatalf("unexpected queue result total=%d len=%d", totalQueue, len(queueItems))
	}
	if queueItems[0].AuthorName != "Author" || queueItems[0].Severity != string(model.FlagSeverityHigh) {
		t.Fatalf("unexpected queue mapping: %+v", queueItems[0])
	}

	reports, totalReports, err := svc.GetReports(ctx, dto.ReportQueryParams{Status: "pending", ReportType: "article", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("GetReports error: %v", err)
	}
	if totalReports != 1 || len(reports) != 1 {
		t.Fatalf("unexpected reports result total=%d len=%d", totalReports, len(reports))
	}
	if reports[0].ReporterName != "Reporter" || reports[0].ReportedUserName != "Author" || reports[0].ContentTitle != "Article One" {
		t.Fatalf("unexpected report mapping: %+v", reports[0])
	}

	blockedUsers, blockedCount, err := svc.GetBlockedUsers(ctx, 1)
	if err != nil {
		t.Fatalf("GetBlockedUsers error: %v", err)
	}
	if blockedCount != 1 || len(blockedUsers) != 1 || blockedUsers[0].BlockedName != "Author" {
		t.Fatalf("unexpected blocked result count=%d data=%+v", blockedCount, blockedUsers)
	}

	strikes, err := svc.GetUserStrikes(ctx, 3, true)
	if err != nil {
		t.Fatalf("GetUserStrikes error: %v", err)
	}
	if len(strikes) != 1 || strikes[0].IssuedByName != "Moderator" {
		t.Fatalf("unexpected strikes mapping: %+v", strikes)
	}

	actions, totalActions, err := svc.GetModeratorActions(ctx, dto.ModeratorActionQueryParams{ModeratorID: 2, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("GetModeratorActions error: %v", err)
	}
	if totalActions == 0 || len(actions) == 0 || actions[0].ModeratorName != "Moderator" {
		t.Fatalf("unexpected actions mapping total=%d data=%+v", totalActions, actions)
	}

	keywords, err := svc.GetCrisisKeywords(ctx)
	if err != nil {
		t.Fatalf("GetCrisisKeywords error: %v", err)
	}
	if len(keywords) != 1 || keywords[0].Language != "id" {
		t.Fatalf("unexpected keywords mapping: %+v", keywords)
	}

	appeals, totalAppeals, err := svc.GetAppeals(ctx, "pending", 1, 10)
	if err != nil {
		t.Fatalf("GetAppeals error: %v", err)
	}
	if totalAppeals != 1 || len(appeals) != 1 {
		t.Fatalf("unexpected appeals result total=%d len=%d", totalAppeals, len(appeals))
	}
	if appeals[0].UserName != "Author" || appeals[0].UserEmail != "author@example.com" {
		t.Fatalf("unexpected appeals mapping: %+v", appeals[0])
	}
}

func TestModerationService_StateChangingSuccess(t *testing.T) {
	svc, db := setupModerationServiceSuccess(t)
	ctx := context.Background()

	if err := svc.AddTriggerWarnings(ctx, "article", 1, 2, []string{"self_harm"}); err != nil {
		t.Fatalf("AddTriggerWarnings article error: %v", err)
	}
	if err := svc.AddTriggerWarnings(ctx, "forum", 1, 2, []string{"abuse"}); err != nil {
		t.Fatalf("AddTriggerWarnings forum error: %v", err)
	}

	if err := svc.AcceptAIDisclaimer(ctx, 1); err != nil {
		t.Fatalf("AcceptAIDisclaimer error: %v", err)
	}
	if err := svc.UpdateContentWarningPreference(ctx, 1, string(model.ContentWarningHideAll)); err != nil {
		t.Fatalf("UpdateContentWarningPreference error: %v", err)
	}

	var reporter model.User
	if err := db.First(&reporter, 1).Error; err != nil {
		t.Fatalf("query reporter error: %v", err)
	}
	if !reporter.HasAcceptedAIDisclaimer || reporter.ContentWarningPreference != model.ContentWarningHideAll {
		t.Fatalf("unexpected reporter moderation preferences: %+v", reporter)
	}

	newKeyword, err := svc.CreateCrisisKeyword(ctx, &dto.CreateCrisisKeywordRequest{Keyword: "darurat", Category: "emergency", Severity: "high"})
	if err != nil {
		t.Fatalf("CreateCrisisKeyword error: %v", err)
	}
	if newKeyword.Language != "id" {
		t.Fatalf("expected default language id, got %q", newKeyword.Language)
	}

	newAppeal, err := svc.CreateAppeal(ctx, 1, &dto.CreateAppealRequest{Reason: "please review my case", Evidence: "timeline"})
	if err != nil {
		t.Fatalf("CreateAppeal error: %v", err)
	}
	if newAppeal.Status != model.AppealStatusPending {
		t.Fatalf("unexpected new appeal status: %s", newAppeal.Status)
	}
	if err := svc.ReviewAppeal(ctx, 1, 2, &dto.ReviewAppealRequest{Status: "approved", Notes: "accepted"}); err != nil {
		t.Fatalf("ReviewAppeal error: %v", err)
	}

	var reviewedAppeal model.Appeal
	if err := db.First(&reviewedAppeal, 1).Error; err != nil {
		t.Fatalf("query reviewed appeal error: %v", err)
	}
	if reviewedAppeal.Status != model.AppealStatusApproved || reviewedAppeal.ReviewerID == nil || *reviewedAppeal.ReviewerID != 2 {
		t.Fatalf("unexpected reviewed appeal: %+v", reviewedAppeal)
	}

	var author model.User
	if err := db.First(&author, 3).Error; err != nil {
		t.Fatalf("query author error: %v", err)
	}
	if author.IsBanned || author.SuspensionEnd != nil {
		t.Fatalf("expected restrictions lifted, got is_banned=%v suspension_end=%v", author.IsBanned, author.SuspensionEnd)
	}

	if err := svc.HandleReport(ctx, 1, 2, &dto.HandleReportRequest{Action: "dismiss", Notes: "no issue"}); err != nil {
		t.Fatalf("HandleReport dismiss error: %v", err)
	}

	var handledReport model.UserReport
	if err := db.First(&handledReport, 1).Error; err != nil {
		t.Fatalf("query handled report error: %v", err)
	}
	if handledReport.Status != model.ReportStatusDismissed || handledReport.HandledByID == nil || *handledReport.HandledByID != 2 {
		t.Fatalf("unexpected handled report: %+v", handledReport)
	}
}

func TestModerationService_ModerateNewArticle_Branches(t *testing.T) {
	svc, db := setupModerationServiceSuccess(t)
	ctx := context.Background()

	call := 0
	svc.aiModerationService.genaiModel = &genai.GenerativeModel{}
	svc.aiModerationService.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		call++
		if call == 1 {
			return &genai.GenerateContentResponse{Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{genai.Text(`{"status":"flagged","confidence":88,"reasons":["sensitif"],"flag_category":"self_harm","severity":"high","suggestions":"revisi nada"}`)}}}}}, nil
		}
		return &genai.GenerateContentResponse{Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{genai.Text(`{"trigger_warnings":["self_harm"],"has_sensitive_content":true}`)}}}}}, nil
	}

	article := &model.Article{ID: 99, Title: "Topik", Content: "konten", ModerationStatus: model.ArticleModerationPending}
	res, err := svc.ModerateNewArticle(ctx, article)
	if err != nil {
		t.Fatalf("ModerateNewArticle flagged path error: %v", err)
	}
	if res.Status != model.ArticleModerationFlagged || article.ModerationStatus != model.ArticleModerationFlagged {
		t.Fatalf("expected flagged status, got res=%s article=%s", res.Status, article.ModerationStatus)
	}
	if !strings.Contains(article.ModerationNotes, "AI Moderation") || !strings.Contains(article.ModerationNotes, "Saran") {
		t.Fatalf("expected moderation notes with reasons+saran, got %q", article.ModerationNotes)
	}
	if len(article.TriggerWarnings) == 0 {
		t.Fatalf("expected trigger warnings, got %+v", article.TriggerWarnings)
	}

	var flagsCount int64
	if err := db.Model(&model.ContentFlag{}).Where("content_type = ? AND content_id = ?", "article", 99).Count(&flagsCount).Error; err != nil {
		t.Fatalf("count flags error: %v", err)
	}
	if flagsCount == 0 {
		t.Fatal("expected content flag to be created for flagged article")
	}

	svc.aiModerationService.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{genai.Text(`{"status":"approved","confidence":95,"reasons":[],"flag_category":"","severity":"low","suggestions":""}`)}}}}}, nil
	}
	article2 := &model.Article{ID: 100, Title: "Topik2", Content: "konten2", ModerationStatus: model.ArticleModerationPending}
	res2, err := svc.ModerateNewArticle(ctx, article2)
	if err != nil {
		t.Fatalf("ModerateNewArticle approved path error: %v", err)
	}
	if res2.Status != model.ArticleModerationApproved || article2.ModerationStatus != model.ArticleModerationApproved {
		t.Fatalf("expected approved status, got res=%s article=%s", res2.Status, article2.ModerationStatus)
	}
}

func TestModerationService_AutoSuspensionAndStats(t *testing.T) {
	svc, db := setupModerationServiceSuccess(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		strike := model.UserStrike{UserID: 1, Reason: "repeat offense", Severity: model.StrikeSeverityMinor, IssuedByID: 2, IsActive: true}
		if err := db.Create(&strike).Error; err != nil {
			t.Fatalf("seed extra strike: %v", err)
		}
	}

	suspended, err := svc.CheckAutoSuspension(ctx, 1)
	if err != nil {
		t.Fatalf("CheckAutoSuspension error: %v", err)
	}
	if !suspended {
		t.Fatal("expected user to be auto-suspended")
	}

	stats, err := svc.GetModerationStats(ctx)
	if err != nil {
		t.Fatalf("GetModerationStats error: %v", err)
	}
	if stats.PendingArticles < 1 || stats.PendingReports < 1 || stats.ActiveStrikes < 1 || stats.SuspendedUsers < 1 || stats.BannedUsers < 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestModerationService_HighImpactBranches(t *testing.T) {
	svc, db := setupModerationServiceSuccess(t)
	ctx := context.Background()

	articles := []model.Article{
		{ID: 2, Title: "Article Two", Slug: "article-two", Content: "x", ArticleCategoryID: 1, UserID: 3, Status: model.ArticleStatusDraft, ModerationStatus: model.ArticleModerationPending, IsUserGenerated: true},
		{ID: 3, Title: "Article Three", Slug: "article-three", Content: "y", ArticleCategoryID: 1, UserID: 3, Status: model.ArticleStatusDraft, ModerationStatus: model.ArticleModerationPending, IsUserGenerated: true},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed extra articles: %v", err)
	}

	flags := []model.ContentFlag{
		{ContentType: "article", ContentID: 2, FlagType: model.ContentFlagTypeManual, FlagCategory: model.ContentFlagCategoryOther, Severity: model.FlagSeverityMedium, AIReason: "manual review", IsResolved: false},
		{ContentType: "article", ContentID: 3, FlagType: model.ContentFlagTypeManual, FlagCategory: model.ContentFlagCategoryOther, Severity: model.FlagSeverityLow, AIReason: "manual review", IsResolved: false},
	}
	if err := db.Create(&flags).Error; err != nil {
		t.Fatalf("seed extra flags: %v", err)
	}

	if err := svc.ModerateArticle(ctx, 1, 2, &dto.ModerateArticleRequest{Action: "approve", Notes: "ok", TriggerWarnings: []string{"tw"}}); err != nil {
		t.Fatalf("ModerateArticle approve error: %v", err)
	}
	if err := svc.ModerateArticle(ctx, 2, 2, &dto.ModerateArticleRequest{Action: "reject", Notes: "not suitable"}); err != nil {
		t.Fatalf("ModerateArticle reject error: %v", err)
	}
	if err := svc.ModerateArticle(ctx, 3, 2, &dto.ModerateArticleRequest{Action: "request_edit", Notes: "needs edits"}); err != nil {
		t.Fatalf("ModerateArticle request_edit error: %v", err)
	}

	var approved model.Article
	if err := db.First(&approved, 1).Error; err != nil {
		t.Fatalf("query approved article: %v", err)
	}
	if approved.ModerationStatus != model.ArticleModerationApproved || approved.Status != model.ArticleStatusPublished {
		t.Fatalf("unexpected approved article state: %+v", approved)
	}

	var rejected model.Article
	if err := db.First(&rejected, 2).Error; err != nil {
		t.Fatalf("query rejected article: %v", err)
	}
	if rejected.ModerationStatus != model.ArticleModerationRejected || rejected.Status != model.ArticleStatusDraft {
		t.Fatalf("unexpected rejected article state: %+v", rejected)
	}

	var editable model.Article
	if err := db.First(&editable, 3).Error; err != nil {
		t.Fatalf("query editable article: %v", err)
	}
	if editable.ModerationStatus != model.ArticleModerationRevisionNeeded {
		t.Fatalf("unexpected revision article state: %+v", editable)
	}

	reportedUserID := uint(3)
	if _, err := svc.CreateReport(ctx, 1, &dto.CreateReportRequest{ReportType: "user", Reason: "harassment", UserID: &reportedUserID}); err != nil {
		t.Fatalf("CreateReport success error: %v", err)
	}
	if _, err := svc.CreateReport(ctx, 1, &dto.CreateReportRequest{ReportType: "user", Reason: "harassment", UserID: &reportedUserID}); err == nil {
		t.Fatal("expected duplicate CreateReport error")
	}

	if err := svc.BlockUser(ctx, 1, 2, "spam"); err != nil {
		t.Fatalf("BlockUser success error: %v", err)
	}
	if err := svc.BlockUser(ctx, 1, 2, "spam"); err == nil {
		t.Fatal("expected BlockUser duplicate error")
	}

	articleID := uint(1)
	reportWarn := model.UserReport{ReporterID: 1, ReportType: model.ReportTypeUser, ReportedUserID: &reportedUserID, Reason: model.ReportReasonHarassment, Status: model.ReportStatusPending}
	reportRemove := model.UserReport{ReporterID: 1, ReportType: model.ReportTypeArticle, ReportedContentID: &articleID, ReportedUserID: &reportedUserID, Reason: model.ReportReasonSpam, Status: model.ReportStatusPending}
	reportSuspend := model.UserReport{ReporterID: 1, ReportType: model.ReportTypeUser, ReportedUserID: &reportedUserID, Reason: model.ReportReasonHarassment, Status: model.ReportStatusPending}
	reportBan := model.UserReport{ReporterID: 1, ReportType: model.ReportTypeUser, ReportedUserID: &reportedUserID, Reason: model.ReportReasonHarassment, Status: model.ReportStatusPending}
	if err := db.Create(&reportWarn).Error; err != nil {
		t.Fatalf("seed warn report: %v", err)
	}
	if err := db.Create(&reportRemove).Error; err != nil {
		t.Fatalf("seed remove report: %v", err)
	}
	if err := db.Create(&reportSuspend).Error; err != nil {
		t.Fatalf("seed suspend report: %v", err)
	}
	if err := db.Create(&reportBan).Error; err != nil {
		t.Fatalf("seed ban report: %v", err)
	}

	if err := svc.HandleReport(ctx, reportWarn.ID, 2, &dto.HandleReportRequest{Action: "warn", Notes: "warning"}); err != nil {
		t.Fatalf("HandleReport warn error: %v", err)
	}
	if err := svc.HandleReport(ctx, reportRemove.ID, 2, &dto.HandleReportRequest{Action: "remove_content", Notes: "remove"}); err != nil {
		t.Fatalf("HandleReport remove_content error: %v", err)
	}
	if err := svc.HandleReport(ctx, reportSuspend.ID, 2, &dto.HandleReportRequest{Action: "suspend", Notes: "suspend", Duration: 2}); err != nil {
		t.Fatalf("HandleReport suspend error: %v", err)
	}
	if err := svc.HandleReport(ctx, reportBan.ID, 2, &dto.HandleReportRequest{Action: "ban", Notes: "ban"}); err != nil {
		t.Fatalf("HandleReport ban error: %v", err)
	}

	var blockedArticle model.Article
	if err := db.First(&blockedArticle, 1).Error; err != nil {
		t.Fatalf("query blocked article: %v", err)
	}
	if blockedArticle.Status != model.ArticleStatusBlocked {
		t.Fatalf("expected blocked article status, got %s", blockedArticle.Status)
	}

	var strikeCount int64
	if err := db.Model(&model.UserStrike{}).Where("user_id = ?", 3).Count(&strikeCount).Error; err != nil {
		t.Fatalf("count strikes: %v", err)
	}
	if strikeCount < 3 {
		t.Fatalf("expected strikes created by report actions, got %d", strikeCount)
	}
}
