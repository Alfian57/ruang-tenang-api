package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPlaylistTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestPlaylistHandler_InvalidRequestBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPlaylistHandler(nil)

	tests := []struct {
		name   string
		call   func(*gin.Context)
		method string
		target string
		body   string
		setup  func(*gin.Context)
		code   int
	}{
		{name: "create-invalid-json", call: h.CreatePlaylist, method: http.MethodPost, target: "/playlists", body: "{", code: http.StatusBadRequest},
		{name: "update-invalid-json", call: h.UpdatePlaylist, method: http.MethodPut, target: "/playlists/x", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "x"}} }, code: http.StatusBadRequest},
		{name: "add-song-invalid-json", call: h.AddSongToPlaylist, method: http.MethodPost, target: "/playlists/x/songs", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "x"}} }, code: http.StatusBadRequest},
		{name: "add-songs-invalid-json", call: h.AddSongsToPlaylist, method: http.MethodPost, target: "/playlists/x/songs/batch", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "x"}} }, code: http.StatusBadRequest},
		{name: "reorder-invalid-json", call: h.ReorderPlaylistItems, method: http.MethodPut, target: "/playlists/x/reorder", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "x"}} }, code: http.StatusBadRequest},
		{name: "remove-song-invalid-id", call: h.RemoveSongFromPlaylist, method: http.MethodDelete, target: "/playlists/x/songs/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "x"}, {Key: "songId", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "remove-item-invalid-id", call: h.RemoveItemFromPlaylist, method: http.MethodDelete, target: "/playlists/x/items/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "x"}, {Key: "itemId", Value: "bad"}} }, code: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newPlaylistTestContext(tc.method, tc.target, tc.body)
			if tc.setup != nil {
				tc.setup(c)
			}
			tc.call(c)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}

func setupPlaylistHandler(t *testing.T, withSchema bool) *PlaylistHandler {
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
		if err := db.Exec(`INSERT INTO songs (id, title, slug, file_path, song_category_id) VALUES (1, 'Song One', 'song-one', '/a.mp3', 1), (2, 'Song Two', 'song-two', '/b.mp3', 1)`).Error; err != nil {
			t.Fatalf("seed songs: %v", err)
		}
		if err := db.Exec(`INSERT INTO playlists (id, uuid, user_id, name, is_public) VALUES (1, '11111111-1111-1111-1111-111111111111', 1, 'P1', 0), (2, '22222222-2222-2222-2222-222222222222', 2, 'P2', 0), (3, '33333333-3333-3333-3333-333333333333', 2, 'P3', 1)`).Error; err != nil {
			t.Fatalf("seed playlists: %v", err)
		}
		if err := db.Exec(`INSERT INTO playlist_items (id, uuid, playlist_id, song_id, position) VALUES (1, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1, 1, 0)`).Error; err != nil {
			t.Fatalf("seed playlist item: %v", err)
		}
	}

	playlistRepo := repository.NewPlaylistRepository(db)
	playlistItemRepo := repository.NewPlaylistItemRepository(db)
	songRepo := repository.NewSongRepository(db)
	svc := service.NewPlaylistService(playlistRepo, playlistItemRepo, songRepo)
	return NewPlaylistHandler(svc)
}

func TestPlaylistHandler_CoreEndpoints_SuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create-and-list-success", func(t *testing.T) {
		h := setupPlaylistHandler(t, true)

		c1, w1 := newPlaylistTestContext(http.MethodPost, "/playlists", `{"name":"My List","description":"desc","is_public":true}`)
		c1.Set("user_id", uint(1))
		h.CreatePlaylist(c1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w1.Code)
		}

		c2, w2 := newPlaylistTestContext(http.MethodGet, "/playlists", "")
		c2.Set("user_id", uint(1))
		h.GetMyPlaylists(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}
	})

	t.Run("create-and-list-internal-error", func(t *testing.T) {
		h := setupPlaylistHandler(t, false)

		c1, w1 := newPlaylistTestContext(http.MethodPost, "/playlists", `{"name":"My List"}`)
		c1.Set("user_id", uint(1))
		h.CreatePlaylist(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w1.Code)
		}

		c2, w2 := newPlaylistTestContext(http.MethodGet, "/playlists", "")
		c2.Set("user_id", uint(1))
		h.GetMyPlaylists(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("get-update-delete-branches", func(t *testing.T) {
		h := setupPlaylistHandler(t, true)

		forbiddenGet, wfg := newPlaylistTestContext(http.MethodGet, "/playlists/2", "")
		forbiddenGet.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		forbiddenGet.Set("user_id", uint(1))
		h.GetPlaylist(forbiddenGet)
		if wfg.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", wfg.Code)
		}

		successGet, wsg := newPlaylistTestContext(http.MethodGet, "/playlists/1", "")
		successGet.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
		successGet.Set("user_id", uint(1))
		h.GetPlaylist(successGet)
		if wsg.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wsg.Code)
		}

		invalidGet, wig := newPlaylistTestContext(http.MethodGet, "/playlists/bad", "")
		invalidGet.Params = gin.Params{{Key: "uuid", Value: "bad-uuid"}}
		invalidGet.Set("user_id", uint(1))
		h.GetPlaylist(invalidGet)
		if wig.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", wig.Code)
		}

		successUpdate, wsu := newPlaylistTestContext(http.MethodPut, "/playlists/1", `{"name":"Updated","description":"x"}`)
		successUpdate.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
		successUpdate.Set("user_id", uint(1))
		h.UpdatePlaylist(successUpdate)
		if wsu.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wsu.Code)
		}

		forbiddenUpdate, wfu := newPlaylistTestContext(http.MethodPut, "/playlists/2", `{"name":"Updated","description":"x"}`)
		forbiddenUpdate.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		forbiddenUpdate.Set("user_id", uint(1))
		h.UpdatePlaylist(forbiddenUpdate)
		if wfu.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", wfu.Code)
		}

		notFoundUpdate, wnu := newPlaylistTestContext(http.MethodPut, "/playlists/missing", `{"name":"Updated","description":"x"}`)
		notFoundUpdate.Params = gin.Params{{Key: "uuid", Value: "99999999-9999-9999-9999-999999999999"}}
		notFoundUpdate.Set("user_id", uint(1))
		h.UpdatePlaylist(notFoundUpdate)
		if wnu.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", wnu.Code)
		}

		successDelete, wsd := newPlaylistTestContext(http.MethodDelete, "/playlists/1", "")
		successDelete.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
		successDelete.Set("user_id", uint(1))
		h.DeletePlaylist(successDelete)
		if wsd.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", wsd.Code)
		}

		forbiddenDelete, wfd := newPlaylistTestContext(http.MethodDelete, "/playlists/2", "")
		forbiddenDelete.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		forbiddenDelete.Set("user_id", uint(1))
		h.DeletePlaylist(forbiddenDelete)
		if wfd.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", wfd.Code)
		}

		notFoundDelete, wnd := newPlaylistTestContext(http.MethodDelete, "/playlists/missing", "")
		notFoundDelete.Params = gin.Params{{Key: "uuid", Value: "99999999-9999-9999-9999-999999999999"}}
		notFoundDelete.Set("user_id", uint(1))
		h.DeletePlaylist(notFoundDelete)
		if wnd.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", wnd.Code)
		}
	})

	t.Run("update-delete-internal-errors", func(t *testing.T) {
		h := setupPlaylistHandler(t, false)

		updateErr, wue := newPlaylistTestContext(http.MethodPut, "/playlists/x", `{"name":"X","description":"Y"}`)
		updateErr.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
		updateErr.Set("user_id", uint(1))
		h.UpdatePlaylist(updateErr)
		if wue.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 update, got %d", wue.Code)
		}

		deleteErr, wde := newPlaylistTestContext(http.MethodDelete, "/playlists/x", "")
		deleteErr.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
		deleteErr.Set("user_id", uint(1))
		h.DeletePlaylist(deleteErr)
		if wde.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 delete, got %d", wde.Code)
		}
	})
}

func TestPlaylistHandler_ItemEndpoints_Branches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupPlaylistHandler(t, true)

	addDup, wDup := newPlaylistTestContext(http.MethodPost, "/playlists/1/songs", `{"song_id":1}`)
	addDup.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	addDup.Set("user_id", uint(1))
	h.AddSongToPlaylist(addDup)
	if wDup.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate, got %d", wDup.Code)
	}

	addMissingSong, wMissing := newPlaylistTestContext(http.MethodPost, "/playlists/1/songs", `{"song_id":999}`)
	addMissingSong.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	addMissingSong.Set("user_id", uint(1))
	h.AddSongToPlaylist(addMissingSong)
	if wMissing.Code != http.StatusNotFound {
		t.Fatalf("expected 404 missing song, got %d", wMissing.Code)
	}

	addForbidden, wForbidden := newPlaylistTestContext(http.MethodPost, "/playlists/2/songs", `{"song_id":2}`)
	addForbidden.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
	addForbidden.Set("user_id", uint(1))
	h.AddSongToPlaylist(addForbidden)
	if wForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 add forbidden, got %d", wForbidden.Code)
	}

	addOk, wAddOk := newPlaylistTestContext(http.MethodPost, "/playlists/1/songs", `{"song_id":2}`)
	addOk.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	addOk.Set("user_id", uint(1))
	h.AddSongToPlaylist(addOk)
	if wAddOk.Code != http.StatusCreated {
		t.Fatalf("expected 201 add success, got %d", wAddOk.Code)
	}

	addBatchForbidden, wBatchForbidden := newPlaylistTestContext(http.MethodPost, "/playlists/2/songs/batch", `{"song_ids":[1,2]}`)
	addBatchForbidden.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
	addBatchForbidden.Set("user_id", uint(1))
	h.AddSongsToPlaylist(addBatchForbidden)
	if wBatchForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 batch forbidden, got %d", wBatchForbidden.Code)
	}

	addBatchOK, wBatchOK := newPlaylistTestContext(http.MethodPost, "/playlists/1/songs/batch", `{"song_ids":[1,2]}`)
	addBatchOK.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	addBatchOK.Set("user_id", uint(1))
	h.AddSongsToPlaylist(addBatchOK)
	if wBatchOK.Code != http.StatusCreated {
		t.Fatalf("expected 201 batch success, got %d", wBatchOK.Code)
	}

	removeSongForbidden, wRemoveForbidden := newPlaylistTestContext(http.MethodDelete, "/playlists/2/songs/1", "")
	removeSongForbidden.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}, {Key: "songId", Value: "1"}}
	removeSongForbidden.Set("userID", uint(1))
	h.RemoveSongFromPlaylist(removeSongForbidden)
	if wRemoveForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 remove song forbidden, got %d", wRemoveForbidden.Code)
	}

	removeSongOK, wRemoveSongOK := newPlaylistTestContext(http.MethodDelete, "/playlists/1/songs/2", "")
	removeSongOK.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}, {Key: "songId", Value: "2"}}
	removeSongOK.Set("userID", uint(1))
	h.RemoveSongFromPlaylist(removeSongOK)
	if wRemoveSongOK.Code != http.StatusOK {
		t.Fatalf("expected 200 remove song success, got %d", wRemoveSongOK.Code)
	}

	removeItemNotFound, wItemNotFound := newPlaylistTestContext(http.MethodDelete, "/playlists/1/items/999", "")
	removeItemNotFound.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}, {Key: "itemId", Value: "999"}}
	removeItemNotFound.Set("userID", uint(1))
	h.RemoveItemFromPlaylist(removeItemNotFound)
	if wItemNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected 404 item not found, got %d", wItemNotFound.Code)
	}

	removeItemForbidden, wItemForbidden := newPlaylistTestContext(http.MethodDelete, "/playlists/2/items/1", "")
	removeItemForbidden.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}, {Key: "itemId", Value: "1"}}
	removeItemForbidden.Set("userID", uint(1))
	h.RemoveItemFromPlaylist(removeItemForbidden)
	if wItemForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 item forbidden, got %d", wItemForbidden.Code)
	}

	removeItemOK, wItemOK := newPlaylistTestContext(http.MethodDelete, "/playlists/1/items/1", "")
	removeItemOK.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}, {Key: "itemId", Value: "1"}}
	removeItemOK.Set("userID", uint(1))
	h.RemoveItemFromPlaylist(removeItemOK)
	if wItemOK.Code != http.StatusOK {
		t.Fatalf("expected 200 item remove success, got %d", wItemOK.Code)
	}

	reorderForbidden, wReorderForbidden := newPlaylistTestContext(http.MethodPut, "/playlists/2/reorder", `{"item_ids":[1]}`)
	reorderForbidden.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
	reorderForbidden.Set("userID", uint(1))
	h.ReorderPlaylistItems(reorderForbidden)
	if wReorderForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 reorder forbidden, got %d", wReorderForbidden.Code)
	}

	reorderOK, wReorderOK := newPlaylistTestContext(http.MethodPut, "/playlists/1/reorder", `{"item_ids":[1,2]}`)
	reorderOK.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	reorderOK.Set("userID", uint(1))
	h.ReorderPlaylistItems(reorderOK)
	if wReorderOK.Code != http.StatusOK {
		t.Fatalf("expected 200 reorder ok, got %d", wReorderOK.Code)
	}

	hErr := setupPlaylistHandler(t, false)

	batchErr, wBatchErr := newPlaylistTestContext(http.MethodPost, "/playlists/1/songs/batch", `{"song_ids":[1]}`)
	batchErr.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	batchErr.Set("user_id", uint(1))
	hErr.AddSongsToPlaylist(batchErr)
	if wBatchErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 batch error, got %d", wBatchErr.Code)
	}

	removeSongErr, wRemoveSongErr := newPlaylistTestContext(http.MethodDelete, "/playlists/1/songs/1", "")
	removeSongErr.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}, {Key: "songId", Value: "1"}}
	removeSongErr.Set("userID", uint(1))
	hErr.RemoveSongFromPlaylist(removeSongErr)
	if wRemoveSongErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 remove song error, got %d", wRemoveSongErr.Code)
	}

	removeItemErr, wRemoveItemErr := newPlaylistTestContext(http.MethodDelete, "/playlists/1/items/1", "")
	removeItemErr.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}, {Key: "itemId", Value: "1"}}
	removeItemErr.Set("userID", uint(1))
	hErr.RemoveItemFromPlaylist(removeItemErr)
	if wRemoveItemErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 remove item error, got %d", wRemoveItemErr.Code)
	}

	reorderErr, wReorderErr := newPlaylistTestContext(http.MethodPut, "/playlists/1/reorder", `{"item_ids":[1]}`)
	reorderErr.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
	reorderErr.Set("userID", uint(1))
	hErr.ReorderPlaylistItems(reorderErr)
	if wReorderErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 reorder error, got %d", wReorderErr.Code)
	}
}

func TestPlaylistHandler_GetPublicPlaylists_Branches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	good := setupPlaylistHandler(t, true)
	c1, w1 := newPlaylistTestContext(http.MethodGet, "/playlists/public?page=0&limit=99", "")
	good.GetPublicPlaylists(c1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	bad := setupPlaylistHandler(t, false)
	c2, w2 := newPlaylistTestContext(http.MethodGet, "/playlists/public", "")
	bad.GetPublicPlaylists(c2)
	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w2.Code)
	}
}
