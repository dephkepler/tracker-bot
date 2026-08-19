package service

import (
	"context"
	"errors"
	"testing"
	"tracker-bot/internal/models"
)

type fakeEntryRepo struct {
	listLimit, listOffset int // records what the last ListPage call received
	count                 int
	tgIDs                 []int64
	deletedID             int64 // records what the last Delete call received
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
func (f *fakeEntryRepo) ListAllTelegramIDs(context.Context) ([]int64, error) { return f.tgIDs, nil }
func (f *fakeEntryRepo) Delete(_ context.Context, dbUserID int64) error {
	f.deletedID = dbUserID
	return nil
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

func TestEntryService_ListAllTelegramIDs(t *testing.T) {
	repo := &fakeEntryRepo{tgIDs: []int64{10, 20, 30}}
	svc := NewEntryService(repo)

	got, err := svc.ListAllTelegramIDs(context.Background())
	if err != nil {
		t.Fatalf("ListAllTelegramIDs: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ListAllTelegramIDs() = %v, want 3 ids", got)
	}
}

func TestEntryService_DeleteUser(t *testing.T) {
	repo := &fakeEntryRepo{}
	svc := NewEntryService(repo)

	if err := svc.DeleteUser(context.Background(), 5); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if repo.deletedID != 5 {
		t.Fatalf("repo.Delete called with %d, want 5", repo.deletedID)
	}
}

func TestEntryService_DeleteUser_InvalidID(t *testing.T) {
	repo := &fakeEntryRepo{}
	svc := NewEntryService(repo)

	if err := svc.DeleteUser(context.Background(), 0); !errors.Is(err, models.ErrUserNotFound) {
		t.Fatalf("DeleteUser(0): got %v, want ErrUserNotFound", err)
	}
	if repo.deletedID != 0 {
		t.Fatalf("repo.Delete should not have been called for an invalid id, got call with %d", repo.deletedID)
	}
}
