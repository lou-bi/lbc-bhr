package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	mockdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao/mock"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	usersclient "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
	usersapi "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg/api"
	mockusers "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg/mock"
)

func TestGetAdByID(t *testing.T) {
	ctx := context.Background()
	ad := domain.Ad{ID: 42, Title: "bike", Price: 5000, UserID: "user-1"}

	t.Run("nominal", func(t *testing.T) {
		m := mockdao.New(t)
		u := mockusers.New(t)
		user := usersclient.User{UserID: ad.UserID, Nickname: "alice"}
		m.On("GetAdByID", ctx, ad.ID).Return(ad, nil)
		u.On("GetUserByID", ctx, ad.UserID).Return(user, nil)
		want := ad
		want.Username = user.Nickname

		got, err := usecase.NewGetAdByIDFunc(m, u)(ctx, ad.ID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error no ad found", func(t *testing.T) {
		m := mockdao.New(t)
		u := mockusers.New(t)
		m.On("GetAdByID", ctx, mock.Anything).Return(domain.Ad{}, adsdao.ErrNotFound)

		_, err := usecase.NewGetAdByIDFunc(m, u)(ctx, ad.ID)

		require.Error(t, err)
		assert.ErrorIs(t, err, adsdao.ErrNotFound)
	})

	t.Run("error from dao", func(t *testing.T) {
		m := mockdao.New(t)
		u := mockusers.New(t)
		dbErr := errors.New("connection reset")
		m.On("GetAdByID", ctx, mock.Anything).Return(domain.Ad{}, dbErr)

		_, err := usecase.NewGetAdByIDFunc(m, u)(ctx, ad.ID)

		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("error no user found", func(t *testing.T) {
		m := mockdao.New(t)
		u := mockusers.New(t)
		m.On("GetAdByID", ctx, ad.ID).Return(ad, nil)
		u.On("GetUserByID", ctx, ad.UserID).Return(usersclient.User{}, usersapi.ErrNotFound)

		_, err := usecase.NewGetAdByIDFunc(m, u)(ctx, ad.ID)

		require.Error(t, err)
		assert.ErrorIs(t, err, usersapi.ErrNotFound)
	})

	t.Run("error from users service", func(t *testing.T) {
		m := mockdao.New(t)
		u := mockusers.New(t)
		usersErr := errors.New("identity service answered 500")
		m.On("GetAdByID", ctx, ad.ID).Return(ad, nil)
		u.On("GetUserByID", ctx, ad.UserID).Return(usersclient.User{}, usersErr)

		got, err := usecase.NewGetAdByIDFunc(m, u)(ctx, ad.ID)

		require.Error(t, err)
		assert.ErrorIs(t, err, usersErr)
		assert.Equal(t, domain.Ad{}, got)
	})
}
