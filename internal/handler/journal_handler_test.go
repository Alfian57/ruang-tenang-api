package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newJournalTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestJournalHandler_GuardBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewJournalHandler(&service.JournalService{})
	h.SetDailyTaskService(&mockDailyTaskSvcForHandler{})

	t.Run("create-journal-invalid-json", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodPost, "/journals", "{")
		c.Set("user_id", uint(1))

		h.CreateJournal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("get-journal-invalid-uuid", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodGet, "/journals/bad", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "bad"}}

		h.GetJournal(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("update-journal-invalid-json", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodPut, "/journals/bad", "{")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "bad"}}

		h.UpdateJournal(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("delete-journal-invalid-uuid", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodDelete, "/journals/bad", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "bad"}}

		h.DeleteJournal(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("search-journals-query-required", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodGet, "/journals/search", "")
		c.Set("user_id", uint(1))

		h.SearchJournals(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("search-journals-service-error", func(t *testing.T) {
		hErr := setupJournalHandlerWithService(t, false)
		c, w := newJournalTestContext(http.MethodGet, "/journals/search?q=tenang&limit=5", "")
		c.Set("user_id", uint(1))

		hErr.SearchJournals(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestJournalHandler_HelperFunctions(t *testing.T) {
	t.Run("split tags trims blanks", func(t *testing.T) {
		tags := splitTags(" calm, focus , ,night ")
		if len(tags) != 3 {
			t.Fatalf("expected 3 tags, got %d", len(tags))
		}
		if tags[0] != "calm" || tags[1] != "focus" || tags[2] != "night" {
			t.Fatalf("unexpected tags: %#v", tags)
		}
	})

	t.Run("split string without separator", func(t *testing.T) {
		parts := splitString("single", ",")
		if len(parts) != 1 || parts[0] != "single" {
			t.Fatalf("unexpected parts: %#v", parts)
		}
	})

	t.Run("trim space tabs and spaces", func(t *testing.T) {
		got := trimSpace("\t  hello world  \t")
		if got != "hello world" {
			t.Fatalf("expected 'hello world', got %q", got)
		}
	})
}

func setupJournalHandlerWithService(t *testing.T, withSchema bool) *JournalHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
		schema := []string{
			`CREATE TABLE journals (
				id INTEGER PRIMARY KEY,
				uuid TEXT,
				user_id INTEGER,
				title TEXT,
				content TEXT,
				mood_id INTEGER,
				tags TEXT,
				is_private BOOLEAN,
				share_with_ai BOOLEAN,
				word_count INTEGER,
				summary TEXT,
				created_at DATETIME,
				updated_at DATETIME,
				deleted_at DATETIME
			)`,
			`INSERT INTO journals (id, uuid, user_id, title, content, share_with_ai, created_at, updated_at) VALUES (1, '11111111-1111-1111-1111-111111111111', 1, 't', 'c', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		}
		for _, q := range schema {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup journal schema failed: %v", err)
			}
		}
	}

	svc := service.NewJournalService(
		repository.NewJournalRepository(db),
		repository.NewJournalSettingsRepository(db),
		repository.NewJournalAIAccessLogRepository(db),
		repository.NewUserMoodRepository(db),
		nil,
	)

	return NewJournalHandler(svc)
}

func TestJournalHandler_MoreEndpointBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list journals error branch", func(t *testing.T) {
		h := setupJournalHandlerWithService(t, false)
		c, w := newJournalTestContext(http.MethodGet, "/journals?page=1&limit=10&tags=a,b&mood=2&start_date=2026-01-01&end_date=2026-01-31", "")
		c.Set("user_id", uint(1))
		h.ListJournals(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("settings endpoints", func(t *testing.T) {
		h := setupJournalHandlerWithService(t, false)

		c1, w1 := newJournalTestContext(http.MethodGet, "/journals/settings", "")
		c1.Set("user_id", uint(1))
		h.GetSettings(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w1.Code)
		}

		c2, w2 := newJournalTestContext(http.MethodPut, "/journals/settings", "{")
		c2.Set("user_id", uint(1))
		h.UpdateSettings(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		c3, w3 := newJournalTestContext(http.MethodPut, "/journals/settings", `{"allow_ai_access":true}`)
		c3.Set("user_id", uint(1))
		h.UpdateSettings(c3)
		if w3.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w3.Code)
		}
	})

	t.Run("ai and analytics endpoints", func(t *testing.T) {
		h := setupJournalHandlerWithService(t, false)

		cases := []struct {
			name   string
			method string
			target string
			call   func(*gin.Context)
			body   string
			code   int
			setup  func(*gin.Context)
		}{
			{name: "get-ai-context", method: http.MethodGet, target: "/journals/ai-context?query=x&max_entries=3&include_summary=true", call: h.GetAIContext, code: http.StatusInternalServerError},
			{name: "get-ai-access-logs", method: http.MethodGet, target: "/journals/ai-access-logs?limit=2", call: h.GetAIAccessLogs, code: http.StatusInternalServerError},
			{name: "get-analytics", method: http.MethodGet, target: "/journals/analytics", call: h.GetAnalytics, code: http.StatusInternalServerError},
			{name: "get-writing-prompt", method: http.MethodGet, target: "/journals/prompt", call: h.GetWritingPrompt, code: http.StatusInternalServerError},
			{name: "get-weekly-summary", method: http.MethodGet, target: "/journals/weekly-summary", call: h.GetWeeklySummary, code: http.StatusInternalServerError},
			{name: "export-invalid-json", method: http.MethodPost, target: "/journals/export", call: h.ExportJournals, body: "{", code: http.StatusBadRequest},
			{name: "export-service-error", method: http.MethodPost, target: "/journals/export", call: h.ExportJournals, body: `{"format":"json"}`, code: http.StatusBadRequest},
			{name: "toggle-ai-share-not-found", method: http.MethodPost, target: "/journals/bad/toggle-ai-share", call: h.ToggleAIShare, code: http.StatusNotFound, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "uuid", Value: "bad"}} }},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				c, w := newJournalTestContext(tc.method, tc.target, tc.body)
				c.Set("user_id", uint(1))
				if tc.setup != nil {
					tc.setup(c)
				}
				tc.call(c)
				if w.Code != tc.code {
					t.Fatalf("expected %d, got %d", tc.code, w.Code)
				}
			})
		}
	})

	t.Run("toggle-ai-share-update-error", func(t *testing.T) {
		h := setupJournalHandlerWithService(t, true)
		c, w := newJournalTestContext(http.MethodPost, "/journals/11111111-1111-1111-1111-111111111111/toggle-ai-share", "")
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: "11111111-1111-1111-1111-111111111111"}}
		h.ToggleAIShare(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("update-journal-service-error", func(t *testing.T) {
		hErr := setupJournalHandlerWithService(t, false)
		known := "11111111-1111-1111-1111-111111111111"
		c, w := newJournalTestContext(http.MethodPut, "/journals/"+known, `{"title":"x"}`)
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: known}}
		hErr.UpdateJournal(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 update journal not found, got %d", w.Code)
		}
	})

	t.Run("export-journals-internal-error", func(t *testing.T) {
		hErr := setupJournalHandlerWithService(t, false)
		c, w := newJournalTestContext(http.MethodPost, "/journals/export", `{"format":"txt"}`)
		c.Set("user_id", uint(1))
		hErr.ExportJournals(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 export internal error, got %d", w.Code)
		}
	})
}

func setupJournalHandlerSuccess(t *testing.T) (*JournalHandler, *gorm.DB, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	knownUUID := uuid.New().String()
	schema := []string{
		`CREATE TABLE user_moods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			mood TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE journals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			user_id INTEGER NOT NULL,
			title TEXT,
			content TEXT,
			summary TEXT,
			mood_id INTEGER,
			tags TEXT,
			is_private BOOLEAN,
			share_with_ai BOOLEAN,
			ai_accessed_at DATETIME,
			word_count INTEGER,
			sentiment_score REAL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE journal_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE,
			allow_ai_access BOOLEAN,
			ai_context_days INTEGER,
			ai_context_max_entries INTEGER,
			default_share_with_ai BOOLEAN,
			is_blocked BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE journal_ai_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			journal_id INTEGER,
			chat_session_id INTEGER,
			accessed_at DATETIME,
			context_type TEXT
		)`,
		`INSERT INTO journal_settings (user_id, allow_ai_access, ai_context_days, ai_context_max_entries, default_share_with_ai, is_blocked, created_at, updated_at)
		 VALUES (1, 1, 7, 5, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO journals (uuid, user_id, title, content, summary, share_with_ai, word_count, created_at, updated_at)
		 VALUES ('` + knownUUID + `', 1, 'Existing', 'existing content body', '', 0, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
	}

	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	svc := service.NewJournalService(
		repository.NewJournalRepository(db),
		repository.NewJournalSettingsRepository(db),
		repository.NewJournalAIAccessLogRepository(db),
		repository.NewUserMoodRepository(db),
		nil,
	)

	return NewJournalHandler(svc), db, knownUUID
}

func TestJournalHandler_SuccessCRUDAndExportBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, knownUUID := setupJournalHandlerSuccess(t)

	t.Run("create-journal-success", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodPost, "/journals", `{"title":"New Entry","content":"hari ini tenang"}`)
		c.Set("user_id", uint(1))
		h.CreateJournal(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
	})

	t.Run("update-journal-success", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodPut, "/journals/"+knownUUID, `{"title":"Updated Existing"}`)
		c.Set("user_id", uint(1))
		c.Params = gin.Params{{Key: "uuid", Value: knownUUID}}
		h.UpdateJournal(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("list-journals-success", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodGet, "/journals?page=1&limit=10", "")
		c.Set("user_id", uint(1))
		h.ListJournals(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("export-journals-success", func(t *testing.T) {
		c, w := newJournalTestContext(http.MethodPost, "/journals/export", `{"format":"txt"}`)
		c.Set("user_id", uint(1))
		h.ExportJournals(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestJournalHandler_SuccessAndExtraErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, db, knownUUID := setupJournalHandlerSuccess(t)

	t.Run("get-and-delete-journal-success", func(t *testing.T) {
		getCtx, getW := newJournalTestContext(http.MethodGet, "/journals/"+knownUUID, "")
		getCtx.Set("user_id", uint(1))
		getCtx.Params = gin.Params{{Key: "uuid", Value: knownUUID}}
		h.GetJournal(getCtx)
		if getW.Code != http.StatusOK {
			t.Fatalf("expected 200 get journal, got %d", getW.Code)
		}

		delCtx, delW := newJournalTestContext(http.MethodDelete, "/journals/"+knownUUID, "")
		delCtx.Set("user_id", uint(1))
		delCtx.Params = gin.Params{{Key: "uuid", Value: knownUUID}}
		h.DeleteJournal(delCtx)
		if delW.Code != http.StatusOK {
			t.Fatalf("expected 200 delete journal, got %d", delW.Code)
		}
	})

	t.Run("settings-ai-analytics-success", func(t *testing.T) {
		ctx1, w1 := newJournalTestContext(http.MethodGet, "/journals/settings", "")
		ctx1.Set("user_id", uint(1))
		h.GetSettings(ctx1)
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 get settings, got %d", w1.Code)
		}

		ctx2, w2 := newJournalTestContext(http.MethodPut, "/journals/settings", `{"allow_ai_access":true,"ai_context_days":7,"ai_context_max_entries":5}`)
		ctx2.Set("user_id", uint(1))
		h.UpdateSettings(ctx2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 update settings, got %d", w2.Code)
		}

		ctx3, w3 := newJournalTestContext(http.MethodGet, "/journals/ai-context?max_entries=2&include_summary=false", "")
		ctx3.Set("user_id", uint(1))
		h.GetAIContext(ctx3)
		if w3.Code != http.StatusOK {
			t.Fatalf("expected 200 get ai context, got %d", w3.Code)
		}

		ctx4, w4 := newJournalTestContext(http.MethodGet, "/journals/ai-access-logs?limit=5", "")
		ctx4.Set("user_id", uint(1))
		h.GetAIAccessLogs(ctx4)
		if w4.Code != http.StatusOK {
			t.Fatalf("expected 200 get ai access logs, got %d", w4.Code)
		}

		ctx5, w5 := newJournalTestContext(http.MethodGet, "/journals/analytics", "")
		ctx5.Set("user_id", uint(1))
		h.GetAnalytics(ctx5)
		if w5.Code != http.StatusOK {
			t.Fatalf("expected 200 get analytics, got %d", w5.Code)
		}

		ctx6, w6 := newJournalTestContext(http.MethodGet, "/journals/prompt", "")
		ctx6.Set("user_id", uint(1))
		h.GetWritingPrompt(ctx6)
		if w6.Code != http.StatusOK {
			t.Fatalf("expected 200 get prompt, got %d", w6.Code)
		}

		ctx7, w7 := newJournalTestContext(http.MethodGet, "/journals/weekly-summary", "")
		ctx7.Set("user_id", uint(1))
		h.GetWeeklySummary(ctx7)
		if w7.Code != http.StatusOK {
			t.Fatalf("expected 200 get weekly summary, got %d", w7.Code)
		}
	})

	t.Run("toggle-ai-share-success", func(t *testing.T) {
		newID := uuid.New().String()
		if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, summary, share_with_ai, word_count, created_at, updated_at) VALUES (?, 1, 'Toggle Entry', 'toggle content body', '', 0, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, newID).Error; err != nil {
			t.Fatalf("seed toggle journal: %v", err)
		}

		ctx, w := newJournalTestContext(http.MethodPost, "/journals/"+newID+"/toggle-ai-share", "")
		ctx.Set("user_id", uint(1))
		ctx.Params = gin.Params{{Key: "uuid", Value: newID}}
		h.ToggleAIShare(ctx)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 toggle ai share, got %d", w.Code)
		}
	})

	t.Run("create-journal-service-error", func(t *testing.T) {
		hErr := setupJournalHandlerWithService(t, false)
		ctx, w := newJournalTestContext(http.MethodPost, "/journals", `{"title":"x","content":"y"}`)
		ctx.Set("user_id", uint(1))
		hErr.CreateJournal(ctx)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 create journal error, got %d", w.Code)
		}
	})
}
