package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type ModerationService struct {
	moderationRepo      *repository.ModerationRepository
	userRepo            *repository.UserRepository
	articleRepo         *repository.ArticleRepository
	forumRepo           repository.ForumRepository
	aiModerationService *AIModerationService
}

func NewModerationService(
	moderationRepo *repository.ModerationRepository,
	userRepo *repository.UserRepository,
	articleRepo *repository.ArticleRepository,
	forumRepo repository.ForumRepository,
	aiModerationService *AIModerationService,
) *ModerationService {
	return &ModerationService{
		moderationRepo:      moderationRepo,
		userRepo:            userRepo,
		articleRepo:         articleRepo,
		forumRepo:           forumRepo,
		aiModerationService: aiModerationService,
	}
}

// ========================
// Article Moderation
// ========================

// ModerateNewArticle runs AI moderation on a new article
func (s *ModerationService) ModerateNewArticle(ctx context.Context, article *model.Article) (*dto.AIModerationResult, error) {
	result, err := s.aiModerationService.ModerateArticle(ctx, article.Title, article.Content)
	if err != nil {
		return nil, err
	}

	// Update article moderation status
	article.ModerationStatus = result.Status
	if len(result.Reasons) > 0 {
		article.ModerationNotes = fmt.Sprintf("AI Moderation: %v", result.Reasons)
	}
	if result.Suggestions != "" {
		article.ModerationNotes += "\n\nSaran: " + result.Suggestions
	}

	// Detect trigger warnings
	warnings, _ := s.aiModerationService.DetectTriggerWarnings(ctx, article.Content)
	if len(warnings) > 0 {
		article.TriggerWarnings = model.TriggerWarnings(warnings)
	}

	// Create content flag if flagged or rejected
	if result.Status == model.ArticleModerationFlagged || result.Status == model.ArticleModerationRejected {
		confidence := result.Confidence
		flag := &model.ContentFlag{
			ContentType:  "article",
			ContentID:    article.ID,
			FlagType:     model.ContentFlagTypeAIModeration,
			FlagCategory: model.ContentFlagCategory(result.FlagCategory),
			Severity:     model.FlagSeverity(result.Severity),
			AIConfidence: &confidence,
			AIReason:     fmt.Sprintf("%v", result.Reasons),
		}
		_ = s.moderationRepo.CreateContentFlag(ctx, flag)
	}

	return result, nil
}

// GetModerationQueue returns articles pending moderation
func (s *ModerationService) GetModerationQueue(ctx context.Context, status string, page, limit int) ([]dto.ModerationQueueItem, int64, error) {
	articles, total, err := s.moderationRepo.GetArticlesPendingModeration(ctx, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var items []dto.ModerationQueueItem
	for _, article := range articles {
		// Get associated flags for reasons
		flags, _ := s.moderationRepo.GetContentFlags(ctx, "article", article.ID)
		var flagReasons []string
		var severity string
		for _, flag := range flags {
			if !flag.IsResolved {
				flagReasons = append(flagReasons, flag.AIReason)
				severity = string(flag.Severity)
			}
		}

		excerpt := article.Content
		if len(excerpt) > 200 {
			excerpt = excerpt[:200] + "..."
		}

		authorName := ""
		if article.Author != nil {
			authorName = article.Author.Name
		}

		items = append(items, dto.ModerationQueueItem{
			ID:               article.ID,
			Title:            article.Title,
			Excerpt:          excerpt,
			AuthorID:         article.UserID,
			AuthorName:       authorName,
			ModerationStatus: article.ModerationStatus,
			Severity:         severity,
			FlagReasons:      flagReasons,
			CreatedAt:        article.CreatedAt,
			UpdatedAt:        article.UpdatedAt,
		})
	}

	return items, total, nil
}

// ModerateArticle performs moderator action on an article
func (s *ModerationService) ModerateArticle(ctx context.Context, articleID, moderatorID uint, req *dto.ModerateArticleRequest) error {
	article, err := s.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return errors.New("article not found")
	}

	// Determine new status based on action
	var newStatus model.ArticleModerationStatus
	var actionType model.ModeratorActionType
	switch req.Action {
	case "approve":
		newStatus = model.ArticleModerationApproved
		actionType = model.ModeratorActionArticleApproved
	case "reject":
		newStatus = model.ArticleModerationRejected
		actionType = model.ModeratorActionArticleRejected
	case "request_edit":
		newStatus = model.ArticleModerationRevisionNeeded
		actionType = model.ModeratorActionArticleRequestEdit
	default:
		return errors.New("invalid action")
	}

	// Store previous state for audit
	prevState, _ := json.Marshal(map[string]interface{}{
		"status":            article.Status,
		"moderation_status": article.ModerationStatus,
	})

	// Update article moderation
	err = s.moderationRepo.UpdateArticleModeration(ctx, articleID, newStatus, req.Notes, moderatorID, req.TriggerWarnings)
	if err != nil {
		return err
	}

	// Resolve associated flags
	flags, _ := s.moderationRepo.GetContentFlags(ctx, "article", articleID)
	for _, flag := range flags {
		if !flag.IsResolved {
			_ = s.moderationRepo.ResolveFlag(ctx, flag.ID, moderatorID, "Resolved via moderation action: "+req.Action)
		}
	}

	// Store new state for audit
	newState, _ := json.Marshal(map[string]interface{}{
		"status":            article.Status,
		"moderation_status": newStatus,
	})

	// Log moderator action
	action := &model.ModeratorAction{
		ModeratorID:   moderatorID,
		ActionType:    actionType,
		TargetType:    "article",
		TargetID:      articleID,
		PreviousState: string(prevState),
		NewState:      string(newState),
		Reason:        req.Notes,
	}
	return s.moderationRepo.CreateModeratorAction(ctx, action)
}

// ========================
// Reports Management
// ========================

// CreateReport creates a new user report
func (s *ModerationService) CreateReport(ctx context.Context, reporterID uint, req *dto.CreateReportRequest) (*model.UserReport, error) {
	// Check for duplicate report
	isDuplicate, err := s.moderationRepo.CheckDuplicateReport(ctx, reporterID, req.ReportType, req.ContentID, req.UserID)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		return nil, errors.New("you have already reported this content")
	}

	report := &model.UserReport{
		ReporterID:  reporterID,
		ReportType:  model.ReportType(req.ReportType),
		Reason:      model.ReportReason(req.Reason),
		Description: req.Description,
		Status:      model.ReportStatusPending,
	}

	if req.ContentID != nil {
		report.ReportedContentID = req.ContentID
	}
	if req.UserID != nil {
		report.ReportedUserID = req.UserID
	}

	if err := s.moderationRepo.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// GetReports returns reports for moderator dashboard
func (s *ModerationService) GetReports(ctx context.Context, params dto.ReportQueryParams) ([]dto.ReportDTO, int64, error) {
	reports, total, err := s.moderationRepo.GetReports(ctx, params.Status, params.ReportType, params.Reason, params.Page, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	var dtos []dto.ReportDTO
	for _, report := range reports {
		dto := dto.ReportDTO{
			ID:                report.ID,
			ReporterID:        report.ReporterID,
			ReportType:        string(report.ReportType),
			ReportedContentID: report.ReportedContentID,
			ReportedUserID:    report.ReportedUserID,
			Reason:            string(report.Reason),
			Description:       report.Description,
			Status:            string(report.Status),
			HandledByID:       report.HandledByID,
			HandledAt:         report.HandledAt,
			ActionTaken:       string(report.ActionTaken),
			ModeratorNotes:    report.ModeratorNotes,
			CreatedAt:         report.CreatedAt,
		}

		if report.Reporter != nil {
			dto.ReporterName = report.Reporter.Name
		}
		if report.ReportedUser != nil {
			dto.ReportedUserName = report.ReportedUser.Name
		}
		if report.HandledBy != nil {
			dto.HandledByName = report.HandledBy.Name
		}

		// Get content preview if applicable
		if report.ReportedContentID != nil {
			switch report.ReportType {
			case model.ReportTypeArticle:
				if article, err := s.articleRepo.FindByID(ctx, *report.ReportedContentID); err == nil {
					dto.ContentTitle = article.Title
					dto.ContentPreview = article.Content
					if len(dto.ContentPreview) > 200 {
						dto.ContentPreview = dto.ContentPreview[:200] + "..."
					}
				}
			}
		}

		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// HandleReport processes a report with moderator action
func (s *ModerationService) HandleReport(ctx context.Context, reportID, moderatorID uint, req *dto.HandleReportRequest) error {
	report, err := s.moderationRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return errors.New("report not found")
	}

	if report.Status == model.ReportStatusResolved || report.Status == model.ReportStatusDismissed {
		return errors.New("report already handled")
	}

	now := time.Now()
	report.HandledByID = &moderatorID
	report.HandledAt = &now
	report.ModeratorNotes = req.Notes

	var targetUserID *uint
	if report.ReportedUserID != nil {
		targetUserID = report.ReportedUserID
	}

	switch req.Action {
	case "dismiss":
		report.Status = model.ReportStatusDismissed
		report.ActionTaken = model.ActionTakenNone

	case "warn":
		report.Status = model.ReportStatusResolved
		report.ActionTaken = model.ActionTakenUserWarned
		if targetUserID != nil {
			// Create strike for the user
			strike := &model.UserStrike{
				UserID:     *targetUserID,
				ReportID:   &reportID,
				Reason:     "Warning: " + req.Notes,
				Severity:   model.StrikeSeverityWarning,
				IssuedByID: moderatorID,
				IsActive:   true,
			}
			_ = s.moderationRepo.CreateStrike(ctx, strike)
		}

	case "remove_content":
		report.Status = model.ReportStatusResolved
		report.ActionTaken = model.ActionTakenContentRemoved
		if report.ReportedContentID != nil {
			// Remove/block the content based on type
			switch report.ReportType {
			case model.ReportTypeArticle:
				_ = s.articleRepo.UpdateStatus(ctx, *report.ReportedContentID, model.ArticleStatusBlocked)
			case model.ReportTypeForum:
				_ = s.moderationRepo.FlagForum(ctx, *report.ReportedContentID, "Removed due to report: "+req.Notes)
			case model.ReportTypeForumPost:
				_ = s.moderationRepo.FlagForumPost(ctx, *report.ReportedContentID, "Removed due to report: "+req.Notes)
			}
		}

	case "suspend":
		report.Status = model.ReportStatusResolved
		report.ActionTaken = model.ActionTakenUserSuspended
		if targetUserID != nil {
			duration := time.Duration(req.Duration) * 24 * time.Hour
			endTime := now.Add(duration)
			_ = s.moderationRepo.SuspendUser(ctx, *targetUserID, endTime, req.Notes)

			// Create major strike
			strike := &model.UserStrike{
				UserID:     *targetUserID,
				ReportID:   &reportID,
				Reason:     fmt.Sprintf("Suspended for %d days: %s", req.Duration, req.Notes),
				Severity:   model.StrikeSeverityMajor,
				IssuedByID: moderatorID,
				IsActive:   true,
			}
			_ = s.moderationRepo.CreateStrike(ctx, strike)
		}

	case "ban":
		report.Status = model.ReportStatusResolved
		report.ActionTaken = model.ActionTakenUserBanned
		if targetUserID != nil {
			_ = s.moderationRepo.BanUser(ctx, *targetUserID, req.Notes)

			// Create major strike
			strike := &model.UserStrike{
				UserID:     *targetUserID,
				ReportID:   &reportID,
				Reason:     "Banned: " + req.Notes,
				Severity:   model.StrikeSeverityMajor,
				IssuedByID: moderatorID,
				IsActive:   true,
			}
			_ = s.moderationRepo.CreateStrike(ctx, strike)
		}

	default:
		return errors.New("invalid action")
	}

	if err := s.moderationRepo.UpdateReport(ctx, report); err != nil {
		return err
	}

	// Log moderator action
	action := &model.ModeratorAction{
		ModeratorID: moderatorID,
		ActionType:  model.ModeratorActionType("report_" + req.Action),
		TargetType:  "report",
		TargetID:    reportID,
		Reason:      req.Notes,
	}
	return s.moderationRepo.CreateModeratorAction(ctx, action)
}

// ========================
// User Blocking
// ========================

// BlockUser blocks another user
func (s *ModerationService) BlockUser(ctx context.Context, blockerID, blockedID uint, reason string) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	// Check if already blocked
	isBlocked, _ := s.moderationRepo.IsBlocked(ctx, blockerID, blockedID)
	if isBlocked {
		return errors.New("user already blocked")
	}

	block := &model.UserBlock{
		BlockerID: blockerID,
		BlockedID: blockedID,
		Reason:    reason,
	}

	return s.moderationRepo.CreateBlock(ctx, block)
}

// UnblockUser removes a block
func (s *ModerationService) UnblockUser(ctx context.Context, blockerID, blockedID uint) error {
	return s.moderationRepo.DeleteBlock(ctx, blockerID, blockedID)
}

// GetBlockedUsers returns list of users blocked by a user
func (s *ModerationService) GetBlockedUsers(ctx context.Context, userID uint) ([]dto.BlockDTO, int64, error) {
	blocks, err := s.moderationRepo.GetBlocksByBlocker(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	count, _ := s.moderationRepo.GetBlocksCount(ctx, userID)

	var dtos []dto.BlockDTO
	for _, block := range blocks {
		dto := dto.BlockDTO{
			ID:        block.ID,
			BlockerID: block.BlockerID,
			BlockedID: block.BlockedID,
			Reason:    block.Reason,
			CreatedAt: block.CreatedAt,
		}
		if block.Blocked != nil {
			dto.BlockedName = block.Blocked.Name
			dto.BlockedAvatar = block.Blocked.Avatar
		}
		dtos = append(dtos, dto)
	}

	return dtos, count, nil
}

// IsUserBlocked checks if userB is blocked by userA
func (s *ModerationService) IsUserBlocked(ctx context.Context, blockerID, blockedID uint) (bool, error) {
	return s.moderationRepo.IsBlocked(ctx, blockerID, blockedID)
}

// GetBlockedUserIDs returns IDs of users blocked by a user
func (s *ModerationService) GetBlockedUserIDs(ctx context.Context, userID uint) ([]uint, error) {
	return s.moderationRepo.GetBlockedUserIDs(ctx, userID)
}

// ========================
// User Strikes
// ========================

// GetUserStrikes returns strikes for a user
func (s *ModerationService) GetUserStrikes(ctx context.Context, userID uint, activeOnly bool) ([]dto.StrikeDTO, error) {
	strikes, err := s.moderationRepo.GetUserStrikes(ctx, userID, activeOnly)
	if err != nil {
		return nil, err
	}

	var dtos []dto.StrikeDTO
	for _, strike := range strikes {
		dto := dto.StrikeDTO{
			ID:         strike.ID,
			UserID:     strike.UserID,
			ReportID:   strike.ReportID,
			Reason:     strike.Reason,
			Severity:   string(strike.Severity),
			IssuedByID: strike.IssuedByID,
			ExpiresAt:  strike.ExpiresAt,
			IsActive:   strike.IsActive,
			Notes:      strike.Notes,
			CreatedAt:  strike.CreatedAt,
		}
		if strike.User != nil {
			dto.UserName = strike.User.Name
		}
		if strike.IssuedBy != nil {
			dto.IssuedByName = strike.IssuedBy.Name
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

// CheckAutoSuspension checks if user should be auto-suspended based on strikes
func (s *ModerationService) CheckAutoSuspension(ctx context.Context, userID uint) (bool, error) {
	count, err := s.moderationRepo.GetActiveStrikesCount(ctx, userID)
	if err != nil {
		return false, err
	}

	// Auto-suspend after 3 strikes
	if count >= 3 {
		endTime := time.Now().Add(7 * 24 * time.Hour) // 7 days
		return true, s.moderationRepo.SuspendUser(ctx, userID, endTime, "Auto-suspended: 3 strikes reached")
	}

	return false, nil
}

// ========================
// Trigger Warnings
// ========================

// AddTriggerWarnings adds trigger warnings to content
func (s *ModerationService) AddTriggerWarnings(ctx context.Context, contentType string, contentID, moderatorID uint, warnings []string) error {
	switch contentType {
	case "article":
		err := s.moderationRepo.UpdateArticleModeration(ctx, contentID, model.ArticleModerationApproved, "", moderatorID, warnings)
		if err != nil {
			return err
		}
	case "forum":
		err := s.moderationRepo.UpdateForumTriggerWarnings(ctx, contentID, warnings)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid content type")
	}

	// Log action
	action := &model.ModeratorAction{
		ModeratorID: moderatorID,
		ActionType:  model.ModeratorActionTriggerWarningAdded,
		TargetType:  contentType,
		TargetID:    contentID,
		NewState:    fmt.Sprintf("%v", warnings),
	}
	return s.moderationRepo.CreateModeratorAction(ctx, action)
}

// ========================
// AI Disclaimer & Preferences
// ========================

// AcceptAIDisclaimer marks that user has accepted the AI disclaimer
func (s *ModerationService) AcceptAIDisclaimer(ctx context.Context, userID uint) error {
	return s.moderationRepo.SetAIDisclaimerAccepted(ctx, userID, true)
}

// UpdateContentWarningPreference updates user's content warning preference
func (s *ModerationService) UpdateContentWarningPreference(ctx context.Context, userID uint, preference string) error {
	pref := model.ContentWarningPreference(preference)
	return s.moderationRepo.SetContentWarningPreference(ctx, userID, pref)
}

// ========================
// Crisis Detection
// ========================

// DetectCrisis checks a message for crisis indicators
func (s *ModerationService) DetectCrisis(ctx context.Context, message string) (*model.CrisisDetectionResult, error) {
	return s.aiModerationService.DetectCrisis(ctx, message)
}

// ========================
// Moderation Statistics
// ========================

// GetModerationStats returns dashboard statistics
func (s *ModerationService) GetModerationStats(ctx context.Context) (*dto.ModerationStatsDTO, error) {
	pendingArticles, err := s.moderationRepo.GetPendingArticlesCount(ctx)
	if err != nil {
		return nil, err
	}
	flaggedArticles, err := s.moderationRepo.GetFlaggedArticlesCount(ctx)
	if err != nil {
		return nil, err
	}
	pendingReports, err := s.moderationRepo.GetPendingReportsCount(ctx)
	if err != nil {
		return nil, err
	}
	resolvedToday, err := s.moderationRepo.GetResolvedReportsTodayCount(ctx, nil)
	if err != nil {
		return nil, err
	}
	activeStrikes, err := s.moderationRepo.GetAllActiveStrikesCount(ctx)
	if err != nil {
		return nil, err
	}
	suspendedUsers, err := s.moderationRepo.GetSuspendedUsersCount(ctx)
	if err != nil {
		return nil, err
	}
	bannedUsers, err := s.moderationRepo.GetBannedUsersCount(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.ModerationStatsDTO{
		PendingArticles:      pendingArticles,
		FlaggedArticles:      flaggedArticles,
		PendingReports:       pendingReports,
		ResolvedReportsToday: resolvedToday,
		ActiveStrikes:        activeStrikes,
		SuspendedUsers:       suspendedUsers,
		BannedUsers:          bannedUsers,
	}, nil
}

// ========================
// Moderator Actions Log
// ========================

// GetModeratorActions returns audit log of moderator actions
func (s *ModerationService) GetModeratorActions(ctx context.Context, params dto.ModeratorActionQueryParams) ([]dto.ModeratorActionDTO, int64, error) {
	var moderatorIDPtr *uint
	if params.ModeratorID > 0 {
		moderatorIDPtr = &params.ModeratorID
	}

	actions, total, err := s.moderationRepo.GetModeratorActions(ctx, moderatorIDPtr, params.ActionType, params.TargetType, params.Page, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	var dtos []dto.ModeratorActionDTO
	for _, action := range actions {
		dto := dto.ModeratorActionDTO{
			ID:            action.ID,
			ModeratorID:   action.ModeratorID,
			ActionType:    string(action.ActionType),
			TargetType:    action.TargetType,
			TargetID:      action.TargetID,
			PreviousState: action.PreviousState,
			NewState:      action.NewState,
			Reason:        action.Reason,
			Notes:         action.Notes,
			CreatedAt:     action.CreatedAt,
		}
		if action.Moderator != nil {
			dto.ModeratorName = action.Moderator.Name
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// ========================
// Crisis Keywords Management
// ========================

// GetCrisisKeywords returns all crisis keywords
func (s *ModerationService) GetCrisisKeywords(ctx context.Context) ([]dto.CrisisKeywordDTO, error) {
	keywords, err := s.moderationRepo.GetAllCrisisKeywords(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []dto.CrisisKeywordDTO
	for _, kw := range keywords {
		dtos = append(dtos, dto.CrisisKeywordDTO{
			ID:       kw.ID,
			Keyword:  kw.Keyword,
			Category: string(kw.Category),
			Severity: string(kw.Severity),
			Language: kw.Language,
			IsActive: kw.IsActive,
			Notes:    kw.Notes,
		})
	}

	return dtos, nil
}

// CreateCrisisKeyword creates a new crisis keyword
func (s *ModerationService) CreateCrisisKeyword(ctx context.Context, req *dto.CreateCrisisKeywordRequest) (*model.CrisisKeyword, error) {
	language := req.Language
	if language == "" {
		language = "id"
	}

	keyword := &model.CrisisKeyword{
		Keyword:  req.Keyword,
		Category: model.CrisisCategory(req.Category),
		Severity: model.CrisisSeverity(req.Severity),
		Language: language,
		IsActive: true,
		Notes:    req.Notes,
	}

	if err := s.moderationRepo.CreateCrisisKeyword(ctx, keyword); err != nil {
		return nil, err
	}

	return keyword, nil
}

// DeleteCrisisKeyword deletes a crisis keyword
// DeleteCrisisKeyword deletes a crisis keyword
func (s *ModerationService) DeleteCrisisKeyword(ctx context.Context, id uint) error {
	return s.moderationRepo.DeleteCrisisKeyword(ctx, id)
}

// ========================
// Appeal Management
// ========================

// CreateAppeal creates a new appeal for a user
func (s *ModerationService) CreateAppeal(ctx context.Context, userID uint, req *dto.CreateAppealRequest) (*model.Appeal, error) {
	// Check if there is already a pending appeal
	appeals, _, err := s.moderationRepo.GetAppeals(ctx, model.AppealStatusPending, userID, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(appeals) > 0 {
		return nil, errors.New("you already have a pending appeal")
	}

	appeal := &model.Appeal{
		UserID:   userID,
		Reason:   req.Reason,
		Evidence: req.Evidence,
		Status:   model.AppealStatusPending,
	}

	if err := s.moderationRepo.CreateAppeal(ctx, appeal); err != nil {
		return nil, err
	}

	return appeal, nil
}

// GetAppeals returns appeals for moderator dashboard
func (s *ModerationService) GetAppeals(ctx context.Context, status string, page, limit int) ([]dto.AppealDTO, int64, error) {
	appeals, total, err := s.moderationRepo.GetAppeals(ctx, model.AppealStatus(status), 0, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var dtos []dto.AppealDTO
	for _, appeal := range appeals {
		dto := dto.AppealDTO{
			ID:            appeal.ID,
			UserID:        appeal.UserID,
			Reason:        appeal.Reason,
			Evidence:      appeal.Evidence,
			Status:        string(appeal.Status),
			ReviewerNotes: appeal.ReviewerNotes,
			ReviewerID:    appeal.ReviewerID,
			CreatedAt:     appeal.CreatedAt,
			UpdatedAt:     appeal.UpdatedAt,
		}

		// User info not directly available in model structs slice in some ORM configs unless preloaded
		// Repo preloads "User" and "Reviewer".
		dto.UserName = appeal.User.Name
		dto.UserEmail = appeal.User.Email
		if appeal.Reviewer != nil {
			dto.ReviewerName = appeal.Reviewer.Name
		}

		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// ReviewAppeal processes an appeal
func (s *ModerationService) ReviewAppeal(ctx context.Context, appealID, reviewerID uint, req *dto.ReviewAppealRequest) error {
	appeal, err := s.moderationRepo.GetAppealByID(ctx, appealID)
	if err != nil {
		return errors.New("appeal not found")
	}

	if appeal.Status != model.AppealStatusPending {
		return errors.New("appeal already reviewed")
	}

	newStatus := model.AppealStatus(req.Status)
	appeal.Status = newStatus
	appeal.ReviewerID = &reviewerID
	appeal.ReviewerNotes = req.Notes

	if err := s.moderationRepo.UpdateAppeal(ctx, appeal); err != nil {
		return err
	}

	// If approved, lift restrictions
	if newStatus == model.AppealStatusApproved {
		// Lift ban
		_ = s.moderationRepo.UnbanUser(ctx, appeal.UserID)
		// Lift suspension
		_ = s.moderationRepo.UnsuspendUser(ctx, appeal.UserID)
	}

	// Log action
	action := &model.ModeratorAction{
		ModeratorID: reviewerID,
		ActionType:  model.ModeratorActionType("appeal_" + req.Status),
		TargetType:  "appeal",
		TargetID:    appealID,
		Reason:      req.Notes,
	}
	return s.moderationRepo.CreateModeratorAction(ctx, action)
}
