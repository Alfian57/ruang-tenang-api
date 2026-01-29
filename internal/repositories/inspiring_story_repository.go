package repositories

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InspiringStoryRepository struct {
	db *gorm.DB
}

func NewInspiringStoryRepository(db *gorm.DB) *InspiringStoryRepository {
	return &InspiringStoryRepository{db: db}
}

// ==========================================
// Story Categories
// ==========================================

// GetAllCategories retrieves all story categories
func (r *InspiringStoryRepository) GetAllCategories() ([]models.StoryCategory, error) {
	var categories []models.StoryCategory
	err := r.db.Where("is_active = ?", true).
		Order("display_order ASC, name ASC").
		Find(&categories).Error
	return categories, err
}

// GetCategoryByID retrieves a category by ID
func (r *InspiringStoryRepository) GetCategoryByID(id uuid.UUID) (*models.StoryCategory, error) {
	var category models.StoryCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoryBySlug retrieves a category by slug
func (r *InspiringStoryRepository) GetCategoryBySlug(slug string) (*models.StoryCategory, error) {
	var category models.StoryCategory
	err := r.db.Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoriesWithCount retrieves categories with story counts
func (r *InspiringStoryRepository) GetCategoriesWithCount() ([]CategoryWithCount, error) {
	var results []CategoryWithCount

	err := r.db.Model(&models.StoryCategory{}).
		Select("story_categories.*, COALESCE(COUNT(DISTINCT story_category_relations.story_id), 0) as story_count").
		Joins("LEFT JOIN story_category_relations ON story_categories.id = story_category_relations.category_id").
		Joins("LEFT JOIN inspiring_stories ON story_category_relations.story_id = inspiring_stories.id AND inspiring_stories.status = 'approved'").
		Where("story_categories.is_active = ?", true).
		Group("story_categories.id").
		Order("story_categories.display_order ASC").
		Scan(&results).Error

	return results, err
}

type CategoryWithCount struct {
	models.StoryCategory
	StoryCount int `json:"story_count"`
}

// ==========================================
// Inspiring Stories
// ==========================================

// Create creates a new inspiring story
func (r *InspiringStoryRepository) Create(story *models.InspiringStory) error {
	return r.db.Create(story).Error
}

// GetByID retrieves a story by ID
func (r *InspiringStoryRepository) GetByID(id uuid.UUID) (*models.InspiringStory, error) {
	var story models.InspiringStory
	err := r.db.First(&story, id).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// GetByIDWithRelations retrieves a story with all relations
func (r *InspiringStoryRepository) GetByIDWithRelations(id uuid.UUID) (*models.InspiringStory, error) {
	var story models.InspiringStory
	err := r.db.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		First(&story, id).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// Update updates a story
func (r *InspiringStoryRepository) Update(story *models.InspiringStory) error {
	return r.db.Save(story).Error
}

// Delete deletes a story
func (r *InspiringStoryRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete category relations
		if err := tx.Where("story_id = ?", id).Delete(&models.StoryCategoryRelation{}).Error; err != nil {
			return err
		}
		// Delete tags
		if err := tx.Where("story_id = ?", id).Delete(&models.StoryTag{}).Error; err != nil {
			return err
		}
		// Delete hearts
		if err := tx.Where("story_id = ?", id).Delete(&models.StoryHeart{}).Error; err != nil {
			return err
		}
		// Delete comment hearts
		if err := tx.Exec("DELETE FROM story_comment_hearts WHERE comment_id IN (SELECT id FROM story_comments WHERE story_id = ?)", id).Error; err != nil {
			return err
		}
		// Delete comments
		if err := tx.Where("story_id = ?", id).Delete(&models.StoryComment{}).Error; err != nil {
			return err
		}
		// Delete story
		return tx.Delete(&models.InspiringStory{}, id).Error
	})
}

// GetApprovedStories retrieves approved stories with pagination
func (r *InspiringStoryRepository) GetApprovedStories(page, limit int, sortBy string) ([]models.InspiringStory, int64, error) {
	var stories []models.InspiringStory
	var total int64

	query := r.db.Model(&models.InspiringStory{}).Where("status = ?", "approved")
	query.Count(&total)

	// Apply sorting
	switch sortBy {
	case "hearts":
		query = query.Order("heart_count DESC, published_at DESC")
	case "featured":
		query = query.Order("is_featured DESC, published_at DESC")
	case "views":
		query = query.Order("view_count DESC, published_at DESC")
	default: // recent
		query = query.Order("published_at DESC")
	}

	err := query.
		Preload("Author").
		Preload("Categories").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&stories).Error

	return stories, total, err
}

// GetStoriesByCategory retrieves stories by category
func (r *InspiringStoryRepository) GetStoriesByCategory(categoryID uuid.UUID, page, limit int) ([]models.InspiringStory, int64, error) {
	var stories []models.InspiringStory
	var total int64

	query := r.db.Model(&models.InspiringStory{}).
		Joins("JOIN story_category_relations ON inspiring_stories.id = story_category_relations.story_id").
		Where("story_category_relations.category_id = ? AND inspiring_stories.status = ?", categoryID, "approved")

	query.Count(&total)

	err := query.
		Preload("Author").
		Preload("Categories").
		Order("published_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&stories).Error

	return stories, total, err
}

// GetStoriesByAuthor retrieves stories by author
func (r *InspiringStoryRepository) GetStoriesByAuthor(authorID uint, status string, page, limit int) ([]models.InspiringStory, int64, error) {
	var stories []models.InspiringStory
	var total int64

	query := r.db.Model(&models.InspiringStory{}).Where("author_id = ?", authorID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.
		Preload("Categories").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&stories).Error

	return stories, total, err
}

// SearchStories searches stories by title or content
func (r *InspiringStoryRepository) SearchStories(search string, page, limit int) ([]models.InspiringStory, int64, error) {
	var stories []models.InspiringStory
	var total int64

	searchPattern := "%" + search + "%"

	query := r.db.Model(&models.InspiringStory{}).
		Where("status = ? AND (title ILIKE ? OR content ILIKE ?)", "approved", searchPattern, searchPattern)

	query.Count(&total)

	err := query.
		Preload("Author").
		Preload("Categories").
		Order("published_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&stories).Error

	return stories, total, err
}

// GetFeaturedStories retrieves featured stories
func (r *InspiringStoryRepository) GetFeaturedStories(limit int) ([]models.InspiringStory, error) {
	var stories []models.InspiringStory
	err := r.db.Where("status = ? AND is_featured = ?", "approved", true).
		Preload("Author").
		Preload("Categories").
		Order("featured_at DESC").
		Limit(limit).
		Find(&stories).Error
	return stories, err
}

// GetPendingStories retrieves stories pending moderation
func (r *InspiringStoryRepository) GetPendingStories(page, limit int) ([]models.InspiringStory, int64, error) {
	var stories []models.InspiringStory
	var total int64

	query := r.db.Model(&models.InspiringStory{}).Where("status = ?", "pending")
	query.Count(&total)

	err := query.
		Preload("Author").
		Preload("Categories").
		Order("created_at ASC"). // Oldest first
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&stories).Error

	return stories, total, err
}

// UpdateStatus updates story status
func (r *InspiringStoryRepository) UpdateStatus(id uuid.UUID, status string, moderatorID uint, feedback string) error {
	updates := map[string]interface{}{
		"status":             status,
		"moderated_by":       moderatorID,
		"moderator_feedback": feedback,
		"moderated_at":       time.Now(),
	}

	if status == "approved" {
		updates["published_at"] = time.Now()
	}

	return r.db.Model(&models.InspiringStory{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementViewCount increments the view count
func (r *InspiringStoryRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&models.InspiringStory{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// SetFeatured sets or unsets featured status
func (r *InspiringStoryRepository) SetFeatured(id uuid.UUID, featured bool, featuredBy uint) error {
	updates := map[string]interface{}{
		"is_featured": featured,
	}
	if featured {
		updates["featured_at"] = time.Now()
		updates["featured_by"] = featuredBy
	} else {
		updates["featured_at"] = nil
		updates["featured_by"] = nil
	}
	return r.db.Model(&models.InspiringStory{}).Where("id = ?", id).Updates(updates).Error
}

// GetAuthorStoriesCount returns count of stories by author in a month
func (r *InspiringStoryRepository) GetAuthorStoriesCount(authorID uint, month, year int) (int64, error) {
	var count int64
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	err := r.db.Model(&models.InspiringStory{}).
		Where("author_id = ? AND created_at >= ? AND created_at < ?", authorID, startOfMonth, endOfMonth).
		Count(&count).Error

	return count, err
}

// ==========================================
// Story Tags
// ==========================================

// SetStoryTags sets tags for a story (replaces existing)
func (r *InspiringStoryRepository) SetStoryTags(storyID uuid.UUID, tags []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing tags
		if err := tx.Where("story_id = ?", storyID).Delete(&models.StoryTag{}).Error; err != nil {
			return err
		}

		// Create new tags
		for _, tag := range tags {
			storyTag := &models.StoryTag{
				StoryID: storyID,
				Tag:     tag,
			}
			if err := tx.Create(storyTag).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetStoryTags gets all tags for a story
func (r *InspiringStoryRepository) GetStoryTags(storyID uuid.UUID) ([]string, error) {
	var tags []models.StoryTag
	err := r.db.Where("story_id = ?", storyID).Find(&tags).Error
	if err != nil {
		return nil, err
	}

	result := make([]string, len(tags))
	for i, t := range tags {
		result[i] = t.Tag
	}
	return result, nil
}

// ==========================================
// Story Category Relations
// ==========================================

// SetStoryCategories sets categories for a story (replaces existing)
func (r *InspiringStoryRepository) SetStoryCategories(storyID uuid.UUID, categoryIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing relations
		if err := tx.Where("story_id = ?", storyID).Delete(&models.StoryCategoryRelation{}).Error; err != nil {
			return err
		}

		// Create new relations
		for _, catID := range categoryIDs {
			relation := &models.StoryCategoryRelation{
				StoryID:    storyID,
				CategoryID: catID,
			}
			if err := tx.Create(relation).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ==========================================
// Story Hearts
// ==========================================

// AddHeart adds a heart to a story
func (r *InspiringStoryRepository) AddHeart(storyID uuid.UUID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		heart := &models.StoryHeart{
			StoryID: storyID,
			UserID:  userID,
		}
		if err := tx.Create(heart).Error; err != nil {
			return err
		}

		// Update heart count
		return tx.Model(&models.InspiringStory{}).
			Where("id = ?", storyID).
			UpdateColumn("heart_count", gorm.Expr("heart_count + 1")).Error
	})
}

// RemoveHeart removes a heart from a story
func (r *InspiringStoryRepository) RemoveHeart(storyID uuid.UUID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("story_id = ? AND user_id = ?", storyID, userID).Delete(&models.StoryHeart{})
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// Update heart count
		return tx.Model(&models.InspiringStory{}).
			Where("id = ?", storyID).
			UpdateColumn("heart_count", gorm.Expr("GREATEST(heart_count - 1, 0)")).Error
	})
}

// HasHearted checks if user has hearted a story
func (r *InspiringStoryRepository) HasHearted(storyID uuid.UUID, userID uint) bool {
	var count int64
	r.db.Model(&models.StoryHeart{}).
		Where("story_id = ? AND user_id = ?", storyID, userID).
		Count(&count)
	return count > 0
}

// GetStoryHeartCount returns the heart count for a story
func (r *InspiringStoryRepository) GetStoryHeartCount(storyID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&models.StoryHeart{}).Where("story_id = ?", storyID).Count(&count).Error
	return int(count), err
}

// ==========================================
// Story Comments
// ==========================================

// CreateComment creates a new comment
func (r *InspiringStoryRepository) CreateComment(comment *models.StoryComment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// Update comment count
		return tx.Model(&models.InspiringStory{}).
			Where("id = ?", comment.StoryID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

// GetCommentByID retrieves a comment by ID
func (r *InspiringStoryRepository) GetCommentByID(id uuid.UUID) (*models.StoryComment, error) {
	var comment models.StoryComment
	err := r.db.Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetStoryComments retrieves comments for a story
func (r *InspiringStoryRepository) GetStoryComments(storyID uuid.UUID, page, limit int, includeHidden bool) ([]models.StoryComment, int64, error) {
	var comments []models.StoryComment
	var total int64

	query := r.db.Model(&models.StoryComment{}).Where("story_id = ?", storyID)
	if !includeHidden {
		query = query.Where("is_hidden = ?", false)
	}

	query.Count(&total)

	err := query.
		Preload("User").
		Order("created_at ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&comments).Error

	return comments, total, err
}

// HideComment hides a comment
func (r *InspiringStoryRepository) HideComment(id uuid.UUID, hidden bool, reason string) error {
	return r.db.Model(&models.StoryComment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_hidden":     hidden,
			"hidden_reason": reason,
		}).Error
}

// DeleteComment deletes a comment
func (r *InspiringStoryRepository) DeleteComment(id uuid.UUID, storyID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete comment hearts
		if err := tx.Where("comment_id = ?", id).Delete(&models.StoryCommentHeart{}).Error; err != nil {
			return err
		}

		// Delete comment
		if err := tx.Delete(&models.StoryComment{}, id).Error; err != nil {
			return err
		}

		// Update comment count
		return tx.Model(&models.InspiringStory{}).
			Where("id = ?", storyID).
			UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error
	})
}

// ==========================================
// Comment Hearts
// ==========================================

// AddCommentHeart adds a heart to a comment
func (r *InspiringStoryRepository) AddCommentHeart(commentID uuid.UUID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		heart := &models.StoryCommentHeart{
			CommentID: commentID,
			UserID:    userID,
		}
		if err := tx.Create(heart).Error; err != nil {
			return err
		}

		return tx.Model(&models.StoryComment{}).
			Where("id = ?", commentID).
			UpdateColumn("heart_count", gorm.Expr("heart_count + 1")).Error
	})
}

// RemoveCommentHeart removes a heart from a comment
func (r *InspiringStoryRepository) RemoveCommentHeart(commentID uuid.UUID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&models.StoryCommentHeart{})
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&models.StoryComment{}).
			Where("id = ?", commentID).
			UpdateColumn("heart_count", gorm.Expr("GREATEST(heart_count - 1, 0)")).Error
	})
}

// HasHeartedComment checks if user has hearted a comment
func (r *InspiringStoryRepository) HasHeartedComment(commentID uuid.UUID, userID uint) bool {
	var count int64
	r.db.Model(&models.StoryCommentHeart{}).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Count(&count)
	return count > 0
}

// ==========================================
// Story Stats
// ==========================================

// GetAuthorStats returns stats for an author
func (r *InspiringStoryRepository) GetAuthorStats(authorID uint) (*AuthorStoryStats, error) {
	stats := &AuthorStoryStats{}

	// Total stories
	r.db.Model(&models.InspiringStory{}).Where("author_id = ?", authorID).Count(&stats.TotalStories)

	// Approved stories
	r.db.Model(&models.InspiringStory{}).Where("author_id = ? AND status = 'approved'", authorID).Count(&stats.ApprovedStories)

	// Pending stories
	r.db.Model(&models.InspiringStory{}).Where("author_id = ? AND status = 'pending'", authorID).Count(&stats.PendingStories)

	// Total hearts received
	r.db.Model(&models.StoryHeart{}).
		Joins("JOIN inspiring_stories ON story_hearts.story_id = inspiring_stories.id").
		Where("inspiring_stories.author_id = ?", authorID).
		Count(&stats.TotalHearts)

	// Total views
	r.db.Model(&models.InspiringStory{}).
		Where("author_id = ?", authorID).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&stats.TotalViews)

	// Total comments received
	r.db.Model(&models.StoryComment{}).
		Joins("JOIN inspiring_stories ON story_comments.story_id = inspiring_stories.id").
		Where("inspiring_stories.author_id = ?", authorID).
		Count(&stats.TotalComments)

	return stats, nil
}

type AuthorStoryStats struct {
	TotalStories    int64
	ApprovedStories int64
	PendingStories  int64
	TotalHearts     int64
	TotalViews      int64
	TotalComments   int64
}

// GetMostAppreciatedStories returns most hearted stories for a period
func (r *InspiringStoryRepository) GetMostAppreciatedStories(month, year, limit int) ([]models.InspiringStory, error) {
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var stories []models.InspiringStory
	err := r.db.Where("status = 'approved' AND published_at >= ? AND published_at < ?", startOfMonth, endOfMonth).
		Preload("Author").
		Preload("Categories").
		Order("heart_count DESC").
		Limit(limit).
		Find(&stories).Error

	return stories, err
}
