package service

import (
	"context"
	"fmt"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// user listing/counting/deletion lives on EntryService; this is cross-domain reporting only
type AdminService interface {
	GetOverviewStats(ctx context.Context) (models.AdminOverviewStats, error)
	GetUserDetail(ctx context.Context, dbUserID int64) (models.AdminUserDetail, error)
}

type adminService struct {
	repo repo.AdminRepository
}

func NewAdminService(repo repo.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) GetOverviewStats(ctx context.Context) (models.AdminOverviewStats, error) {
	return s.repo.GetOverviewStats(ctx)
}

func (s *adminService) GetUserDetail(ctx context.Context, dbUserID int64) (models.AdminUserDetail, error) {
	if dbUserID <= 0 {
		return models.AdminUserDetail{}, fmt.Errorf("get user detail: invalid userID")
	}
	return s.repo.GetUserDetail(ctx, dbUserID)
}
