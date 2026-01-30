package production

import (
	"log"

	"gorm.io/gorm"
)

// SeedAll runs all production seeders (essential data only)
func SeedAll(db *gorm.DB) error {
	log.Println("🏭 Starting PRODUCTION seeding...")
	log.Println("  → Seeding essential application data only")
	log.Println("")

	// Seed in order of dependencies
	seeders := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"Level Configs", SeedLevelConfigs},
		{"Article Categories", SeedArticleCategories},
		{"Song Categories", SeedSongCategories},
		{"Forum Categories", SeedForumCategories},
		{"Story Categories", SeedStoryCategories},
		{"Breathing Techniques", SeedBreathingTechniques},
		{"Feature Definitions", SeedFeatureDefinitions},
		{"Badge Definitions", SeedBadgeDefinitions},
		{"Crisis Keywords", SeedCrisisKeywords},
		{"Admin User", SeedAdminUser},
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
