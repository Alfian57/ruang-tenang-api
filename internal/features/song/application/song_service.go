package application

import (
	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/internal/features/song/infrastructure")

type SongService struct {
	songRepo     *infrastructure.SongRepository
	categoryRepo *infrastructure.SongCategoryRepository
	cacheService *cache.CacheService
}

func NewSongService(songRepo *infrastructure.SongRepository, categoryRepo *infrastructure.SongCategoryRepository, cacheService *cache.CacheService) *SongService {
	return &SongService{
		songRepo:     songRepo,
		categoryRepo: categoryRepo,
		cacheService: cacheService,
	}
}

func (s *SongService) GetCategories(ctx context.Context) ([]dto.SongCategoryDTO, error) {
	// Check cache first
	if cached := s.cacheService.Get(cache.CacheKeySongCategories); cached != nil {
		return cached.([]dto.SongCategoryDTO), nil
	}

	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []dto.SongCategoryDTO
	for _, category := range categories {
		songCount := s.songRepo.CountByCategoryID(ctx, category.ID)
		result = append(result, dto.SongCategoryDTO{
			ID:        category.ID,
			Slug:      category.Slug,
			Name:      category.Name,
			Thumbnail: category.Thumbnail,
			SongCount: int(songCount),
			CreatedAt: category.CreatedAt,
		})
	}

	// Store in cache
	s.cacheService.SetWithTTL(cache.CacheKeySongCategories, result, s.cacheService.CategoryTTL)
	return result, nil
}

func (s *SongService) GetSongsByCategory(ctx context.Context, categoryID uint) ([]dto.SongListDTO, error) {
	songs, err := s.songRepo.FindByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	var result []dto.SongListDTO
	for _, song := range songs {
		result = append(result, dto.SongListDTO{
			ID:         song.ID,
			Slug:       song.Slug,
			Title:      song.Title,
			FilePath:   song.FilePath,
			Thumbnail:  song.Thumbnail,
			CategoryID: song.SongCategoryID,
		})
	}

	return result, nil
}

func (s *SongService) GetSongsByCategoryBySlug(ctx context.Context, categorySlug string) ([]dto.SongListDTO, error) {
	category, err := s.categoryRepo.FindBySlug(ctx, categorySlug)
	if err != nil {
		return nil, err
	}
	return s.GetSongsByCategory(ctx, category.ID)
}

func (s *SongService) GetSongByID(ctx context.Context, id uint) (*dto.SongDTO, error) {
	song, err := s.songRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.SongDTO{
		ID:         song.ID,
		Title:      song.Title,
		Slug:       song.Slug,
		FilePath:   song.FilePath,
		Thumbnail:  song.Thumbnail,
		CategoryID: song.SongCategoryID,
		Category: dto.SongCategoryDTO{
			ID:        song.Category.ID,
			Name:      song.Category.Name,
			Thumbnail: song.Category.Thumbnail,
			CreatedAt: song.Category.CreatedAt,
		},
		CreatedAt: song.CreatedAt,
	}, nil
}

func (s *SongService) GetSongBySlug(ctx context.Context, slug string) (*dto.SongDTO, error) {
	song, err := s.songRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return &dto.SongDTO{
		ID:         song.ID,
		Title:      song.Title,
		Slug:       song.Slug,
		FilePath:   song.FilePath,
		Thumbnail:  song.Thumbnail,
		CategoryID: song.SongCategoryID,
		Category: dto.SongCategoryDTO{
			ID:        song.Category.ID,
			Name:      song.Category.Name,
			Thumbnail: song.Category.Thumbnail,
			CreatedAt: song.Category.CreatedAt,
		},
		CreatedAt: song.CreatedAt,
	}, nil
}

func (s *SongService) CreateCategory(ctx context.Context, category *model.SongCategory) error {
	err := s.categoryRepo.Create(ctx, category)
	if err == nil {
		s.cacheService.Delete(cache.CacheKeySongCategories)
	}
	return err
}

func (s *SongService) CreateSong(ctx context.Context, song *model.Song) error {
	err := s.songRepo.Create(ctx, song)
	if err == nil {
		// Invalidate categories cache as song count changes
		s.cacheService.Delete(cache.CacheKeySongCategories)
	}
	return err
}
