package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newChatTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestChatHandler_InvalidInputBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChatHandler(nil)

	tests := []struct {
		name   string
		call   func(*gin.Context)
		method string
		target string
		body   string
		setup  func(*gin.Context)
		code   int
	}{
		{name: "get-sessions-invalid-query", call: h.GetSessions, method: http.MethodGet, target: "/chat-sessions?page=abc", code: http.StatusBadRequest},
		{name: "create-session-invalid-json", call: h.CreateSession, method: http.MethodPost, target: "/chat-sessions", body: "{", code: http.StatusBadRequest},
		{name: "send-message-invalid-json", call: h.SendMessage, method: http.MethodPost, target: "/chat-sessions/uuid/messages", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "uuid"}} }, code: http.StatusBadRequest},
		{name: "toggle-message-like-invalid-id", call: h.ToggleMessageLike, method: http.MethodPut, target: "/chat-messages/bad/like", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "toggle-message-dislike-invalid-id", call: h.ToggleMessageDislike, method: http.MethodPut, target: "/chat-messages/bad/dislike", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "create-folder-invalid-json", call: h.CreateFolder, method: http.MethodPost, target: "/chat-folders", body: "{", code: http.StatusBadRequest},
		{name: "update-folder-invalid-id", call: h.UpdateFolder, method: http.MethodPut, target: "/chat-folders/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "delete-folder-invalid-id", call: h.DeleteFolder, method: http.MethodDelete, target: "/chat-folders/bad", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "reorder-folders-invalid-json", call: h.ReorderFolders, method: http.MethodPut, target: "/chat-folders/reorder", body: "{", code: http.StatusBadRequest},
		{name: "move-to-folder-invalid-json", call: h.MoveToFolder, method: http.MethodPut, target: "/chat-sessions/uuid/folder", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "uuid"}} }, code: http.StatusBadRequest},
		{name: "toggle-message-pin-invalid-id", call: h.ToggleMessagePin, method: http.MethodPut, target: "/chat-messages/bad/pin", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "export-chat-invalid-json", call: h.ExportChat, method: http.MethodPost, target: "/chat-sessions/uuid/export", body: "{", setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "uuid"}} }, code: http.StatusBadRequest},
		{name: "suggested-prompts-invalid-query", call: h.GetSuggestedPrompts, method: http.MethodGet, target: "/chat-prompts?has_messages=abc", code: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newChatTestContext(tc.method, tc.target, tc.body)
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

func setupChatHandlerService(t *testing.T) *ChatHandler {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	createFolders := `CREATE TABLE chat_folders (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
		name TEXT,
		color TEXT,
		icon TEXT,
		position INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	)`
	createSessions := `CREATE TABLE chat_sessions (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
		folder_id INTEGER,
		title TEXT,
		summary TEXT,
		summary_generated_at DATETIME,
		is_favorite BOOLEAN,
		is_trash BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`
	createMessages := `CREATE TABLE chat_messages (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		chat_session_id INTEGER,
		role TEXT,
		content TEXT,
		type TEXT,
		is_liked BOOLEAN,
		is_disliked BOOLEAN,
		is_pinned BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME
	)`

	for _, query := range []string{createFolders, createSessions, createMessages} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO chat_folders (id, uuid, user_id, name, color, icon, position, created_at, updated_at) VALUES (1, '11111111-1111-1111-1111-111111111111', 1, 'Folder A', '#6366f1', 'folder', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed folder: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, folder_id, title, summary, summary_generated_at, is_favorite, is_trash, created_at, updated_at) VALUES (1, '22222222-2222-2222-2222-222222222222', 1, 1, 'Session 1', 'existing summary', CURRENT_TIMESTAMP, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed session: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, type, is_pinned, created_at, updated_at) VALUES (1, '33333333-3333-3333-3333-333333333333', 1, 'ai', 'important', 'text', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed message: %v", err)
	}

	svc := service.NewChatService(
		repository.NewChatSessionRepository(db),
		repository.NewChatMessageRepository(db),
		&config.Config{GeminiAPIKey: "", AppEnv: "test"},
		service.NewGamificationService(db),
		nil,
	)
	svc.SetFolderRepo(repository.NewChatFolderRepository(db))

	return NewChatHandler(svc)
}

func setupChatHandlerErrorService(t *testing.T) *ChatHandler {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := service.NewChatService(
		repository.NewChatSessionRepository(db),
		repository.NewChatMessageRepository(db),
		&config.Config{GeminiAPIKey: "", AppEnv: "test"},
		service.NewGamificationService(db),
		nil,
	)
	svc.SetFolderRepo(repository.NewChatFolderRepository(db))

	return NewChatHandler(svc)
}

func TestChatHandler_UUIDBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChatHandler(&service.ChatService{})

	tests := []struct {
		name   string
		call   func(*gin.Context)
		method string
		target string
		setup  func(*gin.Context)
		code   int
	}{
		{name: "get-session-invalid-uuid", call: h.GetSession, method: http.MethodGet, target: "/chat-sessions/bad", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusNotFound},
		{name: "toggle-trash-invalid-uuid", call: h.ToggleTrash, method: http.MethodPut, target: "/chat-sessions/bad/trash", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "toggle-favorite-invalid-uuid", call: h.ToggleFavorite, method: http.MethodPut, target: "/chat-sessions/bad/favorite", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "delete-session-invalid-uuid", call: h.DeleteSession, method: http.MethodDelete, target: "/chat-sessions/bad", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "get-pinned-invalid-uuid", call: h.GetPinnedMessages, method: http.MethodGet, target: "/chat-sessions/bad/pinned", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "get-summary-invalid-uuid", call: h.GetSummary, method: http.MethodGet, target: "/chat-sessions/bad/summary", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusNotFound},
		{name: "generate-summary-invalid-uuid", call: h.GenerateSummary, method: http.MethodPost, target: "/chat-sessions/bad/summary", setup: func(c *gin.Context) { c.Set("user_id", uint(1)); c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }, code: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newChatTestContext(tc.method, tc.target, "")
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

func TestChatHandler_SuccessBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get-sessions-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodGet, "/chat-sessions?page=1&limit=10&filter=all", "")
		c.Set("user_id", uint(1))

		h.GetSessions(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("create-session-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodPost, "/chat-sessions", `{"title":"New Session"}`)
		c.Set("user_id", uint(1))

		h.CreateSession(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
	})

	t.Run("send-message-success-and-unauthorized", func(t *testing.T) {
		h := setupChatHandlerService(t)

		c1, w1 := newChatTestContext(http.MethodPost, "/chat-sessions/22222222-2222-2222-2222-222222222222/messages", `{"content":"halo ai"}`)
		c1.Set("user_id", uint(1))
		c1.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		h.SendMessage(c1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 send message success, got %d", w1.Code)
		}

		c2, w2 := newChatTestContext(http.MethodPost, "/chat-sessions/22222222-2222-2222-2222-222222222222/messages", `{"content":"unauthorized"}`)
		c2.Set("user_id", uint(2))
		c2.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		h.SendMessage(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 send message unauthorized, got %d", w2.Code)
		}
	})

	t.Run("get-folders-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodGet, "/chat-folders", "")
		c.Set("user_id", uint(1))

		h.GetFolders(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-session-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodGet, "/chat-sessions/22222222-2222-2222-2222-222222222222", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}

		h.GetSession(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("toggle-trash-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodPut, "/chat-sessions/22222222-2222-2222-2222-222222222222/trash", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}

		h.ToggleTrash(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("toggle-favorite-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodPut, "/chat-sessions/22222222-2222-2222-2222-222222222222/favorite", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}

		h.ToggleFavorite(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("toggle-like-dislike-pin-success-and-error", func(t *testing.T) {
		h := setupChatHandlerService(t)

		cl, wl := newChatTestContext(http.MethodPut, "/chat-messages/1/like", "")
		cl.Set("user_id", uint(1))
		cl.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ToggleMessageLike(cl)
		if wl.Code != http.StatusOK {
			t.Fatalf("expected 200 toggle like, got %d", wl.Code)
		}

		cd, wd := newChatTestContext(http.MethodPut, "/chat-messages/1/dislike", "")
		cd.Set("user_id", uint(1))
		cd.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ToggleMessageDislike(cd)
		if wd.Code != http.StatusOK {
			t.Fatalf("expected 200 toggle dislike, got %d", wd.Code)
		}

		cp, wp := newChatTestContext(http.MethodPut, "/chat-messages/1/pin", "")
		cp.Set("user_id", uint(1))
		cp.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ToggleMessagePin(cp)
		if wp.Code != http.StatusOK {
			t.Fatalf("expected 200 toggle pin, got %d", wp.Code)
		}

		cl2, wl2 := newChatTestContext(http.MethodPut, "/chat-messages/999/like", "")
		cl2.Set("user_id", uint(1))
		cl2.Params = gin.Params{{Key: "id", Value: "999"}}
		h.ToggleMessageLike(cl2)
		if wl2.Code != http.StatusOK {
			t.Fatalf("expected 200 like missing msg path, got %d", wl2.Code)
		}
	})

	t.Run("folder-crud-reorder-move-success", func(t *testing.T) {
		h := setupChatHandlerService(t)

		c1, w1 := newChatTestContext(http.MethodPost, "/chat-folders", `{"name":"Focus","color":"#4f46e5","icon":"sparkles"}`)
		c1.Set("user_id", uint(1))
		h.CreateFolder(c1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("expected 201 create folder, got %d", w1.Code)
		}

		c2, w2 := newChatTestContext(http.MethodPut, "/chat-folders/1", `{"name":"Renamed","color":"#334155","icon":"folder"}`)
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateFolder(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 update folder, got %d", w2.Code)
		}

		c3, w3 := newChatTestContext(http.MethodPut, "/chat-folders/reorder", `{"folder_ids":[1]}`)
		c3.Set("user_id", uint(1))
		h.ReorderFolders(c3)
		if w3.Code != http.StatusOK {
			t.Fatalf("expected 200 reorder folders, got %d", w3.Code)
		}

		c4, w4 := newChatTestContext(http.MethodPut, "/chat-sessions/22222222-2222-2222-2222-222222222222/folder", `{"folder_id":1}`)
		c4.Set("user_id", uint(1))
		c4.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		h.MoveToFolder(c4)
		if w4.Code != http.StatusOK {
			t.Fatalf("expected 200 move to folder, got %d", w4.Code)
		}

		c5, w5 := newChatTestContext(http.MethodDelete, "/chat-folders/1", "")
		c5.Set("user_id", uint(1))
		c5.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteFolder(c5)
		if w5.Code != http.StatusOK {
			t.Fatalf("expected 200 delete folder, got %d", w5.Code)
		}
	})

	t.Run("delete-session-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodDelete, "/chat-sessions/22222222-2222-2222-2222-222222222222", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}

		h.DeleteSession(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-pinned-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodGet, "/chat-sessions/22222222-2222-2222-2222-222222222222/pinned", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}

		h.GetPinnedMessages(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-summary-success", func(t *testing.T) {
		h := setupChatHandlerService(t)
		c, w := newChatTestContext(http.MethodGet, "/chat-sessions/22222222-2222-2222-2222-222222222222/summary", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}

		h.GetSummary(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("export-and-generate-summary-branches", func(t *testing.T) {
		h := setupChatHandlerService(t)

		ce, we := newChatTestContext(http.MethodPost, "/chat-sessions/22222222-2222-2222-2222-222222222222/export", `{"format":"txt","include_metadata":true}`)
		ce.Set("user_id", uint(1))
		ce.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		h.ExportChat(ce)
		if we.Code != http.StatusOK {
			t.Fatalf("expected 200 export chat, got %d", we.Code)
		}

		cg, wg := newChatTestContext(http.MethodPost, "/chat-sessions/22222222-2222-2222-2222-222222222222/summary", "")
		cg.Set("user_id", uint(1))
		cg.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		h.GenerateSummary(cg)
		if wg.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 generate summary path, got %d", wg.Code)
		}
	})

	t.Run("get-suggested-prompts-success", func(t *testing.T) {
		h := NewChatHandler(&service.ChatService{})
		c, w := newChatTestContext(http.MethodGet, "/chat-prompts?has_messages=true&mood=happy&time_of_day=night", "")
		c.Set("user_id", uint(1))

		h.GetSuggestedPrompts(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get-suggested-prompts-service-error-missing-user", func(t *testing.T) {
		h := NewChatHandler(&service.ChatService{})
		c, w := newChatTestContext(http.MethodGet, "/chat-prompts?has_messages=true&mood=happy&time_of_day=night", "")

		h.GetSuggestedPrompts(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestChatHandler_ServiceErrorBranchesExtended(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupChatHandlerErrorService(t)

	t.Run("sessions and create session errors", func(t *testing.T) {
		c1, w1 := newChatTestContext(http.MethodGet, "/chat-sessions?page=0&limit=99", "")
		c1.Set("user_id", uint(1))
		h.GetSessions(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 get sessions, got %d", w1.Code)
		}

		c2, w2 := newChatTestContext(http.MethodPost, "/chat-sessions", `{"title":"new"}`)
		c2.Set("user_id", uint(1))
		h.CreateSession(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 create session, got %d", w2.Code)
		}
	})

	t.Run("session uuid operations errors", func(t *testing.T) {
		uuid := "22222222-2222-2222-2222-222222222222"

		c1, w1 := newChatTestContext(http.MethodGet, "/chat-sessions/"+uuid, "")
		c1.Set("user_id", uint(1))
		c1.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.GetSession(c1)
		if w1.Code != http.StatusNotFound {
			t.Fatalf("expected 404 get session, got %d", w1.Code)
		}

		c2, w2 := newChatTestContext(http.MethodPut, "/chat-sessions/"+uuid+"/trash", "")
		c2.Set("user_id", uint(1))
		c2.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.ToggleTrash(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 toggle trash, got %d", w2.Code)
		}

		c3, w3 := newChatTestContext(http.MethodPut, "/chat-sessions/"+uuid+"/favorite", "")
		c3.Set("user_id", uint(1))
		c3.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.ToggleFavorite(c3)
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 toggle favorite, got %d", w3.Code)
		}

		c4, w4 := newChatTestContext(http.MethodDelete, "/chat-sessions/"+uuid, "")
		c4.Set("user_id", uint(1))
		c4.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.DeleteSession(c4)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 delete session, got %d", w4.Code)
		}

		c5, w5 := newChatTestContext(http.MethodPost, "/chat-sessions/"+uuid+"/messages", `{"content":"halo"}`)
		c5.Set("user_id", uint(1))
		c5.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.SendMessage(c5)
		if w5.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 send message, got %d", w5.Code)
		}

		c6, w6 := newChatTestContext(http.MethodGet, "/chat-sessions/"+uuid+"/pinned", "")
		c6.Set("user_id", uint(1))
		c6.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.GetPinnedMessages(c6)
		if w6.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 get pinned, got %d", w6.Code)
		}

		c7, w7 := newChatTestContext(http.MethodGet, "/chat-sessions/"+uuid+"/summary", "")
		c7.Set("user_id", uint(1))
		c7.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.GetSummary(c7)
		if w7.Code != http.StatusNotFound {
			t.Fatalf("expected 404 get summary, got %d", w7.Code)
		}

		c8, w8 := newChatTestContext(http.MethodPost, "/chat-sessions/"+uuid+"/summary", "")
		c8.Set("user_id", uint(1))
		c8.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.GenerateSummary(c8)
		if w8.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 generate summary, got %d", w8.Code)
		}

		c9, w9 := newChatTestContext(http.MethodPost, "/chat-sessions/"+uuid+"/export", `{"format":"txt"}`)
		c9.Set("user_id", uint(1))
		c9.Params = gin.Params{{Key: "uuid", Value: uuid}}
		h.ExportChat(c9)
		if w9.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 export chat, got %d", w9.Code)
		}
	})

	t.Run("folder and message action errors", func(t *testing.T) {
		c1, w1 := newChatTestContext(http.MethodGet, "/chat-folders", "")
		c1.Set("user_id", uint(1))
		h.GetFolders(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 get folders, got %d", w1.Code)
		}

		c2, w2 := newChatTestContext(http.MethodPost, "/chat-folders", `{"name":"x"}`)
		c2.Set("user_id", uint(1))
		h.CreateFolder(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 create folder, got %d", w2.Code)
		}

		c3, w3 := newChatTestContext(http.MethodPut, "/chat-folders/1", `{"name":"x"}`)
		c3.Set("user_id", uint(1))
		c3.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateFolder(c3)
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 update folder, got %d", w3.Code)
		}

		c4, w4 := newChatTestContext(http.MethodDelete, "/chat-folders/1", "")
		c4.Set("user_id", uint(1))
		c4.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteFolder(c4)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 delete folder, got %d", w4.Code)
		}

		c5, w5 := newChatTestContext(http.MethodPut, "/chat-folders/reorder", `{"folder_ids":[1,2]}`)
		c5.Set("user_id", uint(1))
		h.ReorderFolders(c5)
		if w5.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 reorder folders, got %d", w5.Code)
		}

		c6, w6 := newChatTestContext(http.MethodPut, "/chat-sessions/22222222-2222-2222-2222-222222222222/folder", `{"folder_id":1}`)
		c6.Set("user_id", uint(1))
		c6.Params = gin.Params{{Key: "uuid", Value: "22222222-2222-2222-2222-222222222222"}}
		h.MoveToFolder(c6)
		if w6.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 move to folder, got %d", w6.Code)
		}

		c7, w7 := newChatTestContext(http.MethodPut, "/chat-messages/1/like", "")
		c7.Set("user_id", uint(1))
		c7.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ToggleMessageLike(c7)
		if w7.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 toggle like, got %d", w7.Code)
		}

		c8, w8 := newChatTestContext(http.MethodPut, "/chat-messages/1/dislike", "")
		c8.Set("user_id", uint(1))
		c8.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ToggleMessageDislike(c8)
		if w8.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 toggle dislike, got %d", w8.Code)
		}

		c9, w9 := newChatTestContext(http.MethodPut, "/chat-messages/1/pin", "")
		c9.Set("user_id", uint(1))
		c9.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ToggleMessagePin(c9)
		if w9.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 toggle pin, got %d", w9.Code)
		}
	})
}
