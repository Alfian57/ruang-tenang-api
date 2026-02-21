package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Alfian57/ruang-tenang-api/cmd/seed/production"
	"gorm.io/gorm"
)

var productionSeeders = []seederRunner{
	{"Level Configs", production.SeedLevelConfigs},
	{"Article Categories", production.SeedArticleCategories},
	{"Song Categories", production.SeedSongCategories},
	{"Forum Categories", production.SeedForumCategories},
	{"Story Categories", production.SeedStoryCategories},
	{"Breathing Techniques", production.SeedBreathingTechniques},
	{"Feature Definitions", production.SeedFeatureDefinitions},
	{"Badge Definitions", production.SeedBadgeDefinitions},
	{"Crisis Keywords", production.SeedCrisisKeywords},
}

// runProductionSeeder executes the production seeding strategy
func runProductionSeeder(db *gorm.DB, opts SeedOptions) error {
	log.Println("🏭 Starting PRODUCTION seeding...")
	log.Println("  → Seeding essential application data only")
	log.Println("")

	if opts.Count > 0 {
		log.Printf("ℹ️  --count=%d ignored in production mode", opts.Count)
	}

	// Seed in order of dependencies
	seeders := append([]seederRunner{}, productionSeeders...)

	if os.Getenv("SEED_PROD_CREATE_ADMIN") == "true" {
		if os.Getenv("SEED_ADMIN_EMAIL") == "" || os.Getenv("SEED_ADMIN_PASSWORD") == "" {
			return fmt.Errorf("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are required when SEED_PROD_CREATE_ADMIN=true")
		}
		seeders = append(seeders, seederRunner{"Admin User", production.SeedAdminUser})
	}

	for _, s := range seeders {
		if !shouldRunSeeder(opts.Only, s.name) {
			continue
		}
		log.Printf("📦 %s...", s.name)
		if err := s.fn(db); err != nil {
			log.Printf("  ❌ Failed: %v", err)
			return err
		}
		log.Printf("  ✅ Done")
	}

	log.Println("")
	log.Println("🎉 Production seeding completed!")
	return nil
}
