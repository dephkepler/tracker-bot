package service

import (
	"context"
	"testing"
	"tracker-bot/internal/models"
)

type fakeAdminRepo struct {
	overview  models.AdminOverviewStats
	detail    models.AdminUserDetail
	detailErr error
	gotUserID int64
}

func (f *fakeAdminRepo) GetOverviewStats(context.Context) (models.AdminOverviewStats, error) {
	return f.overview, nil
}

func (f *fakeAdminRepo) GetUserDetail(_ context.Context, dbUserID int64) (models.AdminUserDetail, error) {
	f.gotUserID = dbUserID
	if f.detailErr != nil {
		return models.AdminUserDetail{}, f.detailErr
	}
	return f.detail, nil
}

func TestAdminService_GetOverviewStats_PassesThrough(t *testing.T) {
	repo := &fakeAdminRepo{overview: models.AdminOverviewStats{TotalUsers: 7, ActiveTrackTimers: 3}}
	svc := NewAdminService(repo)

	got, err := svc.GetOverviewStats(context.Background())
	if err != nil {
		t.Fatalf("get overview stats: %v", err)
	}
	if got.TotalUsers != 7 || got.ActiveTrackTimers != 3 {
		t.Fatalf("got %+v, want TotalUsers=7 ActiveTrackTimers=3", got)
	}
}

func TestAdminService_GetUserDetail_InvalidID(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := NewAdminService(repo)

	if _, err := svc.GetUserDetail(context.Background(), 0); err == nil {
		t.Fatal("get user detail with id=0: want error")
	}
	if repo.gotUserID != 0 {
		t.Fatalf("repo should not have been called for an invalid id, got call with %d", repo.gotUserID)
	}
}

func TestAdminService_GetUserDetail_PassesThrough(t *testing.T) {
	repo := &fakeAdminRepo{detail: models.AdminUserDetail{DBID: 5, ActivitiesCount: 2}}
	svc := NewAdminService(repo)

	got, err := svc.GetUserDetail(context.Background(), 5)
	if err != nil {
		t.Fatalf("get user detail: %v", err)
	}
	if got.DBID != 5 || got.ActivitiesCount != 2 {
		t.Fatalf("got %+v, want DBID=5 ActivitiesCount=2", got)
	}
	if repo.gotUserID != 5 {
		t.Fatalf("repo called with userID=%d, want 5", repo.gotUserID)
	}
}
