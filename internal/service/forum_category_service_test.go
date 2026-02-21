package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type mockForumCategoryRepo struct {
	findAllCalls int
	createCalls  int
	updateCalls  int
	deleteCalls  int

	findAllResult []model.ForumCategory
	findByIDModel *model.ForumCategory

	createErr   error
	findAllErr  error
	findByIDErr error
	updateErr   error
	deleteErr   error
}

func (m *mockForumCategoryRepo) Create(_ context.Context, category *model.ForumCategory) error {
	m.createCalls++
	if m.createErr == nil && m.findByIDModel == nil {
		m.findByIDModel = &model.ForumCategory{ID: 1, Name: category.Name}
	}
	return m.createErr
}

func (m *mockForumCategoryRepo) FindAll(_ context.Context) ([]model.ForumCategory, error) {
	m.findAllCalls++
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.findAllResult, nil
}

func (m *mockForumCategoryRepo) FindByID(_ context.Context, _ uint) (*model.ForumCategory, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if m.findByIDModel == nil {
		m.findByIDModel = &model.ForumCategory{ID: 1, Name: "old"}
	}
	return m.findByIDModel, nil
}

func (m *mockForumCategoryRepo) FindBySlug(_ context.Context, _ string) (*model.ForumCategory, error) {
	return nil, errors.New("not implemented")
}

func (m *mockForumCategoryRepo) Update(_ context.Context, _ *model.ForumCategory) error {
	m.updateCalls++
	return m.updateErr
}

func (m *mockForumCategoryRepo) Delete(_ context.Context, _ uint) error {
	m.deleteCalls++
	return m.deleteErr
}

func TestForumCategoryService_GetAllCategoriesUsesCache(t *testing.T) {
	repo := &mockForumCategoryRepo{
		findAllResult: []model.ForumCategory{{ID: 1, Name: "General"}},
	}
	cache := NewCacheService()
	svc := NewForumCategoryService(repo, cache)

	ctx := context.Background()
	first, err := svc.GetAllCategories(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := svc.GetAllCategories(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("unexpected categories: %+v %+v", first, second)
	}
	if repo.findAllCalls != 1 {
		t.Fatalf("expected repo FindAll to be called once, got %d", repo.findAllCalls)
	}
}

func TestForumCategoryService_CreateAndDeleteInvalidateCache(t *testing.T) {
	repo := &mockForumCategoryRepo{
		findAllResult: []model.ForumCategory{{ID: 1, Name: "General"}},
	}
	cache := NewCacheService()
	svc := NewForumCategoryService(repo, cache)
	ctx := context.Background()

	if _, err := svc.GetAllCategories(ctx); err != nil {
		t.Fatalf("unexpected initial get error: %v", err)
	}

	if err := svc.CreateCategory(ctx, "Mindfulness"); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if _, err := svc.GetAllCategories(ctx); err != nil {
		t.Fatalf("unexpected get error after create: %v", err)
	}

	if err := svc.DeleteCategory(ctx, 1); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if _, err := svc.GetAllCategories(ctx); err != nil {
		t.Fatalf("unexpected get error after delete: %v", err)
	}

	if repo.createCalls != 1 || repo.deleteCalls != 1 {
		t.Fatalf("unexpected create/delete calls: create=%d delete=%d", repo.createCalls, repo.deleteCalls)
	}
	if repo.findAllCalls != 3 {
		t.Fatalf("expected cache invalidation to force refetches, got findAllCalls=%d", repo.findAllCalls)
	}
}

func TestForumCategoryService_UpdateCategory(t *testing.T) {
	repo := &mockForumCategoryRepo{
		findByIDModel: &model.ForumCategory{ID: 10, Name: "Old"},
	}
	cache := NewCacheService()
	svc := NewForumCategoryService(repo, cache)

	err := svc.UpdateCategory(context.Background(), 10, "New Name")
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected update call, got %d", repo.updateCalls)
	}
	if repo.findByIDModel.Name != "New Name" {
		t.Fatalf("expected category name to be updated, got %s", repo.findByIDModel.Name)
	}
}
