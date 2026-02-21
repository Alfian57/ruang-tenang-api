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

var (
	seedLoadConfigFn           = config.LoadConfig
	seedConnectDBFn            = database.Connect
	seedRunProductionSeederFn  = runProductionSeeder
	seedRunDevelopmentSeederFn = runDevelopmentSeeder
	seedRunTestingSeederFn     = runTestingSeeder
	resetAllTablesFn           = resetAllTables
	runSeedCLIFn               = runSeedCLI
	seedMainFatalFn            = func(v ...any) { log.Fatal(v...) }
	seedSetenvFn               = os.Setenv
	seedGetenvFn               = os.Getenv
)

type seedCLIArgs struct {
	modeFlag      string
	legacyEnvFlag string
	reset         bool
	count         int
	only          string
}

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

	if err := runSeedCLIFn(seedCLIArgs{
		modeFlag:      *modeFlag,
		legacyEnvFlag: *legacyEnvFlag,
		reset:         *resetFlag,
		count:         *countFlag,
		only:          *onlyFlag,
	}); err != nil {
		seedMainFatalFn(err)
	}
}

func resolveSeedMode(modeFlag string, seedModeEnv string, legacyEnvFlag string) string {
	// Resolve mode priority: --mode > SEED_MODE > --env (legacy) > default
	resolvedMode := strings.ToLower(strings.TrimSpace(modeFlag))
	if resolvedMode == "" {
		resolvedMode = strings.ToLower(strings.TrimSpace(seedModeEnv))
	}
	if resolvedMode == "" {
		resolvedMode = strings.ToLower(strings.TrimSpace(legacyEnvFlag))
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

	return resolvedMode
}

func runSeedCLI(args seedCLIArgs) error {
	resolvedMode := resolveSeedMode(args.modeFlag, seedGetenvFn("SEED_MODE"), args.legacyEnvFlag)

	if resolvedMode != "production" && resolvedMode != "development" && resolvedMode != "testing" {
		return fmt.Errorf("❌ Invalid SEED_MODE/mode: %s (allowed: production|development|testing)", resolvedMode)
	}

	if args.reset && resolvedMode == "production" {
		return fmt.Errorf("❌ --reset is forbidden in production mode")
	}

	if resolvedMode == "production" && strings.TrimSpace(seedGetenvFn("SEED_CONFIRM")) != "YES" {
		return fmt.Errorf("❌ Production seeding requires explicit confirmation: SEED_CONFIRM=YES")
	}

	if args.count > 0 {
		_ = seedSetenvFn("SEED_COUNT", fmt.Sprintf("%d", args.count))
	}

	options := SeedOptions{
		Mode:  resolvedMode,
		Reset: args.reset,
		Count: args.count,
		Only:  normalizeSeederName(args.only),
	}

	// Display banner
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                   RUANG TENANG SEEDER                        ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")

	// Load configuration
	cfg, err := seedLoadConfigFn()
	if err != nil {
		return fmt.Errorf("❌ Failed to load configuration: %w", err)
	}

	// Connect to database
	log.Println("📦 Connecting to database...")
	db, err := seedConnectDBFn(cfg)
	if err != nil {
		return fmt.Errorf("❌ Failed to connect to database: %w", err)
	}
	log.Println("✅ Database connected")
	log.Println("")

	// Run appropriate seeder based on mode
	switch options.Mode {
	case "production":
		if err := seedRunProductionSeederFn(db, options); err != nil {
			return fmt.Errorf("❌ Production seeding failed: %w", err)
		}

	case "development":
		if err := seedRunDevelopmentSeederFn(db, options); err != nil {
			return fmt.Errorf("❌ Development seeding failed: %w", err)
		}

	case "testing":
		if err := seedRunTestingSeederFn(db, options); err != nil {
			return fmt.Errorf("❌ Testing seeding failed: %w", err)
		}

	default:
		return fmt.Errorf("❌ Unknown mode: %s", options.Mode)
	}

	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                    SEEDING COMPLETE ✅                       ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")
	log.Println("📋 Test Accounts (Development only):")
	log.Println("   Admin: admin@ruang-tenang.com / password")
	log.Println("   Member: gading@gmail.com / password")

	return nil
}

func runTestingSeeder(db *gorm.DB, opts SeedOptions) error {
	seeder := seedtesting.NewTestingSeeder(db)

	if opts.Reset {
		if err := resetAllTablesFn(db); err != nil {
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
