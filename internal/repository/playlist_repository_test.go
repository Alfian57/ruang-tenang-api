package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPlaylistRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL,
			deleted_at DATETIME
		)`,
		`CREATE TABLE song_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			thumbnail TEXT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL,
			file_path TEXT NOT NULL,
			thumbnail TEXT,
			song_category_id INTEGER NOT NULL,
			deleted_at DATETIME
		)`,
		`CREATE TABLE playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			thumbnail TEXT,
			is_public BOOLEAN DEFAULT FALSE,
			is_admin_playlist BOOLEAN DEFAULT FALSE,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE playlist_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL,
			playlist_id INTEGER NOT NULL,
			song_id INTEGER NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			added_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}

	return db
}

func TestPlaylistRepository_ErrorPathsOnEmptyDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	playlistRepo := NewPlaylistRepository(db)
	itemRepo := NewPlaylistItemRepository(db)
	ctx := context.Background()

	if err := playlistRepo.Create(ctx, &model.Playlist{}); err == nil {
		t.Fatal("expected playlist Create error")
	}
	if _, err := playlistRepo.FindByID(ctx, 1); err == nil {
		t.Fatal("expected playlist FindByID error")
	}
	if _, err := playlistRepo.FindByUUID(ctx, uuid.New()); err == nil {
		t.Fatal("expected playlist FindByUUID error")
	}
	if _, err := playlistRepo.FindByIDWithItems(ctx, 1); err == nil {
		t.Fatal("expected playlist FindByIDWithItems error")
	}
	if _, err := playlistRepo.FindByUserID(ctx, 1); err == nil {
		t.Fatal("expected playlist FindByUserID error")
	}
	if _, _, err := playlistRepo.FindByUserIDWithItemCount(ctx, 1); err == nil {
		t.Fatal("expected playlist FindByUserIDWithItemCount error")
	}
	if _, _, err := playlistRepo.FindPublicPlaylists(ctx, 10, 0); err == nil {
		t.Fatal("expected playlist FindPublicPlaylists error")
	}
	if err := playlistRepo.Update(ctx, &model.Playlist{}); err == nil {
		t.Fatal("expected playlist Update error")
	}
	if err := playlistRepo.Delete(ctx, 1); err == nil {
		t.Fatal("expected playlist Delete error")
	}
	if _, err := playlistRepo.IsOwner(ctx, 1, 1); err == nil {
		t.Fatal("expected playlist IsOwner error")
	}
	if _, err := playlistRepo.CountByUserID(ctx, 1); err == nil {
		t.Fatal("expected playlist CountByUserID error")
	}

	if err := itemRepo.Create(ctx, &model.PlaylistItem{}); err == nil {
		t.Fatal("expected item Create error")
	}
	if err := itemRepo.CreateBatch(ctx, []model.PlaylistItem{{}}); err == nil {
		t.Fatal("expected item CreateBatch error")
	}
	if _, err := itemRepo.FindByID(ctx, 1); err == nil {
		t.Fatal("expected item FindByID error")
	}
	if _, err := itemRepo.FindByPlaylistID(ctx, 1); err == nil {
		t.Fatal("expected item FindByPlaylistID error")
	}
	if _, err := itemRepo.FindByPlaylistIDAndSongID(ctx, 1, 1); err == nil {
		t.Fatal("expected item FindByPlaylistIDAndSongID error")
	}
	if err := itemRepo.Delete(ctx, 1); err == nil {
		t.Fatal("expected item Delete error")
	}
	if err := itemRepo.DeleteByPlaylistIDAndSongID(ctx, 1, 1); err == nil {
		t.Fatal("expected item DeleteByPlaylistIDAndSongID error")
	}
	if _, err := itemRepo.GetMaxPosition(ctx, 1); err == nil {
		t.Fatal("expected item GetMaxPosition error")
	}
	if err := itemRepo.UpdatePositions(ctx, 1, map[uint]int{1: 0}); err == nil {
		t.Fatal("expected item UpdatePositions error")
	}
	if err := itemRepo.ReorderItems(ctx, 1, []uint{1}); err == nil {
		t.Fatal("expected item ReorderItems error")
	}
	if _, err := itemRepo.CountByPlaylistID(ctx, 1); err == nil {
		t.Fatal("expected item CountByPlaylistID error")
	}
	if _, err := itemRepo.IsSongInPlaylist(ctx, 1, 1); err == nil {
		t.Fatal("expected item IsSongInPlaylist error")
	}
}

func TestPlaylistRepository_SuccessPaths(t *testing.T) {
	db := setupPlaylistRepoDB(t)
	ctx := context.Background()

	playlistRepo := NewPlaylistRepository(db)
	itemRepo := NewPlaylistItemRepository(db)

	const user1ID uint = 1
	const user2ID uint = 2
	if err := db.Exec(`INSERT INTO users (id, name, username, email, password) VALUES (1, 'User One', 'user-one', 'user1@example.com', 'secret')`).Error; err != nil {
		t.Fatalf("insert user1: %v", err)
	}
	if err := db.Exec(`INSERT INTO users (id, name, username, email, password) VALUES (2, 'User Two', 'user-two', 'user2@example.com', 'secret')`).Error; err != nil {
		t.Fatalf("insert user2: %v", err)
	}

	const categoryID uint = 1
	if err := db.Exec(`INSERT INTO song_categories (id, name, slug, thumbnail) VALUES (1, 'Relax', 'relax', '')`).Error; err != nil {
		t.Fatalf("insert category: %v", err)
	}

	const song1ID uint = 1
	const song2ID uint = 2
	if err := db.Exec(`INSERT INTO songs (id, title, slug, file_path, song_category_id) VALUES (1, 'Ocean', 'ocean', 'ocean.mp3', ?)`, categoryID).Error; err != nil {
		t.Fatalf("insert song1: %v", err)
	}
	if err := db.Exec(`INSERT INTO songs (id, title, slug, file_path, song_category_id) VALUES (2, 'Rain', 'rain', 'rain.mp3', ?)`, categoryID).Error; err != nil {
		t.Fatalf("insert song2: %v", err)
	}

	now := time.Now()
	playlist1 := model.Playlist{ID: 1, UUID: uuid.New(), UserID: user1ID, Name: "Morning", IsPublic: true, CreatedAt: now.Add(-time.Hour)}
	playlist2 := model.Playlist{ID: 2, UUID: uuid.New(), UserID: user1ID, Name: "Night", IsPublic: false, CreatedAt: now}
	if err := db.Exec(`INSERT INTO playlists (id, uuid, user_id, name, is_public, is_admin_playlist, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		playlist1.ID, playlist1.UUID.String(), playlist1.UserID, playlist1.Name, playlist1.IsPublic, false, playlist1.CreatedAt, now).Error; err != nil {
		t.Fatalf("insert playlist1: %v", err)
	}
	if err := db.Exec(`INSERT INTO playlists (id, uuid, user_id, name, is_public, is_admin_playlist, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		playlist2.ID, playlist2.UUID.String(), playlist2.UserID, playlist2.Name, playlist2.IsPublic, false, playlist2.CreatedAt, now).Error; err != nil {
		t.Fatalf("insert playlist2: %v", err)
	}

	item1 := model.PlaylistItem{ID: 1, UUID: uuid.New(), PlaylistID: playlist1.ID, SongID: song1ID, Position: 0}
	item2 := model.PlaylistItem{ID: 2, UUID: uuid.New(), PlaylistID: playlist1.ID, SongID: song2ID, Position: 1}
	if err := db.Exec(`INSERT INTO playlist_items (id, uuid, playlist_id, song_id, position, added_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item1.ID, item1.UUID.String(), item1.PlaylistID, item1.SongID, item1.Position, now, now, now).Error; err != nil {
		t.Fatalf("insert item1: %v", err)
	}
	if err := db.Exec(`INSERT INTO playlist_items (id, uuid, playlist_id, song_id, position, added_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item2.ID, item2.UUID.String(), item2.PlaylistID, item2.SongID, item2.Position, now, now, now).Error; err != nil {
		t.Fatalf("insert item2: %v", err)
	}

	loaded, err := playlistRepo.FindByIDWithItems(ctx, playlist1.ID)
	if err != nil {
		t.Fatalf("find by id with items: %v", err)
	}
	if len(loaded.Items) != 2 || loaded.Items[0].Position != 0 || loaded.Items[1].Position != 1 {
		t.Fatalf("expected ordered items, got %#v", loaded.Items)
	}

	loadedByID, err := playlistRepo.FindByID(ctx, playlist1.ID)
	if err != nil || loadedByID == nil || loadedByID.ID != playlist1.ID {
		t.Fatalf("find by id success failed: err=%v playlist=%+v", err, loadedByID)
	}
	loadedByUUID, err := playlistRepo.FindByUUID(ctx, playlist1.UUID)
	if err != nil || loadedByUUID == nil || loadedByUUID.ID != playlist1.ID {
		t.Fatalf("find by uuid success failed: err=%v playlist=%+v", err, loadedByUUID)
	}

	if _, err := playlistRepo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected playlist FindByID missing error")
	}
	if _, err := playlistRepo.FindByUUID(ctx, uuid.New()); err == nil {
		t.Fatal("expected playlist FindByUUID missing error")
	}
	if _, err := playlistRepo.FindByIDWithItems(ctx, 999999); err == nil {
		t.Fatal("expected playlist FindByIDWithItems missing error")
	}

	playlists, itemCounts, err := playlistRepo.FindByUserIDWithItemCount(ctx, user1ID)
	if err != nil {
		t.Fatalf("find by user with item count: %v", err)
	}
	if len(playlists) != 2 {
		t.Fatalf("expected 2 playlists, got %d", len(playlists))
	}
	if itemCounts[playlist1.ID] != 2 || itemCounts[playlist2.ID] != 0 {
		t.Fatalf("unexpected item counts: %#v", itemCounts)
	}

	publicPlaylists, totalPublic, err := playlistRepo.FindPublicPlaylists(ctx, 10, 0)
	if err != nil {
		t.Fatalf("find public playlists: %v", err)
	}
	if totalPublic != 1 || len(publicPlaylists) != 1 || publicPlaylists[0].ID != playlist1.ID {
		t.Fatalf("unexpected public result total=%d len=%d", totalPublic, len(publicPlaylists))
	}

	owner, err := playlistRepo.IsOwner(ctx, playlist1.ID, user1ID)
	if err != nil || !owner {
		t.Fatalf("expected owner true, err=%v owner=%v", err, owner)
	}
	notOwner, err := playlistRepo.IsOwner(ctx, playlist1.ID, user2ID)
	if err != nil || notOwner {
		t.Fatalf("expected owner false, err=%v owner=%v", err, notOwner)
	}

	playlistCount, err := playlistRepo.CountByUserID(ctx, user1ID)
	if err != nil || playlistCount != 2 {
		t.Fatalf("count by user failed: err=%v count=%d", err, playlistCount)
	}

	maxPosNoItems, err := itemRepo.GetMaxPosition(ctx, playlist2.ID)
	if err != nil || maxPosNoItems != -1 {
		t.Fatalf("expected max position -1 for empty playlist, err=%v max=%d", err, maxPosNoItems)
	}
	maxPos, err := itemRepo.GetMaxPosition(ctx, playlist1.ID)
	if err != nil || maxPos != 1 {
		t.Fatalf("expected max position 1, err=%v max=%d", err, maxPos)
	}

	countItems, err := itemRepo.CountByPlaylistID(ctx, playlist1.ID)
	if err != nil || countItems != 2 {
		t.Fatalf("count by playlist failed: err=%v count=%d", err, countItems)
	}
	inPlaylist, err := itemRepo.IsSongInPlaylist(ctx, playlist1.ID, song1ID)
	if err != nil || !inPlaylist {
		t.Fatalf("song should be in playlist, err=%v in=%v", err, inPlaylist)
	}
	notInPlaylist, err := itemRepo.IsSongInPlaylist(ctx, playlist1.ID, 99999)
	if err != nil || notInPlaylist {
		t.Fatalf("song should not be in playlist, err=%v in=%v", err, notInPlaylist)
	}

	if err := itemRepo.UpdatePositions(ctx, playlist1.ID, map[uint]int{item1.ID: 2, item2.ID: 3}); err != nil {
		t.Fatalf("update positions: %v", err)
	}
	if err := itemRepo.ReorderItems(ctx, playlist1.ID, []uint{item2.ID, item1.ID}); err != nil {
		t.Fatalf("reorder items: %v", err)
	}

	updatedItems, err := itemRepo.FindByPlaylistID(ctx, playlist1.ID)
	if err != nil {
		t.Fatalf("find by playlist after reorder: %v", err)
	}
	if len(updatedItems) != 2 || updatedItems[0].ID != item2.ID || updatedItems[0].Position != 0 {
		t.Fatalf("unexpected reordered items: %#v", updatedItems)
	}

	itemByID, err := itemRepo.FindByID(ctx, item1.ID)
	if err != nil || itemByID == nil || itemByID.ID != item1.ID {
		t.Fatalf("find item by id success failed: err=%v item=%+v", err, itemByID)
	}
	itemByPlaylistSong, err := itemRepo.FindByPlaylistIDAndSongID(ctx, playlist1.ID, song1ID)
	if err != nil || itemByPlaylistSong == nil || itemByPlaylistSong.ID != item1.ID {
		t.Fatalf("find item by playlist+song success failed: err=%v item=%+v", err, itemByPlaylistSong)
	}

	if _, err := itemRepo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected playlist item FindByID missing error")
	}
	if _, err := itemRepo.FindByPlaylistIDAndSongID(ctx, playlist1.ID, 999999); err == nil {
		t.Fatal("expected FindByPlaylistIDAndSongID missing error")
	}

	if err := itemRepo.DeleteByPlaylistIDAndSongID(ctx, playlist1.ID, song2ID); err != nil {
		t.Fatalf("delete by playlist and song: %v", err)
	}
	remaining, err := itemRepo.CountByPlaylistID(ctx, playlist1.ID)
	if err != nil || remaining != 1 {
		t.Fatalf("unexpected remaining count err=%v count=%d", err, remaining)
	}
}
