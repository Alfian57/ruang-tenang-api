package seed

import (
	"fmt"
	"log"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"gorm.io/gorm"
)

// ---- Public types ----

// SeedOptions holds CLI flags shared by every seeder binary.
type SeedOptions struct {
	Reset bool
	Count int
	Only  string
}

// SeederRunner pairs a human-readable name with a seed function.
type SeederRunner struct {
	Name string
	Fn   func(*gorm.DB) error
}

// ---- Dependency injection (for testing) ----

var (
	LoadConfigFn = config.LoadConfig
	ConnectDBFn  = database.Connect
)

// ---- Shared helpers ----

// NormalizeSeederName lowercases and strips spaces for matching.
func NormalizeSeederName(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	return strings.ReplaceAll(normalized, " ", "")
}

// ShouldRunSeeder returns true when the seeder name matches the --only filter.
func ShouldRunSeeder(only string, seederName string) bool {
	if only == "" || only == "all" {
		return true
	}
	return only == NormalizeSeederName(seederName)
}

// ResetAllTables truncates every table except schema_migrations.
func ResetAllTables(db *gorm.DB) error {
	var tables []string

	switch db.Dialector.Name() {
	case "sqlite":
		if err := db.Raw("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&tables).Error; err != nil {
			return err
		}
	default:
		if err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = current_schema() ORDER BY tablename").Scan(&tables).Error; err != nil {
			return err
		}
	}

	if len(tables) == 0 {
		return nil
	}

	if db.Dialector.Name() == "sqlite" {
		for _, table := range tables {
			if table == "schema_migrations" {
				continue
			}
			if err := db.Exec(fmt.Sprintf("DELETE FROM \"%s\"", table)).Error; err != nil {
				return err
			}
		}
		var sqliteSequenceCount int64
		if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'sqlite_sequence'").Scan(&sqliteSequenceCount).Error; err == nil && sqliteSequenceCount > 0 {
			_ = db.Exec("DELETE FROM sqlite_sequence").Error
		}
		return nil
	}

	quoted := make([]string, 0, len(tables))
	for _, table := range tables {
		if table == "schema_migrations" {
			continue
		}
		quoted = append(quoted, fmt.Sprintf("\"%s\"", table))
	}

	if len(quoted) == 0 {
		return nil
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(quoted, ", "))
	return db.Exec(query).Error
}

// ConnectAndRun loads config, connects to the DB, and invokes runFn.
func ConnectAndRun(runFn func(*gorm.DB) error) error {
	cfg, err := LoadConfigFn()
	if err != nil {
		return fmt.Errorf("❌ Failed to load configuration: %w", err)
	}

	log.Println("📦 Connecting to database...")
	db, err := ConnectDBFn(cfg)
	if err != nil {
		return fmt.Errorf("❌ Failed to connect to database: %w", err)
	}
	log.Println("✅ Database connected")
	log.Println("")

	return runFn(db)
}
