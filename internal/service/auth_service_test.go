package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAuthServiceForTest(t *testing.T) (*AuthService, *repository.UserRepository) {
	t.Helper()
	config.AppConfig = &config.Config{JWTSecret: "test-secret", JWTExpiryHours: 24}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	repo := repository.NewUserRepository(db)
	return NewAuthService(repo), repo
}

func TestAuthService_RegisterAndLoginBranches(t *testing.T) {
	ctx := context.Background()
	svc, repo := newAuthServiceForTest(t)

	if _, err := svc.Register(ctx, &dto.RegisterRequest{Name: "A", Email: "a@test.local", Password: "secret123"}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if _, err := svc.Register(ctx, &dto.RegisterRequest{Name: "A2", Email: "a@test.local", Password: "secret123"}); err == nil {
		t.Fatal("expected duplicate email register error")
	}

	if _, err := svc.Login(ctx, &dto.LoginRequest{Email: "a@test.local", Password: "wrong"}); err == nil {
		t.Fatal("expected invalid password error")
	}

	user, err := repo.FindByEmail(ctx, "a@test.local")
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	user.IsBlocked = true
	_ = repo.Update(ctx, user)
	if _, err := svc.Login(ctx, &dto.LoginRequest{Email: "a@test.local", Password: "secret123"}); err == nil {
		t.Fatal("expected blocked user login error")
	}

	user.IsBlocked = false
	_ = repo.Update(ctx, user)
	resp, err := svc.Login(ctx, &dto.LoginRequest{Email: "a@test.local", Password: "secret123", RememberMe: true})
	if err != nil {
		t.Fatalf("login should succeed: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestAuthService_ProfilePasswordAndResetFlows(t *testing.T) {
	ctx := context.Background()
	svc, repo := newAuthServiceForTest(t)

	hash1, _ := utils.HashPassword("secret123")
	hash2, _ := utils.HashPassword("secret999")
	u1 := &model.User{Name: "User One", Email: "one@test.local", Password: hash1}
	u2 := &model.User{Name: "User Two", Email: "two@test.local", Password: hash2}
	_ = repo.Create(ctx, u1)
	_ = repo.Create(ctx, u2)

	profile, err := svc.GetProfile(ctx, u1.ID)
	if err != nil {
		t.Fatalf("expected get profile success: %v", err)
	}
	if profile.Email != "one@test.local" {
		t.Fatalf("unexpected profile email: %s", profile.Email)
	}
	if _, err := svc.GetProfile(ctx, 99999); err == nil {
		t.Fatal("expected get profile not found error")
	}

	if _, err := svc.UpdateProfile(ctx, 99999, &dto.UpdateProfileRequest{Name: "N", Email: "x@test.local"}); err == nil {
		t.Fatal("expected user not found error")
	}
	if _, err := svc.UpdateProfile(ctx, u1.ID, &dto.UpdateProfileRequest{Name: "New", Email: "two@test.local"}); err == nil {
		t.Fatal("expected email already taken error")
	}
	updatedProfile, err := svc.UpdateProfile(ctx, u1.ID, &dto.UpdateProfileRequest{Name: "User One Updated", Email: "one.new@test.local", Avatar: "avatar.png"})
	if err != nil {
		t.Fatalf("expected update profile success: %v", err)
	}
	if updatedProfile.Name != "User One Updated" || updatedProfile.Email != "one.new@test.local" || updatedProfile.Avatar != "avatar.png" {
		t.Fatalf("unexpected updated profile: %+v", updatedProfile)
	}

	if err := svc.UpdatePassword(ctx, u1.ID, &dto.UpdatePasswordRequest{CurrentPassword: "wrong", NewPassword: "newsecret123"}); err == nil {
		t.Fatal("expected wrong current password error")
	}
	if err := svc.UpdatePassword(ctx, u1.ID, &dto.UpdatePasswordRequest{CurrentPassword: "secret123", NewPassword: "newsecret123"}); err != nil {
		t.Fatalf("expected password update success: %v", err)
	}

	if err := svc.ForgotPassword(ctx, &dto.ForgotPasswordRequest{Email: "missing@test.local"}); err != nil {
		t.Fatalf("forgot password should ignore missing email: %v", err)
	}
	if err := svc.ForgotPassword(ctx, &dto.ForgotPasswordRequest{Email: "one.new@test.local"}); err != nil {
		t.Fatalf("forgot password existing user failed: %v", err)
	}

	updated, err := repo.FindByEmail(ctx, "one.new@test.local")
	if err != nil {
		t.Fatalf("find updated user: %v", err)
	}
	if updated.ResetToken == "" {
		t.Fatal("expected reset token to be set")
	}

	if err := svc.ResetPassword(ctx, &dto.ResetPasswordRequest{Token: "invalid-token", NewPassword: "abc12345"}); err == nil {
		t.Fatal("expected invalid token error")
	}
	if err := svc.ResetPassword(ctx, &dto.ResetPasswordRequest{Token: updated.ResetToken, NewPassword: "abc12345"}); err != nil {
		t.Fatalf("expected reset password success: %v", err)
	}

	finalUser, _ := repo.FindByID(ctx, updated.ID)
	if finalUser.ResetToken != "" {
		t.Fatal("expected reset token to be cleared")
	}
}

func TestAuthService_UpdateProfile_UpdateError(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	repo := repository.NewUserRepository(db)
	svc := NewAuthService(repo)

	user := &model.User{Name: "User One", Email: "one@test.local", Password: "hashed"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	if err := db.Migrator().DropTable(&model.User{}); err != nil {
		t.Fatalf("drop users table failed: %v", err)
	}

	if _, err := svc.UpdateProfile(ctx, user.ID, &dto.UpdateProfileRequest{Name: "Updated", Email: "updated@test.local"}); err == nil {
		t.Fatal("expected failed to update profile error")
	}
}

func TestAuthService_UpdatePassword_AdditionalBranches(t *testing.T) {
	ctx := context.Background()
	svc, repo := newAuthServiceForTest(t)

	if err := svc.UpdatePassword(ctx, 99999, &dto.UpdatePasswordRequest{CurrentPassword: "x", NewPassword: "newsecret123"}); err == nil {
		t.Fatal("expected user not found error")
	}

	hash, _ := utils.HashPassword("secret123")
	u := &model.User{Name: "User", Email: "pw@test.local", Password: hash}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	repo2 := repository.NewUserRepository(db)
	if err := repo2.Create(ctx, &model.User{Name: "User2", Email: "pw2@test.local", Password: hash}); err != nil {
		t.Fatalf("seed user2 failed: %v", err)
	}
	if err := db.Migrator().DropTable(&model.User{}); err != nil {
		t.Fatalf("drop users table failed: %v", err)
	}

	svc2 := NewAuthService(repo2)
	if err := svc2.UpdatePassword(ctx, 1, &dto.UpdatePasswordRequest{CurrentPassword: "secret123", NewPassword: "newsecret123"}); err == nil {
		t.Fatal("expected failed to update password error")
	}
}

func TestAuthService_Register_AdditionalErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("hash password error for too-long input", func(t *testing.T) {
		svc, _ := newAuthServiceForTest(t)
		tooLongPassword := strings.Repeat("a", 100)
		if _, err := svc.Register(ctx, &dto.RegisterRequest{Name: "Too Long", Email: "long@test.local", Password: tooLongPassword}); err == nil {
			t.Fatal("expected failed to hash password error")
		}
	})

	t.Run("create user error when users table missing", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := db.AutoMigrate(&model.User{}); err != nil {
			t.Fatalf("migrate users: %v", err)
		}

		repo := repository.NewUserRepository(db)
		svc := NewAuthService(repo)

		if err := db.Migrator().DropTable(&model.User{}); err != nil {
			t.Fatalf("drop users table: %v", err)
		}

		if _, err := svc.Register(ctx, &dto.RegisterRequest{Name: "No Table", Email: "notable@test.local", Password: "secret123"}); err == nil {
			t.Fatal("expected failed to create user error")
		}
	})
}

func TestAuthService_ResetPassword_ClearTokenFailureIgnored(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	repo := repository.NewUserRepository(db)
	svc := NewAuthService(repo)

	hash, _ := utils.HashPassword("secret123")
	u := &model.User{Name: "User", Email: "reset-clear@test.local", Password: hash, ResetToken: "tok-clear", ResetTokenExpiry: time.Now().Add(time.Hour)}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	if err := db.Exec(`
		CREATE TRIGGER fail_clear_reset_token
		BEFORE UPDATE OF reset_token ON users
		WHEN NEW.reset_token IS NULL
		BEGIN
			SELECT RAISE(FAIL, 'fail clear reset token');
		END;
	`).Error; err != nil {
		t.Fatalf("create trigger failed: %v", err)
	}

	if err := svc.ResetPassword(ctx, &dto.ResetPasswordRequest{Token: "tok-clear", NewPassword: "newsecret123"}); err != nil {
		t.Fatalf("expected ResetPassword success when clear token fails, got %v", err)
	}

	updated, err := repo.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("find updated user failed: %v", err)
	}
	if updated.ResetToken == "" {
		t.Fatal("expected reset token to remain when clear step fails")
	}
}

func TestAuthService_AdditionalFailureBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("update password hash failure", func(t *testing.T) {
		svc, repo := newAuthServiceForTest(t)
		hash, _ := utils.HashPassword("secret123")
		u := &model.User{Name: "User", Email: "upd-hash@test.local", Password: hash}
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("seed user: %v", err)
		}

		err := svc.UpdatePassword(ctx, u.ID, &dto.UpdatePasswordRequest{CurrentPassword: "secret123", NewPassword: strings.Repeat("x", 100)})
		if err == nil || !strings.Contains(err.Error(), "failed to hash password") {
			t.Fatalf("expected failed to hash password, got %v", err)
		}
	})

	t.Run("forgot password save reset token failure", func(t *testing.T) {
		_, repo := newAuthServiceForTest(t)
		u := &model.User{Name: "User", Email: "forgot-save@test.local", Password: "hashed"}
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("seed user: %v", err)
		}

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := db.AutoMigrate(&model.User{}); err != nil {
			t.Fatalf("migrate users: %v", err)
		}
		repo2 := repository.NewUserRepository(db)
		svc2 := NewAuthService(repo2)
		if err := repo2.Create(ctx, &model.User{Name: "User2", Email: "forgot-save-2@test.local", Password: "hashed"}); err != nil {
			t.Fatalf("seed user2: %v", err)
		}
		if err := db.Exec(`
			CREATE TRIGGER fail_set_reset_token
			BEFORE UPDATE OF reset_token ON users
			WHEN NEW.reset_token IS NOT NULL
			BEGIN
				SELECT RAISE(FAIL, 'fail set reset token');
			END;
		`).Error; err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		err = svc2.ForgotPassword(ctx, &dto.ForgotPasswordRequest{Email: "forgot-save-2@test.local"})
		if err == nil || !strings.Contains(err.Error(), "failed to save reset token") {
			t.Fatalf("expected failed to save reset token, got %v", err)
		}
	})

	t.Run("reset password hash failure", func(t *testing.T) {
		svc, repo := newAuthServiceForTest(t)
		u := &model.User{Name: "User", Email: "reset-hash@test.local", Password: "hashed", ResetToken: "tok-hash", ResetTokenExpiry: time.Now().Add(time.Hour)}
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("seed user: %v", err)
		}

		err := svc.ResetPassword(ctx, &dto.ResetPasswordRequest{Token: "tok-hash", NewPassword: strings.Repeat("y", 100)})
		if err == nil || !strings.Contains(err.Error(), "failed to hash password") {
			t.Fatalf("expected failed to hash password, got %v", err)
		}
	})

	t.Run("reset password update failure", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := db.AutoMigrate(&model.User{}); err != nil {
			t.Fatalf("migrate users: %v", err)
		}
		repo := repository.NewUserRepository(db)
		svc := NewAuthService(repo)

		u := &model.User{Name: "User", Email: "reset-upd@test.local", Password: "hashed", ResetToken: "tok-upd", ResetTokenExpiry: time.Now().Add(time.Hour)}
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("seed user: %v", err)
		}
		if err := db.Exec(`
			CREATE TRIGGER fail_update_password
			BEFORE UPDATE OF password ON users
			BEGIN
				SELECT RAISE(FAIL, 'fail update password');
			END;
		`).Error; err != nil {
			t.Fatalf("create trigger: %v", err)
		}

		err = svc.ResetPassword(ctx, &dto.ResetPasswordRequest{Token: "tok-upd", NewPassword: "newsecret123"})
		if err == nil || !strings.Contains(err.Error(), "failed to update password") {
			t.Fatalf("expected failed to update password, got %v", err)
		}
	})
}
