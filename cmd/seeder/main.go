package main

import (
	"fmt"
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
)

func main() {
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

	log.Println("🔥 Dropping all tables to clear data...")
	migrator := db.Migrator()
	if err := migrator.DropTable(
		&models.UserMood{},
		&models.ChatMessage{},
		&models.ChatSession{},
		&models.Song{},
		&models.SongCategory{},
		&models.Article{},
		&models.ArticleCategory{},
		&models.User{},
		&models.Forum{},
		&models.ForumCategory{},
		&models.ForumPost{},
		&models.ForumLike{},
		&models.UserActivity{},
		&models.LevelConfig{},
		&models.ExpHistory{},
	); err != nil {
		log.Printf("⚠️ Failed to drop tables (might not exist): %v", err)
	}

	log.Println("🔄 Running migrations (AutoMigrate)...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.ArticleCategory{},
		&models.Article{},
		&models.SongCategory{},
		&models.Song{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.UserMood{},
		&models.ForumCategory{},
		&models.Forum{},
		&models.ForumPost{},
		&models.ForumLike{},
		&models.UserActivity{},
		&models.LevelConfig{},
		&models.ExpHistory{},
	); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	log.Println("🌱 Starting database seeder...")

	seedLevels(db)
	seedUsers(db)
	seedArticles(db)
	seedSongs(db)
	seedForums(db)
	seedChats(db)
	seedActivity(db)

	fmt.Println("\n✅ Database seeding completed!")
	fmt.Println("\n📋 Test Accounts:")
	fmt.Println("   Admin: admin@ruangtenang.id / admin123")
	fmt.Println("   Member: john@example.com / member123")
	fmt.Println("\n📊 Level Configurations:")
	fmt.Println("   Level 1: 0 EXP - Beginner 🌱")
	fmt.Println("   Level 2: 100 EXP - Explorer 🌿")
	fmt.Println("   Level 3: 300 EXP - Learner 📚")
	fmt.Println("   Level 4: 600 EXP - Intermediate 🌳")
	fmt.Println("   Level 5: 1000 EXP - Advanced 🏆")
	fmt.Println("   Level 6: 1500 EXP - Expert 💎")
	fmt.Println("   Level 7: 2000 EXP - Master ⭐")
	fmt.Println("   Level 8: 3000 EXP - Grandmaster 👑")
}
