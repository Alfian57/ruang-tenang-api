package services

import (
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(req *dto.RegisterRequest) (*models.User, error) {
	// Check if email exists
	if s.userRepo.ExistsByEmail(req.Email) {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     models.RoleMember,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
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
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Role:      string(user.Role),
			Exp:       user.Exp,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

func (s *AuthService) GetProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *AuthService) UpdateProfile(userID uint, req *dto.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if new email is taken by another user
	if req.Email != user.Email && s.userRepo.ExistsByEmailExcept(req.Email, userID) {
		return nil, errors.New("email already taken")
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Avatar = req.Avatar

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return user, nil
}

func (s *AuthService) UpdatePassword(userID uint, req *dto.UpdatePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
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

	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *AuthService) ForgotPassword(req *dto.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(req.Email)
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

	if err := s.userRepo.UpdateResetToken(user.Email, token, expiry); err != nil {
		return errors.New("failed to save reset token")
	}

	// Mock email sending - Log the token
	// In production, send email here
	println("RESET TOKEN for " + req.Email + ": " + token)

	return nil
}

func (s *AuthService) ResetPassword(req *dto.ResetPasswordRequest) error {
	user, err := s.userRepo.FindByResetToken(req.Token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	// Clear token
	if err := s.userRepo.ClearResetToken(user.ID); err != nil {
		// Log error but don't fail properly finished process
	}

	return nil
}
