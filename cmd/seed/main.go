package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	seedtesting "github.com/Alfian57/ruang-tenang-api/internal/seed/testing"
	"gorm.io/gorm"
)

type SeedOptions struct {
	Mode  string
	Reset bool
	Count int
	Only  string
}

func main() {
	// Parse command line flags
	modeFlag := flag.String("mode", "", "Seed mode: production | development | testing")
	legacyEnvFlag := flag.String("env", "", "Legacy alias for --mode")
	resetFlag := flag.Bool("reset", false, "Reset DB before seeding (development/testing only)")
	countFlag := flag.Int("count", 0, "Optional count for sample/fake data (development/testing)")
	onlyFlag := flag.String("only", "", "Optional seeder group/table filter")
	flag.Parse()

	// Resolve mode priority: --mode > SEED_MODE > --env (legacy) > default
	resolvedMode := strings.ToLower(strings.TrimSpace(*modeFlag))
	if resolvedMode == "" {
		resolvedMode = strings.ToLower(strings.TrimSpace(os.Getenv("SEED_MODE")))
	}
	if resolvedMode == "" {
		resolvedMode = strings.ToLower(strings.TrimSpace(*legacyEnvFlag))
	}
	if resolvedMode == "" {
		resolvedMode = "development"
	}

	if resolvedMode == "prod" {
		resolvedMode = "production"
	}
	if resolvedMode == "dev" {
		resolvedMode = "development"
	}

	if resolvedMode != "production" && resolvedMode != "development" && resolvedMode != "testing" {
		log.Fatalf("❌ Invalid SEED_MODE/mode: %s (allowed: production|development|testing)", resolvedMode)
	}

	if *resetFlag && resolvedMode == "production" {
		log.Fatal("❌ --reset is forbidden in production mode")
	}

	if resolvedMode == "production" && strings.TrimSpace(os.Getenv("SEED_CONFIRM")) != "YES" {
		log.Fatal("❌ Production seeding requires explicit confirmation: SEED_CONFIRM=YES")
	}

	if *countFlag > 0 {
		_ = os.Setenv("SEED_COUNT", fmt.Sprintf("%d", *countFlag))
	}

	options := SeedOptions{
		Mode:  resolvedMode,
		Reset: *resetFlag,
		Count: *countFlag,
		Only:  normalizeSeederName(*onlyFlag),
	}

	// Display banner
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                   RUANG TENANG SEEDER                        ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Connect to database
	log.Println("📦 Connecting to database...")
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected")
	log.Println("")

	// Run appropriate seeder based on mode
	switch options.Mode {
	case "production":
		if err := runProductionSeeder(db, options); err != nil {
			log.Fatalf("❌ Production seeding failed: %v", err)
		}

	case "development":
		if err := runDevelopmentSeeder(db, options); err != nil {
			log.Fatalf("❌ Development seeding failed: %v", err)
		}

	case "testing":
		if err := runTestingSeeder(db, options); err != nil {
			log.Fatalf("❌ Testing seeding failed: %v", err)
		}

	default:
		log.Printf("❌ Unknown mode: %s", options.Mode)
		log.Println("   Valid options: production, development, testing")
		os.Exit(1)
	}

	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                    SEEDING COMPLETE ✅                       ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")
	log.Println("📋 Test Accounts (Development only):")
	log.Println("   Admin: admin@ruang-tenang.com / password")
	log.Println("   Member: gading@gmail.com / password")
}

func runTestingSeeder(db *gorm.DB, opts SeedOptions) error {
	seeder := seedtesting.NewTestingSeeder(db)

	if opts.Reset {
		if err := resetAllTables(db); err != nil {
			return fmt.Errorf("failed to reset testing database: %w", err)
		}
	}

	if opts.Only != "" && opts.Only != "all" && opts.Only != "seed" {
		return fmt.Errorf("testing mode currently supports --only=all|seed (received: %s)", opts.Only)
	}

	return seeder.Seed()
}

func resetAllTables(db *gorm.DB) error {
	var tables []string
	if err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = current_schema() ORDER BY tablename").Scan(&tables).Error; err != nil {
		return err
	}

	if len(tables) == 0 {
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

func normalizeSeederName(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	return strings.ReplaceAll(normalized, " ", "")
}

func shouldRunSeeder(only string, seederName string) bool {
	if only == "" || only == "all" {
		return true
	}
	return only == normalizeSeederName(seederName)
}
