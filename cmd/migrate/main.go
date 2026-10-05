package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	Down() error
	Steps(n int) error
	Drop() error
	Version() (uint, bool, error)
	Force(version int) error
}

var (
	loadConfigFn  = config.LoadConfig
	initLoggerFn  = logger.Init
	connectDBFn   = database.Connect
	runMigrateFn  = runMigrate
	logFatalFn    = func(v ...any) { log.Fatal(v...) }
	logInfoFn     = func(v ...any) { log.Println(v...) }
	newMigratorFn = func(sourceURL, databaseURL string) (migrator, error) {
		return migrate.New(sourceURL, databaseURL)
	}
)

func main() {
	args := os.Args[1:]
	cmd := "up"
	var cmdArgs []string
	if len(args) > 0 {
		cmd = strings.ToLower(args[0])
		cmdArgs = args[1:]
	}

	if err := runMigrateFn(cmd, cmdArgs); err != nil {
		logFatalFn(err)
		return
	}
}

func runMigrate(cmd string, args []string) error {
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

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate up failed: %w", err)
		}
		logInfoFn("Migrations applied successfully (up-to-date)")

	case "down":
		steps := 1
		if len(args) > 0 {
			if args[0] == "all" {
				if err := m.Down(); err != nil && err != migrate.ErrNoChange {
					return fmt.Errorf("migrate down all failed: %w", err)
				}
				logInfoFn("All migrations rolled back successfully")
				return nil
			}
			parsedSteps, err := strconv.Atoi(args[0])
			if err != nil || parsedSteps <= 0 {
				return fmt.Errorf("invalid step count: %s (must be positive integer or 'all')", args[0])
			}
			steps = parsedSteps
		}
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate down %d step(s) failed: %w", steps, err)
		}
		logInfoFn(fmt.Sprintf("Rolled back %d migration step(s) successfully", steps))

	case "fresh":
		logInfoFn("Dropping all database tables...")
		if err := m.Drop(); err != nil {
			return fmt.Errorf("migrate fresh drop failed: %w", err)
		}
		logInfoFn("All tables dropped. Re-applying all migrations from scratch...")
		mFresh, err := newMigratorFn(resolveMigrationsSourceURL(), cfg.DatabaseURL)
		if err != nil {
			return err
		}
		if err := mFresh.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate fresh up failed: %w", err)
		}
		logInfoFn("Fresh migrations applied successfully")

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				logInfoFn("No migrations applied yet (version: 0)")
				return nil
			}
			return fmt.Errorf("failed to get migration version: %w", err)
		}
		logInfoFn(fmt.Sprintf("Current migration version: %d (dirty: %t)", v, dirty))

	case "force":
		if len(args) == 0 {
			return fmt.Errorf("usage: force <version>")
		}
		v, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version: %s (must be integer)", args[0])
		}
		if err := m.Force(v); err != nil {
			return fmt.Errorf("migrate force %d failed: %w", v, err)
		}
		logInfoFn(fmt.Sprintf("Migration forced to version %d", v))

	default:
		return fmt.Errorf("unknown command '%s'. Available commands: up, down [steps|all], fresh, version, force <version>", cmd)
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
