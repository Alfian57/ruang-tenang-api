package application

import (
	"context"
	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure")

type UserService struct {
	userRepo *infrastructure.UserRepository
}

func NewUserService(userRepo *infrastructure.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetLeaderboard(ctx context.Context, limit int) ([]model.User, error) {
	return s.userRepo.GetTopUsers(ctx, limit)
}
