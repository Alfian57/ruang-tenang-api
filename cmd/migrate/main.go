package main

import (
	"flag"
	"log"
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
)

func main() {
	action := flag.String("action", "up", "Migration action: up (default) or down")
	force := flag.Bool("force", false, "Force action (required for down)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	migrator := db.Migrator()

	modelsList := []interface{}{
		&models.User{},
		&models.UserActivity{},
		&models.UserMood{},
		&models.UserBlock{},
		&models.UserReport{},
		&models.UserStrike{},
		&models.LevelConfig{},
		&models.ExpHistory{},
		&models.BadgeDefinition{},
		&models.UserBadge{},
		&models.ArticleCategory{},
		&models.Article{},
		&models.SongCategory{},
		&models.Song{},
		&models.Playlist{},
		&models.PlaylistItem{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.ChatFolder{},
		&models.Forum{},
		&models.ForumCategory{},
		&models.ForumPost{},
		&models.ForumLike{},
		&models.ForumPostVote{},
		&models.ForumPostReport{},
		&models.StoryCategory{},
		&models.InspiringStory{},
		&models.StoryTag{},
		&models.StoryHeart{},
		&models.StoryComment{},
		&models.StoryCommentHeart{},
		&models.BreathingTechnique{},
		&models.BreathingSession{},
		&models.BreathingPreference{},
		&models.BreathingFavorite{},
		&models.Journal{},
		&models.JournalSettings{},
		&models.JournalAIAccessLog{},
		&models.DailyTaskConfig{},
		&models.DailyTask{},
		&models.DailyTaskSummary{},
		&models.FeatureDefinition{},
		&models.UserFeatureUnlock{},
		&models.CrisisKeyword{},
		&models.CrisisDetectionResult{},
		&models.ContentFlag{},
		&models.ModeratorAction{},
	}

	switch *action {
	case "up":
		log.Println("🔄 Running migrations (AutoMigrate)...")
		if err := db.AutoMigrate(modelsList...); err != nil {
			log.Fatalf("❌ Failed to migrate database: %v", err)
		}
		log.Println("✅ Migration completed successfully!")

	case "down":
		if !*force {
			log.Println("⚠️  Warning: 'down' action requires -force flag to execute.")
			log.Println("   Use: ./migrate -action=down -force")
			os.Exit(1)
		}

		log.Println("🔥 Dropping all tables...")
		if err := migrator.DropTable(modelsList...); err != nil {
			log.Printf("⚠️ Failed to drop some tables: %v", err)
		}
		log.Println("✅ All tables dropped successfully!")

	default:
		log.Fatalf("❌ Unknown action: %s. Use 'up' or 'down'", *action)
	}
}
