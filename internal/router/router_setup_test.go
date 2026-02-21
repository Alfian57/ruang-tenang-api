package router

import (
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetupRouter_InitializesWithInMemoryDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	cfg := &config.Config{
		AppEnv:             "test",
		GeminiAPIKey:       "",
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	r := SetupRouter(cfg)
	if r == nil {
		t.Fatal("expected router instance")
	}

	routes := r.Routes()
	if len(routes) == 0 {
		t.Fatal("expected registered routes")
	}

	if !hasRoute(routes, "GET", "/health") {
		t.Fatal("expected health route")
	}
	if !hasRoute(routes, "GET", "/api/v1/articles") {
		t.Fatal("expected api route")
	}
}
