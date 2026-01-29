package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type ModerationService struct {
	moderationRepo      *repositories.ModerationRepository
	userRepo            *repositories.UserRepository
	articleRepo         *repositories.ArticleRepository
	forumRepo           repositories.ForumRepository
	aiModerationService *AIModerationService
}

func NewModerationService(
	moderationRepo *repositories.ModerationRepository,
	userRepo *repositories.UserRepository,
	articleRepo *repositories.ArticleRepository,
	forumRepo repositories.ForumRepository,
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
func (s *ModerationService) ModerateNewArticle(article *models.Article) (*dto.AIModerationResult, error) {
	result, err := s.aiModerationService.ModerateArticle(article.Title, article.Content)
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
	warnings, _ := s.aiModerationService.DetectTriggerWarnings(article.Content)
	if len(warnings) > 0 {
		article.TriggerWarnings = models.TriggerWarnings(warnings)
	}

	// Create content flag if flagged or rejected
	if result.Status == models.ArticleModerationFlagged || result.Status == models.ArticleModerationRejected {
		confidence := result.Confidence
		flag := &models.ContentFlag{
			ContentType:  "article",
			ContentID:    article.ID,
			FlagType:     models.ContentFlagTypeAIModeration,
			FlagCategory: models.ContentFlagCategory(result.FlagCategory),
			Severity:     models.FlagSeverity(result.Severity),
			AIConfidence: &confidence,
			AIReason:     fmt.Sprintf("%v", result.Reasons),
		}
		_ = s.moderationRepo.CreateContentFlag(flag)
	}

	return result, nil
}

// GetModerationQueue returns articles pending moderation
func (s *ModerationService) GetModerationQueue(status string, page, limit int) ([]dto.ModerationQueueItem, int64, error) {
	articles, total, err := s.moderationRepo.GetArticlesPendingModeration(status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	var items []dto.ModerationQueueItem
	for _, article := range articles {
		// Get associated flags for reasons
		flags, _ := s.moderationRepo.GetContentFlags("article", article.ID)
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
func (s *ModerationService) ModerateArticle(articleID, moderatorID uint, req *dto.ModerateArticleRequest) error {
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return errors.New("article not found")
	}

	// Determine new status based on action
	var newStatus models.ArticleModerationStatus
	var actionType models.ModeratorActionType
	switch req.Action {
	case "approve":
		newStatus = models.ArticleModerationApproved
		actionType = models.ModeratorActionArticleApproved
	case "reject":
		newStatus = models.ArticleModerationRejected
		actionType = models.ModeratorActionArticleRejected
	case "request_edit":
		newStatus = models.ArticleModerationRevisionNeeded
		actionType = models.ModeratorActionArticleRequestEdit
	default:
		return errors.New("invalid action")
	}

	// Store previous state for audit
	prevState, _ := json.Marshal(map[string]interface{}{
		"status":            article.Status,
		"moderation_status": article.ModerationStatus,
	})

	// Update article moderation
	err = s.moderationRepo.UpdateArticleModeration(articleID, newStatus, req.Notes, moderatorID, req.TriggerWarnings)
	if err != nil {
		return err
	}

	// Resolve associated flags
	flags, _ := s.moderationRepo.GetContentFlags("article", articleID)
	for _, flag := range flags {
		if !flag.IsResolved {
			_ = s.moderationRepo.ResolveFlag(flag.ID, moderatorID, "Resolved via moderation action: "+req.Action)
		}
	}

	// Store new state for audit
	newState, _ := json.Marshal(map[string]interface{}{
		"status":            article.Status,
		"moderation_status": newStatus,
	})

	// Log moderator action
	action := &models.ModeratorAction{
		ModeratorID:   moderatorID,
		ActionType:    actionType,
		TargetType:    "article",
		TargetID:      articleID,
		PreviousState: string(prevState),
		NewState:      string(newState),
		Reason:        req.Notes,
	}
	return s.moderationRepo.CreateModeratorAction(action)
}

// ========================
// Reports Management
// ========================

// CreateReport creates a new user report
func (s *ModerationService) CreateReport(reporterID uint, req *dto.CreateReportRequest) (*models.UserReport, error) {
	// Check for duplicate report
	isDuplicate, err := s.moderationRepo.CheckDuplicateReport(reporterID, req.ReportType, req.ContentID, req.UserID)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		return nil, errors.New("you have already reported this content")
	}

	report := &models.UserReport{
		ReporterID:  reporterID,
		ReportType:  models.ReportType(req.ReportType),
		Reason:      models.ReportReason(req.Reason),
		Description: req.Description,
		Status:      models.ReportStatusPending,
	}

	if req.ContentID != nil {
		report.ReportedContentID = req.ContentID
	}
	if req.UserID != nil {
		report.ReportedUserID = req.UserID
	}

	if err := s.moderationRepo.CreateReport(report); err != nil {
		return nil, err
	}

	return report, nil
}

// GetReports returns reports for moderator dashboard
func (s *ModerationService) GetReports(params dto.ReportQueryParams) ([]dto.ReportDTO, int64, error) {
	reports, total, err := s.moderationRepo.GetReports(params.Status, params.ReportType, params.Reason, params.Page, params.Limit)
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
			case models.ReportTypeArticle:
				if article, err := s.articleRepo.FindByID(*report.ReportedContentID); err == nil {
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
func (s *ModerationService) HandleReport(reportID, moderatorID uint, req *dto.HandleReportRequest) error {
	report, err := s.moderationRepo.GetReportByID(reportID)
	if err != nil {
		return errors.New("report not found")
	}

	if report.Status == models.ReportStatusResolved || report.Status == models.ReportStatusDismissed {
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
		report.Status = models.ReportStatusDismissed
		report.ActionTaken = models.ActionTakenNone

	case "warn":
		report.Status = models.ReportStatusResolved
		report.ActionTaken = models.ActionTakenUserWarned
		if targetUserID != nil {
			// Create strike for the user
			strike := &models.UserStrike{
				UserID:     *targetUserID,
				ReportID:   &reportID,
				Reason:     "Warning: " + req.Notes,
				Severity:   models.StrikeSeverityWarning,
				IssuedByID: moderatorID,
				IsActive:   true,
			}
			_ = s.moderationRepo.CreateStrike(strike)
		}

	case "remove_content":
		report.Status = models.ReportStatusResolved
		report.ActionTaken = models.ActionTakenContentRemoved
		if report.ReportedContentID != nil {
			// Remove/block the content based on type
			switch report.ReportType {
			case models.ReportTypeArticle:
				_ = s.articleRepo.UpdateStatus(*report.ReportedContentID, models.ArticleStatusBlocked)
			case models.ReportTypeForum:
				_ = s.moderationRepo.FlagForum(*report.ReportedContentID, "Removed due to report: "+req.Notes)
			case models.ReportTypeForumPost:
				_ = s.moderationRepo.FlagForumPost(*report.ReportedContentID, "Removed due to report: "+req.Notes)
			}
		}

	case "suspend":
		report.Status = models.ReportStatusResolved
		report.ActionTaken = models.ActionTakenUserSuspended
		if targetUserID != nil {
			duration := time.Duration(req.Duration) * 24 * time.Hour
			endTime := now.Add(duration)
			_ = s.moderationRepo.SuspendUser(*targetUserID, endTime, req.Notes)

			// Create major strike
			strike := &models.UserStrike{
				UserID:     *targetUserID,
				ReportID:   &reportID,
				Reason:     fmt.Sprintf("Suspended for %d days: %s", req.Duration, req.Notes),
				Severity:   models.StrikeSeverityMajor,
				IssuedByID: moderatorID,
				IsActive:   true,
			}
			_ = s.moderationRepo.CreateStrike(strike)
		}

	case "ban":
		report.Status = models.ReportStatusResolved
		report.ActionTaken = models.ActionTakenUserBanned
		if targetUserID != nil {
			_ = s.moderationRepo.BanUser(*targetUserID, req.Notes)

			// Create major strike
			strike := &models.UserStrike{
				UserID:     *targetUserID,
				ReportID:   &reportID,
				Reason:     "Banned: " + req.Notes,
				Severity:   models.StrikeSeverityMajor,
				IssuedByID: moderatorID,
				IsActive:   true,
			}
			_ = s.moderationRepo.CreateStrike(strike)
		}

	default:
		return errors.New("invalid action")
	}

	if err := s.moderationRepo.UpdateReport(report); err != nil {
		return err
	}

	// Log moderator action
	action := &models.ModeratorAction{
		ModeratorID: moderatorID,
		ActionType:  models.ModeratorActionType("report_" + req.Action),
		TargetType:  "report",
		TargetID:    reportID,
		Reason:      req.Notes,
	}
	return s.moderationRepo.CreateModeratorAction(action)
}

// ========================
// User Blocking
// ========================

// BlockUser blocks another user
func (s *ModerationService) BlockUser(blockerID, blockedID uint, reason string) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	// Check if already blocked
	isBlocked, _ := s.moderationRepo.IsBlocked(blockerID, blockedID)
	if isBlocked {
		return errors.New("user already blocked")
	}

	block := &models.UserBlock{
		BlockerID: blockerID,
		BlockedID: blockedID,
		Reason:    reason,
	}

	return s.moderationRepo.CreateBlock(block)
}

// UnblockUser removes a block
func (s *ModerationService) UnblockUser(blockerID, blockedID uint) error {
	return s.moderationRepo.DeleteBlock(blockerID, blockedID)
}

// GetBlockedUsers returns list of users blocked by a user
func (s *ModerationService) GetBlockedUsers(userID uint) ([]dto.BlockDTO, int64, error) {
	blocks, err := s.moderationRepo.GetBlocksByBlocker(userID)
	if err != nil {
		return nil, 0, err
	}

	count, _ := s.moderationRepo.GetBlocksCount(userID)

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
func (s *ModerationService) IsUserBlocked(blockerID, blockedID uint) (bool, error) {
	return s.moderationRepo.IsBlocked(blockerID, blockedID)
}

// GetBlockedUserIDs returns IDs of users blocked by a user
func (s *ModerationService) GetBlockedUserIDs(userID uint) ([]uint, error) {
	return s.moderationRepo.GetBlockedUserIDs(userID)
}

// ========================
// User Strikes
// ========================

// GetUserStrikes returns strikes for a user
func (s *ModerationService) GetUserStrikes(userID uint, activeOnly bool) ([]dto.StrikeDTO, error) {
	strikes, err := s.moderationRepo.GetUserStrikes(userID, activeOnly)
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
func (s *ModerationService) CheckAutoSuspension(userID uint) (bool, error) {
	count, err := s.moderationRepo.GetActiveStrikesCount(userID)
	if err != nil {
		return false, err
	}

	// Auto-suspend after 3 strikes
	if count >= 3 {
		endTime := time.Now().Add(7 * 24 * time.Hour) // 7 days
		return true, s.moderationRepo.SuspendUser(userID, endTime, "Auto-suspended: 3 strikes reached")
	}

	return false, nil
}

// ========================
// Trigger Warnings
// ========================

// AddTriggerWarnings adds trigger warnings to content
func (s *ModerationService) AddTriggerWarnings(contentType string, contentID, moderatorID uint, warnings []string) error {
	switch contentType {
	case "article":
		err := s.moderationRepo.UpdateArticleModeration(contentID, models.ArticleModerationApproved, "", moderatorID, warnings)
		if err != nil {
			return err
		}
	case "forum":
		err := s.moderationRepo.UpdateForumTriggerWarnings(contentID, warnings)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid content type")
	}

	// Log action
	action := &models.ModeratorAction{
		ModeratorID: moderatorID,
		ActionType:  models.ModeratorActionTriggerWarningAdded,
		TargetType:  contentType,
		TargetID:    contentID,
		NewState:    fmt.Sprintf("%v", warnings),
	}
	return s.moderationRepo.CreateModeratorAction(action)
}

// ========================
// AI Disclaimer & Preferences
// ========================

// AcceptAIDisclaimer marks that user has accepted the AI disclaimer
func (s *ModerationService) AcceptAIDisclaimer(userID uint) error {
	return s.moderationRepo.SetAIDisclaimerAccepted(userID, true)
}

// UpdateContentWarningPreference updates user's content warning preference
func (s *ModerationService) UpdateContentWarningPreference(userID uint, preference string) error {
	pref := models.ContentWarningPreference(preference)
	return s.moderationRepo.SetContentWarningPreference(userID, pref)
}

// ========================
// Crisis Detection
// ========================

// DetectCrisis checks a message for crisis indicators
func (s *ModerationService) DetectCrisis(message string) (*models.CrisisDetectionResult, error) {
	return s.aiModerationService.DetectCrisis(message)
}

// ========================
// Moderation Statistics
// ========================

// GetModerationStats returns dashboard statistics
func (s *ModerationService) GetModerationStats() (*dto.ModerationStatsDTO, error) {
	pendingArticles, _ := s.moderationRepo.GetPendingArticlesCount()
	flaggedArticles, _ := s.moderationRepo.GetFlaggedArticlesCount()
	pendingReports, _ := s.moderationRepo.GetPendingReportsCount()
	resolvedToday, _ := s.moderationRepo.GetResolvedReportsTodayCount(nil)
	activeStrikes, _ := s.moderationRepo.GetAllActiveStrikesCount()
	suspendedUsers, _ := s.moderationRepo.GetSuspendedUsersCount()
	bannedUsers, _ := s.moderationRepo.GetBannedUsersCount()

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
func (s *ModerationService) GetModeratorActions(params dto.ModeratorActionQueryParams) ([]dto.ModeratorActionDTO, int64, error) {
	var moderatorIDPtr *uint
	if params.ModeratorID > 0 {
		moderatorIDPtr = &params.ModeratorID
	}

	actions, total, err := s.moderationRepo.GetModeratorActions(moderatorIDPtr, params.ActionType, params.TargetType, params.Page, params.Limit)
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
func (s *ModerationService) GetCrisisKeywords() ([]dto.CrisisKeywordDTO, error) {
	keywords, err := s.moderationRepo.GetAllCrisisKeywords()
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
func (s *ModerationService) CreateCrisisKeyword(req *dto.CreateCrisisKeywordRequest) (*models.CrisisKeyword, error) {
	language := req.Language
	if language == "" {
		language = "id"
	}

	keyword := &models.CrisisKeyword{
		Keyword:  req.Keyword,
		Category: models.CrisisCategory(req.Category),
		Severity: models.CrisisSeverity(req.Severity),
		Language: language,
		IsActive: true,
		Notes:    req.Notes,
	}

	if err := s.moderationRepo.CreateCrisisKeyword(keyword); err != nil {
		return nil, err
	}

	return keyword, nil
}

// DeleteCrisisKeyword deletes a crisis keyword
func (s *ModerationService) DeleteCrisisKeyword(id uint) error {
	return s.moderationRepo.DeleteCrisisKeyword(id)
}
