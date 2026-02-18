package repository

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlaylistRepository handles database operations for playlists
type PlaylistRepository struct {
	db *gorm.DB
}

// NewPlaylistRepository creates a new playlist repository
func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

// Create creates a new playlist
func (r *PlaylistRepository) Create(ctx context.Context, playlist *model.Playlist) error {
	return r.db.WithContext(ctx).Create(playlist).Error
}

// FindByID finds a playlist by ID
func (r *PlaylistRepository) FindByID(ctx context.Context, id uint) (*model.Playlist, error) {
	var playlist model.Playlist
	err := r.db.WithContext(ctx).Preload("User").First(&playlist, id).Error
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

// FindByUUID finds a playlist by UUID
func (r *PlaylistRepository) FindByUUID(ctx context.Context, playlistUUID uuid.UUID) (*model.Playlist, error) {
	var playlist model.Playlist
	err := r.db.WithContext(ctx).Preload("User").Where("uuid = ?", playlistUUID).First(&playlist).Error
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

// FindByIDWithItems finds a playlist by ID with all items preloaded
func (r *PlaylistRepository) FindByIDWithItems(ctx context.Context, id uint) (*model.Playlist, error) {
	var playlist model.Playlist
	err := r.db.WithContext(ctx).Preload("User").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Items.Song").
		Preload("Items.Song.Category").
		First(&playlist, id).Error
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

// FindByUserID finds all playlists for a user
func (r *PlaylistRepository) FindByUserID(ctx context.Context, userID uint) ([]model.Playlist, error) {
	var playlists []model.Playlist
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&playlists).Error
	return playlists, err
}

// FindByUserIDWithItemCount finds all playlists for a user with item count
func (r *PlaylistRepository) FindByUserIDWithItemCount(ctx context.Context, userID uint) ([]model.Playlist, map[uint]int, error) {
	var playlists []model.Playlist
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&playlists).Error
	if err != nil {
		return nil, nil, err
	}

	// Get item counts for each playlist
	itemCounts := make(map[uint]int)
	for _, p := range playlists {
		var count int64
		r.db.WithContext(ctx).Model(&model.PlaylistItem{}).Where("playlist_id = ?", p.ID).Count(&count)
		itemCounts[p.ID] = int(count)
	}

	return playlists, itemCounts, nil
}

// FindPublicPlaylists finds all public playlists
func (r *PlaylistRepository) FindPublicPlaylists(ctx context.Context, limit, offset int) ([]model.Playlist, int64, error) {
	var playlists []model.Playlist
	var total int64

	r.db.WithContext(ctx).Model(&model.Playlist{}).Where("is_public = ?", true).Count(&total)

	err := r.db.WithContext(ctx).Where("is_public = ?", true).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&playlists).Error

	return playlists, total, err
}

// Update updates a playlist
func (r *PlaylistRepository) Update(ctx context.Context, playlist *model.Playlist) error {
	return r.db.WithContext(ctx).Save(playlist).Error
}

// Delete soft deletes a playlist
func (r *PlaylistRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Playlist{}, id).Error
}

// IsOwner checks if a user owns a playlist
func (r *PlaylistRepository) IsOwner(ctx context.Context, playlistID, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Playlist{}).
		Where("id = ? AND user_id = ?", playlistID, userID).
		Count(&count).Error
	return count > 0, err
}

// CountByUserID counts the number of playlists for a user
func (r *PlaylistRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Playlist{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// PlaylistItemRepository handles database operations for playlist items
type PlaylistItemRepository struct {
	db *gorm.DB
}

// NewPlaylistItemRepository creates a new playlist item repository
func NewPlaylistItemRepository(db *gorm.DB) *PlaylistItemRepository {
	return &PlaylistItemRepository{db: db}
}

// Create creates a new playlist item
func (r *PlaylistItemRepository) Create(ctx context.Context, item *model.PlaylistItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// CreateBatch creates multiple playlist items
func (r *PlaylistItemRepository) CreateBatch(ctx context.Context, items []model.PlaylistItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

// FindByID finds a playlist item by ID
func (r *PlaylistItemRepository) FindByID(ctx context.Context, id uint) (*model.PlaylistItem, error) {
	var item model.PlaylistItem
	err := r.db.WithContext(ctx).Preload("Song").Preload("Song.Category").First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindByPlaylistID finds all items in a playlist
func (r *PlaylistItemRepository) FindByPlaylistID(ctx context.Context, playlistID uint) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem
	err := r.db.WithContext(ctx).Where("playlist_id = ?", playlistID).
		Preload("Song").
		Preload("Song.Category").
		Order("position ASC").
		Find(&items).Error
	return items, err
}

// FindByPlaylistIDAndSongID finds an item by playlist and song ID
func (r *PlaylistItemRepository) FindByPlaylistIDAndSongID(ctx context.Context, playlistID, songID uint) (*model.PlaylistItem, error) {
	var item model.PlaylistItem
	err := r.db.WithContext(ctx).Where("playlist_id = ? AND song_id = ?", playlistID, songID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Delete soft deletes a playlist item
func (r *PlaylistItemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.PlaylistItem{}, id).Error
}

// DeleteByPlaylistIDAndSongID deletes an item by playlist and song ID
func (r *PlaylistItemRepository) DeleteByPlaylistIDAndSongID(ctx context.Context, playlistID, songID uint) error {
	return r.db.WithContext(ctx).Where("playlist_id = ? AND song_id = ?", playlistID, songID).
		Delete(&model.PlaylistItem{}).Error
}

// GetMaxPosition gets the maximum position in a playlist
func (r *PlaylistItemRepository) GetMaxPosition(ctx context.Context, playlistID uint) (int, error) {
	var maxPos *int
	err := r.db.WithContext(ctx).Model(&model.PlaylistItem{}).
		Where("playlist_id = ?", playlistID).
		Select("MAX(position)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	if maxPos == nil {
		return -1, nil
	}
	return *maxPos, nil
}

// UpdatePositions updates positions for multiple items
func (r *PlaylistItemRepository) UpdatePositions(ctx context.Context, playlistID uint, itemPositions map[uint]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for itemID, position := range itemPositions {
			if err := tx.Model(&model.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", itemID, playlistID).
				Update("position", position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ReorderItems reorders items based on the provided order
func (r *PlaylistItemRepository) ReorderItems(ctx context.Context, playlistID uint, itemIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, itemID := range itemIDs {
			if err := tx.Model(&model.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", itemID, playlistID).
				Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CountByPlaylistID counts items in a playlist
func (r *PlaylistItemRepository) CountByPlaylistID(ctx context.Context, playlistID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PlaylistItem{}).Where("playlist_id = ?", playlistID).Count(&count).Error
	return count, err
}

// IsSongInPlaylist checks if a song is already in a playlist
func (r *PlaylistItemRepository) IsSongInPlaylist(ctx context.Context, playlistID, songID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PlaylistItem{}).
		Where("playlist_id = ? AND song_id = ?", playlistID, songID).
		Count(&count).Error
	return count > 0, err
}
