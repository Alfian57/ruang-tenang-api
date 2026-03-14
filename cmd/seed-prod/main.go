package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Alfian57/ruang-tenang-api/cmd/seed-prod/production"
	"github.com/Alfian57/ruang-tenang-api/internal/seed"
	"gorm.io/gorm"
)

var (
	runCLIFn    = runCLI
	mainFatalFn = func(v ...any) { log.Fatal(v...) }
	getenvFn    = os.Getenv
)

var productionSeeders = []seed.SeederRunner{
	{Name: "Level Configs", Fn: production.SeedLevelConfigs},
	{Name: "Article Categories", Fn: production.SeedArticleCategories},
	{Name: "Song Categories", Fn: production.SeedSongCategories},
	{Name: "Forum Categories", Fn: production.SeedForumCategories},
	{Name: "Story Categories", Fn: production.SeedStoryCategories},
	{Name: "Breathing Techniques", Fn: production.SeedBreathingTechniques},
	{Name: "Feature Definitions", Fn: production.SeedFeatureDefinitions},
	{Name: "Badge Definitions", Fn: production.SeedBadgeDefinitions},
	{Name: "Crisis Keywords", Fn: production.SeedCrisisKeywords},
	{Name: "Default Accounts", Fn: production.SeedDefaultAccounts},
	{Name: "Articles", Fn: production.SeedArticles},
	{Name: "Map Regions", Fn: production.SeedMapRegions},
	{Name: "League Divisions", Fn: production.SeedLeagueDivisions},
	{Name: "Spin Rewards", Fn: production.SeedSpinRewards},
	{Name: "Streak Societies", Fn: production.SeedStreakSocieties},
	{Name: "Timed Challenge Templates", Fn: production.SeedTimedChallengeTemplates},
	{Name: "Rewards", Fn: production.SeedRewards},
	{Name: "League Seasons", Fn: production.SeedLeagueSeasons},
}

func main() {
	onlyFlag := flag.String("only", "", "Optional seeder group/table filter")
	flag.Parse()

	if err := runCLIFn(seed.SeedOptions{
		Only: seed.NormalizeSeederName(*onlyFlag),
	}); err != nil {
		mainFatalFn(err)
	}
}

func runCLI(opts seed.SeedOptions) error {
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║              RUANG TENANG SEEDER (PRODUCTION)               ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")

	return seed.ConnectAndRun(func(db *gorm.DB) error {
		return runProductionSeeder(db, opts)
	})
}

func runProductionSeeder(db *gorm.DB, opts seed.SeedOptions) error {
	log.Println("🏭 Starting PRODUCTION seeding...")
	log.Println("  → Seeding essential application data only")
	log.Println("")

	seeders := append([]seed.SeederRunner{}, productionSeeders...)

	if getenvFn("SEED_PROD_CREATE_ADMIN") == "true" {
		if getenvFn("SEED_ADMIN_EMAIL") == "" || getenvFn("SEED_ADMIN_PASSWORD") == "" {
			return fmt.Errorf("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are required when SEED_PROD_CREATE_ADMIN=true")
		}
		seeders = append(seeders, seed.SeederRunner{Name: "Admin User", Fn: production.SeedAdminUser})
	}

	for _, s := range seeders {
		if !seed.ShouldRunSeeder(opts.Only, s.Name) {
			continue
		}
		log.Printf("📦 %s...", s.Name)
		if err := s.Fn(db); err != nil {
			log.Printf("  ❌ Failed: %v", err)
			return err
		}
		log.Printf("  ✅ Done")
	}

	log.Println("")
	log.Println("🎉 Production seeding completed!")
	return nil
}
