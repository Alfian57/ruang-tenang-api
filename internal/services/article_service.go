package services

import (
	"errors"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type ArticleService struct {
	articleRepo  *repositories.ArticleRepository
	categoryRepo *repositories.ArticleCategoryRepository
}

func NewArticleService(articleRepo *repositories.ArticleRepository, categoryRepo *repositories.ArticleCategoryRepository) *ArticleService {
	return &ArticleService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
	}
}

// GetPublishedArticles returns only published articles for public view
func (s *ArticleService) GetPublishedArticles(params *dto.ArticleQueryParams) ([]dto.ArticleListDTO, int64, error) {
	articles, total, err := s.articleRepo.FindPublished(params.CategoryID, params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	return s.articlesToListDTO(articles), total, nil
}

// GetArticles returns articles with optional filters (for admin)
func (s *ArticleService) GetArticles(params *dto.ArticleQueryParams) ([]dto.ArticleListDTO, int64, error) {
	articles, total, err := s.articleRepo.FindAll(params.CategoryID, params.Search, params.Page, params.Limit, params.Status, params.UserID)
	if err != nil {
		return nil, 0, err
	}

	return s.articlesToListDTO(articles), total, nil
}

// GetUserArticles returns articles owned by a specific user
func (s *ArticleService) GetUserArticles(userID uint, page, limit int) ([]dto.ArticleListDTO, int64, error) {
	articles, total, err := s.articleRepo.FindByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return s.articlesToListDTO(articles), total, nil
}

func (s *ArticleService) articlesToListDTO(articles []models.Article) []dto.ArticleListDTO {
	var result []dto.ArticleListDTO
	for _, article := range articles {
		excerpt := article.Content
		if len(excerpt) > 150 {
			excerpt = excerpt[:150] + "..."
		}
		// Remove HTML tags from excerpt
		excerpt = strings.ReplaceAll(excerpt, "<p>", "")
		excerpt = strings.ReplaceAll(excerpt, "</p>", " ")

		item := dto.ArticleListDTO{
			ID:         article.ID,
			Title:      article.Title,
			Thumbnail:  article.Thumbnail,
			Excerpt:    excerpt,
			CategoryID: article.ArticleCategoryID,
			Category: dto.ArticleCategoryDTO{
				ID:        article.Category.ID,
				Name:      article.Category.Name,
				CreatedAt: article.Category.CreatedAt,
			},
			UserID:    article.UserID,
			Status:    string(article.Status),
			CreatedAt: article.CreatedAt,
		}

		if article.Author != nil {
			item.Author = &dto.ArticleAuthorDTO{
				ID:   article.Author.ID,
				Name: article.Author.Name,
			}
		}

		result = append(result, item)
	}

	return result
}

func (s *ArticleService) GetArticleByID(id uint) (*dto.ArticleDTO, error) {
	article, err := s.articleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	result := &dto.ArticleDTO{
		ID:         article.ID,
		Title:      article.Title,
		Thumbnail:  article.Thumbnail,
		Content:    article.Content,
		CategoryID: article.ArticleCategoryID,
		Category: dto.ArticleCategoryDTO{
			ID:        article.Category.ID,
			Name:      article.Category.Name,
			CreatedAt: article.Category.CreatedAt,
		},
		UserID:    article.UserID,
		Status:    string(article.Status),
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}

	if article.Author != nil {
		result.Author = &dto.ArticleAuthorDTO{
			ID:   article.Author.ID,
			Name: article.Author.Name,
		}
	}

	return result, nil
}

// GetPublishedArticleByID returns an article only if it's published
func (s *ArticleService) GetPublishedArticleByID(id uint) (*dto.ArticleDTO, error) {
	article, err := s.GetArticleByID(id)
	if err != nil {
		return nil, err
	}

	if article.Status != string(models.ArticleStatusPublished) {
		return nil, errors.New("article not found")
	}

	return article, nil
}

func (s *ArticleService) GetCategories() ([]dto.ArticleCategoryDTO, error) {
	categories, err := s.categoryRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var result []dto.ArticleCategoryDTO
	for _, category := range categories {
		result = append(result, dto.ArticleCategoryDTO{
			ID:        category.ID,
			Name:      category.Name,
			CreatedAt: category.CreatedAt,
		})
	}

	return result, nil
}

// CreateArticle creates a new article (admin)
func (s *ArticleService) CreateArticle(article *models.Article) error {
	return s.articleRepo.Create(article)
}

// CreateUserArticle creates a new article for a user
func (s *ArticleService) CreateUserArticle(userID uint, req *dto.CreateUserArticleRequest) (*models.Article, error) {
	article := &models.Article{
		Title:             req.Title,
		Thumbnail:         req.Thumbnail,
		Content:           req.Content,
		ArticleCategoryID: req.CategoryID,
		UserID:            &userID,
		Status:            models.ArticleStatusPublished,
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, err
	}

	return article, nil
}

// UpdateUserArticle updates an article owned by the user
func (s *ArticleService) UpdateUserArticle(userID uint, articleID uint, req *dto.UpdateUserArticleRequest) (*models.Article, error) {
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if article.UserID == nil || *article.UserID != userID {
		return nil, errors.New("not authorized to update this article")
	}

	// Check if blocked
	if article.Status == models.ArticleStatusBlocked {
		return nil, errors.New("cannot update blocked article")
	}

	article.Title = req.Title
	article.Thumbnail = req.Thumbnail
	article.Content = req.Content
	article.ArticleCategoryID = req.CategoryID

	if err := s.articleRepo.Update(article); err != nil {
		return nil, err
	}

	return article, nil
}

// DeleteUserArticle deletes an article owned by the user
func (s *ArticleService) DeleteUserArticle(userID uint, articleID uint) error {
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return err
	}

	// Check ownership
	if article.UserID == nil || *article.UserID != userID {
		return errors.New("not authorized to delete this article")
	}

	return s.articleRepo.Delete(articleID)
}

// BlockArticle blocks an article (admin only)
func (s *ArticleService) BlockArticle(articleID uint) error {
	return s.articleRepo.UpdateStatus(articleID, models.ArticleStatusBlocked)
}

// UnblockArticle unblocks an article (admin only)
func (s *ArticleService) UnblockArticle(articleID uint) error {
	return s.articleRepo.UpdateStatus(articleID, models.ArticleStatusPublished)
}

func (s *ArticleService) CreateCategory(category *models.ArticleCategory) error {
	return s.categoryRepo.Create(category)
}

func (s *ArticleService) UpdateArticle(article *models.Article) error {
	return s.articleRepo.Update(article)
}

func (s *ArticleService) DeleteArticle(articleID uint) error {
	return s.articleRepo.Delete(articleID)
}
