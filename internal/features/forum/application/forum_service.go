package application

import (
	"context"
	"errors"
	"time"

	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/contentctx"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	gamificationpkg "github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/Alfian57/ruang-tenang-api/internal/features/forum/infrastructure"
)

type ForumService interface {
	CreateForum(ctx context.Context, userID uint, title, content string, categoryID *uint) (*model.Forum, error)
	GetForums(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error)
	GetForumByID(ctx context.Context, userID, id uint) (*model.Forum, error)
	GetForumBySlug(ctx context.Context, userID uint, slug string) (*model.Forum, error)
	DeleteForum(ctx context.Context, userID uint, userRole string, forumID uint) error
	DeleteForumBySlug(ctx context.Context, userID uint, userRole string, slug string) error

	CreateForumPost(ctx context.Context, userID uint, forumID uint, content string) error
	CreateForumPostBySlug(ctx context.Context, userID uint, forumSlug string, content string) error
	GetForumPosts(ctx context.Context, forumID uint, limit, offset int) ([]model.ForumPost, int64, error)
	GetForumPostsSorted(ctx context.Context, forumID uint, limit, offset int, sort string, userID uint) ([]model.ForumPost, int64, error)
	GetForumPostsSortedBySlug(ctx context.Context, forumSlug string, limit, offset int, sort string, userID uint) ([]model.ForumPost, int64, error)
	DeleteForumPost(ctx context.Context, userID uint, userRole string, postID uint) error

	ToggleLike(ctx context.Context, userID, forumID uint) (bool, error)
	ToggleLikeBySlug(ctx context.Context, userID uint, forumSlug string) (bool, error)
	GetForumStats(ctx context.Context, forumID uint) (int64, error) // Likes count

	// New voting methods
	VotePost(ctx context.Context, userID, postID uint, voteType string) error
	RemovePostVote(ctx context.Context, userID, postID uint) error
	GetPostVoteStatus(ctx context.Context, userID, postID uint) (*model.ForumPostVote, error)

	// Best answer methods
	MarkAsAcceptedAnswer(ctx context.Context, userID uint, userRole string, postID uint) error
	UnmarkAcceptedAnswer(ctx context.Context, userID uint, userRole string, forumID uint) error
	UnmarkAcceptedAnswerBySlug(ctx context.Context, userID uint, userRole string, forumSlug string) error
	GetAcceptedAnswer(ctx context.Context, forumID uint) (*model.ForumPost, error)
	GetAcceptedAnswerBySlug(ctx context.Context, forumSlug string) (*model.ForumPost, error)

	// Report methods
	ReportPost(ctx context.Context, userID, postID uint, reason, description string) error
	GetPendingPostReports(ctx context.Context, limit, offset int) ([]model.ForumPostReport, int64, error)
	ReviewPostReport(ctx context.Context, reviewerID, reportID uint, status string, notes string) error
}

var ErrForumBlocked = errors.New("user is blocked from forum access")
var ErrCannotLikeOwnForum = errors.New("cannot like your own forum")
var ErrCannotVoteOwnPost = errors.New("cannot vote on your own post")

type forumService struct {
	repo                  infrastructure.ForumRepository
	userRepo              *authinfra.UserRepository
	gamificationService   *gamificationapp.GamificationService
	contentContextService *contentctx.ContentContextService
}

func NewForumService(repo infrastructure.ForumRepository, userRepo *authinfra.UserRepository, gamificationService *gamificationapp.GamificationService, contentContextService *contentctx.ContentContextService) ForumService {
	return &forumService{repo, userRepo, gamificationService, contentContextService}
}

// checkForumBlocked returns an error if the user is blocked from the forum
func (s *forumService) checkForumBlocked(ctx context.Context, userID uint) error {
	if s.userRepo == nil {
		return nil
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.IsForumBlocked {
		return ErrForumBlocked
	}
	return nil
}

func (s *forumService) awardExp(ctx context.Context, userID uint, activity gamificationpkg.ActivityType, points int64) {
	if s.gamificationService == nil || points == 0 {
		return
	}
	if err := s.gamificationService.AwardExp(ctx, userID, activity, points); err != nil {
		logger.Warn("forum: failed to award exp",
			zap.Uint("user_id", userID), zap.String("activity", string(activity)), zap.Error(err))
	}
}

func (s *forumService) enrichForumDetail(ctx context.Context, forum *model.Forum, userID uint) {
	if forum == nil {
		return
	}

	count, _ := s.repo.GetLikesCount(ctx, forum.ID)
	forum.LikesCount = count

	repliesCount, _ := s.repo.GetRepliesCount(ctx, forum.ID)
	forum.RepliesCount = repliesCount

	liked, _ := s.repo.HasUserLiked(ctx, userID, forum.ID)
	forum.IsLiked = liked
}

func (s *forumService) createForumPost(ctx context.Context, userID uint, forumID uint, content string) error {
	post := &model.ForumPost{
		UserID:  userID,
		ForumID: forumID,
		Content: content,
	}

	if err := s.repo.CreateForumPost(ctx, post); err != nil {
		return err
	}

	s.awardExp(ctx, userID, gamificationpkg.ActivityForumComment, gamificationpkg.ExpForumComment)
	return nil
}

func (s *forumService) CreateForum(ctx context.Context, userID uint, title, content string, categoryID *uint) (*model.Forum, error) {
	// Check if user is blocked from forum
	if err := s.checkForumBlocked(ctx, userID); err != nil {
		return nil, err
	}

	forum := &model.Forum{
		UserID:     userID,
		Title:      title,
		Content:    content,
		CategoryID: categoryID,
	}
	err := s.repo.CreateForum(ctx, forum)
	if err == nil && s.contentContextService != nil {
		s.contentContextService.NotifyForumChange(ctx, forum)
	}
	if err != nil {
		return nil, err
	}
	return forum, nil
}

func (s *forumService) GetForums(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error) {
	forums, total, err := s.repo.GetForums(ctx, limit, offset, search, categoryID)
	if err != nil {
		return nil, 0, err
	}

	// Populate likes count for each forum
	// Note: N+1 problem here, but acceptable for small scale. Better approach: join or subquery in repo.
	for i := range forums {
		count, _ := s.repo.GetLikesCount(ctx, forums[i].ID)
		forums[i].LikesCount = count

		repliesCount, _ := s.repo.GetRepliesCount(ctx, forums[i].ID)
		forums[i].RepliesCount = repliesCount
	}

	return forums, total, nil
}

func (s *forumService) GetForumByID(ctx context.Context, userID, id uint) (*model.Forum, error) {
	forum, err := s.repo.GetForumByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.enrichForumDetail(ctx, forum, userID)

	return forum, nil
}

func (s *forumService) GetForumBySlug(ctx context.Context, userID uint, slug string) (*model.Forum, error) {
	forum, err := s.repo.GetForumBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	s.enrichForumDetail(ctx, forum, userID)

	return forum, nil
}

func (s *forumService) DeleteForum(ctx context.Context, userID uint, userRole string, forumID uint) error {
	forum, err := s.repo.GetForumByID(ctx, forumID)
	if err != nil {
		return err
	}

	// Allow if user is owner or admin
	if forum.UserID != userID && userRole != "admin" {
		return errors.New("unauthorized")
	}

	err = s.repo.DeleteForum(ctx, forumID)
	if err == nil && s.contentContextService != nil {
		s.contentContextService.NotifyForumDelete(ctx, forumID)
	}
	return err
}

func (s *forumService) DeleteForumBySlug(ctx context.Context, userID uint, userRole string, slug string) error {
	forum, err := s.repo.GetForumBySlug(ctx, slug)
	if err != nil {
		return err
	}

	// Allow if user is owner or admin
	if forum.UserID != userID && userRole != "admin" {
		return errors.New("unauthorized")
	}

	err = s.repo.DeleteForum(ctx, forum.ID)
	if err == nil && s.contentContextService != nil {
		s.contentContextService.NotifyForumDelete(ctx, forum.ID)
	}
	return err
}

func (s *forumService) CreateForumPost(ctx context.Context, userID uint, forumID uint, content string) error {
	// Check if user is blocked from forum
	if err := s.checkForumBlocked(ctx, userID); err != nil {
		return err
	}

	return s.createForumPost(ctx, userID, forumID, content)
}

func (s *forumService) CreateForumPostBySlug(ctx context.Context, userID uint, forumSlug string, content string) error {
	// Check if user is blocked from forum
	if err := s.checkForumBlocked(ctx, userID); err != nil {
		return err
	}

	forum, err := s.repo.GetForumBySlug(ctx, forumSlug)
	if err != nil {
		return err
	}

	return s.createForumPost(ctx, userID, forum.ID, content)
}

func (s *forumService) GetForumPosts(ctx context.Context, forumID uint, limit, offset int) ([]model.ForumPost, int64, error) {
	return s.repo.GetForumPosts(ctx, forumID, limit, offset)
}

func (s *forumService) GetForumPostsBySlug(ctx context.Context, forumSlug string, limit, offset int) ([]model.ForumPost, int64, error) {
	forum, err := s.repo.GetForumBySlug(ctx, forumSlug)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.GetForumPosts(ctx, forum.ID, limit, offset)
}

func (s *forumService) GetForumPostsSorted(ctx context.Context, forumID uint, limit, offset int, sort string, userID uint) ([]model.ForumPost, int64, error) {
	sortOption := infrastructure.SortByTop // default
	switch sort {
	case "newest":
		sortOption = infrastructure.SortByNewest
	case "oldest":
		sortOption = infrastructure.SortByOldest
	case "top":
		sortOption = infrastructure.SortByTop
	}
	return s.repo.GetForumPostsSorted(ctx, forumID, limit, offset, sortOption, userID)
}

func (s *forumService) DeleteForumPost(ctx context.Context, userID uint, userRole string, postID uint) error {
	post, err := s.repo.GetForumPostByID(ctx, postID)
	if err != nil {
		return err
	}

	// Allow if user is owner or admin
	if post.UserID != userID && userRole != "admin" {
		return errors.New("unauthorized")
	}

	return s.repo.DeleteForumPost(ctx, postID)
}

func (s *forumService) ToggleLike(ctx context.Context, userID, forumID uint) (bool, error) {
	// Check if user is blocked from forum
	if err := s.checkForumBlocked(ctx, userID); err != nil {
		return false, err
	}

	forum, err := s.repo.GetForumByID(ctx, forumID)
	if err != nil {
		return false, err
	}

	if forum.UserID == userID {
		return false, ErrCannotLikeOwnForum
	}

	return s.repo.ToggleLike(ctx, userID, forumID)
}

func (s *forumService) ToggleLikeBySlug(ctx context.Context, userID uint, forumSlug string) (bool, error) {
	// Check if user is blocked from forum
	if err := s.checkForumBlocked(ctx, userID); err != nil {
		return false, err
	}

	forum, err := s.repo.GetForumBySlug(ctx, forumSlug)
	if err != nil {
		return false, err
	}

	if forum.UserID == userID {
		return false, ErrCannotLikeOwnForum
	}

	return s.repo.ToggleLike(ctx, userID, forum.ID)
}

func (s *forumService) GetForumPostsSortedBySlug(ctx context.Context, forumSlug string, limit, offset int, sort string, userID uint) ([]model.ForumPost, int64, error) {
	forum, err := s.repo.GetForumBySlug(ctx, forumSlug)
	if err != nil {
		return nil, 0, err
	}
	return s.GetForumPostsSorted(ctx, forum.ID, limit, offset, sort, userID)
}

func (s *forumService) GetForumStats(ctx context.Context, forumID uint) (int64, error) {
	return s.repo.GetLikesCount(ctx, forumID)
}

// ==================== Voting Methods ====================

func (s *forumService) VotePost(ctx context.Context, userID, postID uint, voteType string) error {
	// Validate vote type (only upvote allowed for now - mental health community)
	var vt model.VoteType
	switch voteType {
	case "upvote":
		vt = model.VoteTypeUpvote
	case "downvote":
		vt = model.VoteTypeDownvote
	default:
		return errors.New("invalid vote type")
	}

	// Get the post to check ownership and get author ID
	post, err := s.repo.GetForumPostByID(ctx, postID)
	if err != nil {
		return errors.New("post not found")
	}

	// Check if user is blocked from forum
	if err := s.checkForumBlocked(ctx, userID); err != nil {
		return err
	}

	// Rate limit: max 30 votes per hour
	voteCount, err := s.repo.CountUserVotesInLastHour(ctx, userID)
	if err == nil && voteCount >= 30 {
		return errors.New("vote rate limit exceeded, please try again later")
	}

	// Minimum account age: 1 day (Temporarily disabled for testing/development)
	// if s.userRepo != nil {
	// 	user, err := s.userRepo.FindByID(ctx, userID)
	// 	if err == nil && time.Since(user.CreatedAt).Hours() < 24 {
	// 		return errors.New("your account must be at least 1 day old to vote")
	// 	}
	// }

	// Spam detection: burst voting (10+ votes in 5 minutes)
	burstCount, err := s.repo.CountUserVotesInLastMinutes(ctx, userID, 5)
	if err == nil && burstCount >= 10 {
		return errors.New("voting too quickly, please slow down")
	}

	// Spam detection: same-author targeting (5+ votes on same author in 1 hour)
	authorVotes, err := s.repo.CountUserVotesOnAuthorInLastHour(ctx, userID, post.UserID)
	if err == nil && authorVotes >= 5 {
		return errors.New("too many votes on the same user's posts, please diversify")
	}

	// Prevent self-voting
	if post.UserID == userID {
		return ErrCannotVoteOwnPost
	}

	// Check if user already voted
	existingVote, _ := s.repo.GetUserPostVote(ctx, userID, postID)
	isNewVote := existingVote == nil
	isChangingVote := existingVote != nil && existingVote.VoteType != vt

	// Perform the vote
	if err := s.repo.VotePost(ctx, userID, postID, vt); err != nil {
		return err
	}

	// Award XP to post author for upvotes
	if vt == model.VoteTypeUpvote && (isNewVote || isChangingVote) {
		s.awardExp(ctx, post.UserID, gamificationpkg.ActivityPostUpvoteGiven, gamificationpkg.ExpPostUpvoteGiven)
	}

	// Update community favorite status
	go func() {
		_ = s.repo.UpdateCommunityFavorite(ctx, post.ForumID)
	}()

	return nil
}

func (s *forumService) RemovePostVote(ctx context.Context, userID, postID uint) error {
	// Get the existing vote to check if it was an upvote
	existingVote, err := s.repo.GetUserPostVote(ctx, userID, postID)
	if err != nil || existingVote == nil {
		return errors.New("no vote to remove")
	}

	// Get the post for author info
	post, err := s.repo.GetForumPostByID(ctx, postID)
	if err != nil {
		return errors.New("post not found")
	}

	// Remove the vote
	if err := s.repo.RemovePostVote(ctx, userID, postID); err != nil {
		return err
	}

	// Note: We could deduct XP when upvote is removed, but for mental health community,
	// it's better to not punish users. Just don't give additional XP.

	// Update community favorite status
	go func() {
		_ = s.repo.UpdateCommunityFavorite(ctx, post.ForumID)
	}()

	return nil
}

func (s *forumService) GetPostVoteStatus(ctx context.Context, userID, postID uint) (*model.ForumPostVote, error) {
	return s.repo.GetUserPostVote(ctx, userID, postID)
}

// ==================== Best Answer Methods ====================

func (s *forumService) MarkAsAcceptedAnswer(ctx context.Context, userID uint, userRole string, postID uint) error {
	// Get the post
	post, err := s.repo.GetForumPostByID(ctx, postID)
	if err != nil {
		return errors.New("post not found")
	}

	// Get the forum to check ownership
	forum, err := s.repo.GetForumByID(ctx, post.ForumID)
	if err != nil {
		return errors.New("forum not found")
	}

	// Only OP (thread creator) or admin can mark accepted answer
	if forum.UserID != userID && userRole != "admin" {
		return errors.New("only the thread creator can mark an accepted answer")
	}

	// Cannot mark own answer as accepted (unless admin)
	if post.UserID == userID && userRole != "admin" {
		return errors.New("cannot mark your own answer as accepted")
	}

	// Check if this post was already accepted
	wasAlreadyAccepted := post.IsAcceptedAnswer

	// Mark as accepted
	if err := s.repo.MarkAsAcceptedAnswer(ctx, postID); err != nil {
		return err
	}

	// Award XP to the answer author (only if not already accepted)
	if !wasAlreadyAccepted {
		s.awardExp(ctx, post.UserID, gamificationpkg.ActivityAcceptedAnswer, gamificationpkg.ExpAcceptedAnswer)
	}

	return nil
}

func (s *forumService) UnmarkAcceptedAnswer(ctx context.Context, userID uint, userRole string, forumID uint) error {
	// Get the forum to check ownership
	forum, err := s.repo.GetForumByID(ctx, forumID)
	if err != nil {
		return errors.New("forum not found")
	}

	// Only OP or admin can unmark accepted answer
	if forum.UserID != userID && userRole != "admin" {
		return errors.New("only the thread creator can unmark an accepted answer")
	}

	return s.repo.UnmarkAcceptedAnswer(ctx, forumID)
}

func (s *forumService) GetAcceptedAnswer(ctx context.Context, forumID uint) (*model.ForumPost, error) {
	return s.repo.GetAcceptedAnswer(ctx, forumID)
}

func (s *forumService) GetAcceptedAnswerBySlug(ctx context.Context, forumSlug string) (*model.ForumPost, error) {
	forum, err := s.repo.GetForumBySlug(ctx, forumSlug)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAcceptedAnswer(ctx, forum.ID)
}

func (s *forumService) UnmarkAcceptedAnswerBySlug(ctx context.Context, userID uint, userRole string, forumSlug string) error {
	forum, err := s.repo.GetForumBySlug(ctx, forumSlug)
	if err != nil {
		return errors.New("forum not found")
	}
	return s.UnmarkAcceptedAnswer(ctx, userID, userRole, forum.ID)
}

// ==================== Report Methods ====================

func (s *forumService) ReportPost(ctx context.Context, userID, postID uint, reason, description string) error {
	// Validate reason
	if !model.IsValidPostReportReason(reason) {
		return errors.New("invalid report reason")
	}

	// Check if post exists
	post, err := s.repo.GetForumPostByID(ctx, postID)
	if err != nil {
		return errors.New("post not found")
	}

	// Cannot report own post
	if post.UserID == userID {
		return errors.New("cannot report your own post")
	}

	// Check if user already reported this post
	hasReported, _ := s.repo.HasUserReportedPost(ctx, userID, postID)
	if hasReported {
		return errors.New("you have already reported this post")
	}

	// Create report
	report := &model.ForumPostReport{
		PostID:      postID,
		ReporterID:  userID,
		Reason:      model.ForumPostReportReason(reason),
		Description: description,
		Status:      model.PostReportStatusPending,
	}

	return s.repo.CreatePostReport(ctx, report)
}

func (s *forumService) GetPendingPostReports(ctx context.Context, limit, offset int) ([]model.ForumPostReport, int64, error) {
	return s.repo.GetPendingPostReports(ctx, limit, offset)
}

func (s *forumService) ReviewPostReport(ctx context.Context, reviewerID, reportID uint, status string, notes string) error {
	// Validate status
	var reportStatus model.ForumPostReportStatus
	switch status {
	case "reviewed":
		reportStatus = model.PostReportStatusReviewed
	case "dismissed":
		reportStatus = model.PostReportStatusDismissed
	case "actioned":
		reportStatus = model.PostReportStatusActioned
	default:
		return errors.New("invalid report status")
	}

	// Get the report
	report, err := s.repo.GetPostReportByID(ctx, reportID)
	if err != nil {
		return errors.New("report not found")
	}

	// Update report
	now := time.Now()
	report.Status = reportStatus
	report.ReviewedBy = &reviewerID
	report.ReviewedAt = &now
	report.ModeratorNotes = notes

	// If actioned, flag the post
	if reportStatus == model.PostReportStatusActioned {
		post, err := s.repo.GetForumPostByID(ctx, report.PostID)
		if err == nil {
			post.IsFlagged = true
			post.FlaggedReason = string(report.Reason) + ": " + notes
			_ = s.repo.UpdateForumPost(ctx, post)
		}
	}

	return s.repo.UpdatePostReport(ctx, report)
}
