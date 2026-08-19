package service

import (
	"context"
	"errors"

	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
	"tracker-bot/pkg/apptime"
)

type EntryService interface {
	// EnsureUser loads or creates the user and reports whether this is their
	// very first time (used to decide between a welcome message and a plain
	// "back home" one).
	EnsureUser(ctx context.Context, user *models.UserInput) (dbID int64, isNew bool, err error)
	// CountUsers returns the total number of registered users (admin stat).
	CountUsers(ctx context.Context) (int, error)
	// ListUsersPage returns one page of registered users, newest first
	// (admin listing).
	ListUsersPage(ctx context.Context, limit, offset int) ([]models.AdminUserRow, error)
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
		// No per-user timezone picker is wired yet (the "📍 Time zone" profile
		// button has no handler) — every user is currently treated as being in
		// this zone, matching apptime.Location which drives all calendar math.
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

// CountUsers returns the total number of registered users.
func (s *entryService) CountUsers(ctx context.Context) (int, error) {
	return s.repo.CountAll(ctx)
}

// ListUsersPage returns one page of registered users, clamping limit/offset
// to sane bounds.
func (s *entryService) ListUsersPage(ctx context.Context, limit, offset int) ([]models.AdminUserRow, error) {
	if limit <= 0 {
		limit = 15
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPage(ctx, limit, offset)
}
