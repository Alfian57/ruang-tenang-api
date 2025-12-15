package services

import (
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/Alfian57/ruang-tenang-api/pkg/gamification"
)

type ForumService interface {
	CreateForum(userID uint, title, content string, categoryID *uint) error
	GetForums(limit, offset int, search string, categoryID *uint) ([]models.Forum, int64, error)
	GetForumByID(userID, id uint) (*models.Forum, error)
	DeleteForum(userID uint, userRole string, forumID uint) error

	CreateForumPost(userID uint, forumID uint, content string) error
	GetForumPosts(forumID uint, limit, offset int) ([]models.ForumPost, int64, error)
	DeleteForumPost(userID uint, userRole string, postID uint) error

	ToggleLike(userID, forumID uint) (bool, error)
	GetForumStats(forumID uint) (int64, error) // Likes count
}

type forumService struct {
	repo                repositories.ForumRepository
	gamificationService *GamificationService
}

func NewForumService(repo repositories.ForumRepository, gamificationService *GamificationService) ForumService {
	return &forumService{repo, gamificationService}
}

func (s *forumService) CreateForum(userID uint, title, content string, categoryID *uint) error {
	forum := &models.Forum{
		UserID:     userID,
		Title:      title,
		Content:    content,
		CategoryID: categoryID,
	}
	return s.repo.CreateForum(forum)
}

func (s *forumService) GetForums(limit, offset int, search string, categoryID *uint) ([]models.Forum, int64, error) {
	forums, total, err := s.repo.GetForums(limit, offset, search, categoryID)
	if err != nil {
		return nil, 0, err
	}

	// Populate likes count for each forum
	// Note: N+1 problem here, but acceptable for small scale. Better approach: join or subquery in repo.
	for i := range forums {
		count, _ := s.repo.GetLikesCount(forums[i].ID)
		forums[i].LikesCount = count
	}

	return forums, total, nil
}

func (s *forumService) GetForumByID(userID, id uint) (*models.Forum, error) {
	forum, err := s.repo.GetForumByID(id)
	if err != nil {
		return nil, err
	}
	// Get likes count
	count, _ := s.repo.GetLikesCount(id)
	forum.LikesCount = count

	// Check if liked
	liked, _ := s.repo.HasUserLiked(userID, id)
	forum.IsLiked = liked

	return forum, nil
}

func (s *forumService) DeleteForum(userID uint, userRole string, forumID uint) error {
	forum, err := s.repo.GetForumByID(forumID)
	if err != nil {
		return err
	}

	// Allow if user is owner or admin
	if forum.UserID != userID && userRole != "admin" {
		return errors.New("unauthorized")
	}

	return s.repo.DeleteForum(forumID)
}

func (s *forumService) CreateForumPost(userID uint, forumID uint, content string) error {
	post := &models.ForumPost{
		UserID:  userID,
		ForumID: forumID,
		Content: content,
	}
	err := s.repo.CreateForumPost(post)
	if err != nil {
		return err
	}

	// Award EXP for commenting
	go func() {
		// We ignore error here since it's a side effect and shouldn't block the main flow
		_ = s.gamificationService.AwardExp(userID, gamification.ActivityForumComment, gamification.ExpForumComment)
	}()

	return nil
}

func (s *forumService) GetForumPosts(forumID uint, limit, offset int) ([]models.ForumPost, int64, error) {
	return s.repo.GetForumPosts(forumID, limit, offset)
}

func (s *forumService) DeleteForumPost(userID uint, userRole string, postID uint) error {
	post, err := s.repo.GetForumPostByID(postID)
	if err != nil {
		return err
	}

	// Allow if user is owner or admin
	if post.UserID != userID && userRole != "admin" {
		return errors.New("unauthorized")
	}

	return s.repo.DeleteForumPost(postID)
}

func (s *forumService) ToggleLike(userID, forumID uint) (bool, error) {
	return s.repo.ToggleLike(userID, forumID)
}

func (s *forumService) GetForumStats(forumID uint) (int64, error) {
	return s.repo.GetLikesCount(forumID)
}
