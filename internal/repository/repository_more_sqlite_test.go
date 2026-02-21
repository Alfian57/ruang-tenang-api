package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepoSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}, &model.UserMood{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestUserRepository_CRUDFindersAndExistence(t *testing.T) {
	ctx := context.Background()
	r := NewUserRepository(newRepoSQLiteDB(t))

	u := &model.User{Name: "User A", Username: "usera", Email: "usera@example.com", Password: "secret"}
	if err := r.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}

	if got, err := r.FindByID(ctx, u.ID); err != nil || got.Email != u.Email {
		t.Fatalf("find by id failed: got=%+v err=%v", got, err)
	}
	if got, err := r.FindByEmail(ctx, u.Email); err != nil || got.ID != u.ID {
		t.Fatalf("find by email failed: got=%+v err=%v", got, err)
	}
	if got, err := r.FindByUsername(ctx, u.Username); err != nil || got.ID != u.ID {
		t.Fatalf("find by username failed: got=%+v err=%v", got, err)
	}

	u.Name = "User A Updated"
	if err := r.Update(ctx, u); err != nil {
		t.Fatalf("update: %v", err)
	}

	if !r.ExistsByEmail(ctx, u.Email) {
		t.Fatalf("expected email exists")
	}

	u2 := &model.User{Name: "User B", Username: "userb", Email: "userb@example.com", Password: "secret", Exp: 200}
	if err := r.Create(ctx, u2); err != nil {
		t.Fatalf("create second user: %v", err)
	}

	if !r.ExistsByEmailExcept(ctx, u.Email, u2.ID) {
		t.Fatalf("expected email exists except other user")
	}

	top, err := r.GetTopUsers(ctx, 1)
	if err != nil || len(top) != 1 || top[0].ID != u2.ID {
		t.Fatalf("get top users failed: users=%+v err=%v", top, err)
	}

	if err := r.UpdateResetToken(ctx, u.Email, "tok", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("update reset token: %v", err)
	}
	if _, err := r.FindByResetToken(ctx, "tok"); err != nil {
		t.Fatalf("find by reset token: %v", err)
	}
	if err := r.ClearResetToken(ctx, u.ID); err != nil {
		t.Fatalf("clear reset token: %v", err)
	}

	if err := r.Delete(ctx, u.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
}

func TestLevelConfigRepository_FullFlow(t *testing.T) {
	ctx := context.Background()
	r := NewLevelConfigRepository(newRepoSQLiteDB(t))

	l1 := &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱", TierName: "Bronze", TierColor: "#aaa"}
	l2 := &model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐", TierName: "Silver", TierColor: "#bbb"}
	if err := r.Create(ctx, l1); err != nil {
		t.Fatalf("create l1: %v", err)
	}
	if err := r.Create(ctx, l2); err != nil {
		t.Fatalf("create l2: %v", err)
	}

	all, err := r.GetAll(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("get all failed: len=%d err=%v", len(all), err)
	}

	if got, err := r.GetByID(ctx, l1.ID); err != nil || got.Level != 1 {
		t.Fatalf("get by id failed: got=%+v err=%v", got, err)
	}
	if got, err := r.GetByLevel(ctx, 2); err != nil || got.MinExp != 100 {
		t.Fatalf("get by level failed: got=%+v err=%v", got, err)
	}
	if got, err := r.GetLevelByExp(ctx, 150); err != nil || got.Level != 2 {
		t.Fatalf("get level by exp failed: got=%+v err=%v", got, err)
	}
	if got, err := r.GetNextLevel(ctx, 1); err != nil || got.Level != 2 {
		t.Fatalf("get next level failed: got=%+v err=%v", got, err)
	}

	if !r.ExistsByLevel(ctx, 1) || !r.ExistsByLevelExcept(ctx, 1, l2.ID) {
		t.Fatalf("expected exists checks true")
	}

	l1.BadgeName = "Pemula+"
	if err := r.Update(ctx, l1); err != nil {
		t.Fatalf("update level config: %v", err)
	}

	count, err := r.Count(ctx)
	if err != nil || count != 2 {
		t.Fatalf("count failed: count=%d err=%v", count, err)
	}

	if err := r.Delete(ctx, l2.ID); err != nil {
		t.Fatalf("delete level config: %v", err)
	}
}

func TestUserMoodRepository_FullFlow(t *testing.T) {
	ctx := context.Background()
	db := newRepoSQLiteDB(t)
	ur := NewUserRepository(db)
	mr := NewUserMoodRepository(db)

	u := &model.User{Name: "Mood User", Username: "mooduser", Email: "mood@example.com", Password: "secret"}
	if err := ur.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	m1 := &model.UserMood{UserID: u.ID, Mood: model.MoodHappy}
	m2 := &model.UserMood{UserID: u.ID, Mood: model.MoodSad}
	if err := mr.Create(ctx, m1); err != nil {
		t.Fatalf("create mood1: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := mr.Create(ctx, m2); err != nil {
		t.Fatalf("create mood2: %v", err)
	}

	start := time.Now().AddDate(0, 0, -1)
	end := time.Now().AddDate(0, 0, 1)
	moods, total, err := mr.FindByUserID(ctx, u.ID, &start, &end, 1, 10)
	if err != nil || total < 2 || len(moods) < 2 {
		t.Fatalf("find by user failed: total=%d len=%d err=%v", total, len(moods), err)
	}

	latest, err := mr.GetLatestByUserID(ctx, u.ID)
	if err != nil || latest.ID != m2.ID {
		t.Fatalf("latest mood failed: got=%+v err=%v", latest, err)
	}

	today, err := mr.FindTodayByUserID(ctx, u.ID)
	if err != nil || today == nil {
		t.Fatalf("today mood failed: mood=%+v err=%v", today, err)
	}

	m2.Mood = model.MoodNeutral
	if err := mr.Update(ctx, m2); err != nil {
		t.Fatalf("update mood: %v", err)
	}

	stats, err := mr.GetMoodStats(ctx, u.ID, 7)
	if err != nil {
		t.Fatalf("mood stats failed: %v", err)
	}
	if len(stats) == 0 {
		t.Fatalf("expected non-empty mood stats")
	}
}
