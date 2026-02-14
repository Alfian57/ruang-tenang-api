package service

import (
	"context"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetLeaderboard(ctx context.Context, limit int) ([]model.User, error) {
	return s.userRepo.GetTopUsers(ctx, limit)
}
