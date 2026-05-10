package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"

	"github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
)

type AuthService struct {
	userRepo *infrastructure.UserRepository
}

func NewAuthService(userRepo *infrastructure.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error) {
	// Normalize email: trim whitespace and convert to lowercase
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Check if email exists
	if s.userRepo.ExistsByEmail(ctx, req.Email) {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     model.RoleUser,
	}

	if req.Role != "" {
		normalizedRole := model.UserRole(strings.TrimSpace(strings.ToLower(req.Role)))
		if normalizedRole != model.RoleUser {
			return nil, errors.New("invalid role")
		}
		user.Role = model.RoleUser
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Normalize email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Check if user can access (blocked, banned, or suspended)
	if user.IsBlocked {
		return nil, errors.New("akun Anda telah diblokir, silakan hubungi administrator")
	}

	if user.IsBanned {
		return nil, errors.New("akun Anda telah dibanned, silakan hubungi administrator")
	}

	if user.IsSuspended() {
		return nil, errors.New("akun Anda sedang disuspend, silakan coba lagi nanti")
	}

	tokenExpiry := time.Duration(config.AppConfig.JWTExpiryHours) * time.Hour
	if req.RememberMe {
		tokenExpiry = 30 * 24 * time.Hour // 30 days
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), tokenExpiry)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserDTO{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email,
			Avatar:       user.Avatar,
			Role:         string(user.Role),
			Exp:          user.Exp,
			GoldCoins:    user.GoldCoins,
			IsPremium:    user.IsPremium && (user.PremiumExpiresAt == nil || user.PremiumExpiresAt.After(time.Now())),
			PremiumUntil: formatTimePointer(user.PremiumExpiresAt),
			CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

func formatTimePointer(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02T15:04:05Z")
}

func (s *AuthService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if new email is taken by another user
	if req.Email != user.Email && s.userRepo.ExistsByEmailExcept(ctx, req.Email, userID) {
		return nil, errors.New("email already taken")
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Avatar = req.Avatar

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return user, nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, userID uint, req *dto.UpdatePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Return nil to avoid email enumeration
		return nil
	}

	// Generate reset token (simple random string)
	token, err := utils.GenerateRandomString(32)
	if err != nil {
		return errors.New("failed to generate token")
	}

	expiry := time.Now().Add(1 * time.Hour)

	if err := s.userRepo.UpdateResetToken(ctx, user.Email, token, expiry); err != nil {
		return errors.New("failed to save reset token")
	}

	// TODO: Implement actual email sending
	// In production, send email with the reset token
	// For now, token is saved in DB and can be used via /reset-password endpoint
	_ = token

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	user, err := s.userRepo.FindByResetToken(ctx, req.Token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.New("failed to update password")
	}

	// Clear token
	if err := s.userRepo.ClearResetToken(ctx, user.ID); err != nil {
		// Log error but don't fail properly finished process
	}

	return nil
}
