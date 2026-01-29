package services

import (
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"gorm.io/gorm"
)

// PlaylistService handles playlist business logic
type PlaylistService struct {
	playlistRepo     *repositories.PlaylistRepository
	playlistItemRepo *repositories.PlaylistItemRepository
	songRepo         *repositories.SongRepository
}

// NewPlaylistService creates a new playlist service
func NewPlaylistService(
	playlistRepo *repositories.PlaylistRepository,
	playlistItemRepo *repositories.PlaylistItemRepository,
	songRepo *repositories.SongRepository,
) *PlaylistService {
	return &PlaylistService{
		playlistRepo:     playlistRepo,
		playlistItemRepo: playlistItemRepo,
		songRepo:         songRepo,
	}
}

// CreatePlaylist creates a new playlist for a user
func (s *PlaylistService) CreatePlaylist(userID uint, req *dto.CreatePlaylistRequest) (*dto.PlaylistDTO, error) {
	playlist := &models.Playlist{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		IsPublic:    req.IsPublic,
	}

	if err := s.playlistRepo.Create(playlist); err != nil {
		return nil, err
	}

	return s.toPlaylistDTO(playlist, 0), nil
}

// GetUserPlaylists gets all playlists for a user
func (s *PlaylistService) GetUserPlaylists(userID uint) ([]dto.PlaylistListDTO, error) {
	playlists, itemCounts, err := s.playlistRepo.FindByUserIDWithItemCount(userID)
	if err != nil {
		return nil, err
	}

	var result []dto.PlaylistListDTO
	for _, p := range playlists {
		result = append(result, dto.PlaylistListDTO{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Thumbnail:   p.Thumbnail,
			IsPublic:    p.IsPublic,
			ItemCount:   itemCounts[p.ID],
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return result, nil
}

// GetPlaylist gets a playlist by ID
func (s *PlaylistService) GetPlaylist(playlistID, userID uint) (*dto.PlaylistDTO, error) {
	playlist, err := s.playlistRepo.FindByIDWithItems(playlistID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Check access: user must be owner or playlist must be public
	if playlist.UserID != userID && !playlist.IsPublic {
		return nil, ErrForbidden
	}

	return s.toPlaylistDTOWithItems(playlist), nil
}

// UpdatePlaylist updates a playlist
func (s *PlaylistService) UpdatePlaylist(playlistID, userID uint, req *dto.UpdatePlaylistRequest) (*dto.PlaylistDTO, error) {
	playlist, err := s.playlistRepo.FindByID(playlistID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Check ownership
	if playlist.UserID != userID {
		return nil, ErrForbidden
	}

	playlist.Name = req.Name
	playlist.Description = req.Description
	playlist.Thumbnail = req.Thumbnail
	playlist.IsPublic = req.IsPublic

	if err := s.playlistRepo.Update(playlist); err != nil {
		return nil, err
	}

	itemCount, _ := s.playlistItemRepo.CountByPlaylistID(playlistID)
	return s.toPlaylistDTO(playlist, int(itemCount)), nil
}

// DeletePlaylist deletes a playlist
func (s *PlaylistService) DeletePlaylist(playlistID, userID uint) error {
	playlist, err := s.playlistRepo.FindByID(playlistID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Check ownership
	if playlist.UserID != userID {
		return ErrForbidden
	}

	return s.playlistRepo.Delete(playlistID)
}

// AddSongToPlaylist adds a song to a playlist
func (s *PlaylistService) AddSongToPlaylist(playlistID, userID, songID uint) (*dto.PlaylistItemDTO, error) {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(playlistID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrForbidden
	}

	// Check if song exists
	song, err := s.songRepo.FindByID(songID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Check if song is already in playlist
	exists, err := s.playlistItemRepo.IsSongInPlaylist(playlistID, songID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("song already in playlist")
	}

	// Get max position
	maxPos, err := s.playlistItemRepo.GetMaxPosition(playlistID)
	if err != nil {
		return nil, err
	}

	item := &models.PlaylistItem{
		PlaylistID: playlistID,
		SongID:     songID,
		Position:   maxPos + 1,
		AddedAt:    time.Now(),
	}

	if err := s.playlistItemRepo.Create(item); err != nil {
		return nil, err
	}

	return &dto.PlaylistItemDTO{
		ID:         item.ID,
		PlaylistID: item.PlaylistID,
		SongID:     item.SongID,
		Position:   item.Position,
		AddedAt:    item.AddedAt,
		Song:       s.toSongDTO(song),
	}, nil
}

// AddSongsToPlaylist adds multiple songs to a playlist
func (s *PlaylistService) AddSongsToPlaylist(playlistID, userID uint, songIDs []uint) ([]dto.PlaylistItemDTO, error) {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(playlistID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrForbidden
	}

	// Get max position
	maxPos, err := s.playlistItemRepo.GetMaxPosition(playlistID)
	if err != nil {
		return nil, err
	}

	var items []models.PlaylistItem
	var addedItems []dto.PlaylistItemDTO
	position := maxPos + 1

	for _, songID := range songIDs {
		// Check if song exists
		song, err := s.songRepo.FindByID(songID)
		if err != nil {
			continue // Skip invalid songs
		}

		// Check if song is already in playlist
		exists, err := s.playlistItemRepo.IsSongInPlaylist(playlistID, songID)
		if err != nil || exists {
			continue // Skip duplicates
		}

		item := models.PlaylistItem{
			PlaylistID: playlistID,
			SongID:     songID,
			Position:   position,
			AddedAt:    time.Now(),
		}
		items = append(items, item)

		addedItems = append(addedItems, dto.PlaylistItemDTO{
			SongID:   songID,
			Position: position,
			Song:     s.toSongDTO(song),
		})

		position++
	}

	if len(items) > 0 {
		if err := s.playlistItemRepo.CreateBatch(items); err != nil {
			return nil, err
		}
	}

	return addedItems, nil
}

// RemoveSongFromPlaylist removes a song from a playlist
func (s *PlaylistService) RemoveSongFromPlaylist(playlistID, userID, songID uint) error {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(playlistID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}

	return s.playlistItemRepo.DeleteByPlaylistIDAndSongID(playlistID, songID)
}

// RemoveItemFromPlaylist removes an item from a playlist by item ID
func (s *PlaylistService) RemoveItemFromPlaylist(playlistID, userID, itemID uint) error {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(playlistID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}

	// Check if item belongs to the playlist
	item, err := s.playlistItemRepo.FindByID(itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if item.PlaylistID != playlistID {
		return ErrForbidden
	}

	return s.playlistItemRepo.Delete(itemID)
}

// ReorderPlaylistItems reorders items in a playlist
func (s *PlaylistService) ReorderPlaylistItems(playlistID, userID uint, itemIDs []uint) error {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(playlistID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}

	return s.playlistItemRepo.ReorderItems(playlistID, itemIDs)
}

// GetPublicPlaylists gets public playlists
func (s *PlaylistService) GetPublicPlaylists(page, limit int) ([]dto.PlaylistDTO, int64, error) {
	offset := (page - 1) * limit
	playlists, total, err := s.playlistRepo.FindPublicPlaylists(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.PlaylistDTO
	for _, p := range playlists {
		itemCount, _ := s.playlistItemRepo.CountByPlaylistID(p.ID)
		result = append(result, *s.toPlaylistDTO(&p, int(itemCount)))
	}

	return result, total, nil
}

// Helper functions for DTO conversion

func (s *PlaylistService) toPlaylistDTO(playlist *models.Playlist, itemCount int) *dto.PlaylistDTO {
	result := &dto.PlaylistDTO{
		ID:          playlist.ID,
		UserID:      playlist.UserID,
		Name:        playlist.Name,
		Description: playlist.Description,
		Thumbnail:   playlist.Thumbnail,
		IsPublic:    playlist.IsPublic,
		ItemCount:   itemCount,
		TotalSongs:  itemCount,
		CreatedAt:   playlist.CreatedAt,
		UpdatedAt:   playlist.UpdatedAt,
	}

	if playlist.User.ID != 0 {
		result.User = &dto.UserBasicDTO{
			ID:     playlist.User.ID,
			Name:   playlist.User.Name,
			Avatar: playlist.User.Avatar,
		}
	}

	return result
}

func (s *PlaylistService) toPlaylistDTOWithItems(playlist *models.Playlist) *dto.PlaylistDTO {
	result := s.toPlaylistDTO(playlist, len(playlist.Items))

	var items []dto.PlaylistItemDTO
	for _, item := range playlist.Items {
		items = append(items, dto.PlaylistItemDTO{
			ID:         item.ID,
			PlaylistID: item.PlaylistID,
			SongID:     item.SongID,
			Position:   item.Position,
			AddedAt:    item.AddedAt,
			Song:       s.toSongDTO(&item.Song),
		})
	}
	result.Items = items

	return result
}

func (s *PlaylistService) toSongDTO(song *models.Song) *dto.SongDTO {
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
	}
}
