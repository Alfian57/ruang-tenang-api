package service

import (
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type SongService struct {
	songRepo     *repository.SongRepository
	categoryRepo *repository.SongCategoryRepository
	cacheService *CacheService
}

func NewSongService(songRepo *repository.SongRepository, categoryRepo *repository.SongCategoryRepository, cacheService *CacheService) *SongService {
	return &SongService{
		songRepo:     songRepo,
		categoryRepo: categoryRepo,
		cacheService: cacheService,
	}
}

func (s *SongService) GetCategories() ([]dto.SongCategoryDTO, error) {
	// Check cache first
	if cached := s.cacheService.Get(CacheKeySongCategories); cached != nil {
		return cached.([]dto.SongCategoryDTO), nil
	}

	categories, err := s.categoryRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var result []dto.SongCategoryDTO
	for _, category := range categories {
		songCount := s.songRepo.CountByCategoryID(category.ID)
		result = append(result, dto.SongCategoryDTO{
			ID:        category.ID,
			Name:      category.Name,
			Thumbnail: category.Thumbnail,
			SongCount: int(songCount),
			CreatedAt: category.CreatedAt,
		})
	}

	// Store in cache
	s.cacheService.SetWithTTL(CacheKeySongCategories, result, s.cacheService.CategoryTTL)
	return result, nil
}

func (s *SongService) GetSongsByCategory(categoryID uint) ([]dto.SongListDTO, error) {
	songs, err := s.songRepo.FindByCategoryID(categoryID)
	if err != nil {
		return nil, err
	}

	var result []dto.SongListDTO
	for _, song := range songs {
		result = append(result, dto.SongListDTO{
			ID:         song.ID,
			Title:      song.Title,
			FilePath:   song.FilePath,
			Thumbnail:  song.Thumbnail,
			CategoryID: song.SongCategoryID,
		})
	}

	return result, nil
}

func (s *SongService) GetSongByID(id uint) (*dto.SongDTO, error) {
	song, err := s.songRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return &dto.SongDTO{
		ID:         song.ID,
		Title:      song.Title,
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

func (s *SongService) CreateCategory(category *model.SongCategory) error {
	err := s.categoryRepo.Create(category)
	if err == nil {
		s.cacheService.Delete(CacheKeySongCategories)
	}
	return err
}

func (s *SongService) CreateSong(song *model.Song) error {
	err := s.songRepo.Create(song)
	if err == nil {
		// Invalidate categories cache as song count changes
		s.cacheService.Delete(CacheKeySongCategories)
	}
	return err
}
