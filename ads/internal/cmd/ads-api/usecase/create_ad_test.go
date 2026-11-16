package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	mockdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestCreateAd(t *testing.T) {
	ctx := context.Background()

	ucReq := usecase.CreateAdRequest{
		Title:      "mountain bike",
		PriceCents: 15000,
		PhotoURL:   "https://example.com/bike.jpg",
		OwnerID:    "user-1",
	}
	daoReq := adsdao.CreateAdRequest{
		Title:      "mountain bike",
		PriceCents: 15000,
		PhotoURL:   "https://example.com/bike.jpg",
		OwnerID:    "user-1",
	}

	t.Run("nominal", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(0, time.Time{}, nil)
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(7), nil)
		m.On("Commit").Return(nil)
		m.On("Rollback").Return(nil)

		got, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		require.NoError(t, err)
		assert.Equal(t, domain.AdID(7), got)
	})

	t.Run("accepts the last ad of the quota", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(4, time.Now().Add(-time.Hour), nil)
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(7), nil)
		m.On("Commit").Return(nil)
		m.On("Rollback").Return(nil)

		got, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		require.NoError(t, err)
		assert.Equal(t, domain.AdID(7), got)
	})

	t.Run("error quota exceeded", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(5, time.Now().Add(-23*time.Hour), nil)
		m.On("Rollback").Return(nil)

		_, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		var quotaErr usecase.QuotaExceededError
		require.ErrorAs(t, err, &quotaErr)
		// The oldest ad leaves the window in about one hour.
		assert.InDelta(t, float64(time.Hour), float64(quotaErr.RetryAfter), float64(time.Minute))
		m.AssertNotCalled(t, "CreateAd", ctx, daoReq)
		m.AssertNotCalled(t, "Commit")
	})

	t.Run("error quota exceeded never waits a negative time", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		// Clock skew: the oldest ad already looks out of the window.
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(5, time.Now().Add(-25*time.Hour), nil)
		m.On("Rollback").Return(nil)

		_, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		var quotaErr usecase.QuotaExceededError
		require.ErrorAs(t, err, &quotaErr)
		assert.Zero(t, quotaErr.RetryAfter)
	})

	t.Run("error on begin", func(t *testing.T) {
		m := mockdao.New(t)
		beginErr := errors.New("connection refused")
		m.On("Begin", ctx).Return(beginErr)

		_, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		assert.ErrorIs(t, err, beginErr)
	})

	t.Run("error on lock", func(t *testing.T) {
		m := mockdao.New(t)
		lockErr := errors.New("lock timeout")
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(lockErr)
		m.On("Rollback").Return(nil)

		_, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		assert.ErrorIs(t, err, lockErr)
		m.AssertNotCalled(t, "CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour)
		m.AssertNotCalled(t, "Commit")
	})

	t.Run("error on count", func(t *testing.T) {
		m := mockdao.New(t)
		countErr := errors.New("connection reset")
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(0, time.Time{}, countErr)
		m.On("Rollback").Return(nil)

		_, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		assert.ErrorIs(t, err, countErr)
		m.AssertNotCalled(t, "CreateAd", ctx, daoReq)
		m.AssertNotCalled(t, "Commit")
	})

	t.Run("error from dao", func(t *testing.T) {
		m := mockdao.New(t)
		dbErr := errors.New("connection reset")
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(0, time.Time{}, nil)
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(0), dbErr)
		m.On("Rollback").Return(nil)

		_, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		m.AssertNotCalled(t, "Commit")
	})

	t.Run("error on commit", func(t *testing.T) {
		m := mockdao.New(t)
		commitErr := errors.New("connection reset")
		m.On("Begin", ctx).Return(nil)
		m.On("LockOwner", ctx, ucReq.OwnerID).Return(nil)
		m.On("CountRecentAds", ctx, ucReq.OwnerID, 24*time.Hour).Return(0, time.Time{}, nil)
		m.On("CreateAd", ctx, daoReq).Return(domain.AdID(7), nil)
		m.On("Commit").Return(commitErr)
		m.On("Rollback").Return(nil)

		got, err := usecase.NewCreateAdFunc(m)(ctx, ucReq)

		assert.ErrorIs(t, err, commitErr)
		assert.Zero(t, got)
	})
}
