package service

import (
	"context"
	"errors"

	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
	"tracker-bot/pkg/apptime"
)

type EntryService interface {
	// isNew decides welcome message vs plain "back home" for the caller.
	EnsureUser(ctx context.Context, user *models.UserInput) (dbID int64, isNew bool, err error)
	CountUsers(ctx context.Context) (int, error)
	ListUsersPage(ctx context.Context, limit, offset int) ([]models.AdminUserRow, error)
	ListAllTelegramIDs(ctx context.Context) ([]int64, error)
	DeleteUser(ctx context.Context, dbUserID int64) error
}

type entryService struct {
	repo repo.EntryRepository
}

func NewEntryService(repo repo.EntryRepository) EntryService {
	return &entryService{
		repo: repo,
	}
}

func (s *entryService) EnsureUser(ctx context.Context, user *models.UserInput) (int64, bool, error) {
	_, err := s.repo.GetByID(ctx, user.TgUserID)
	if err == nil {
		dbID, err := s.repo.GetDBIDByTgUserID(ctx, user.TgUserID)
		return dbID, false, err
	}

	if !errors.Is(err, models.ErrUserNotFound) {
		return 0, false, err
	}

	if user.Language == nil || *user.Language == "" {
		v := "en"
		user.Language = &v
	}
	if user.TimeZone == nil || *user.TimeZone == "" {
		// no per-user tz picker yet; every user defaults to apptime.Location,
		// which drives all calendar math.
		v := apptime.Location.String()
		user.TimeZone = &v
	}

	dbID, err := s.repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, models.ErrUserExists) {
			return dbID, false, err
		}
		return 0, false, err
	}

	return dbID, true, nil
}

func (s *entryService) CountUsers(ctx context.Context) (int, error) {
	return s.repo.CountAll(ctx)
}

func (s *entryService) ListUsersPage(ctx context.Context, limit, offset int) ([]models.AdminUserRow, error) {
	if limit <= 0 {
		limit = 15
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPage(ctx, limit, offset)
}

func (s *entryService) ListAllTelegramIDs(ctx context.Context) ([]int64, error) {
	return s.repo.ListAllTelegramIDs(ctx)
}

func (s *entryService) DeleteUser(ctx context.Context, dbUserID int64) error {
	if dbUserID <= 0 {
		return models.ErrUserNotFound
	}
	return s.repo.Delete(ctx, dbUserID)
}
