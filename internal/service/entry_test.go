package service

import (
	"context"
	"testing"
	"tracker-bot/internal/models"
)

type fakeEntryRepo struct {
	listLimit, listOffset int // records what the last ListPage call received
	count                 int
}

func (f *fakeEntryRepo) GetByID(context.Context, int64) (*models.User, error) { return nil, nil }
func (f *fakeEntryRepo) GetDBIDByTgUserID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (f *fakeEntryRepo) Create(context.Context, *models.UserInput) (int64, error) { return 0, nil }
func (f *fakeEntryRepo) CountAll(context.Context) (int, error)                    { return f.count, nil }
func (f *fakeEntryRepo) ListPage(_ context.Context, limit, offset int) ([]models.AdminUserRow, error) {
	f.listLimit, f.listOffset = limit, offset
	return nil, nil
}

func TestEntryService_ListUsersPage_ClampsLimitAndOffset(t *testing.T) {
	repo := &fakeEntryRepo{}
	svc := NewEntryService(repo)
	ctx := context.Background()

	if _, err := svc.ListUsersPage(ctx, 0, -5); err != nil {
		t.Fatalf("ListUsersPage: %v", err)
	}
	if repo.listLimit != 15 {
		t.Fatalf("limit passed to repo = %d, want default 15 for limit<=0", repo.listLimit)
	}
	if repo.listOffset != 0 {
		t.Fatalf("offset passed to repo = %d, want clamped to 0 for negative input", repo.listOffset)
	}

	if _, err := svc.ListUsersPage(ctx, 10, 20); err != nil {
		t.Fatalf("ListUsersPage: %v", err)
	}
	if repo.listLimit != 10 || repo.listOffset != 20 {
		t.Fatalf("valid limit/offset got mangled: got (%d, %d), want (10, 20)", repo.listLimit, repo.listOffset)
	}
}

func TestEntryService_CountUsers(t *testing.T) {
	repo := &fakeEntryRepo{count: 42}
	svc := NewEntryService(repo)

	got, err := svc.CountUsers(context.Background())
	if err != nil {
		t.Fatalf("CountUsers: %v", err)
	}
	if got != 42 {
		t.Fatalf("CountUsers() = %d, want 42", got)
	}
}
