package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"

	"github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrInvalidPhoneVerification = errors.New("kode verifikasi tidak valid atau kedaluwarsa")

type AuthService struct {
	userRepo *infrastructure.UserRepository
	whatsApp WhatsAppSender
}

func NewAuthService(userRepo *infrastructure.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) SetWhatsAppSender(sender WhatsAppSender) { s.whatsApp = sender }

func normalizeWhatsAppNumber(raw string) (string, error) {
	number := strings.TrimSpace(raw)
	number = strings.TrimPrefix(number, "+")
	if strings.HasPrefix(number, "0") {
		number = "62" + number[1:]
	}
	if !strings.HasPrefix(number, "62") || len(number) < 11 || len(number) > 16 || number[2] != '8' {
		return "", errors.New("nomor WhatsApp Indonesia tidak valid")
	}
	for _, ch := range number {
		if ch < '0' || ch > '9' {
			return "", errors.New("nomor WhatsApp Indonesia tidak valid")
		}
	}
	return number, nil
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*model.User, error) {
	// Normalize email: trim whitespace and convert to lowercase
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	whatsAppNumber, err := normalizeWhatsAppNumber(req.WhatsAppNumber)
	if err != nil {
		return nil, err
	}

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
		Name:           req.Name,
		Email:          req.Email,
		WhatsAppNumber: whatsAppNumber,
		Password:       hashedPassword,
		Role:           model.RoleUser,
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
	if user.WhatsAppNumber == "" || user.WhatsAppVerifiedAt == nil {
		return s.startPhoneVerification(ctx, user, req.RememberMe)
	}
	return s.loginResponse(user, req.RememberMe)
}

func (s *AuthService) loginResponse(user *model.User, rememberMe bool) (*dto.LoginResponse, error) {
	tokenExpiry := time.Duration(config.AppConfig.JWTExpiryHours) * time.Hour
	if rememberMe {
		tokenExpiry = 30 * 24 * time.Hour // 30 days
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), tokenExpiry, user.WhatsAppVerifiedAt != nil)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &dto.LoginResponse{
		Token: token,
		User: &dto.UserDTO{
			ID:               user.ID,
			Name:             user.Name,
			Email:            user.Email,
			WhatsAppNumber:   user.WhatsAppNumber,
			WhatsAppVerified: user.WhatsAppVerifiedAt != nil,
			Avatar:           user.Avatar,
			Role:             string(user.Role),
			Exp:              user.Exp,
			GoldCoins:        user.GoldCoins,
			IsPremium:        user.IsPremium && (user.PremiumExpiresAt == nil || user.PremiumExpiresAt.After(time.Now())),
			PremiumUntil:     formatTimePointer(user.PremiumExpiresAt),
			CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

func phoneChallengeHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func phoneCodeHash(challenge, code string) string {
	mac := hmac.New(sha256.New, []byte(config.AppConfig.JWTSecret))
	_, _ = mac.Write([]byte(challenge + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func newPhoneCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func (s *AuthService) startPhoneVerification(ctx context.Context, user *model.User, rememberMe bool) (*dto.LoginResponse, error) {
	if s.whatsApp == nil || !s.whatsApp.IsConfigured() {
		return nil, errors.New("pengiriman WhatsApp belum dikonfigurasi")
	}
	challenge, err := utils.GenerateRandomString(48)
	if err != nil {
		return nil, err
	}
	verification := &model.PhoneVerification{
		UserID: user.ID, ChallengeHash: phoneChallengeHash(challenge),
		RememberMe: rememberMe, ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if user.WhatsAppNumber != "" {
		code, err := newPhoneCode()
		if err != nil {
			return nil, err
		}
		verification.CodeHash = phoneCodeHash(challenge, code)
		if err := s.userRepo.CreatePhoneVerification(ctx, verification); err != nil {
			return nil, err
		}
		if err := s.whatsApp.Send(ctx, user.WhatsAppNumber, "Kode verifikasi Ruang Tenang: "+code+". Berlaku 10 menit. Jangan bagikan kode ini."); err != nil {
			return nil, errors.New("gagal mengirim kode verifikasi WhatsApp")
		}
	} else if err := s.userRepo.CreatePhoneVerification(ctx, verification); err != nil {
		return nil, err
	}
	return &dto.LoginResponse{VerificationRequired: true, VerificationToken: challenge, PhoneRequired: user.WhatsAppNumber == ""}, nil
}

func (s *AuthService) SetVerificationPhone(ctx context.Context, req *dto.SetVerificationPhoneRequest) error {
	number, err := normalizeWhatsAppNumber(req.WhatsAppNumber)
	if err != nil {
		return err
	}
	code, err := newPhoneCode()
	if err != nil {
		return err
	}
	challengeHash := phoneChallengeHash(req.VerificationToken)
	err = s.userRepo.WithPhoneVerification(ctx, challengeHash, func(tx *gorm.DB, verification *model.PhoneVerification, user *model.User) error {
		if time.Now().After(verification.ExpiresAt) {
			return ErrInvalidPhoneVerification
		}
		if user.WhatsAppNumber != "" {
			return ErrInvalidPhoneVerification
		}
		user.WhatsAppNumber = number
		user.WhatsAppVerifiedAt = nil
		verification.CodeHash = phoneCodeHash(req.VerificationToken, code)
		verification.Attempts = 0
		verification.ExpiresAt = time.Now().Add(10 * time.Minute)
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return tx.Save(verification).Error
	})
	if err != nil {
		return err
	}
	if err := s.whatsApp.Send(ctx, number, "Kode verifikasi Ruang Tenang: "+code+". Berlaku 10 menit. Jangan bagikan kode ini."); err != nil {
		return errors.New("gagal mengirim kode verifikasi WhatsApp")
	}
	return nil
}

func (s *AuthService) VerifyPhone(ctx context.Context, req *dto.VerifyPhoneRequest) (*dto.LoginResponse, error) {
	var verifiedUser *model.User
	var rememberMe bool
	err := s.userRepo.WithPhoneVerification(ctx, phoneChallengeHash(req.VerificationToken), func(tx *gorm.DB, verification *model.PhoneVerification, user *model.User) error {
		if time.Now().After(verification.ExpiresAt) || verification.Attempts >= 5 || verification.CodeHash == "" || user.WhatsAppNumber == "" {
			return ErrInvalidPhoneVerification
		}
		verification.Attempts++
		if err := tx.Save(verification).Error; err != nil {
			return err
		}
		expected := phoneCodeHash(req.VerificationToken, req.Code)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(verification.CodeHash)) != 1 {
			return nil
		}
		if user.IsBlocked || user.IsBanned || user.IsSuspended() {
			return ErrInvalidPhoneVerification
		}
		now := time.Now()
		user.WhatsAppVerifiedAt = &now
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		if err := tx.Delete(verification).Error; err != nil {
			return err
		}
		verifiedUser = user
		rememberMe = verification.RememberMe
		return nil
	})
	if err != nil || verifiedUser == nil {
		return nil, ErrInvalidPhoneVerification
	}
	return s.loginResponse(verifiedUser, rememberMe)
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
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Check if new email is taken by another user
	if req.Email != "" && req.Email != user.Email && s.userRepo.ExistsByEmailExcept(ctx, req.Email, userID) {
		return nil, errors.New("email already taken")
	}

	if req.Name != "" {
		user.Name = strings.TrimSpace(req.Name)
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.WhatsAppNumber != "" {
		number, err := normalizeWhatsAppNumber(req.WhatsAppNumber)
		if err != nil {
			return nil, err
		}
		if user.WhatsAppNumber != number {
			user.WhatsAppNumber = number
			user.WhatsAppVerifiedAt = nil
		}
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

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
	if s.whatsApp == nil || !s.whatsApp.IsConfigured() {
		return errors.New("pengiriman WhatsApp belum dikonfigurasi")
	}
	user, err := s.userRepo.FindByEmail(ctx, strings.TrimSpace(strings.ToLower(req.Email)))
	if err != nil {
		// Keep the same response for unknown accounts and accounts without a number.
		return nil
	}
	if user.WhatsAppNumber == "" || user.WhatsAppVerifiedAt == nil {
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

	message := "Kode reset kata sandi Ruang Tenang: " + token + "\nBerlaku 1 jam. Jangan bagikan kode ini."
	if frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_URL")); frontendURL != "" {
		message += "\n" + strings.TrimRight(frontendURL, "/") + "/reset-password?token=" + url.QueryEscape(token)
	}
	if err := s.whatsApp.Send(ctx, user.WhatsAppNumber, message); err != nil {
		_ = s.userRepo.ClearResetToken(ctx, user.ID)
		logger.Warn("auth: WhatsApp reset delivery failed", zap.Error(err))
		return nil
	}

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
