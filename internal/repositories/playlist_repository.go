package repositories

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
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
func (r *PlaylistRepository) Create(playlist *models.Playlist) error {
	return r.db.Create(playlist).Error
}

// FindByID finds a playlist by ID
func (r *PlaylistRepository) FindByID(id uint) (*models.Playlist, error) {
	var playlist models.Playlist
	err := r.db.Preload("User").First(&playlist, id).Error
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

// FindByIDWithItems finds a playlist by ID with all items preloaded
func (r *PlaylistRepository) FindByIDWithItems(id uint) (*models.Playlist, error) {
	var playlist models.Playlist
	err := r.db.Preload("User").
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
func (r *PlaylistRepository) FindByUserID(userID uint) ([]models.Playlist, error) {
	var playlists []models.Playlist
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&playlists).Error
	return playlists, err
}

// FindByUserIDWithItemCount finds all playlists for a user with item count
func (r *PlaylistRepository) FindByUserIDWithItemCount(userID uint) ([]models.Playlist, map[uint]int, error) {
	var playlists []models.Playlist
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&playlists).Error
	if err != nil {
		return nil, nil, err
	}

	// Get item counts for each playlist
	itemCounts := make(map[uint]int)
	for _, p := range playlists {
		var count int64
		r.db.Model(&models.PlaylistItem{}).Where("playlist_id = ?", p.ID).Count(&count)
		itemCounts[p.ID] = int(count)
	}

	return playlists, itemCounts, nil
}

// FindPublicPlaylists finds all public playlists
func (r *PlaylistRepository) FindPublicPlaylists(limit, offset int) ([]models.Playlist, int64, error) {
	var playlists []models.Playlist
	var total int64

	r.db.Model(&models.Playlist{}).Where("is_public = ?", true).Count(&total)

	err := r.db.Where("is_public = ?", true).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&playlists).Error

	return playlists, total, err
}

// Update updates a playlist
func (r *PlaylistRepository) Update(playlist *models.Playlist) error {
	return r.db.Save(playlist).Error
}

// Delete soft deletes a playlist
func (r *PlaylistRepository) Delete(id uint) error {
	return r.db.Delete(&models.Playlist{}, id).Error
}

// IsOwner checks if a user owns a playlist
func (r *PlaylistRepository) IsOwner(playlistID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Playlist{}).
		Where("id = ? AND user_id = ?", playlistID, userID).
		Count(&count).Error
	return count > 0, err
}

// CountByUserID counts the number of playlists for a user
func (r *PlaylistRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Playlist{}).Where("user_id = ?", userID).Count(&count).Error
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
func (r *PlaylistItemRepository) Create(item *models.PlaylistItem) error {
	return r.db.Create(item).Error
}

// CreateBatch creates multiple playlist items
func (r *PlaylistItemRepository) CreateBatch(items []models.PlaylistItem) error {
	return r.db.Create(&items).Error
}

// FindByID finds a playlist item by ID
func (r *PlaylistItemRepository) FindByID(id uint) (*models.PlaylistItem, error) {
	var item models.PlaylistItem
	err := r.db.Preload("Song").Preload("Song.Category").First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindByPlaylistID finds all items in a playlist
func (r *PlaylistItemRepository) FindByPlaylistID(playlistID uint) ([]models.PlaylistItem, error) {
	var items []models.PlaylistItem
	err := r.db.Where("playlist_id = ?", playlistID).
		Preload("Song").
		Preload("Song.Category").
		Order("position ASC").
		Find(&items).Error
	return items, err
}

// FindByPlaylistIDAndSongID finds an item by playlist and song ID
func (r *PlaylistItemRepository) FindByPlaylistIDAndSongID(playlistID, songID uint) (*models.PlaylistItem, error) {
	var item models.PlaylistItem
	err := r.db.Where("playlist_id = ? AND song_id = ?", playlistID, songID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Delete soft deletes a playlist item
func (r *PlaylistItemRepository) Delete(id uint) error {
	return r.db.Delete(&models.PlaylistItem{}, id).Error
}

// DeleteByPlaylistIDAndSongID deletes an item by playlist and song ID
func (r *PlaylistItemRepository) DeleteByPlaylistIDAndSongID(playlistID, songID uint) error {
	return r.db.Where("playlist_id = ? AND song_id = ?", playlistID, songID).
		Delete(&models.PlaylistItem{}).Error
}

// GetMaxPosition gets the maximum position in a playlist
func (r *PlaylistItemRepository) GetMaxPosition(playlistID uint) (int, error) {
	var maxPos *int
	err := r.db.Model(&models.PlaylistItem{}).
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
func (r *PlaylistItemRepository) UpdatePositions(playlistID uint, itemPositions map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for itemID, position := range itemPositions {
			if err := tx.Model(&models.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", itemID, playlistID).
				Update("position", position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ReorderItems reorders items based on the provided order
func (r *PlaylistItemRepository) ReorderItems(playlistID uint, itemIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, itemID := range itemIDs {
			if err := tx.Model(&models.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", itemID, playlistID).
				Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CountByPlaylistID counts items in a playlist
func (r *PlaylistItemRepository) CountByPlaylistID(playlistID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.PlaylistItem{}).Where("playlist_id = ?", playlistID).Count(&count).Error
	return count, err
}

// IsSongInPlaylist checks if a song is already in a playlist
func (r *PlaylistItemRepository) IsSongInPlaylist(playlistID, songID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PlaylistItem{}).
		Where("playlist_id = ? AND song_id = ?", playlistID, songID).
		Count(&count).Error
	return count > 0, err
}
