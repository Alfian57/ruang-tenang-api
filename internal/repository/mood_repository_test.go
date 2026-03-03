package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMoodRepositoryTest(t *testing.T) (*UserMoodRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.UserMood{}); err != nil {
		t.Fatalf("migrate user_moods: %v", err)
	}

	repo := NewUserMoodRepository(db)
	now := time.Now()
	seed := []model.UserMood{
		{UserID: 1, Mood: model.MoodHappy, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour)},
		{UserID: 1, Mood: model.MoodSad, CreatedAt: now.Add(-26 * time.Hour), UpdatedAt: now.Add(-26 * time.Hour)},
		{UserID: 2, Mood: model.MoodNeutral, CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
	}
	for i := range seed {
		if err := repo.Create(context.Background(), &seed[i]); err != nil {
			t.Fatalf("seed mood %d: %v", i+1, err)
		}
	}

	return repo, db
}

func TestMoodRepository_Branches(t *testing.T) {
	repo, db := setupMoodRepositoryTest(t)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-48 * time.Hour)
	end := now
	moods, total, err := repo.FindByUserID(ctx, 1, &start, &end, 1, 10)
	if err != nil || total == 0 || len(moods) == 0 {
		t.Fatalf("FindByUserID unexpected err=%v total=%d len=%d", err, total, len(moods))
	}

	latest, err := repo.GetLatestByUserID(ctx, 1)
	if err != nil || latest.UserID != 1 {
		t.Fatalf("GetLatestByUserID unexpected mood=%+v err=%v", latest, err)
	}
	if _, err := repo.GetLatestByUserID(ctx, 999); err == nil {
		t.Fatal("expected GetLatestByUserID missing error")
	}

	stats, err := repo.GetMoodStats(ctx, 1, 30)
	if err != nil {
		t.Fatalf("GetMoodStats: %v", err)
	}
	if stats[string(model.MoodHappy)] == 0 {
		t.Fatalf("expected happy stats present, got %+v", stats)
	}

	todayMood, err := repo.FindTodayByUserID(ctx, 1)
	if err != nil || todayMood.UserID != 1 {
		t.Fatalf("FindTodayByUserID unexpected mood=%+v err=%v", todayMood, err)
	}
	if _, err := repo.FindTodayByUserID(ctx, 999); err == nil {
		t.Fatal("expected FindTodayByUserID missing error")
	}

	todayMood.Mood = model.MoodNeutral
	if err := repo.Update(ctx, todayMood); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := repo.GetLatestByUserID(ctx, 1)
	if err != nil || updated.Mood != model.MoodNeutral {
		t.Fatalf("expected updated mood neutral, got mood=%v err=%v", updated.Mood, err)
	}

	if err := db.Exec(`DROP TABLE user_moods`).Error; err != nil {
		t.Fatalf("drop user_moods: %v", err)
	}
	if _, err := repo.GetMoodStats(ctx, 1, 7); err == nil {
		t.Fatal("expected GetMoodStats error when table missing")
	}
}
