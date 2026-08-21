package service

import (
	"context"

	"tracker-bot/internal/repo"
)

type UIStateService interface {
	GetScreen(ctx context.Context, userID int64) (string, error)
	SetScreen(ctx context.Context, userID int64, screen string) error
}

type uiStateService struct {
	repo repo.UIStateRepository
}

func NewUIStateService(repo repo.UIStateRepository) UIStateService {
	return &uiStateService{repo: repo}
}

func (s *uiStateService) GetScreen(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}
	return s.repo.GetScreen(ctx, userID)
}

func (s *uiStateService) SetScreen(ctx context.Context, userID int64, screen string) error {
	if userID <= 0 {
		return nil
	}
	return s.repo.SetScreen(ctx, userID, screen)
}
