package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	origLoadConfigFn   = loadConfigFn
	origLoadTimezoneFn = loadTimezoneFn
	origInitLoggerFn   = initLoggerFn
	origConnectDBFn    = connectDBFn
	origSetupRouterFn  = setupRouterFn
	origNewServerFn    = newServerFn
	origListenServerFn = listenServerFn
	origShutdownFn     = shutdownFn
	origNotifySignalFn = notifySignalFn
	origLogInfoFn      = logInfoFn
	origLogFatalFn     = logFatalFn
	origLoggerSyncFn   = loggerSyncFn
	origRunServerFn    = runServerFn
)

func resetRunDeps() {
	loadConfigFn = origLoadConfigFn
	loadTimezoneFn = origLoadTimezoneFn
	initLoggerFn = origInitLoggerFn
	connectDBFn = origConnectDBFn
	setupRouterFn = origSetupRouterFn
	newServerFn = origNewServerFn
	listenServerFn = origListenServerFn
	shutdownFn = origShutdownFn
	notifySignalFn = origNotifySignalFn
	logInfoFn = origLogInfoFn
	logFatalFn = origLogFatalFn
	loggerSyncFn = origLoggerSyncFn
	runServerFn = origRunServerFn
}

func TestMain_PanicsWhenConfigInvalid(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic from invalid config")
		}
		if !strings.Contains(fmt.Sprint(recovered), "Failed to load config") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	main()
}

func TestRunServer_Success(t *testing.T) {
	t.Cleanup(resetRunDeps)

	loadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test", AppPort: "8080"}, nil }
	loadTimezoneFn = func() {}
	initLoggerFn = func(string) error { return nil }
	loggerSyncFn = func() {}
	connectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	setupRouterFn = func(*config.Config) *gin.Engine { return gin.New() }
	newServerFn = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	listenServerFn = func(*http.Server) error { return http.ErrServerClosed }
	shutdownFn = func(*http.Server, context.Context) error { return nil }
	logInfoFn = func(string, ...zap.Field) {}
	logFatalFn = func(string, ...zap.Field) {}

	quit := make(chan os.Signal, 1)
	quit <- os.Interrupt

	if err := runServer(quit); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestRunServer_ErrorBranches(t *testing.T) {
	t.Cleanup(resetRunDeps)

	loadConfigFn = func() (*config.Config, error) { return nil, errors.New("cfg") }
	if err := runServer(make(chan os.Signal, 1)); err == nil || !strings.Contains(err.Error(), "Failed to load config") {
		t.Fatalf("expected config error, got %v", err)
	}

	loadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test", AppPort: "8080"}, nil }
	loadTimezoneFn = func() {}
	initLoggerFn = func(string) error { return errors.New("logger") }
	if err := runServer(make(chan os.Signal, 1)); err == nil || !strings.Contains(err.Error(), "Failed to initialize logger") {
		t.Fatalf("expected logger error, got %v", err)
	}

	initLoggerFn = func(string) error { return nil }
	loggerSyncFn = func() {}
	logInfoFn = func(string, ...zap.Field) {}
	logFatalFn = func(string, ...zap.Field) {}
	connectDBFn = func(*config.Config) (*gorm.DB, error) { return nil, errors.New("db") }
	if err := runServer(make(chan os.Signal, 1)); err == nil || !strings.Contains(err.Error(), "Failed to connect to database") {
		t.Fatalf("expected db error, got %v", err)
	}

	connectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	setupRouterFn = func(*config.Config) *gin.Engine { return gin.New() }
	newServerFn = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	listenServerFn = func(*http.Server) error { return http.ErrServerClosed }
	shutdownFn = func(*http.Server, context.Context) error { return errors.New("shutdown") }
	logInfoFn = func(string, ...zap.Field) {}
	logFatalFn = func(string, ...zap.Field) {}
	quit := make(chan os.Signal, 1)
	quit <- os.Interrupt
	if err := runServer(quit); err == nil || !strings.Contains(err.Error(), "Server forced to shutdown") {
		t.Fatalf("expected shutdown error, got %v", err)
	}
}

func TestRunServer_NilQuitUsesNotifySignal(t *testing.T) {
	t.Cleanup(resetRunDeps)

	loadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test", AppPort: "8080"}, nil }
	loadTimezoneFn = func() {}
	initLoggerFn = func(string) error { return nil }
	loggerSyncFn = func() {}
	connectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	setupRouterFn = func(*config.Config) *gin.Engine { return gin.New() }
	newServerFn = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	listenServerFn = func(*http.Server) error { return http.ErrServerClosed }
	shutdownFn = func(*http.Server, context.Context) error { return nil }
	logInfoFn = func(string, ...zap.Field) {}
	logFatalFn = func(string, ...zap.Field) {}

	notifyCalled := false
	notifySignalFn = func(ch chan os.Signal) {
		notifyCalled = true
		ch <- os.Interrupt
	}

	if err := runServer(nil); err != nil {
		t.Fatalf("expected success with nil quit, got %v", err)
	}
	if !notifyCalled {
		t.Fatal("expected notifySignalFn to be used when quit is nil")
	}
}

func TestRunServer_ListenErrorCallsFatal(t *testing.T) {
	t.Cleanup(resetRunDeps)

	loadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test", AppPort: "8080"}, nil }
	loadTimezoneFn = func() {}
	initLoggerFn = func(string) error { return nil }
	loggerSyncFn = func() {}
	connectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	setupRouterFn = func(*config.Config) *gin.Engine { return gin.New() }
	newServerFn = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	shutdownFn = func(*http.Server, context.Context) error { return nil }
	logInfoFn = func(string, ...zap.Field) {}

	called := make(chan struct{}, 1)
	var once sync.Once
	logFatalFn = func(string, ...zap.Field) {
		once.Do(func() { called <- struct{}{} })
	}
	listenServerFn = func(*http.Server) error { return errors.New("listen boom") }

	quit := make(chan os.Signal, 1)
	quit <- os.Interrupt

	if err := runServer(quit); err != nil {
		t.Fatalf("expected runServer to return nil, got %v", err)
	}

	select {
	case <-called:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected logFatalFn to be called on listen error")
	}
}

func TestMain_Branches(t *testing.T) {
	t.Cleanup(resetRunDeps)

	fatalCalled := false
	logFatalFn = func(string, ...zap.Field) { fatalCalled = true }

	runServerFn = func(<-chan os.Signal) error { return errors.New("runtime failure") }
	main()
	if !fatalCalled {
		t.Fatal("expected logFatalFn called for runtime failure")
	}

	runServerFn = func(<-chan os.Signal) error { return errors.New("Failed to initialize logger: boom") }
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on logger init failure")
		}
	}()
	main()
}
