package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/router"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"go.uber.org/zap"

	_ "github.com/Alfian57/ruang-tenang-api/docs"
)

var (
	loadConfigFn   = config.LoadConfig
	loadTimezoneFn = timeutil.LoadTimezone
	initLoggerFn   = logger.Init
	connectDBFn    = database.Connect
	setupRouterFn  = router.SetupRouter
	newServerFn    = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	listenServerFn = func(s *http.Server) error { return s.ListenAndServe() }
	shutdownFn     = func(s *http.Server, ctx context.Context) error { return s.Shutdown(ctx) }
	notifySignalFn = func(ch chan os.Signal) { signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM) }
	logInfoFn      = logger.Info
	logFatalFn     = logger.Fatal
	loggerSyncFn   = logger.Sync
	runServerFn    = runServer
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
	if err := runServerFn(nil); err != nil {
		if strings.HasPrefix(err.Error(), "Failed to load config:") || strings.HasPrefix(err.Error(), "Failed to initialize logger:") {
			panic(err.Error())
		}
		logFatalFn(err.Error())
	}
}

func runServer(quit <-chan os.Signal) error {
	info := logInfoFn
	fatal := logFatalFn
	listen := listenServerFn
	shutdown := shutdownFn
	syncLogger := loggerSyncFn

	cfg, err := loadConfigFn()
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	loadTimezoneFn()

	if err := initLoggerFn(cfg.AppEnv); err != nil {
		return fmt.Errorf("Failed to initialize logger: %v", err)
	}
	defer syncLogger()

	info("Starting Ruang Tenang API...")

	_, err = connectDBFn(cfg)
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}

	info("Database connected successfully")

	r := setupRouterFn(cfg)
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := newServerFn(addr, r)

	go func() {
		info(fmt.Sprintf("Server running on http://localhost%s", addr))
		info(fmt.Sprintf("Swagger docs at http://localhost%s/swagger/index.html", addr))
		if err := listen(srv); err != nil && err != http.ErrServerClosed {
			fatal(fmt.Sprintf("Failed to start server: %v", err))
		}
	}()

	if quit == nil {
		localQuit := make(chan os.Signal, 1)
		notifySignalFn(localQuit)
		quit = localQuit
	}

	<-quit

	info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := shutdown(srv, ctx); err != nil {
		return fmt.Errorf("Server forced to shutdown: %v", err)
	}

	info("Server exited gracefully")
	return nil
}

type loggerInfoFunc = func(msg string, fields ...zap.Field)
