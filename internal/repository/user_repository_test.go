package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserRepositoryTest(t *testing.T) (*UserRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}

	repo := NewUserRepository(db)
	now := time.Now()
	seed := []model.User{
		{
			Name:             "User One",
			Username:         "userone",
			Email:            "user1@test.local",
			Password:         "x",
			Exp:              300,
			ResetToken:       "token-valid",
			ResetTokenExpiry: now.Add(2 * time.Hour),
		},
		{
			Name:             "User Two",
			Username:         "usertwo",
			Email:            "user2@test.local",
			Password:         "x",
			Exp:              100,
			ResetToken:       "token-expired",
			ResetTokenExpiry: now.Add(-2 * time.Hour),
		},
	}
	for i := range seed {
		if err := repo.Create(context.Background(), &seed[i]); err != nil {
			t.Fatalf("seed user %d: %v", i+1, err)
		}
	}

	return repo, db
}

func TestUserRepository_Branches(t *testing.T) {
	repo, _ := setupUserRepositoryTest(t)
	ctx := context.Background()

	u1, err := repo.FindByEmail(ctx, "user1@test.local")
	if err != nil {
		t.Fatalf("FindByEmail existing: %v", err)
	}

	if _, err := repo.FindByID(ctx, u1.ID); err != nil {
		t.Fatalf("FindByID existing: %v", err)
	}
	if _, err := repo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected FindByID missing error")
	}

	if _, err := repo.FindByEmail(ctx, "missing@test.local"); err == nil {
		t.Fatal("expected FindByEmail missing error")
	}

	if _, err := repo.FindByUsername(ctx, "userone"); err != nil {
		t.Fatalf("FindByUsername existing: %v", err)
	}
	if _, err := repo.FindByUsername(ctx, "missing-user"); err == nil {
		t.Fatal("expected FindByUsername missing error")
	}

	top, err := repo.GetTopUsers(ctx, 1)
	if err != nil {
		t.Fatalf("GetTopUsers: %v", err)
	}
	if len(top) != 1 || top[0].Email != "user1@test.local" {
		t.Fatalf("unexpected top users result: %+v", top)
	}

	if !repo.ExistsByEmail(ctx, "user1@test.local") {
		t.Fatal("expected ExistsByEmail true")
	}
	if repo.ExistsByEmail(ctx, "nope@test.local") {
		t.Fatal("expected ExistsByEmail false")
	}

	if !repo.ExistsByEmailExcept(ctx, "user2@test.local", u1.ID) {
		t.Fatal("expected ExistsByEmailExcept true")
	}
	if repo.ExistsByEmailExcept(ctx, "user1@test.local", u1.ID) {
		t.Fatal("expected ExistsByEmailExcept false for same user")
	}

	if _, err := repo.FindByResetToken(ctx, "token-valid"); err != nil {
		t.Fatalf("FindByResetToken valid: %v", err)
	}
	if _, err := repo.FindByResetToken(ctx, "token-expired"); err == nil {
		t.Fatal("expected FindByResetToken expired error")
	}

	u1.Name = "User One Updated"
	if err := repo.Update(ctx, u1); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := repo.UpdateResetToken(ctx, "user2@test.local", "token-new", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("UpdateResetToken: %v", err)
	}
	if _, err := repo.FindByResetToken(ctx, "token-new"); err != nil {
		t.Fatalf("FindByResetToken token-new: %v", err)
	}

	u2, err := repo.FindByEmail(ctx, "user2@test.local")
	if err != nil {
		t.Fatalf("FindByEmail user2: %v", err)
	}
	if err := repo.ClearResetToken(ctx, u2.ID); err != nil {
		t.Fatalf("ClearResetToken: %v", err)
	}
	if _, err := repo.FindByResetToken(ctx, "token-new"); err == nil {
		t.Fatal("expected FindByResetToken missing after clear")
	}

	if err := repo.Delete(ctx, u1.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, u1.ID); err == nil {
		t.Fatal("expected FindByID error after delete")
	}
}

func TestUserRepository_GetTopUsers_ErrorBranch(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)
	ctx := context.Background()

	if err := db.Exec(`DROP TABLE users`).Error; err != nil {
		t.Fatalf("drop users table: %v", err)
	}

	if _, err := repo.GetTopUsers(ctx, 5); err == nil {
		t.Fatal("expected GetTopUsers error when users table is missing")
	}
}
