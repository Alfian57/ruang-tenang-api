package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migrator interface {
	Up() error
}

var (
	loadConfigFn  = config.LoadConfig
	initLoggerFn  = logger.Init
	connectDBFn   = database.Connect
	runMigrateFn  = runMigrate
	logFatalFn    = func(v ...any) { log.Fatal(v...) }
	logSuccessFn  = func() { log.Println("Migrations applied successfully") }
	newMigratorFn = func(sourceURL, databaseURL string) (migrator, error) {
		return migrate.New(sourceURL, databaseURL)
	}
)

func main() {
	if err := runMigrateFn(); err != nil {
		logFatalFn(err)
		return
	}

	logSuccessFn()
}

func runMigrate() error {
	cfg, err := loadConfigFn()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := initLoggerFn(cfg.AppEnv); err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	if _, err := connectDBFn(cfg); err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	m, err := newMigratorFn(resolveMigrationsSourceURL(), cfg.DatabaseURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func resolveMigrationsSourceURL() string {
	if raw := strings.TrimSpace(os.Getenv("MIGRATIONS_PATH")); raw != "" {
		if strings.HasPrefix(raw, "file://") {
			return raw
		}

		if abs, err := filepath.Abs(raw); err == nil {
			return "file://" + filepath.ToSlash(abs)
		}
	}

	candidates := []string{"./migrations", "/app/migrations"}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}

		if abs, err := filepath.Abs(candidate); err == nil {
			return "file://" + filepath.ToSlash(abs)
		}
	}

	return "file://migrations"
}
