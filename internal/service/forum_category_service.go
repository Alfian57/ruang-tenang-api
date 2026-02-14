package service

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type ForumCategoryService interface {
	CreateCategory(name string) error
	GetAllCategories() ([]model.ForumCategory, error)
	UpdateCategory(id uint, name string) error
	DeleteCategory(id uint) error
}

type forumCategoryService struct {
	repo         repository.ForumCategoryRepository
	cacheService *CacheService
}

func NewForumCategoryService(repo repository.ForumCategoryRepository, cacheService *CacheService) ForumCategoryService {
	return &forumCategoryService{repo: repo, cacheService: cacheService}
}

func (s *forumCategoryService) CreateCategory(name string) error {
	category := &model.ForumCategory{
		Name: name,
	}
	err := s.repo.Create(category)
	if err == nil {
		s.cacheService.Delete(CacheKeyForumCategories)
	}
	return err
}

func (s *forumCategoryService) GetAllCategories() ([]model.ForumCategory, error) {
	// Check cache first
	if cached := s.cacheService.Get(CacheKeyForumCategories); cached != nil {
		return cached.([]model.ForumCategory), nil
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
