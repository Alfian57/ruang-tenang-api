package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/seed"
	"github.com/Alfian57/ruang-tenang-api/internal/seed/presentation"
	"gorm.io/gorm"
)

var (
	runCLIFn      = runCLI
	mainFatalFn   = func(v ...any) { log.Fatal(v...) }
	resetTablesFn = seed.ResetAllTables
	setenvFn      = os.Setenv
)

var presentationSeeders = []seed.SeederRunner{
	{Name: "Level Configs", Fn: presentation.SeedLevelConfigs},
	{Name: "Article Categories", Fn: presentation.SeedArticleCategories},
	{Name: "Song Categories", Fn: presentation.SeedSongCategories},
	{Name: "Forum Categories", Fn: presentation.SeedForumCategories},
	{Name: "Story Categories", Fn: presentation.SeedStoryCategories},
	{Name: "Breathing Techniques", Fn: presentation.SeedBreathingTechniques},
	{Name: "Feature Definitions", Fn: presentation.SeedFeatureDefinitions},
	{Name: "Badge Definitions", Fn: presentation.SeedBadgeDefinitions},
	{Name: "Crisis Keywords", Fn: presentation.SeedCrisisKeywords},
	{Name: "Map Regions", Fn: presentation.SeedMapRegions},
	{Name: "League Divisions", Fn: presentation.SeedLeagueDivisions},
	{Name: "Spin Rewards", Fn: presentation.SeedSpinRewards},
	{Name: "Streak Societies", Fn: presentation.SeedStreakSocieties},
	{Name: "Timed Challenge Templates", Fn: presentation.SeedTimedChallengeTemplates},
	{Name: "Rewards", Fn: presentation.SeedRewards},
	{Name: "Premium Catalog", Fn: presentation.SeedPremiumCatalog},
	{Name: "B2B Plans", Fn: presentation.SeedB2BPlans},
	{Name: "League Seasons", Fn: presentation.SeedLeagueSeasons},
	{Name: "Default Accounts", Fn: presentation.SeedDefaultAccounts},
	{Name: "Presentation Users", Fn: presentation.SeedTestUsers},
	{Name: "Moderation State Users", Fn: presentation.SeedModerationStateUsers},
	{Name: "Inspiring Stories", Fn: presentation.SeedInspiringStories},
	{Name: "Articles", Fn: presentation.SeedArticles},
	{Name: "Songs", Fn: presentation.SeedSongs},
	{Name: "Forums", Fn: presentation.SeedForums},
	{Name: "Chat Sessions", Fn: presentation.SeedChatSessions},
	{Name: "Chat Context Preferences", Fn: presentation.SeedChatContextPreferences},
	{Name: "User Moods", Fn: presentation.SeedUserMoods},
	{Name: "Community Data", Fn: presentation.SeedCommunityData},
	{Name: "Journals", Fn: presentation.SeedJournals},
	{Name: "Breathing Sessions", Fn: presentation.SeedBreathingSessions},
	{Name: "Playlists", Fn: presentation.SeedPlaylists},
	{Name: "Premium Topup Data", Fn: presentation.SeedPremiumAndTopupData},
	{Name: "B2B Organizations", Fn: presentation.SeedB2BOrganizations},
	{Name: "Guilds", Fn: presentation.SeedGuilds},
	{Name: "Gamification", Fn: presentation.SeedGamification},
	{Name: "Presentation Completeness", Fn: presentation.SeedPresentationCompleteness},
	{Name: "Demo Profile", Fn: presentation.SeedDemoProfile},
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
	log.Println("==============================================================")
	log.Println("              RUANG TENANG SEEDER (PRESENTATION)             ")
	log.Println("==============================================================")
	log.Println("")

	return seed.ConnectAndRun(func(db *gorm.DB) error {
		return runPresentationSeeder(db, opts)
	})
}

func runPresentationSeeder(db *gorm.DB, opts seed.SeedOptions) error {
	log.Println("Starting presentation seeding...")
	log.Println("  -> One complete dataset for demo/judging")
	log.Println("  -> Includes catalogs, accounts, content, community, billing, B2B, moderation, and demo state")
	log.Println("")

	if opts.Reset {
		log.Println("--reset enabled: truncating all tables before seeding...")
		if err := resetTablesFn(db); err != nil {
			return err
		}
		log.Println("Database reset complete")
		log.Println("")
	}

	if opts.Count > 0 {
		log.Printf("Using --count=%d (available to seeders via SEED_COUNT env)", opts.Count)
	}

	for _, s := range presentationSeeders {
		if !seed.ShouldRunSeeder(opts.Only, s.Name) {
			continue
		}
		log.Printf("%s...", s.Name)
		if err := s.Fn(db); err != nil {
			log.Printf("  Failed: %v", err)
			return err
		}
		log.Printf("  Done")
	}

	log.Println("")
	log.Println("Presentation seeding completed")
	log.Println("")
	log.Println("Test Accounts:")
	log.Println("   Admin: admin@ruang-tenang.com / password")
	log.Println("   Mitra: mitra@ruang-tenang.com / password")
	log.Println("   User Premium: gading@gmail.com / password")
	log.Println("   User B2B Premium: dery@gmail.com / password")
	log.Println("   User Freemium: andhika@gmail.com / password")
	return nil
}
