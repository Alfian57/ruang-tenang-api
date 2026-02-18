package service

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlaylistService handles playlist business logic
type PlaylistService struct {
	playlistRepo     *repository.PlaylistRepository
	playlistItemRepo *repository.PlaylistItemRepository
	songRepo         *repository.SongRepository
}

// NewPlaylistService creates a new playlist service
func NewPlaylistService(
	playlistRepo *repository.PlaylistRepository,
	playlistItemRepo *repository.PlaylistItemRepository,
	songRepo *repository.SongRepository,
) *PlaylistService {
	return &PlaylistService{
		playlistRepo:     playlistRepo,
		playlistItemRepo: playlistItemRepo,
		songRepo:         songRepo,
	}
}

// CreatePlaylist creates a new playlist for a user
func (s *PlaylistService) CreatePlaylist(ctx context.Context, userID uint, req *dto.CreatePlaylistRequest) (*dto.PlaylistDTO, error) {
	playlist := &model.Playlist{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		IsPublic:    req.IsPublic,
	}

	if err := s.playlistRepo.Create(ctx, playlist); err != nil {
		return nil, err
	}

	return s.toPlaylistDTO(ctx, playlist, 0), nil
}

// GetUserPlaylists gets all playlists for a user
func (s *PlaylistService) GetUserPlaylists(ctx context.Context, userID uint) ([]dto.PlaylistListDTO, error) {
	playlists, itemCounts, err := s.playlistRepo.FindByUserIDWithItemCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.PlaylistListDTO
	for _, p := range playlists {
		result = append(result, dto.PlaylistListDTO{
			ID:          p.ID,
			UUID:        p.UUID.String(),
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
func (s *PlaylistService) GetPlaylist(ctx context.Context, playlistID, userID uint) (*dto.PlaylistDTO, error) {
	playlist, err := s.playlistRepo.FindByIDWithItems(ctx, playlistID)
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

	return s.toPlaylistDTOWithItems(ctx, playlist), nil
}

// GetPlaylistByUUID gets a playlist by UUID
func (s *PlaylistService) GetPlaylistByUUID(ctx context.Context, playlistUUID string, userID uint) (*dto.PlaylistDTO, error) {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetPlaylist(ctx, meta.ID, userID)
}

// UpdatePlaylist updates a playlist
func (s *PlaylistService) UpdatePlaylist(ctx context.Context, playlistID, userID uint, req *dto.UpdatePlaylistRequest) (*dto.PlaylistDTO, error) {
	playlist, err := s.playlistRepo.FindByID(ctx, playlistID)
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

	if err := s.playlistRepo.Update(ctx, playlist); err != nil {
		return nil, err
	}

	itemCount, _ := s.playlistItemRepo.CountByPlaylistID(ctx, playlistID)
	return s.toPlaylistDTO(ctx, playlist, int(itemCount)), nil
}

// UpdatePlaylistByUUID updates a playlist by UUID
func (s *PlaylistService) UpdatePlaylistByUUID(ctx context.Context, playlistUUID string, userID uint, req *dto.UpdatePlaylistRequest) (*dto.PlaylistDTO, error) {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.UpdatePlaylist(ctx, meta.ID, userID, req)
}

// DeletePlaylist deletes a playlist
func (s *PlaylistService) DeletePlaylist(ctx context.Context, playlistID, userID uint) error {
	playlist, err := s.playlistRepo.FindByID(ctx, playlistID)
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

	return s.playlistRepo.Delete(ctx, playlistID)
}

// DeletePlaylistByUUID deletes a playlist by UUID
func (s *PlaylistService) DeletePlaylistByUUID(ctx context.Context, playlistUUID string, userID uint) error {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	return s.DeletePlaylist(ctx, meta.ID, userID)
}

// AddSongToPlaylist adds a song to a playlist
func (s *PlaylistService) AddSongToPlaylist(ctx context.Context, playlistID, userID, songID uint) (*dto.PlaylistItemDTO, error) {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(ctx, playlistID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrForbidden
	}

	// Check if song exists
	song, err := s.songRepo.FindByID(ctx, songID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Check if song is already in playlist
	exists, err := s.playlistItemRepo.IsSongInPlaylist(ctx, playlistID, songID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("song already in playlist")
	}

	// Get max position
	maxPos, err := s.playlistItemRepo.GetMaxPosition(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	item := &model.PlaylistItem{
		PlaylistID: playlistID,
		SongID:     songID,
		Position:   maxPos + 1,
		AddedAt:    time.Now(),
	}

	if err := s.playlistItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return &dto.PlaylistItemDTO{
		ID:         item.ID,
		UUID:       item.UUID.String(),
		PlaylistID: item.PlaylistID,
		SongID:     item.SongID,
		Position:   item.Position,
		AddedAt:    item.AddedAt,
		Song:       s.toSongDTO(ctx, song),
	}, nil
}

// AddSongToPlaylistByUUID adds a song to a playlist by UUID
func (s *PlaylistService) AddSongToPlaylistByUUID(ctx context.Context, playlistUUID string, userID, songID uint) (*dto.PlaylistItemDTO, error) {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.AddSongToPlaylist(ctx, meta.ID, userID, songID)
}

// AddSongsToPlaylist adds multiple songs to a playlist
func (s *PlaylistService) AddSongsToPlaylist(ctx context.Context, playlistID, userID uint, songIDs []uint) ([]dto.PlaylistItemDTO, error) {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(ctx, playlistID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrForbidden
	}

	// Get max position
	maxPos, err := s.playlistItemRepo.GetMaxPosition(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	var items []model.PlaylistItem
	var addedItems []dto.PlaylistItemDTO
	position := maxPos + 1

	for _, songID := range songIDs {
		// Check if song exists
		song, err := s.songRepo.FindByID(ctx, songID)
		if err != nil {
			continue // Skip invalid songs
		}

		// Check if song is already in playlist
		exists, err := s.playlistItemRepo.IsSongInPlaylist(ctx, playlistID, songID)
		if err != nil || exists {
			continue // Skip duplicates
		}

		item := model.PlaylistItem{
			PlaylistID: playlistID,
			SongID:     songID,
			Position:   position,
			AddedAt:    time.Now(),
		}
		items = append(items, item)

		addedItems = append(addedItems, dto.PlaylistItemDTO{
			SongID:   songID,
			Position: position,
			Song:     s.toSongDTO(ctx, song),
		})

		position++
	}

	if len(items) > 0 {
		if err := s.playlistItemRepo.CreateBatch(ctx, items); err != nil {
			return nil, err
		}
	}

	return addedItems, nil
}

// AddSongsToPlaylistByUUID adds multiple songs to a playlist by UUID
func (s *PlaylistService) AddSongsToPlaylistByUUID(ctx context.Context, playlistUUID string, userID uint, songIDs []uint) ([]dto.PlaylistItemDTO, error) {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.AddSongsToPlaylist(ctx, meta.ID, userID, songIDs)
}

// RemoveSongFromPlaylist removes a song from a playlist
func (s *PlaylistService) RemoveSongFromPlaylist(ctx context.Context, playlistID, userID, songID uint) error {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}

	return s.playlistItemRepo.DeleteByPlaylistIDAndSongID(ctx, playlistID, songID)
}

// RemoveSongFromPlaylistByUUID removes a song from a playlist by UUID
func (s *PlaylistService) RemoveSongFromPlaylistByUUID(ctx context.Context, playlistUUID string, userID, songID uint) error {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	return s.RemoveSongFromPlaylist(ctx, meta.ID, userID, songID)
}

// RemoveItemFromPlaylist removes an item from a playlist by item ID
func (s *PlaylistService) RemoveItemFromPlaylist(ctx context.Context, playlistID, userID, itemID uint) error {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}

	// Check if item belongs to the playlist
	item, err := s.playlistItemRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if item.PlaylistID != playlistID {
		return ErrForbidden
	}

	return s.playlistItemRepo.Delete(ctx, itemID)
}

// RemoveItemFromPlaylistByUUID removes an item from a playlist by UUIDs (playlistUUID and itemID)
func (s *PlaylistService) RemoveItemFromPlaylistByUUID(ctx context.Context, playlistUUID string, userID, itemID uint) error {
	// Parse UUID
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	// Resolve UUID
	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	return s.RemoveItemFromPlaylist(ctx, meta.ID, userID, itemID)
}

// ReorderPlaylistItems reorders items in a playlist
func (s *PlaylistService) ReorderPlaylistItems(ctx context.Context, playlistID, userID uint, itemIDs []uint) error {
	// Check playlist ownership
	isOwner, err := s.playlistRepo.IsOwner(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}

	return s.playlistItemRepo.ReorderItems(ctx, playlistID, itemIDs)
}

// ReorderPlaylistItemsByUUID reorders items in a playlist by UUID
func (s *PlaylistService) ReorderPlaylistItemsByUUID(ctx context.Context, playlistUUID string, userID uint, itemIDs []uint) error {
	id, err := uuid.Parse(playlistUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	meta, err := s.playlistRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	return s.ReorderPlaylistItems(ctx, meta.ID, userID, itemIDs)
}

// GetPublicPlaylists gets public playlists
func (s *PlaylistService) GetPublicPlaylists(ctx context.Context, page, limit int) ([]dto.PlaylistDTO, int64, error) {
	offset := (page - 1) * limit
	playlists, total, err := s.playlistRepo.FindPublicPlaylists(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.PlaylistDTO
	for _, p := range playlists {
		itemCount, _ := s.playlistItemRepo.CountByPlaylistID(ctx, p.ID)
		result = append(result, *s.toPlaylistDTO(ctx, &p, int(itemCount)))
	}

	return result, total, nil
}

// Helper functions for DTO conversion

func (s *PlaylistService) toPlaylistDTO(ctx context.Context, playlist *model.Playlist, itemCount int) *dto.PlaylistDTO {
	result := &dto.PlaylistDTO{
		ID:          playlist.ID,
		UUID:        playlist.UUID.String(),
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

func (s *PlaylistService) toPlaylistDTOWithItems(ctx context.Context, playlist *model.Playlist) *dto.PlaylistDTO {
	result := s.toPlaylistDTO(ctx, playlist, len(playlist.Items))

	var items []dto.PlaylistItemDTO
	for _, item := range playlist.Items {
		items = append(items, dto.PlaylistItemDTO{
			ID:         item.ID,
			UUID:       item.UUID.String(),
			PlaylistID: item.PlaylistID,
			SongID:     item.SongID,
			Position:   item.Position,
			AddedAt:    item.AddedAt,
			Song:       s.toSongDTO(ctx, &item.Song),
		})
	}
	result.Items = items

	return result
}

func (s *PlaylistService) toSongDTO(ctx context.Context, song *model.Song) *dto.SongDTO {
	return &dto.SongDTO{
		ID:         song.ID,
		Slug:       song.Slug,
		Title:      song.Title,
		FilePath:   song.FilePath,
		Thumbnail:  song.Thumbnail,
		CategoryID: song.SongCategoryID,
		Category: dto.SongCategoryDTO{
			ID:        song.Category.ID,
			Slug:      song.Category.Slug,
			Name:      song.Category.Name,
			Thumbnail: song.Category.Thumbnail,
			CreatedAt: song.Category.CreatedAt,
		},
		CreatedAt: song.CreatedAt,
	}
}
