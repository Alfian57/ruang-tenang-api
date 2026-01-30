package development

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/cmd/seeder/production"
	"gorm.io/gorm"
)

// SeedAll runs all development seeders (includes production data + test data)
func SeedAll(db *gorm.DB) error {
	log.Println("🧪 Starting DEVELOPMENT seeding...")
	log.Println("  → Seeding production data + development test data")
	log.Println("")

	// First, seed production data
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📦 Phase 1: Seeding Production Data")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	productionSeeders := []struct {
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

	for _, s := range productionSeeders {
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

	developmentSeeders := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"Test Users", SeedTestUsers},
		{"Articles", SeedArticles},
		{"Songs", SeedSongs},
		{"Forums", SeedForums},
		{"Chat Sessions", SeedChatSessions},
		{"User Moods", SeedUserMoods},
	}

	for _, s := range developmentSeeders {
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
