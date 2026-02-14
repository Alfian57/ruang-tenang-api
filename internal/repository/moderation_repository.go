package repository

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type ModerationRepository struct {
	db *gorm.DB
}

func NewModerationRepository(db *gorm.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

// ========================
// Content Flag Operations
// ========================

func (r *ModerationRepository) CreateContentFlag(flag *model.ContentFlag) error {
	return r.db.Create(flag).Error
}

func (r *ModerationRepository) GetContentFlagByID(id uint) (*model.ContentFlag, error) {
	var flag model.ContentFlag
	err := r.db.Preload("FlaggedBy").Preload("ResolvedBy").First(&flag, id).Error
	return &flag, err
}

func (r *ModerationRepository) GetContentFlags(contentType string, contentID uint) ([]model.ContentFlag, error) {
	var flags []model.ContentFlag
	query := r.db.Preload("FlaggedBy").Preload("ResolvedBy")
	if contentType != "" && contentID > 0 {
		query = query.Where("content_type = ? AND content_id = ?", contentType, contentID)
	}
	err := query.Find(&flags).Error
	return flags, err
}

func (r *ModerationRepository) GetUnresolvedFlags(page, limit int) ([]model.ContentFlag, int64, error) {
	var flags []model.ContentFlag
	var total int64

	query := r.db.Model(&model.ContentFlag{}).Where("is_resolved = ?", false)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("FlaggedBy").
		Order("CASE severity WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END").
		Order("created_at DESC").
		Offset(offset).Limit(limit).Find(&flags).Error

	return flags, total, err
}

func (r *ModerationRepository) ResolveFlag(id, resolvedByID uint, notes string) error {
	now := time.Now()
	return r.db.Model(&model.ContentFlag{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_resolved":      true,
		"resolved_by_id":   resolvedByID,
		"resolved_at":      now,
		"resolution_notes": notes,
	}).Error
}

func (r *ModerationRepository) DeleteContentFlag(id uint) error {
	return r.db.Delete(&model.ContentFlag{}, id).Error
}

// ========================
// User Report Operations
// ========================

func (r *ModerationRepository) CreateReport(report *model.UserReport) error {
	return r.db.Create(report).Error
}

func (r *ModerationRepository) GetReportByID(id uint) (*model.UserReport, error) {
	var report model.UserReport
	err := r.db.Preload("Reporter").Preload("ReportedUser").Preload("HandledBy").First(&report, id).Error
	return &report, err
}

func (r *ModerationRepository) GetReports(status, reportType, reason string, page, limit int) ([]model.UserReport, int64, error) {
	var reports []model.UserReport
	var total int64

	query := r.db.Model(&model.UserReport{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}
	if reason != "" {
		query = query.Where("reason = ?", reason)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Reporter").Preload("ReportedUser").Preload("HandledBy").
		Order("created_at DESC").
		Offset(offset).Limit(limit).Find(&reports).Error

	return reports, total, err
}

func (r *ModerationRepository) GetPendingReportsCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.UserReport{}).Where("status = ?", model.ReportStatusPending).Count(&count).Error
	return count, err
}

func (r *ModerationRepository) UpdateReport(report *model.UserReport) error {
	return r.db.Save(report).Error
}

func (r *ModerationRepository) CheckDuplicateReport(reporterID uint, reportType string, contentID, userID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&model.UserReport{}).
		Where("reporter_id = ? AND report_type = ? AND status IN (?, ?)",
			reporterID, reportType, model.ReportStatusPending, model.ReportStatusReviewing)

	if contentID != nil {
		query = query.Where("reported_content_id = ?", *contentID)
	}
	if userID != nil {
		query = query.Where("reported_user_id = ?", *userID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *ModerationRepository) GetResolvedReportsTodayCount(moderatorID *uint) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	query := r.db.Model(&model.UserReport{}).
		Where("status = ? AND handled_at >= ?", model.ReportStatusResolved, today)

	if moderatorID != nil {
		query = query.Where("handled_by_id = ?", *moderatorID)
	}

	err := query.Count(&count).Error
	return count, err
}

// ========================
// User Block Operations
// ========================

func (r *ModerationRepository) CreateBlock(block *model.UserBlock) error {
	return r.db.Create(block).Error
}

func (r *ModerationRepository) GetBlockByPair(blockerID, blockedID uint) (*model.UserBlock, error) {
	var block model.UserBlock
	err := r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).First(&block).Error
	return &block, err
}

func (r *ModerationRepository) GetBlocksByBlocker(blockerID uint) ([]model.UserBlock, error) {
	var blocks []model.UserBlock
	err := r.db.Preload("Blocked").Where("blocker_id = ?", blockerID).Order("created_at DESC").Find(&blocks).Error
	return blocks, err
}

func (r *ModerationRepository) GetBlockedUserIDs(blockerID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.UserBlock{}).Where("blocker_id = ?", blockerID).Pluck("blocked_id", &ids).Error
	return ids, err
}

func (r *ModerationRepository) IsBlocked(blockerID, blockedID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&count).Error
	return count > 0, err
}

func (r *ModerationRepository) DeleteBlock(blockerID, blockedID uint) error {
	return r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&model.UserBlock{}).Error
}

func (r *ModerationRepository) GetBlocksCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserBlock{}).Where("blocker_id = ?", userID).Count(&count).Error
	return count, err
}

// ========================
// User Strike Operations
// ========================

func (r *ModerationRepository) CreateStrike(strike *model.UserStrike) error {
	return r.db.Create(strike).Error
}

func (r *ModerationRepository) GetStrikeByID(id uint) (*model.UserStrike, error) {
	var strike model.UserStrike
	err := r.db.Preload("User").Preload("IssuedBy").First(&strike, id).Error
	return &strike, err
}

func (r *ModerationRepository) GetUserStrikes(userID uint, activeOnly bool) ([]model.UserStrike, error) {
	var strikes []model.UserStrike
	query := r.db.Preload("IssuedBy").Where("user_id = ?", userID)
	if activeOnly {
		query = query.Where("is_active = ?", true).
			Where("expires_at IS NULL OR expires_at > ?", time.Now())
	}
	err := query.Order("created_at DESC").Find(&strikes).Error
	return strikes, err
}

func (r *ModerationRepository) GetActiveStrikesCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserStrike{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}

func (r *ModerationRepository) GetAllActiveStrikesCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.UserStrike{}).
		Where("is_active = ?", true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}

func (r *ModerationRepository) DeactivateStrike(id uint) error {
	return r.db.Model(&model.UserStrike{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *ModerationRepository) DeactivateExpiredStrikes() error {
	return r.db.Model(&model.UserStrike{}).
		Where("is_active = ? AND expires_at IS NOT NULL AND expires_at <= ?", true, time.Now()).
		Update("is_active", false).Error
}

// ========================
// Moderator Action Operations
// ========================

func (r *ModerationRepository) CreateModeratorAction(action *model.ModeratorAction) error {
	return r.db.Create(action).Error
}

func (r *ModerationRepository) GetModeratorActions(moderatorID *uint, actionType, targetType string, page, limit int) ([]model.ModeratorAction, int64, error) {
	var actions []model.ModeratorAction
	var total int64

	query := r.db.Model(&model.ModeratorAction{})
	if moderatorID != nil {
		query = query.Where("moderator_id = ?", *moderatorID)
	}
	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Moderator").
		Order("created_at DESC").
		Offset(offset).Limit(limit).Find(&actions).Error

	return actions, total, err
}

// ========================
// Crisis Keyword Operations
// ========================

func (r *ModerationRepository) GetActiveCrisisKeywords(language string) ([]model.CrisisKeyword, error) {
	var keywords []model.CrisisKeyword
	query := r.db.Where("is_active = ?", true)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Find(&keywords).Error
	return keywords, err
}

func (r *ModerationRepository) CreateCrisisKeyword(keyword *model.CrisisKeyword) error {
	return r.db.Create(keyword).Error
}

func (r *ModerationRepository) UpdateCrisisKeyword(keyword *model.CrisisKeyword) error {
	return r.db.Save(keyword).Error
}

func (r *ModerationRepository) DeleteCrisisKeyword(id uint) error {
	return r.db.Delete(&model.CrisisKeyword{}, id).Error
}

func (r *ModerationRepository) GetAllCrisisKeywords() ([]model.CrisisKeyword, error) {
	var keywords []model.CrisisKeyword
	err := r.db.Order("category, severity DESC").Find(&keywords).Error
	return keywords, err
}

// ========================
// Article Moderation Operations
// ========================

func (r *ModerationRepository) GetArticlesPendingModeration(status string, page, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).Where("is_user_generated = ?", true)
	if status != "" {
		query = query.Where("moderation_status = ?", status)
	} else {
		query = query.Where("moderation_status IN (?, ?)",
			model.ArticleModerationPending, model.ArticleModerationFlagged)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Author").Preload("Category").
		Order("CASE moderation_status WHEN 'flagged' THEN 1 WHEN 'pending' THEN 2 ELSE 3 END").
		Order("created_at ASC").
		Offset(offset).Limit(limit).Find(&articles).Error

	return articles, total, err
}

func (r *ModerationRepository) UpdateArticleModeration(articleID uint, status model.ArticleModerationStatus, notes string, moderatorID uint, triggerWarnings []string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"moderation_status": status,
		"moderation_notes":  notes,
		"moderated_by_id":   moderatorID,
		"moderated_at":      now,
	}

	if len(triggerWarnings) > 0 {
		updates["trigger_warnings"] = model.TriggerWarnings(triggerWarnings)
	}

	// If approved, also publish
	if status == model.ArticleModerationApproved {
		updates["status"] = model.ArticleStatusPublished
	}
	// If rejected, set to draft
	if status == model.ArticleModerationRejected {
		updates["status"] = model.ArticleStatusDraft
	}

	return r.db.Model(&model.Article{}).Where("id = ?", articleID).Updates(updates).Error
}

func (r *ModerationRepository) GetPendingArticlesCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.Article{}).
		Where("is_user_generated = ? AND moderation_status = ?", true, model.ArticleModerationPending).
		Count(&count).Error
	return count, err
}

func (r *ModerationRepository) GetFlaggedArticlesCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.Article{}).
		Where("is_user_generated = ? AND moderation_status = ?", true, model.ArticleModerationFlagged).
		Count(&count).Error
	return count, err
}

// ========================
// User Suspension/Ban Operations
// ========================

func (r *ModerationRepository) SuspendUser(userID uint, endTime time.Time, reason string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"suspension_end":    endTime,
		"suspension_reason": reason,
	}).Error
}

func (r *ModerationRepository) UnsuspendUser(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"suspension_end":    nil,
		"suspension_reason": "",
	}).Error
}

func (r *ModerationRepository) BanUser(userID uint, reason string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_banned":  true,
		"ban_reason": reason,
	}).Error
}

func (r *ModerationRepository) UnbanUser(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_banned":  false,
		"ban_reason": "",
	}).Error
}

func (r *ModerationRepository) GetSuspendedUsersCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Where("suspension_end IS NOT NULL AND suspension_end > ?", time.Now()).
		Count(&count).Error
	return count, err
}

func (r *ModerationRepository) GetBannedUsersCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("is_banned = ?", true).Count(&count).Error
	return count, err
}

// ========================
// User AI Disclaimer Operations
// ========================

func (r *ModerationRepository) SetAIDisclaimerAccepted(userID uint, accepted bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		Update("has_accepted_ai_disclaimer", accepted).Error
}

func (r *ModerationRepository) SetContentWarningPreference(userID uint, preference model.ContentWarningPreference) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		Update("content_warning_preference", preference).Error
}

// ========================
// Forum Moderation Operations
// ========================

func (r *ModerationRepository) UpdateForumTriggerWarnings(forumID uint, warnings []string) error {
	return r.db.Model(&model.Forum{}).Where("id = ?", forumID).
		Update("trigger_warnings", model.TriggerWarnings(warnings)).Error
}

func (r *ModerationRepository) FlagForum(forumID uint, reason string) error {
	return r.db.Model(&model.Forum{}).Where("id = ?", forumID).Updates(map[string]interface{}{
		"is_flagged":     true,
		"flagged_reason": reason,
	}).Error
}

func (r *ModerationRepository) UnflagForum(forumID uint) error {
	return r.db.Model(&model.Forum{}).Where("id = ?", forumID).Updates(map[string]interface{}{
		"is_flagged":     false,
		"flagged_reason": "",
	}).Error
}

func (r *ModerationRepository) FlagForumPost(postID uint, reason string) error {
	return r.db.Model(&model.ForumPost{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"is_flagged":     true,
		"flagged_reason": reason,
	}).Error
}

func (r *ModerationRepository) UnflagForumPost(postID uint) error {
	return r.db.Model(&model.ForumPost{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"is_flagged":     false,
		"flagged_reason": "",
	}).Error
}
