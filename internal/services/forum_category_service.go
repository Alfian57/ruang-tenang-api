package services

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type ForumCategoryService interface {
	CreateCategory(name string) error
	GetAllCategories() ([]models.ForumCategory, error)
	UpdateCategory(id uint, name string) error
	DeleteCategory(id uint) error
}

type forumCategoryService struct {
	repo         repositories.ForumCategoryRepository
	cacheService *CacheService
}

func NewForumCategoryService(repo repositories.ForumCategoryRepository, cacheService *CacheService) ForumCategoryService {
	return &forumCategoryService{repo: repo, cacheService: cacheService}
}

func (s *forumCategoryService) CreateCategory(name string) error {
	category := &models.ForumCategory{
		Name: name,
	}
	err := s.repo.Create(category)
	if err == nil {
		s.cacheService.Delete(CacheKeyForumCategories)
	}
	return err
}

func (s *forumCategoryService) GetAllCategories() ([]models.ForumCategory, error) {
	// Check cache first
	if cached := s.cacheService.Get(CacheKeyForumCategories); cached != nil {
		return cached.([]models.ForumCategory), nil
	}

	categories, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cacheService.SetWithTTL(CacheKeyForumCategories, categories, s.cacheService.CategoryTTL)
	return categories, nil
}

func (s *forumCategoryService) UpdateCategory(id uint, name string) error {
	category, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	category.Name = name
	err = s.repo.Update(category)
	if err == nil {
		s.cacheService.Delete(CacheKeyForumCategories)
	}
	return err
}

func (s *forumCategoryService) DeleteCategory(id uint) error {
	err := s.repo.Delete(id)
	if err == nil {
		s.cacheService.Delete(CacheKeyForumCategories)
	}
	return err
}
