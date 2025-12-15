package repositories

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"

	"gorm.io/gorm"
)

type ForumRepository interface {
	CreateForum(forum *models.Forum) error
	GetForums(limit, offset int, search string, categoryID *uint) ([]models.Forum, int64, error)
	GetForumByID(id uint) (*models.Forum, error)
	DeleteForum(id uint) error

	CreateForumPost(post *models.ForumPost) error
	GetForumPosts(forumID uint, limit, offset int) ([]models.ForumPost, int64, error)
	DeleteForumPost(id uint) error
	GetForumPostByID(id uint) (*models.ForumPost, error)

	ToggleLike(userID, forumID uint) (bool, error)
	GetLikesCount(forumID uint) (int64, error)
	HasUserLiked(userID, forumID uint) (bool, error)
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
	err := r.db.First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
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
