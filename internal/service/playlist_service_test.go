package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPlaylistServiceWithDB(t *testing.T, withSchema bool) (*PlaylistService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		queries := []string{
			`CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				name TEXT,
				username TEXT,
				email TEXT,
				password TEXT,
				role TEXT,
				exp INTEGER,
				avatar TEXT,
				is_blocked BOOLEAN,
				is_forum_blocked BOOLEAN,
				reset_token TEXT,
				reset_token_expiry DATETIME,
				suspension_end DATETIME,
				suspension_reason TEXT,
				is_banned BOOLEAN,
				ban_reason TEXT,
				has_accepted_ai_disclaimer BOOLEAN,
				content_warning_preference TEXT,
				profile_theme TEXT,
				profile_banner TEXT,
				avatar_border_color TEXT,
				tagline TEXT,
				bio TEXT,
				current_streak INTEGER,
				longest_streak INTEGER,
				last_activity_date DATETIME,
				total_activities INTEGER,
				last_login_date DATETIME,
				login_streak INTEGER,
				streak_freeze_available BOOLEAN,
				streak_freeze_used_at DATETIME,
				created_at DATETIME,
				updated_at DATETIME,
				deleted_at DATETIME
			)`,
			`CREATE TABLE playlists (id INTEGER PRIMARY KEY, uuid TEXT, user_id INTEGER, name TEXT, description TEXT, thumbnail TEXT, is_public BOOLEAN, is_admin_playlist BOOLEAN, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
			`CREATE TABLE playlist_items (id INTEGER PRIMARY KEY, uuid TEXT, playlist_id INTEGER, song_id INTEGER, position INTEGER, added_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
			`CREATE TABLE song_categories (id INTEGER PRIMARY KEY, name TEXT, slug TEXT, thumbnail TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
			`CREATE TABLE songs (id INTEGER PRIMARY KEY, title TEXT, slug TEXT, file_path TEXT, thumbnail TEXT, song_category_id INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		}
		for _, q := range queries {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("create table failed: %v", err)
			}
		}

		if err := db.Exec(`INSERT INTO users (id, name, avatar) VALUES (1, 'User One', ''), (2, 'User Two', '')`).Error; err != nil {
			t.Fatalf("seed users: %v", err)
		}
		if err := db.Exec(`INSERT INTO song_categories (id, name, slug) VALUES (1, 'Focus', 'focus')`).Error; err != nil {
			t.Fatalf("seed song category: %v", err)
		}
		if err := db.Exec(`INSERT INTO songs (id, title, slug, file_path, song_category_id) VALUES (1, 'Song One', 'song-one', '/a.mp3', 1), (2, 'Song Two', 'song-two', '/b.mp3', 1), (3, 'Song Three', 'song-three', '/c.mp3', 1)`).Error; err != nil {
			t.Fatalf("seed songs: %v", err)
		}
		if err := db.Exec(`INSERT INTO playlists (id, uuid, user_id, name, is_public) VALUES (1, '11111111-1111-1111-1111-111111111111', 1, 'P1', 0), (2, '22222222-2222-2222-2222-222222222222', 2, 'P2', 0), (3, '33333333-3333-3333-3333-333333333333', 2, 'P3', 1)`).Error; err != nil {
			t.Fatalf("seed playlists: %v", err)
		}
		if err := db.Exec(`INSERT INTO playlist_items (id, uuid, playlist_id, song_id, position) VALUES (1, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1, 1, 0)`).Error; err != nil {
			t.Fatalf("seed playlist item: %v", err)
		}
	}

	svc := NewPlaylistService(
		repository.NewPlaylistRepository(db),
		repository.NewPlaylistItemRepository(db),
		repository.NewSongRepository(db),
	)

	return svc, db
}

func setupPlaylistService(t *testing.T, withSchema bool) *PlaylistService {
	t.Helper()
	svc, _ := setupPlaylistServiceWithDB(t, withSchema)
	return svc
}

func TestPlaylistService_Basics(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, true)

	created, err := svc.CreatePlaylist(ctx, 1, &dto.CreatePlaylistRequest{Name: "New P", Description: "d"})
	if err != nil {
		t.Fatalf("create playlist failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created playlist id")
	}

	mine, err := svc.GetUserPlaylists(ctx, 1)
	if err != nil || len(mine) == 0 {
		t.Fatalf("get user playlists failed: %v", err)
	}

	pub, total, err := svc.GetPublicPlaylists(ctx, 1, 10)
	if err != nil {
		t.Fatalf("get public playlists failed: %v", err)
	}
	if total == 0 || len(pub) == 0 {
		t.Fatalf("expected public playlists, total=%d len=%d", total, len(pub))
	}
}

func TestPlaylistService_GetUpdateDelete_Branches(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, true)

	if _, err := svc.GetPlaylist(ctx, 999, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := svc.GetPlaylist(ctx, 2, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if _, err := svc.GetPlaylist(ctx, 1, 1); err != nil {
		t.Fatalf("expected get playlist success: %v", err)
	}

	if _, err := svc.GetPlaylistByUUID(ctx, "bad-uuid", 1); err == nil {
		t.Fatal("expected invalid uuid error")
	}
	if _, err := svc.GetPlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("expected get playlist by uuid success: %v", err)
	}

	if _, err := svc.UpdatePlaylist(ctx, 2, 1, &dto.UpdatePlaylistRequest{Name: "x"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden on update, got %v", err)
	}
	if _, err := svc.UpdatePlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, &dto.UpdatePlaylistRequest{Name: "renamed", Description: "ok"}); err != nil {
		t.Fatalf("expected update success: %v", err)
	}

	if err := svc.DeletePlaylist(ctx, 2, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden on delete, got %v", err)
	}
	if err := svc.DeletePlaylist(ctx, 999, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on delete missing playlist, got %v", err)
	}
	if err := svc.DeletePlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("expected delete success: %v", err)
	}
}

func TestPlaylistService_ItemOps_Branches(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, true)

	if _, err := svc.AddSongToPlaylist(ctx, 2, 1, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected add song forbidden, got %v", err)
	}
	if _, err := svc.AddSongToPlaylist(ctx, 1, 1, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected add song not found, got %v", err)
	}
	if _, err := svc.AddSongToPlaylist(ctx, 1, 1, 1); err == nil || err.Error() != "song already in playlist" {
		t.Fatalf("expected duplicate song error, got %v", err)
	}
	addedItem, err := svc.AddSongToPlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, 2)
	if err != nil {
		t.Fatalf("expected add song by uuid success: %v", err)
	}

	if _, err := svc.AddSongsToPlaylist(ctx, 2, 1, []uint{1, 2}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected add songs forbidden, got %v", err)
	}
	items, err := svc.AddSongsToPlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, []uint{3, 999})
	if err != nil {
		t.Fatalf("expected add songs by uuid success: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one added item")
	}

	if err := svc.RemoveItemFromPlaylist(ctx, 1, 1, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected remove item not found, got %v", err)
	}
	if err := svc.RemoveItemFromPlaylist(ctx, 2, 1, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected remove item forbidden, got %v", err)
	}
	if err := svc.RemoveItemFromPlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, addedItem.ID); err != nil {
		t.Fatalf("expected remove item success: %v", err)
	}

	if err := svc.RemoveSongFromPlaylist(ctx, 2, 1, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected remove song forbidden, got %v", err)
	}
	if err := svc.RemoveSongFromPlaylistByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, 1); err != nil {
		t.Fatalf("expected remove song success: %v", err)
	}

	if err := svc.ReorderPlaylistItems(ctx, 2, 1, []uint{1}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected reorder forbidden, got %v", err)
	}
	if err := svc.ReorderPlaylistItemsByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, []uint{2}); err != nil {
		t.Fatalf("expected reorder by uuid success: %v", err)
	}
}

func TestPlaylistService_DBErrorFallback(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, false)

	if _, err := svc.CreatePlaylist(ctx, 1, &dto.CreatePlaylistRequest{Name: "x"}); err == nil {
		t.Fatal("expected create error on missing schema")
	}
	if _, err := svc.GetUserPlaylists(ctx, 1); err == nil {
		t.Fatal("expected get user playlists error on missing schema")
	}
	if _, _, err := svc.GetPublicPlaylists(ctx, 1, 10); err == nil {
		t.Fatal("expected get public playlists error on missing schema")
	}
	if err := svc.DeletePlaylist(ctx, 1, 1); err == nil {
		t.Fatal("expected delete playlist error on missing schema")
	}
	if _, err := svc.UpdatePlaylist(ctx, 1, 1, &dto.UpdatePlaylistRequest{Name: "x"}); err == nil {
		t.Fatal("expected update playlist error on missing schema")
	}
	if _, err := svc.AddSongToPlaylist(ctx, 1, 1, 1); err == nil {
		t.Fatal("expected add song error on missing schema")
	}
	if _, err := svc.AddSongsToPlaylist(ctx, 1, 1, []uint{1, 2}); err == nil {
		t.Fatal("expected add songs error on missing schema")
	}
	if err := svc.RemoveSongFromPlaylist(ctx, 1, 1, 1); err == nil {
		t.Fatal("expected remove song error on missing schema")
	}
	if err := svc.RemoveItemFromPlaylist(ctx, 1, 1, 1); err == nil {
		t.Fatal("expected remove item error on missing schema")
	}
	if err := svc.ReorderPlaylistItems(ctx, 1, 1, []uint{1}); err == nil {
		t.Fatal("expected reorder items error on missing schema")
	}
}

func TestPlaylistService_AddSongToPlaylist_GranularErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("is-song-in-playlist query error", func(t *testing.T) {
		svc, db := setupPlaylistServiceWithDB(t, true)
		if err := db.Exec(`DROP TABLE playlist_items`).Error; err != nil {
			t.Fatalf("drop playlist_items: %v", err)
		}
		if err := db.Exec(`CREATE TABLE playlist_items (
			id INTEGER PRIMARY KEY,
			uuid TEXT,
			playlist_id INTEGER,
			position INTEGER,
			added_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`).Error; err != nil {
			t.Fatalf("recreate playlist_items without song_id: %v", err)
		}

		if _, err := svc.AddSongToPlaylist(ctx, 1, 1, 2); err == nil {
			t.Fatal("expected IsSongInPlaylist query error")
		}
	})

	t.Run("get-max-position query error", func(t *testing.T) {
		svc, db := setupPlaylistServiceWithDB(t, true)
		if err := db.Exec(`DROP TABLE playlist_items`).Error; err != nil {
			t.Fatalf("drop playlist_items: %v", err)
		}
		if err := db.Exec(`CREATE TABLE playlist_items (
			id INTEGER PRIMARY KEY,
			uuid TEXT,
			playlist_id INTEGER,
			song_id INTEGER,
			added_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`).Error; err != nil {
			t.Fatalf("recreate playlist_items without position: %v", err)
		}

		if _, err := svc.AddSongToPlaylist(ctx, 1, 1, 2); err == nil {
			t.Fatal("expected GetMaxPosition query error")
		}
	})

	t.Run("create playlist item error", func(t *testing.T) {
		svc, db := setupPlaylistServiceWithDB(t, true)
		if err := db.Exec(`CREATE TRIGGER fail_playlist_item_insert
			BEFORE INSERT ON playlist_items
			BEGIN
				SELECT RAISE(FAIL, 'insert denied');
			END;`).Error; err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		if _, err := svc.AddSongToPlaylist(ctx, 1, 1, 2); err == nil {
			t.Fatal("expected create playlist item error")
		}
	})
}

func TestPlaylistService_InvalidUUIDBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, true)

	if _, err := svc.UpdatePlaylistByUUID(ctx, "bad", 1, &dto.UpdatePlaylistRequest{Name: "x"}); err == nil {
		t.Fatal("expected invalid uuid on update by uuid")
	}
	if err := svc.DeletePlaylistByUUID(ctx, "bad", 1); err == nil {
		t.Fatal("expected invalid uuid on delete by uuid")
	}
	if _, err := svc.AddSongToPlaylistByUUID(ctx, "bad", 1, 1); err == nil {
		t.Fatal("expected invalid uuid on add song by uuid")
	}
	if _, err := svc.AddSongsToPlaylistByUUID(ctx, "bad", 1, []uint{1}); err == nil {
		t.Fatal("expected invalid uuid on add songs by uuid")
	}
	if err := svc.RemoveSongFromPlaylistByUUID(ctx, "bad", 1, 1); err == nil {
		t.Fatal("expected invalid uuid on remove song by uuid")
	}
	if err := svc.RemoveItemFromPlaylistByUUID(ctx, "bad", 1, 1); err == nil {
		t.Fatal("expected invalid uuid on remove item by uuid")
	}
	if err := svc.ReorderPlaylistItemsByUUID(ctx, "bad", 1, []uint{1}); err == nil {
		t.Fatal("expected invalid uuid on reorder by uuid")
	}
}

func TestPlaylistService_UUIDResolveErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, true)

	missing := "44444444-4444-4444-4444-444444444444"

	if _, err := svc.GetPlaylistByUUID(ctx, missing, 1); err == nil {
		t.Fatal("expected resolve uuid error on get by uuid")
	}
	if _, err := svc.UpdatePlaylistByUUID(ctx, missing, 1, &dto.UpdatePlaylistRequest{Name: "x"}); err == nil {
		t.Fatal("expected resolve uuid error on update by uuid")
	}
	if err := svc.DeletePlaylistByUUID(ctx, missing, 1); err == nil {
		t.Fatal("expected resolve uuid error on delete by uuid")
	}
	if _, err := svc.AddSongToPlaylistByUUID(ctx, missing, 1, 1); err == nil {
		t.Fatal("expected resolve uuid error on add song by uuid")
	}
	if _, err := svc.AddSongsToPlaylistByUUID(ctx, missing, 1, []uint{1}); err == nil {
		t.Fatal("expected resolve uuid error on add songs by uuid")
	}
	if err := svc.RemoveSongFromPlaylistByUUID(ctx, missing, 1, 1); err == nil {
		t.Fatal("expected resolve uuid error on remove song by uuid")
	}
	if err := svc.RemoveItemFromPlaylistByUUID(ctx, missing, 1, 1); err == nil {
		t.Fatal("expected resolve uuid error on remove item by uuid")
	}
	if err := svc.ReorderPlaylistItemsByUUID(ctx, missing, 1, []uint{1}); err == nil {
		t.Fatal("expected resolve uuid error on reorder by uuid")
	}
}

func TestPlaylistService_PublicAccessAndItemMismatchBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupPlaylistService(t, true)

	if _, err := svc.GetPlaylist(ctx, 3, 1); err != nil {
		t.Fatalf("expected non-owner access for public playlist, got %v", err)
	}

	if _, err := svc.UpdatePlaylist(ctx, 999, 1, &dto.UpdatePlaylistRequest{Name: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected update missing playlist to return ErrNotFound, got %v", err)
	}

	item := &model.PlaylistItem{PlaylistID: 2, SongID: 2, Position: 1}
	if err := svc.playlistItemRepo.Create(ctx, item); err != nil {
		t.Fatalf("seed mismatched playlist item failed: %v", err)
	}

	if err := svc.RemoveItemFromPlaylist(ctx, 1, 1, item.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected item-playlist mismatch to return ErrForbidden, got %v", err)
	}
}
