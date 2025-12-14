package services

import (
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type SongService struct {
	songRepo     *repositories.SongRepository
	categoryRepo *repositories.SongCategoryRepository
}

func NewSongService(songRepo *repositories.SongRepository, categoryRepo *repositories.SongCategoryRepository) *SongService {
	return &SongService{
		songRepo:     songRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *SongService) GetCategories() ([]dto.SongCategoryDTO, error) {
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

func (s *SongService) CreateCategory(category *models.SongCategory) error {
	return s.categoryRepo.Create(category)
}

func (s *SongService) CreateSong(song *models.Song) error {
	return s.songRepo.Create(song)
}
