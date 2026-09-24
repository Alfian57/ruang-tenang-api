package dto

// Auth DTOs
type RegisterRequest struct {
	Name           string `json:"name" binding:"required,min=2,max=100"`
	Email          string `json:"email" binding:"required,email"`
	WhatsAppNumber string `json:"whatsapp_number" binding:"required"`
	Password       string `json:"password" binding:"required,min=6"`
	Role           string `json:"role" binding:"omitempty,oneof=user"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	Token                string   `json:"token,omitempty"`
	User                 *UserDTO `json:"user,omitempty"`
	VerificationRequired bool     `json:"verification_required,omitempty"`
	VerificationToken    string   `json:"verification_token,omitempty"`
	PhoneRequired        bool     `json:"phone_required,omitempty"`
}

type SetVerificationPhoneRequest struct {
	VerificationToken string `json:"verification_token" binding:"required"`
	WhatsAppNumber    string `json:"whatsapp_number" binding:"required"`
}

type VerifyPhoneRequest struct {
	VerificationToken string `json:"verification_token" binding:"required"`
	Code              string `json:"code" binding:"required,len=6,numeric"`
}

type UpdateProfileRequest struct {
	Name           string `json:"name" binding:"omitempty,min=2,max=100"`
	Email          string `json:"email" binding:"omitempty,email"`
	WhatsAppNumber string `json:"whatsapp_number"`
	Avatar         string `json:"avatar"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ForgotPassword & ResetPassword
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token                string `json:"token" binding:"required"`
	NewPassword          string `json:"new_password" binding:"required,min=6"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=NewPassword"`
}

// User DTO
type UserDTO struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	WhatsAppNumber   string `json:"whatsapp_number,omitempty"`
	WhatsAppVerified bool   `json:"whatsapp_verified"`
	Avatar           string `json:"avatar"`
	Role             string `json:"role"`
	Exp              int64  `json:"exp"`
	GoldCoins        int64  `json:"gold_coins"`
	IsPremium        bool   `json:"is_premium"`
	PremiumUntil     string `json:"premium_until,omitempty"`
	Level            int    `json:"level"`
	BadgeName        string `json:"badge_name"`
	BadgeIcon        string `json:"badge_icon"`
	ProfileTheme     string `json:"profile_theme"`
	CreatedAt        string `json:"created_at"`
}
