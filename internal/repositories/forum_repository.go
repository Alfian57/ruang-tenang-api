package repositories

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"

	"gorm.io/gorm"
)

// PostSortOption represents sorting options for forum posts
type PostSortOption string

const (
	SortByTop    PostSortOption = "top"    // Most upvoted
	SortByNewest PostSortOption = "newest" // Most recent
	SortByOldest PostSortOption = "oldest" // Chronological
)

type ForumRepository interface {
	CreateForum(forum *models.Forum) error
	GetForums(limit, offset int, search string, categoryID *uint) ([]models.Forum, int64, error)
	GetForumByID(id uint) (*models.Forum, error)
	DeleteForum(id uint) error
	UpdateForum(forum *models.Forum) error

	CreateForumPost(post *models.ForumPost) error
	GetForumPosts(forumID uint, limit, offset int) ([]models.ForumPost, int64, error)
	GetForumPostsSorted(forumID uint, limit, offset int, sort PostSortOption, userID uint) ([]models.ForumPost, int64, error)
	DeleteForumPost(id uint) error
	GetForumPostByID(id uint) (*models.ForumPost, error)
	UpdateForumPost(post *models.ForumPost) error

	// Forum like methods (existing)
	ToggleLike(userID, forumID uint) (bool, error)
	GetLikesCount(forumID uint) (int64, error)
	GetRepliesCount(forumID uint) (int64, error)
	HasUserLiked(userID, forumID uint) (bool, error)
	FindUpdatedSince(since time.Time) ([]models.Forum, error)

	// Post voting methods (new)
	VotePost(userID, postID uint, voteType models.VoteType) error
	RemovePostVote(userID, postID uint) error
	GetUserPostVote(userID, postID uint) (*models.ForumPostVote, error)
	GetPostVotesCount(postID uint, voteType models.VoteType) (int64, error)
	UpdatePostVoteCounts(postID uint) error

	// Best answer methods
	MarkAsAcceptedAnswer(postID uint) error
	UnmarkAcceptedAnswer(forumID uint) error
	GetAcceptedAnswer(forumID uint) (*models.ForumPost, error)
	UpdateCommunityFavorite(forumID uint) error

	// Report methods
	CreatePostReport(report *models.ForumPostReport) error
	GetPostReports(postID uint) ([]models.ForumPostReport, error)
	GetPendingPostReports(limit, offset int) ([]models.ForumPostReport, int64, error)
	GetPostReportByID(id uint) (*models.ForumPostReport, error)
	UpdatePostReport(report *models.ForumPostReport) error
	HasUserReportedPost(userID, postID uint) (bool, error)
	GetPostReportsCount(postID uint) (int64, error)

	// Batch operations for efficiency
	GetUserVotesForPosts(userID uint, postIDs []uint) (map[uint]*models.ForumPostVote, error)
}

type forumRepository struct {
	db *gorm.DB
}

func NewForumRepository(db *gorm.DB) ForumRepository {
	return &forumRepository{db}
}

// Forum Methods

func (r *forumRepository) CreateForum(forum *models.Forum) error {
	return r.db.Create(forum).Error
}

func (r *forumRepository) GetForums(limit, offset int, search string, categoryID *uint) ([]models.Forum, int64, error) {
	var forums []models.Forum
	var total int64

	query := r.db.Model(&models.Forum{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("User").
		Preload("Category").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&forums).Error

	return forums, total, err
}

func (r *forumRepository) GetForumByID(id uint) (*models.Forum, error) {
	var forum models.Forum
	err := r.db.Preload("User").Preload("Category").First(&forum, id).Error
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (r *forumRepository) DeleteForum(id uint) error {
	return r.db.Delete(&models.Forum{}, id).Error
}

func (r *forumRepository) UpdateForum(forum *models.Forum) error {
	return r.db.Save(forum).Error
}

// Post Methods

func (r *forumRepository) CreateForumPost(post *models.ForumPost) error {
	return r.db.Create(post).Error
}

func (r *forumRepository) GetForumPosts(forumID uint, limit, offset int) ([]models.ForumPost, int64, error) {
	var posts []models.ForumPost
	var total int64

	err := r.db.Model(&models.ForumPost{}).Where("forum_id = ?", forumID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("User").
		Where("forum_id = ?", forumID).
		Order("created_at asc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, total, err
}

func (r *forumRepository) DeleteForumPost(id uint) error {
	return r.db.Delete(&models.ForumPost{}, id).Error
}

func (r *forumRepository) GetForumPostByID(id uint) (*models.ForumPost, error) {
	var post models.ForumPost
	err := r.db.Preload("User").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *forumRepository) UpdateForumPost(post *models.ForumPost) error {
	return r.db.Save(post).Error
}

// GetForumPostsSorted retrieves forum posts with sorting and user vote information
func (r *forumRepository) GetForumPostsSorted(forumID uint, limit, offset int, sort PostSortOption, userID uint) ([]models.ForumPost, int64, error) {
	var posts []models.ForumPost
	var total int64

	err := r.db.Model(&models.ForumPost{}).Where("forum_id = ?", forumID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("User").Where("forum_id = ?", forumID)

	// Apply sorting
	switch sort {
	case SortByTop:
		// Accepted answer first, then community favorite, then by upvotes
		query = query.Order("is_accepted_answer DESC, is_community_favorite DESC, upvotes_count DESC, created_at ASC")
	case SortByNewest:
		query = query.Order("created_at DESC")
	case SortByOldest:
		query = query.Order("created_at ASC")
	default:
		// Default: accepted first, then by upvotes
		query = query.Order("is_accepted_answer DESC, upvotes_count DESC, created_at ASC")
	}

	err = query.Limit(limit).Offset(offset).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	// Populate user vote information if userID is provided
	if userID > 0 && len(posts) > 0 {
		postIDs := make([]uint, len(posts))
		for i, p := range posts {
			postIDs[i] = p.ID
		}

		userVotes, _ := r.GetUserVotesForPosts(userID, postIDs)
		for i := range posts {
			posts[i].NetVotes = posts[i].CalculateNetVotes()
			posts[i].IsAutoHidden = posts[i].ShouldAutoHide()
			if vote, exists := userVotes[posts[i].ID]; exists {
				posts[i].HasUserVoted = true
				posts[i].UserVoteType = vote.VoteType
			}
		}
	} else {
		for i := range posts {
			posts[i].NetVotes = posts[i].CalculateNetVotes()
			posts[i].IsAutoHidden = posts[i].ShouldAutoHide()
		}
	}

	return posts, total, err
}

// Like Methods

func (r *forumRepository) ToggleLike(userID, forumID uint) (bool, error) {
	var like models.ForumLike
	err := r.db.Where("user_id = ? AND forum_id = ?", userID, forumID).First(&like).Error

	if err == nil {
		// Like exists, delete it (unlike)
		err = r.db.Delete(&like).Error
		return false, err
	} else if err == gorm.ErrRecordNotFound {
		// Like doesn't exist, create it (like)
		newLike := models.ForumLike{
			UserID:  userID,
			ForumID: forumID,
		}
		err = r.db.Create(&newLike).Error
		return true, err
	}

	return false, err
}

func (r *forumRepository) GetLikesCount(forumID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumLike{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

func (r *forumRepository) HasUserLiked(userID, forumID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ForumLike{}).Where("user_id = ? AND forum_id = ?", userID, forumID).Count(&count).Error
	return count > 0, err
}

func (r *forumRepository) GetRepliesCount(forumID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumPost{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

// FindUpdatedSince retrieves forums updated since the given time (for incremental cache sync)
func (r *forumRepository) FindUpdatedSince(since time.Time) ([]models.Forum, error) {
	var forums []models.Forum
	err := r.db.Model(&models.Forum{}).
		Where("updated_at > ?", since).
		Find(&forums).Error
	return forums, err
}

// ==================== Post Voting Methods ====================

// VotePost adds or updates a vote on a post
func (r *forumRepository) VotePost(userID, postID uint, voteType models.VoteType) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existingVote models.ForumPostVote
		err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingVote).Error

		if err == gorm.ErrRecordNotFound {
			// Create new vote
			vote := models.ForumPostVote{
				UserID:   userID,
				PostID:   postID,
				VoteType: voteType,
			}
			if err := tx.Create(&vote).Error; err != nil {
				return err
			}
		} else if err == nil {
			// Update existing vote if different
			if existingVote.VoteType != voteType {
				existingVote.VoteType = voteType
				if err := tx.Save(&existingVote).Error; err != nil {
					return err
				}
			}
		} else {
			return err
		}

		// Update cached vote counts on the post
		return r.updatePostVoteCountsInTx(tx, postID)
	})
}

// RemovePostVote removes a vote from a post
func (r *forumRepository) RemovePostVote(userID, postID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.ForumPostVote{})
		if result.Error != nil {
			return result.Error
		}

		// Update cached vote counts on the post
		return r.updatePostVoteCountsInTx(tx, postID)
	})
}

// GetUserPostVote gets the user's vote on a post
func (r *forumRepository) GetUserPostVote(userID, postID uint) (*models.ForumPostVote, error) {
	var vote models.ForumPostVote
	err := r.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&vote).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// GetPostVotesCount gets the count of a specific vote type on a post
func (r *forumRepository) GetPostVotesCount(postID uint, voteType models.VoteType) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumPostVote{}).
		Where("post_id = ? AND vote_type = ?", postID, voteType).
		Count(&count).Error
	return count, err
}

// UpdatePostVoteCounts updates the cached vote counts on a post
func (r *forumRepository) UpdatePostVoteCounts(postID uint) error {
	return r.updatePostVoteCountsInTx(r.db, postID)
}

func (r *forumRepository) updatePostVoteCountsInTx(tx *gorm.DB, postID uint) error {
	var upvotes, downvotes int64

	tx.Model(&models.ForumPostVote{}).
		Where("post_id = ? AND vote_type = ?", postID, models.VoteTypeUpvote).
		Count(&upvotes)

	tx.Model(&models.ForumPostVote{}).
		Where("post_id = ? AND vote_type = ?", postID, models.VoteTypeDownvote).
		Count(&downvotes)

	return tx.Model(&models.ForumPost{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"upvotes_count":   upvotes,
			"downvotes_count": downvotes,
		}).Error
}

// GetUserVotesForPosts gets all user votes for a list of posts (for batch loading)
func (r *forumRepository) GetUserVotesForPosts(userID uint, postIDs []uint) (map[uint]*models.ForumPostVote, error) {
	var votes []models.ForumPostVote
	err := r.db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&votes).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]*models.ForumPostVote)
	for i := range votes {
		result[votes[i].PostID] = &votes[i]
	}
	return result, nil
}

// ==================== Best Answer Methods ====================

// MarkAsAcceptedAnswer marks a post as the accepted answer
func (r *forumRepository) MarkAsAcceptedAnswer(postID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get the post to find its forum
		var post models.ForumPost
		if err := tx.First(&post, postID).Error; err != nil {
			return err
		}

		// Unmark any existing accepted answer in this forum
		if err := tx.Model(&models.ForumPost{}).
			Where("forum_id = ? AND is_accepted_answer = ?", post.ForumID, true).
			Update("is_accepted_answer", false).Error; err != nil {
			return err
		}

		// Mark this post as accepted
		if err := tx.Model(&models.ForumPost{}).
			Where("id = ?", postID).
			Update("is_accepted_answer", true).Error; err != nil {
			return err
		}

		// Update forum's has_accepted_answer flag
		return tx.Model(&models.Forum{}).
			Where("id = ?", post.ForumID).
			Update("has_accepted_answer", true).Error
	})
}

// UnmarkAcceptedAnswer removes the accepted answer mark from all posts in a forum
func (r *forumRepository) UnmarkAcceptedAnswer(forumID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Unmark all accepted answers in this forum
		if err := tx.Model(&models.ForumPost{}).
			Where("forum_id = ? AND is_accepted_answer = ?", forumID, true).
			Update("is_accepted_answer", false).Error; err != nil {
			return err
		}

		// Update forum's has_accepted_answer flag
		return tx.Model(&models.Forum{}).
			Where("id = ?", forumID).
			Update("has_accepted_answer", false).Error
	})
}

// GetAcceptedAnswer gets the accepted answer for a forum
func (r *forumRepository) GetAcceptedAnswer(forumID uint) (*models.ForumPost, error) {
	var post models.ForumPost
	err := r.db.Preload("User").
		Where("forum_id = ? AND is_accepted_answer = ?", forumID, true).
		First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

// UpdateCommunityFavorite updates the community favorite status based on vote counts
func (r *forumRepository) UpdateCommunityFavorite(forumID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Reset all community favorite flags in this forum
		if err := tx.Model(&models.ForumPost{}).
			Where("forum_id = ?", forumID).
			Update("is_community_favorite", false).Error; err != nil {
			return err
		}

		// Find the post with the highest upvotes (minimum 3 upvotes to qualify)
		var topPost models.ForumPost
		err := tx.Where("forum_id = ? AND upvotes_count >= ?", forumID, 3).
			Order("upvotes_count DESC").
			First(&topPost).Error

		if err == gorm.ErrRecordNotFound {
			return nil // No qualifying post
		}
		if err != nil {
			return err
		}

		// Mark as community favorite
		return tx.Model(&models.ForumPost{}).
			Where("id = ?", topPost.ID).
			Update("is_community_favorite", true).Error
	})
}

// ==================== Report Methods ====================

// CreatePostReport creates a new report for a post
func (r *forumRepository) CreatePostReport(report *models.ForumPostReport) error {
	return r.db.Create(report).Error
}

// GetPostReports gets all reports for a specific post
func (r *forumRepository) GetPostReports(postID uint) ([]models.ForumPostReport, error) {
	var reports []models.ForumPostReport
	err := r.db.Preload("Reporter").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&reports).Error
	return reports, err
}

// GetPendingPostReports gets all pending reports for moderation
func (r *forumRepository) GetPendingPostReports(limit, offset int) ([]models.ForumPostReport, int64, error) {
	var reports []models.ForumPostReport
	var total int64

	err := r.db.Model(&models.ForumPostReport{}).
		Where("status = ?", models.PostReportStatusPending).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Post.User").Preload("Post.Forum").Preload("Reporter").
		Where("status = ?", models.PostReportStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	return reports, total, err
}

// GetPostReportByID gets a specific report by ID
func (r *forumRepository) GetPostReportByID(id uint) (*models.ForumPostReport, error) {
	var report models.ForumPostReport
	err := r.db.Preload("Post.User").Preload("Post.Forum").Preload("Reporter").Preload("Reviewer").
		First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// UpdatePostReport updates a report
func (r *forumRepository) UpdatePostReport(report *models.ForumPostReport) error {
	return r.db.Save(report).Error
}

// HasUserReportedPost checks if a user has already reported a post
func (r *forumRepository) HasUserReportedPost(userID, postID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ForumPostReport{}).
		Where("reporter_id = ? AND post_id = ? AND status = ?", userID, postID, models.PostReportStatusPending).
		Count(&count).Error
	return count > 0, err
}

// GetPostReportsCount gets the total number of reports for a post
func (r *forumRepository) GetPostReportsCount(postID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ForumPostReport{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	return count, err
}
