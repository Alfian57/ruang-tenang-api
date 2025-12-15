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
	repo repositories.ForumCategoryRepository
}

func NewForumCategoryService(repo repositories.ForumCategoryRepository) ForumCategoryService {
	return &forumCategoryService{repo}
}

func (s *forumCategoryService) CreateCategory(name string) error {
	category := &models.ForumCategory{
		Name: name,
	}
	return s.repo.Create(category)
}

func (s *forumCategoryService) GetAllCategories() ([]models.ForumCategory, error) {
	return s.repo.FindAll()
}

func (s *forumCategoryService) UpdateCategory(id uint, name string) error {
	category, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	category.Name = name
	return s.repo.Update(category)
}

func (s *forumCategoryService) DeleteCategory(id uint) error {
	return s.repo.Delete(id)
}
