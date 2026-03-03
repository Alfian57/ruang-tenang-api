package development

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCommunitySeedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.ForumCategory{},
		&model.Forum{},
		&model.ForumPost{},
		&model.ForumLike{},
		&model.UserMood{},
	); err != nil {
		t.Fatalf("automigrate error: %v", err)
	}

	schema := []string{
		`CREATE TABLE chat_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT,
			user_id INTEGER,
			folder_id INTEGER,
			title TEXT,
			summary TEXT,
			summary_generated_at DATETIME,
			is_favorite BOOLEAN,
			is_trash BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT,
			chat_session_id INTEGER,
			role TEXT,
			content TEXT,
			type TEXT,
			is_liked BOOLEAN,
			is_disliked BOOLEAN,
			is_pinned BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}

	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}
	return db
}

func seedCommunityBaseData(t *testing.T, db *gorm.DB) {
	t.Helper()
	users := []model.User{
		{Name: "John", Username: "john", Email: "john@example.com", Password: "x", Role: model.RoleMember},
		{Name: "Member 2", Username: "member2", Email: "member2@example.com", Password: "x", Role: model.RoleMember},
		{Name: "Member 3", Username: "member3", Email: "member3@example.com", Password: "x", Role: model.RoleMember},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	categories := []model.ForumCategory{
		{Name: "Kesehatan Mental di Tempat Kerja"},
		{Name: "Kisah Inspiratif"},
		{Name: "Curhat & Keluh Kesah"},
		{Name: "Diskusi Umum"},
		{Name: "Tips Mengelola Stres"},
		{Name: "Pertanyaan & Jawaban"},
		{Name: "Kesehatan Mental di Sekolah"},
	}
	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			t.Fatalf("seed forum category: %v", err)
		}
	}
}

func TestDevelopmentSeeders_ChatForumMood(t *testing.T) {
	db := setupCommunitySeedDB(t)
	seedCommunityBaseData(t, db)

	if err := SeedChatSessions(db); err != nil {
		t.Fatalf("SeedChatSessions failed: %v", err)
	}
	var sessionCount int64
	if err := db.Model(&model.ChatSession{}).Count(&sessionCount).Error; err != nil {
		t.Fatalf("count chat sessions: %v", err)
	}
	if sessionCount == 0 {
		t.Fatal("expected chat sessions to be seeded")
	}

	if err := SeedForums(db); err != nil {
		t.Fatalf("SeedForums failed: %v", err)
	}
	var forumCount int64
	if err := db.Model(&model.Forum{}).Count(&forumCount).Error; err != nil {
		t.Fatalf("count forums: %v", err)
	}
	if forumCount == 0 {
		t.Fatal("expected forums to be seeded")
	}

	if err := SeedUserMoods(db); err != nil {
		t.Fatalf("SeedUserMoods failed: %v", err)
	}
	var moodCount int64
	if err := db.Model(&model.UserMood{}).Count(&moodCount).Error; err != nil {
		t.Fatalf("count user moods: %v", err)
	}
	if moodCount == 0 {
		t.Fatal("expected user moods to be seeded")
	}
}

func TestSeedChatSessions_NoUserAndCreateErrorBranches(t *testing.T) {
	t.Run("no-user-skip", func(t *testing.T) {
		db := setupCommunitySeedDB(t)
		if err := SeedChatSessions(db); err != nil {
			t.Fatalf("expected nil when john user missing, got %v", err)
		}
	})

	t.Run("create-error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := db.AutoMigrate(&model.User{}); err != nil {
			t.Fatalf("migrate users: %v", err)
		}
		if err := db.Create(&model.User{Name: "John", Username: "john", Email: "john@example.com", Password: "x", Role: model.RoleMember}).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}

		if err := SeedChatSessions(db); err == nil {
			t.Fatal("expected error when chat tables are missing")
		}
	})
}

func TestSeedSongs_MissingThumbnailAndAudioBranches(t *testing.T) {
	withTempWorkingDir(t)
	db := setupSeedDevDB(t)

	categories := []model.SongCategory{{Name: "Alam"}, {Name: "Piano"}, {Name: "Hujan"}, {Name: "Laut"}, {Name: "Meditasi"}}
	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			t.Fatalf("seed category: %v", err)
		}
	}

	if err := SeedSongs(db); err == nil {
		t.Fatal("expected missing thumbnail error")
	}

	createDummySeedAssets(t)
	if err := os.Remove(filepath.Join("assets", "audio", "song-1.mp3")); err != nil {
		t.Fatalf("remove sample audio: %v", err)
	}

	if err := SeedSongs(db); err == nil {
		t.Fatal("expected missing audio error")
	}
}

func TestSeedSongs_IdempotentUpdateExistingBranch(t *testing.T) {
	withTempWorkingDir(t)
	createDummySeedAssets(t)
	db := setupSeedDevDB(t)

	categories := []model.SongCategory{{Name: "Alam"}, {Name: "Piano"}, {Name: "Hujan"}, {Name: "Laut"}, {Name: "Meditasi"}}
	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			t.Fatalf("seed category: %v", err)
		}
	}

	if err := SeedSongs(db); err != nil {
		t.Fatalf("first SeedSongs failed: %v", err)
	}

	var firstCount int64
	if err := db.Model(&model.Song{}).Count(&firstCount).Error; err != nil {
		t.Fatalf("count songs after first run: %v", err)
	}
	if firstCount != 15 {
		t.Fatalf("expected 15 songs after first run, got %d", firstCount)
	}

	var before model.Song
	if err := db.Where("title = ?", "Forest Birds Morning").First(&before).Error; err != nil {
		t.Fatalf("query song before second run: %v", err)
	}

	if err := SeedSongs(db); err != nil {
		t.Fatalf("second SeedSongs failed: %v", err)
	}

	var secondCount int64
	if err := db.Model(&model.Song{}).Count(&secondCount).Error; err != nil {
		t.Fatalf("count songs after second run: %v", err)
	}
	if secondCount != firstCount {
		t.Fatalf("expected song count unchanged after second run, got %d (before %d)", secondCount, firstCount)
	}

	var after model.Song
	if err := db.Where("title = ?", "Forest Birds Morning").First(&after).Error; err != nil {
		t.Fatalf("query song after second run: %v", err)
	}
	if after.FilePath == "" || after.Thumbnail == "" {
		t.Fatalf("expected populated file path and thumbnail, got %+v", after)
	}
	if after.FilePath == before.FilePath && after.Thumbnail == before.Thumbnail {
		t.Fatalf("expected existing song to be updated on second run")
	}
}
