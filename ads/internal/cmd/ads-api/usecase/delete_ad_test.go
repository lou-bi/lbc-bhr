package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
)

func TestDeleteAd(t *testing.T) {
	ctx := context.Background()

	ucReq := usecase.DeleteAdRequest{
		ID: 100,
		OwnerID: "00000000-0000-0000-0000-000000000001",
	}
	daoReq := adsdao.DeleteAdRequest{
		ID: 100,
		OwnerID: "00000000-0000-0000-0000-000000000001",
	}

	t.Run("nominal", func(t *testing.T) {
		m := mock.New(t)
		m.On("DeleteAd", ctx, daoReq).Return(nil)

		got := usecase.NewDeleteAdFunc(m)(ctx, ucReq)
		require.NoError(t, got)
	})

	t.Run("error no ad found", func(t *testing.T) {
		m := mock.New(t)
		m.On("DeleteAd", ctx, daoReq).Return(adsdao.ErrNotFound)

		err := usecase.NewDeleteAdFunc(m)(ctx, ucReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, adsdao.ErrNotFound)
	})

	t.Run("error from dao", func(t *testing.T) {
		m := mock.New(t)
		dbErr := errors.New("connection reset")
		m.On("DeleteAd", ctx, daoReq).Return(dbErr)

		err := usecase.NewDeleteAdFunc(m)(ctx, ucReq)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
	})
}
