package main

import (
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/router"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"

	_ "github.com/Alfian57/ruang-tenang-api/docs"
)

// @title Ruang Tenang API
// @version 1.0
// @description API untuk aplikasi Ruang Tenang - Platform Kesehatan Mental
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@ruangtenang.id

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	if err := logger.Init(cfg.AppEnv); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Ruang Tenang API...")

	// Connect to database
	_, err = database.Connect(cfg)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	logger.Info("Database connected successfully")

	// Setup router
	r := router.SetupRouter()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Info(fmt.Sprintf("Server running on http://localhost%s", addr))
	logger.Info(fmt.Sprintf("Swagger docs at http://localhost%s/swagger/index.html", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to start server: %v", err))
	}
}
