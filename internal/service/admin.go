package service

import (
	"context"
	"fmt"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// AdminService contains cross-domain admin use-cases: bot-wide reporting and
// per-user drill-down. User listing/counting/deletion lives on EntryService
// (it only ever touches the users table); this service is for reporting
// that crosses into track/learning data.
type AdminService interface {
	GetOverviewStats(ctx context.Context) (models.AdminOverviewStats, error)
	GetUserDetail(ctx context.Context, dbUserID int64) (models.AdminUserDetail, error)
}

type adminService struct {
	repo repo.AdminRepository
}

// NewAdminService creates admin service.
func NewAdminService(repo repo.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

// GetOverviewStats aggregates bot-wide usage numbers.
func (s *adminService) GetOverviewStats(ctx context.Context) (models.AdminOverviewStats, error) {
	return s.repo.GetOverviewStats(ctx)
}

// GetUserDetail loads one user's profile fields plus usage counts.
func (s *adminService) GetUserDetail(ctx context.Context, dbUserID int64) (models.AdminUserDetail, error) {
	if dbUserID <= 0 {
		return models.AdminUserDetail{}, fmt.Errorf("get user detail: invalid userID")
	}
	return s.repo.GetUserDetail(ctx, dbUserID)
}
