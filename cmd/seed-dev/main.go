package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Alfian57/ruang-tenang-api/cmd/seed-dev/development"
	"github.com/Alfian57/ruang-tenang-api/cmd/seed-prod/production"
	"github.com/Alfian57/ruang-tenang-api/internal/seed"
	"gorm.io/gorm"
)

var (
	runCLIFn      = runCLI
	mainFatalFn   = func(v ...any) { log.Fatal(v...) }
	resetTablesFn = seed.ResetAllTables
	setenvFn      = os.Setenv
)

var devProductionSeeders = []seed.SeederRunner{
	{Name: "Level Configs", Fn: production.SeedLevelConfigs},
	{Name: "Article Categories", Fn: production.SeedArticleCategories},
	{Name: "Song Categories", Fn: production.SeedSongCategories},
	{Name: "Forum Categories", Fn: production.SeedForumCategories},
	{Name: "Story Categories", Fn: production.SeedStoryCategories},
	{Name: "Breathing Techniques", Fn: production.SeedBreathingTechniques},
	{Name: "Feature Definitions", Fn: production.SeedFeatureDefinitions},
	{Name: "Badge Definitions", Fn: production.SeedBadgeDefinitions},
	{Name: "Crisis Keywords", Fn: production.SeedCrisisKeywords},
	{Name: "Admin User", Fn: production.SeedAdminUser},
	{Name: "Map Regions", Fn: production.SeedMapRegions},
	{Name: "League Divisions", Fn: production.SeedLeagueDivisions},
	{Name: "Spin Rewards", Fn: production.SeedSpinRewards},
	{Name: "Streak Societies", Fn: production.SeedStreakSocieties},
	{Name: "Timed Challenge Templates", Fn: production.SeedTimedChallengeTemplates},
	{Name: "Rewards", Fn: production.SeedRewards},
	{Name: "Inspiring Stories", Fn: production.SeedInspiringStories},
	{Name: "League Seasons", Fn: production.SeedLeagueSeasons},
}

var devTestSeeders = []seed.SeederRunner{
	{Name: "Test Users", Fn: development.SeedTestUsers},
	{Name: "Articles", Fn: development.SeedArticles},
	{Name: "Songs", Fn: development.SeedSongs},
	{Name: "Forums", Fn: development.SeedForums},
	{Name: "Chat Sessions", Fn: development.SeedChatSessions},
	{Name: "User Moods", Fn: development.SeedUserMoods},
	{Name: "Community Data", Fn: development.SeedCommunityData},
	{Name: "Journals", Fn: development.SeedJournals},
	{Name: "Breathing Sessions", Fn: development.SeedBreathingSessions},
	{Name: "Playlists", Fn: development.SeedPlaylists},
	{Name: "Guilds", Fn: development.SeedGuilds},
	{Name: "Gamification", Fn: development.SeedGamification},
}

func main() {
	resetFlag := flag.Bool("reset", false, "Reset DB before seeding")
	countFlag := flag.Int("count", 0, "Optional count for sample/fake data")
	onlyFlag := flag.String("only", "", "Optional seeder group/table filter")
	flag.Parse()

	if *countFlag > 0 {
		_ = setenvFn("SEED_COUNT", fmt.Sprintf("%d", *countFlag))
	}

	if err := runCLIFn(seed.SeedOptions{
		Reset: *resetFlag,
		Count: *countFlag,
		Only:  seed.NormalizeSeederName(*onlyFlag),
	}); err != nil {
		mainFatalFn(err)
	}
}

func runCLI(opts seed.SeedOptions) error {
	log.Println("")
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║             RUANG TENANG SEEDER (DEVELOPMENT)               ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println("")

	return seed.ConnectAndRun(func(db *gorm.DB) error {
		return runDevelopmentSeeder(db, opts)
	})
}

func runDevelopmentSeeder(db *gorm.DB, opts seed.SeedOptions) error {
	log.Println("🧪 Starting DEVELOPMENT seeding...")
	log.Println("  → Seeding production data + development test data")
	log.Println("")

	if opts.Reset {
		log.Println("⚠️  --reset enabled: truncating all tables before seeding...")
		if err := resetTablesFn(db); err != nil {
			return err
		}
		log.Println("✅ Database reset complete")
		log.Println("")
	}

	if opts.Count > 0 {
		log.Printf("ℹ️  Using --count=%d (available to seeders via SEED_COUNT env)", opts.Count)
	}

	// Phase 1: Production data
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📦 Phase 1: Seeding Production Data")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, s := range devProductionSeeders {
		if !seed.ShouldRunSeeder(opts.Only, s.Name) {
			continue
		}
		log.Printf("  📦 %s...", s.Name)
		if err := s.Fn(db); err != nil {
			log.Printf("    ❌ Failed: %v", err)
			return err
		}
		log.Printf("    ✅ Done")
	}

	// Phase 2: Development test data
	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🧪 Phase 2: Seeding Development Test Data")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, s := range devTestSeeders {
		if !seed.ShouldRunSeeder(opts.Only, s.Name) {
			continue
		}
		log.Printf("  🧪 %s...", s.Name)
		if err := s.Fn(db); err != nil {
			log.Printf("    ❌ Failed: %v", err)
			return err
		}
		log.Printf("    ✅ Done")
	}

	log.Println("")
	log.Println("🎉 Development seeding completed!")
	log.Println("")
	log.Println("📋 Test Accounts:")
	log.Println("   Admin: admin@ruang-tenang.com / password")
	log.Println("   Member: gading@gmail.com / password")
	return nil
}
