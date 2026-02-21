package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/cmd/seed/development"
	"github.com/Alfian57/ruang-tenang-api/cmd/seed/production"
	"gorm.io/gorm"
)

type seederRunner struct {
	name string
	fn   func(*gorm.DB) error
}

var developmentProductionSeeders = []seederRunner{
	{"Level Configs", production.SeedLevelConfigs},
	{"Article Categories", production.SeedArticleCategories},
	{"Song Categories", production.SeedSongCategories},
	{"Forum Categories", production.SeedForumCategories},
	{"Story Categories", production.SeedStoryCategories},
	{"Breathing Techniques", production.SeedBreathingTechniques},
	{"Feature Definitions", production.SeedFeatureDefinitions},
	{"Badge Definitions", production.SeedBadgeDefinitions},
	{"Crisis Keywords", production.SeedCrisisKeywords},
	{"Admin User", production.SeedAdminUser},
}

var developmentTestSeeders = []seederRunner{
	{"Test Users", development.SeedTestUsers},
	{"Articles", development.SeedArticles},
	{"Songs", development.SeedSongs},
	{"Forums", development.SeedForums},
	{"Chat Sessions", development.SeedChatSessions},
	{"User Moods", development.SeedUserMoods},
}

// runDevelopmentSeeder executes the development seeding strategy
func runDevelopmentSeeder(db *gorm.DB, opts SeedOptions) error {
	log.Println("🧪 Starting DEVELOPMENT seeding...")
	log.Println("  → Seeding production data + development test data")
	log.Println("")

	if opts.Reset {
		log.Println("⚠️  --reset enabled: truncating all tables before seeding...")
		if err := resetAllTablesFn(db); err != nil {
			return err
		}
		log.Println("✅ Database reset complete")
		log.Println("")
	}

	if opts.Count > 0 {
		log.Printf("ℹ️  Using --count=%d (available to seeders via SEED_COUNT env)", opts.Count)
	}

	// First, seed production data
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📦 Phase 1: Seeding Production Data")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, s := range developmentProductionSeeders {
		if !shouldRunSeeder(opts.Only, s.name) {
			continue
		}
		log.Printf("  📦 %s...", s.name)
		if err := s.fn(db); err != nil {
			log.Printf("    ❌ Failed: %v", err)
			return err
		}
		log.Printf("    ✅ Done")
	}

	// Then, seed development test data
	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🧪 Phase 2: Seeding Development Test Data")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, s := range developmentTestSeeders {
		if !shouldRunSeeder(opts.Only, s.name) {
			continue
		}
		log.Printf("  🧪 %s...", s.name)
		if err := s.fn(db); err != nil {
			log.Printf("    ❌ Failed: %v", err)
			return err
		}
		log.Printf("    ✅ Done")
	}

	log.Println("")
	log.Println("🎉 Development seeding completed!")
	return nil
}
