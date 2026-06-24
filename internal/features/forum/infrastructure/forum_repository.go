package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostSortOption represents sorting options for forum posts
type PostSortOption string

const (
	SortByTop    PostSortOption = "top"    // Most upvoted
	SortByNewest PostSortOption = "newest" // Most recent
	SortByOldest PostSortOption = "oldest" // Chronological
)

type ForumRepository interface {
	CreateForum(ctx context.Context, forum *model.Forum) error
	GetForums(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error)
	GetForumByID(ctx context.Context, id uint) (*model.Forum, error)
	GetForumBySlug(ctx context.Context, slug string) (*model.Forum, error)
	DeleteForum(ctx context.Context, id uint) error
	UpdateForum(ctx context.Context, forum *model.Forum) error

	CreateForumPost(ctx context.Context, post *model.ForumPost) error
	GetForumPosts(ctx context.Context, forumID uint, limit, offset int) ([]model.ForumPost, int64, error)
	GetForumPostsSorted(ctx context.Context, forumID uint, limit, offset int, sort PostSortOption, userID uint) ([]model.ForumPost, int64, error)
	DeleteForumPost(ctx context.Context, id uint) error
	GetForumPostByID(ctx context.Context, id uint) (*model.ForumPost, error)
	UpdateForumPost(ctx context.Context, post *model.ForumPost) error

	// Forum like methods (existing)
	ToggleLike(ctx context.Context, userID, forumID uint) (bool, error)
	GetLikesCount(ctx context.Context, forumID uint) (int64, error)
	GetRepliesCount(ctx context.Context, forumID uint) (int64, error)
	HasUserLiked(ctx context.Context, userID, forumID uint) (bool, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Forum, error)

	// Post voting methods (new)
	VotePost(ctx context.Context, userID, postID uint, voteType model.VoteType) error
	RemovePostVote(ctx context.Context, userID, postID uint) error
	GetUserPostVote(ctx context.Context, userID, postID uint) (*model.ForumPostVote, error)
	GetPostVotesCount(ctx context.Context, postID uint, voteType model.VoteType) (int64, error)
	UpdatePostVoteCounts(ctx context.Context, postID uint) error

	// Best answer methods
	MarkAsAcceptedAnswer(ctx context.Context, postID uint) error
	UnmarkAcceptedAnswer(ctx context.Context, forumID uint) error
	GetAcceptedAnswer(ctx context.Context, forumID uint) (*model.ForumPost, error)
	UpdateCommunityFavorite(ctx context.Context, forumID uint) error

	// Report methods
	CreatePostReport(ctx context.Context, report *model.ForumPostReport) error
	GetPostReports(ctx context.Context, postID uint) ([]model.ForumPostReport, error)
	GetPendingPostReports(ctx context.Context, limit, offset int) ([]model.ForumPostReport, int64, error)
	GetPostReportByID(ctx context.Context, id uint) (*model.ForumPostReport, error)
	UpdatePostReport(ctx context.Context, report *model.ForumPostReport) error
	HasUserReportedPost(ctx context.Context, userID, postID uint) (bool, error)
	GetPostReportsCount(ctx context.Context, postID uint) (int64, error)

	// Batch operations for efficiency
	GetUserVotesForPosts(ctx context.Context, userID uint, postIDs []uint) (map[uint]*model.ForumPostVote, error)

	// Anti-abuse
	CountUserVotesInLastHour(ctx context.Context, userID uint) (int64, error)
	CountUserVotesInLastMinutes(ctx context.Context, userID uint, minutes int) (int64, error)
	CountUserVotesOnAuthorInLastHour(ctx context.Context, voterID, authorID uint) (int64, error)
}

type forumRepository struct {
	db *gorm.DB
}

func NewForumRepository(db *gorm.DB) ForumRepository {
	return &forumRepository{db}
}

// Forum Methods

func (r *forumRepository) CreateForum(ctx context.Context, forum *model.Forum) error {
	return r.db.WithContext(ctx).Create(forum).Error
}

func (r *forumRepository) GetForums(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error) {
	var forums []model.Forum
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Forum{})

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

func (r *forumRepository) GetForumByID(ctx context.Context, id uint) (*model.Forum, error) {
	var forum model.Forum
	err := r.db.WithContext(ctx).Preload("User").Preload("Category").First(&forum, id).Error
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (r *forumRepository) GetForumBySlug(ctx context.Context, slug string) (*model.Forum, error) {
	var forum model.Forum
	err := r.db.WithContext(ctx).Preload("User").Preload("Category").Where("slug = ?", slug).First(&forum).Error
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (r *forumRepository) DeleteForum(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Forum{}, id).Error
}

func (r *forumRepository) UpdateForum(ctx context.Context, forum *model.Forum) error {
	return r.db.WithContext(ctx).Save(forum).Error
}

// Post Methods

func (r *forumRepository) CreateForumPost(ctx context.Context, post *model.ForumPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *forumRepository) GetForumPosts(ctx context.Context, forumID uint, limit, offset int) ([]model.ForumPost, int64, error) {
	var posts []model.ForumPost
	var total int64

	err := r.db.WithContext(ctx).Model(&model.ForumPost{}).Where("forum_id = ?", forumID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Preload("User").
		Where("forum_id = ?", forumID).
		Order("created_at asc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, total, err
}

func (r *forumRepository) DeleteForumPost(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ForumPost{}, id).Error
}

func (r *forumRepository) GetForumPostByID(ctx context.Context, id uint) (*model.ForumPost, error) {
	var post model.ForumPost
	err := r.db.WithContext(ctx).Preload("User").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *forumRepository) UpdateForumPost(ctx context.Context, post *model.ForumPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// GetForumPostsSorted retrieves forum posts with sorting and user vote information
func (r *forumRepository) GetForumPostsSorted(ctx context.Context, forumID uint, limit, offset int, sort PostSortOption, userID uint) ([]model.ForumPost, int64, error) {
	var posts []model.ForumPost
	var total int64

	err := r.db.WithContext(ctx).Model(&model.ForumPost{}).Where("forum_id = ?", forumID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	query := r.db.WithContext(ctx).Preload("User").Where("forum_id = ?", forumID)

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

		userVotes, _ := r.GetUserVotesForPosts(ctx, userID, postIDs)
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

func (r *forumRepository) ToggleLike(ctx context.Context, userID, forumID uint) (bool, error) {
	var liked bool

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the existing like row (if any) so two concurrent toggles on the
		// same user+forum serialize instead of racing into a double-create.
		var like model.ForumLike
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND forum_id = ?", userID, forumID).
			First(&like).Error

		if err == nil {
			// Like exists, delete it (unlike)
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			liked = false
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		// Like doesn't exist, create it (like)
		newLike := model.ForumLike{
			UserID:  userID,
			ForumID: forumID,
		}
		if err := tx.Create(&newLike).Error; err != nil {
			return err
		}
		liked = true
		return nil
	})

	return liked, err
}

func (r *forumRepository) GetLikesCount(ctx context.Context, forumID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ForumLike{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

func (r *forumRepository) HasUserLiked(ctx context.Context, userID, forumID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ForumLike{}).Where("user_id = ? AND forum_id = ?", userID, forumID).Count(&count).Error
	return count > 0, err
}

func (r *forumRepository) GetRepliesCount(ctx context.Context, forumID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ForumPost{}).Where("forum_id = ?", forumID).Count(&count).Error
	return count, err
}

// FindUpdatedSince retrieves forums updated since the given time (for incremental cache sync)
func (r *forumRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Forum, error) {
	var forums []model.Forum
	err := r.db.WithContext(ctx).Model(&model.Forum{}).
		Where("updated_at > ?", since).
		Find(&forums).Error
	return forums, err
}

// ==================== Post Voting Methods ====================

// VotePost adds or updates a vote on a post
func (r *forumRepository) VotePost(ctx context.Context, userID, postID uint, voteType model.VoteType) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingVote model.ForumPostVote
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND post_id = ?", userID, postID).
			First(&existingVote).Error

		if err == gorm.ErrRecordNotFound {
			// Create new vote
			vote := model.ForumPostVote{
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
		return r.updatePostVoteCountsInTx(ctx, tx, postID)
	})
}

// RemovePostVote removes a vote from a post
func (r *forumRepository) RemovePostVote(ctx context.Context, userID, postID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.ForumPostVote{})
		if result.Error != nil {
			return result.Error
		}

		// Update cached vote counts on the post
		return r.updatePostVoteCountsInTx(ctx, tx, postID)
	})
}

// GetUserPostVote gets the user's vote on a post
func (r *forumRepository) GetUserPostVote(ctx context.Context, userID, postID uint) (*model.ForumPostVote, error) {
	var vote model.ForumPostVote
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).First(&vote).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// GetPostVotesCount gets the count of a specific vote type on a post
func (r *forumRepository) GetPostVotesCount(ctx context.Context, postID uint, voteType model.VoteType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ForumPostVote{}).
		Where("post_id = ? AND vote_type = ?", postID, voteType).
		Count(&count).Error
	return count, err
}

// UpdatePostVoteCounts updates the cached vote counts on a post
func (r *forumRepository) UpdatePostVoteCounts(ctx context.Context, postID uint) error {
	return r.updatePostVoteCountsInTx(ctx, r.db, postID)
}

func (r *forumRepository) updatePostVoteCountsInTx(ctx context.Context, tx *gorm.DB, postID uint) error {
	// Compute counts and write in a single UPDATE per column via correlated
	// subqueries. This avoids the read-COUNT-then-write race where two
	// concurrent votes could each read the stale count and overwrite each other.
	return tx.Model(&model.ForumPost{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"upvotes_count": gorm.Expr("(SELECT COUNT(*) FROM forum_post_votes WHERE post_id = ? AND vote_type = ?)", postID, model.VoteTypeUpvote),
			"downvotes_count": gorm.Expr("(SELECT COUNT(*) FROM forum_post_votes WHERE post_id = ? AND vote_type = ?)", postID, model.VoteTypeDownvote),
		}).Error
}

// GetUserVotesForPosts gets all user votes for a list of posts (for batch loading)
func (r *forumRepository) GetUserVotesForPosts(ctx context.Context, userID uint, postIDs []uint) (map[uint]*model.ForumPostVote, error) {
	var votes []model.ForumPostVote
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&votes).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]*model.ForumPostVote)
	for i := range votes {
		result[votes[i].PostID] = &votes[i]
	}
	return result, nil
}

// ==================== Best Answer Methods ====================

// MarkAsAcceptedAnswer marks a post as the accepted answer
func (r *forumRepository) MarkAsAcceptedAnswer(ctx context.Context, postID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the post to find its forum
		var post model.ForumPost
		if err := tx.First(&post, postID).Error; err != nil {
			return err
		}

		// Unmark any existing accepted answer in this forum
		if err := tx.Model(&model.ForumPost{}).
			Where("forum_id = ? AND is_accepted_answer = ?", post.ForumID, true).
			Update("is_accepted_answer", false).Error; err != nil {
			return err
		}

		// Mark this post as accepted
		if err := tx.Model(&model.ForumPost{}).
			Where("id = ?", postID).
			Update("is_accepted_answer", true).Error; err != nil {
			return err
		}

		// Update forum's has_accepted_answer flag
		return tx.Model(&model.Forum{}).
			Where("id = ?", post.ForumID).
			Update("has_accepted_answer", true).Error
	})
}

// UnmarkAcceptedAnswer removes the accepted answer mark from all posts in a forum
func (r *forumRepository) UnmarkAcceptedAnswer(ctx context.Context, forumID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unmark all accepted answers in this forum
		if err := tx.Model(&model.ForumPost{}).
			Where("forum_id = ? AND is_accepted_answer = ?", forumID, true).
			Update("is_accepted_answer", false).Error; err != nil {
			return err
		}

		// Update forum's has_accepted_answer flag
		return tx.Model(&model.Forum{}).
			Where("id = ?", forumID).
			Update("has_accepted_answer", false).Error
	})
}

// GetAcceptedAnswer gets the accepted answer for a forum
func (r *forumRepository) GetAcceptedAnswer(ctx context.Context, forumID uint) (*model.ForumPost, error) {
	var post model.ForumPost
	err := r.db.WithContext(ctx).Preload("User").
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
func (r *forumRepository) UpdateCommunityFavorite(ctx context.Context, forumID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reset all community favorite flags in this forum
		if err := tx.Model(&model.ForumPost{}).
			Where("forum_id = ?", forumID).
			Update("is_community_favorite", false).Error; err != nil {
			return err
		}

		// Find the post with the highest upvotes (minimum 3 upvotes to qualify)
		var topPost model.ForumPost
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
		return tx.Model(&model.ForumPost{}).
			Where("id = ?", topPost.ID).
			Update("is_community_favorite", true).Error
	})
}

// ==================== Report Methods ====================

// CreatePostReport creates a new report for a post
func (r *forumRepository) CreatePostReport(ctx context.Context, report *model.ForumPostReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// GetPostReports gets all reports for a specific post
func (r *forumRepository) GetPostReports(ctx context.Context, postID uint) ([]model.ForumPostReport, error) {
	var reports []model.ForumPostReport
	err := r.db.WithContext(ctx).Preload("Reporter").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&reports).Error
	return reports, err
}

// GetPendingPostReports gets all pending reports for moderation
func (r *forumRepository) GetPendingPostReports(ctx context.Context, limit, offset int) ([]model.ForumPostReport, int64, error) {
	var reports []model.ForumPostReport
	var total int64

	err := r.db.WithContext(ctx).Model(&model.ForumPostReport{}).
		Where("status = ?", model.PostReportStatusPending).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Preload("Post.User").Preload("Post.Forum").Preload("Reporter").
		Where("status = ?", model.PostReportStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	return reports, total, err
}

// GetPostReportByID gets a specific report by ID
func (r *forumRepository) GetPostReportByID(ctx context.Context, id uint) (*model.ForumPostReport, error) {
	var report model.ForumPostReport
	err := r.db.WithContext(ctx).Preload("Post.User").Preload("Post.Forum").Preload("Reporter").Preload("Reviewer").
		First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// UpdatePostReport updates a report
func (r *forumRepository) UpdatePostReport(ctx context.Context, report *model.ForumPostReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// HasUserReportedPost checks if a user has already reported a post
func (r *forumRepository) HasUserReportedPost(ctx context.Context, userID, postID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ForumPostReport{}).
		Where("reporter_id = ? AND post_id = ? AND status = ?", userID, postID, model.PostReportStatusPending).
		Count(&count).Error
	return count > 0, err
}

// GetPostReportsCount gets the total number of reports for a post
func (r *forumRepository) GetPostReportsCount(ctx context.Context, postID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ForumPostReport{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	return count, err
}

// CountUserVotesInLastHour counts votes made by a user in the last hour
func (r *forumRepository) CountUserVotesInLastHour(ctx context.Context, userID uint) (int64, error) {
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	err := r.db.WithContext(ctx).Model(&model.ForumPostVote{}).
		Where("user_id = ? AND created_at >= ?", userID, oneHourAgo).
		Count(&count).Error
	return count, err
}

// CountUserVotesInLastMinutes counts votes made by a user in the last N minutes (burst detection)
func (r *forumRepository) CountUserVotesInLastMinutes(ctx context.Context, userID uint, minutes int) (int64, error) {
	var count int64
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.db.WithContext(ctx).Model(&model.ForumPostVote{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

// CountUserVotesOnAuthorInLastHour counts votes a user made on posts by a specific author in the last hour
func (r *forumRepository) CountUserVotesOnAuthorInLastHour(ctx context.Context, voterID, authorID uint) (int64, error) {
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	err := r.db.WithContext(ctx).Model(&model.ForumPostVote{}).
		Joins("JOIN forum_posts ON forum_posts.id = forum_post_votes.post_id").
		Where("forum_post_votes.user_id = ? AND forum_posts.user_id = ? AND forum_post_votes.created_at >= ?", voterID, authorID, oneHourAgo).
		Count(&count).Error
	return count, err
}
