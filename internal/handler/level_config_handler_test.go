package handler

import (
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

func TestLevelConfigHandler_GetAllAndAdminGetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newLevelConfigHandlerForTest(t)

	// Seed one config
	c1, _ := newJSONContext(http.MethodPost, "/", `{"level":1,"min_exp":0,"badge_name":"Pemula","badge_icon":"🌱"}`)
	h.CreateConfig(c1)

	c, w := newJSONContext(http.MethodGet, "/", "")
	h.GetAllConfigs(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for get all, got %d", w.Code)
	}

	cAdmin, wAdmin := newJSONContext(http.MethodGet, "/", "")
	h.AdminGetAllConfigs(cAdmin)
	if wAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin get all, got %d", wAdmin.Code)
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

	// Invalid JSON/body
	{
		c, w := newJSONContext(http.MethodPost, "/", `{"level":0}`)
		h.CreateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid create body, got %d", w.Code)
		}
	}

	// Success
	{
		c, w := newJSONContext(http.MethodPost, "/", `{"level":1,"min_exp":0,"badge_name":"Pemula","badge_icon":"🌱"}`)
		h.CreateConfig(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 create success, got %d", w.Code)
		}
	}

	// Duplicate level
	{
		c, w := newJSONContext(http.MethodPost, "/", `{"level":1,"min_exp":10,"badge_name":"Pemula2","badge_icon":"🌿"}`)
		h.CreateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 duplicate level, got %d", w.Code)
		}
	}

	// Internal error (non-duplicate) branch
	{
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		repo := repository.NewLevelConfigRepository(db)
		svc := service.NewLevelConfigService(repo, service.NewCacheService())
		hErr := NewLevelConfigHandler(svc)

		c, w := newJSONContext(http.MethodPost, "/", `{"level":9,"min_exp":900,"badge_name":"Err","badge_icon":"x"}`)
		hErr.CreateConfig(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 create internal error, got %d", w.Code)
		}
	}
}

func TestLevelConfigHandler_UpdateAndDeleteBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newLevelConfigHandlerForTest(t)

	// Seed two configs
	seed1, _ := newJSONContext(http.MethodPost, "/", `{"level":1,"min_exp":0,"badge_name":"Pemula","badge_icon":"🌱"}`)
	h.CreateConfig(seed1)
	seed2, _ := newJSONContext(http.MethodPost, "/", `{"level":2,"min_exp":100,"badge_name":"Naik","badge_icon":"⭐"}`)
	h.CreateConfig(seed2)

	// Invalid ID
	{
		c, w := newJSONContext(http.MethodPut, "/", `{"level":3,"min_exp":200,"badge_name":"Lanjut","badge_icon":"🚀"}`)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid id, got %d", w.Code)
		}
	}

	// Invalid body
	{
		c, w := newJSONContext(http.MethodPut, "/", `{"level":0}`)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid body, got %d", w.Code)
		}
	}

	// Duplicate level conflict (update id=1 to level=2)
	{
		c, w := newJSONContext(http.MethodPut, "/", `{"level":2,"min_exp":0,"badge_name":"Pemula","badge_icon":"🌱"}`)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 duplicate level on update, got %d", w.Code)
		}
	}

	// Not found ID path -> service update error => 500
	{
		c, w := newJSONContext(http.MethodPut, "/", `{"level":3,"min_exp":200,"badge_name":"Lanjut","badge_icon":"🚀"}`)
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 for missing id update, got %d", w.Code)
		}
	}

	// Success update
	{
		c, w := newJSONContext(http.MethodPut, "/", `{"level":1,"min_exp":5,"badge_name":"Pemula Baru","badge_icon":"🌱"}`)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateConfig(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 update success, got %d", w.Code)
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
