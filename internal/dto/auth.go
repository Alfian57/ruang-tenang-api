package dto

// Auth DTOs
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type UpdateProfileRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100"`
	Email  string `json:"email" binding:"required,email"`
	Avatar string `json:"avatar"`
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
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// User DTO
type UserDTO struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	Exp       int64  `json:"exp"`
	CreatedAt string `json:"created_at"`
}
