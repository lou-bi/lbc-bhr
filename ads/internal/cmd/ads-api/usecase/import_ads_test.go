package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	mockdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestImportAds(t *testing.T) {
	ctx := context.Background()
	req := usecase.ImportAdsRequest{
		OwnerID: "user-1",
		Ads: []usecase.ImportAdItem{
			{Title: "bike", PriceCents: 5000, PhotoURL: "https://example.com/bike.jpg"},
			{Title: "scooter", PriceCents: 8000, PhotoURL: "https://example.com/scooter.jpg"},
			{Title: "car", PriceCents: 200000, PhotoURL: "https://example.com/car.jpg"},
		},
	}
	daoReq := func(i int) adsdao.CreateAdRequest {
		return adsdao.CreateAdRequest{
			Title:      req.Ads[i].Title,
			PriceCents: req.Ads[i].PriceCents,
			PhotoURL:   req.Ads[i].PhotoURL,
			OwnerID:    req.OwnerID,
		}
	}

	t.Run("nominal", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("Begin", ctx).Return(nil)
		m.On("CreateAd", ctx, daoReq(0)).Return(domain.AdID(10), nil)
		m.On("CreateAd", ctx, daoReq(1)).Return(domain.AdID(11), nil)
		m.On("CreateAd", ctx, daoReq(2)).Return(domain.AdID(12), nil)
		m.On("Commit").Return(nil)
		m.On("Rollback").Return(nil)

		got, err := usecase.NewImportAdsFunc(m)(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, []domain.AdID{10, 11, 12}, got)
	})

	t.Run("no ads", func(t *testing.T) {
		m := mockdao.New(t)

		got, err := usecase.NewImportAdsFunc(m)(ctx, usecase.ImportAdsRequest{OwnerID: req.OwnerID})

		require.NoError(t, err)
		assert.Empty(t, got)
		assert.NotNil(t, got)
	})

	t.Run("error on begin", func(t *testing.T) {
		m := mockdao.New(t)
		beginErr := errors.New("connection refused")
		m.On("Begin", ctx).Return(beginErr)

		got, err := usecase.NewImportAdsFunc(m)(ctx, req)

		assert.ErrorIs(t, err, beginErr)
		assert.Nil(t, got)
	})

	t.Run("unknown owner rolls back and never commits", func(t *testing.T) {
		m := mockdao.New(t)
		m.On("Begin", ctx).Return(nil)
		m.On("CreateAd", ctx, daoReq(0)).Return(domain.AdID(10), nil)
		m.On("CreateAd", ctx, daoReq(1)).Return(domain.AdID(0), adsdao.ErrUnknownOwner)
		m.On("Rollback").Return(nil)

		got, err := usecase.NewImportAdsFunc(m)(ctx, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, usecase.ErrUnknownOwner)
		assert.Nil(t, got)
		m.AssertNotCalled(t, "Commit")
		m.AssertNotCalled(t, "CreateAd", ctx, daoReq(2))
	})

	t.Run("error on commit", func(t *testing.T) {
		m := mockdao.New(t)
		commitErr := errors.New("connection reset")
		m.On("Begin", ctx).Return(nil)
		m.On("CreateAd", ctx, daoReq(0)).Return(domain.AdID(10), nil)
		m.On("CreateAd", ctx, daoReq(1)).Return(domain.AdID(11), nil)
		m.On("CreateAd", ctx, daoReq(2)).Return(domain.AdID(12), nil)
		m.On("Commit").Return(commitErr)
		m.On("Rollback").Return(nil)

		got, err := usecase.NewImportAdsFunc(m)(ctx, req)

		assert.ErrorIs(t, err, commitErr)
		assert.Nil(t, got)
	})
}
