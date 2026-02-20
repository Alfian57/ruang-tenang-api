package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/router"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"

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
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	timeutil.LoadTimezone()

	if err := logger.Init(cfg.AppEnv); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Ruang Tenang API...")

	_, err = database.Connect(cfg)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	logger.Info("Database connected successfully")

	r := router.SetupRouter(cfg)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Info(fmt.Sprintf("Server running on http://localhost%s", addr))
		logger.Info(fmt.Sprintf("Swagger docs at http://localhost%s/swagger/index.html", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(fmt.Sprintf("Failed to start server: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	logger.Info("Server exited gracefully")
}
