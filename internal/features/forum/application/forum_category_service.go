package application

import (
	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
	"context"
	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/internal/features/forum/infrastructure")

type ForumCategoryService interface {
	CreateCategory(ctx context.Context, name string) error
	GetAllCategories(ctx context.Context, ) ([]model.ForumCategory, error)
	UpdateCategory(ctx context.Context, id uint, name string) error
	DeleteCategory(ctx context.Context, id uint) error
}

type forumCategoryService struct {
	repo         infrastructure.ForumCategoryRepository
	cacheService *cache.CacheService
}

func NewForumCategoryService(repo infrastructure.ForumCategoryRepository, cacheService *cache.CacheService) ForumCategoryService {
	return &forumCategoryService{repo: repo, cacheService: cacheService}
}

func (s *forumCategoryService) CreateCategory(ctx context.Context, name string) error {
	category := &model.ForumCategory{
		Name: name,
	}
	err := s.repo.Create(ctx, category)
	if err == nil {
		s.cacheService.Delete(cache.CacheKeyForumCategories)
	}
	return err
}

func (s *forumCategoryService) GetAllCategories(ctx context.Context) ([]model.ForumCategory, error) {
	// Check cache first
	if cached := s.cacheService.Get(cache.CacheKeyForumCategories); cached != nil {
		return cached.([]model.ForumCategory), nil
	}

	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cacheService.SetWithTTL(cache.CacheKeyForumCategories, categories, s.cacheService.CategoryTTL)
	return categories, nil
}

func (s *forumCategoryService) UpdateCategory(ctx context.Context, id uint, name string) error {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	category.Name = name
	err = s.repo.Update(ctx, category)
	if err == nil {
		s.cacheService.Delete(cache.CacheKeyForumCategories)
	}
	return err
}

func (s *forumCategoryService) DeleteCategory(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err == nil {
		s.cacheService.Delete(cache.CacheKeyForumCategories)
	}
	return err
}
