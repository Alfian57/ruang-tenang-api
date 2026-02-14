package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/cmd/seed/production"
	"gorm.io/gorm"
)

// runProductionSeeder executes the production seeding strategy
func runProductionSeeder(db *gorm.DB) error {
	log.Println("🏭 Starting PRODUCTION seeding...")
	log.Println("  → Seeding essential application data only")
	log.Println("")

	// Seed in order of dependencies
	seeders := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
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

	for _, s := range seeders {
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
