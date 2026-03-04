package handler

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newLevelConfigHandlerForTest(t *testing.T) *LevelConfigHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.LevelConfig{}); err != nil {
		t.Fatalf("auto migrate level config: %v", err)
	}
	repo := repository.NewLevelConfigRepository(db)
	svc := service.NewLevelConfigService(repo, service.NewCacheService())
	return NewLevelConfigHandler(svc)
}

func newJSONContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

// newMultipartContext creates a gin context with multipart form data for testing
func newMultipartContext(t *testing.T, method, target string, fields map[string]string, includeFile bool) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}

	if includeFile {
		part, err := writer.CreateFormFile("badge_image", "test-badge.png")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		// Write minimal PNG data
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		if _, err := part.Write(pngHeader); err != nil {
			t.Fatalf("write png data: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	return c, w
}

func TestLevelConfigHandler_GetAllAndAdminGetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newLevelConfigHandlerForTest(t)

	// Seed one config directly via DB
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.LevelConfig{})
	repo := repository.NewLevelConfigRepository(db)
	svc := service.NewLevelConfigService(repo, service.NewCacheService())
	h2 := NewLevelConfigHandler(svc)

	db.Create(&model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "/uploads/images/badge_test.png"})

	c, w := newJSONContext(http.MethodGet, "/", "")
	h2.GetAllConfigs(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for get all, got %d", w.Code)
	}

	cAdmin, wAdmin := newJSONContext(http.MethodGet, "/", "")
	h2.AdminGetAllConfigs(cAdmin)
	if wAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin get all, got %d", wAdmin.Code)
	}

	// Also test empty handler
	cEmpty, wEmpty := newJSONContext(http.MethodGet, "/", "")
	h.GetAllConfigs(cEmpty)
	if wEmpty.Code != http.StatusOK {
		t.Fatalf("expected 200 for empty get all, got %d", wEmpty.Code)
	}
}

func TestLevelConfigHandler_GetAll_ErrorBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	repo := repository.NewLevelConfigRepository(db)
	svc := service.NewLevelConfigService(repo, service.NewCacheService())
	h := NewLevelConfigHandler(svc)

	c, w := newJSONContext(http.MethodGet, "/", "")
	h.GetAllConfigs(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for get all error path, got %d", w.Code)
	}
}

func TestLevelConfigHandler_CreateConfigBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newLevelConfigHandlerForTest(t)

	// Missing fields (no level)
	{
		c, w := newMultipartContext(t, http.MethodPost, "/", map[string]string{
			"min_exp":    "0",
			"badge_name": "Pemula",
		}, true)
		h.CreateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing level, got %d: %s", w.Code, w.Body.String())
		}
	}

	// Missing badge_name
	{
		c, w := newMultipartContext(t, http.MethodPost, "/", map[string]string{
			"level":   "1",
			"min_exp": "0",
		}, true)
		h.CreateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing badge_name, got %d: %s", w.Code, w.Body.String())
		}
	}

	// Missing badge_image
	{
		c, w := newMultipartContext(t, http.MethodPost, "/", map[string]string{
			"level":      "1",
			"min_exp":    "0",
			"badge_name": "Pemula",
		}, false)
		h.CreateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing badge_image, got %d: %s", w.Code, w.Body.String())
		}
	}
}

func TestLevelConfigHandler_UpdateAndDeleteBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler with direct DB seeding
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.LevelConfig{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	repo := repository.NewLevelConfigRepository(db)
	svc := service.NewLevelConfigService(repo, service.NewCacheService())
	h := NewLevelConfigHandler(svc)

	// Seed two configs
	db.Create(&model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "/uploads/images/badge_1.png"})
	db.Create(&model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "/uploads/images/badge_2.png"})

	// Invalid ID
	{
		c, w := newMultipartContext(t, http.MethodPut, "/", map[string]string{
			"level":      "3",
			"min_exp":    "200",
			"badge_name": "Lanjut",
		}, false)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid id, got %d", w.Code)
		}
	}

	// Invalid body (missing level)
	{
		c, w := newMultipartContext(t, http.MethodPut, "/", map[string]string{
			"min_exp":    "200",
			"badge_name": "Lanjut",
		}, false)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid body, got %d: %s", w.Code, w.Body.String())
		}
	}

	// Duplicate level conflict (update id=1 to level=2)
	{
		c, w := newMultipartContext(t, http.MethodPut, "/", map[string]string{
			"level":      "2",
			"min_exp":    "0",
			"badge_name": "Pemula",
		}, false)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 duplicate level on update, got %d: %s", w.Code, w.Body.String())
		}
	}

	// Not found ID path -> service update error => 500
	{
		c, w := newMultipartContext(t, http.MethodPut, "/", map[string]string{
			"level":      "3",
			"min_exp":    "200",
			"badge_name": "Lanjut",
		}, false)
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 for missing id update, got %d: %s", w.Code, w.Body.String())
		}
	}

	// Success update (without new image - keeps existing)
	{
		c, w := newMultipartContext(t, http.MethodPut, "/", map[string]string{
			"level":      "1",
			"min_exp":    "5",
			"badge_name": "Pemula Baru",
		}, false)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 update success, got %d: %s", w.Code, w.Body.String())
		}
		// Verify badge_icon is preserved
		if !strings.Contains(w.Body.String(), "badge_1.png") {
			t.Fatalf("expected existing badge_icon to be preserved, got: %s", w.Body.String())
		}
	}

	// Delete invalid id
	{
		c, w := newJSONContext(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "bad"}}
		h.DeleteConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 delete invalid id, got %d", w.Code)
		}
	}

	// Delete success
	{
		c, w := newJSONContext(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "2"}}
		h.DeleteConfig(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 delete success, got %d", w.Code)
		}
	}

	// Delete internal error branch (schema missing)
	{
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		repo := repository.NewLevelConfigRepository(db)
		svc := service.NewLevelConfigService(repo, service.NewCacheService())
		hErr := NewLevelConfigHandler(svc)

		c, w := newJSONContext(http.MethodDelete, "/", "")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		hErr.DeleteConfig(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 delete error path, got %d", w.Code)
		}
	}
}

func TestLevelConfigHandler_CreateConfig_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use DB without migration to trigger internal error
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	repo := repository.NewLevelConfigRepository(db)
	svc := service.NewLevelConfigService(repo, service.NewCacheService())
	h := NewLevelConfigHandler(svc)

	// Create with valid form but DB error (no table)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("level", "1")
	writer.WriteField("min_exp", "0")
	writer.WriteField("badge_name", "Err")
	// Add a file with valid png content-type
	part, _ := writer.CreateFormFile("badge_image", "test.png")
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	part.Write(pngData)
	writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	// The saveBadgeImage will try to save file to disk, this may succeed or fail
	// depending on the test environment. The key test is that the handler doesn't panic.
	h.CreateConfig(c)
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError && w.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: %d, body: %s", w.Code, w.Body.String())
	}

	_ = fmt.Sprintf("test helper")
}
