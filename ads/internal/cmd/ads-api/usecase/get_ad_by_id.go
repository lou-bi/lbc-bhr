package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	usersclient "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
)

type (
	// GetAdByIDFunc returns a single ad. It fails with adsdao.ErrNotFound when no ad matches the identifier,
	// and with the users client error when the author cannot be fetched.
	GetAdByIDFunc func(ctx context.Context, id domain.AdID) (domain.Ad, error)

	getAdByID struct {
		adDAO      adsdao.DAO
		userClient usersclient.UserClient
	}
)

// NewGetAdByIDFunc returns a GetAdByIDFunc backed by the given DAO and user client.
func NewGetAdByIDFunc(adDAO adsdao.DAO, userClient usersclient.UserClient) GetAdByIDFunc {
	return getAdByID{
		adDAO:      adDAO,
		userClient: userClient,
	}.getAdByID
}

func (uc getAdByID) getAdByID(ctx context.Context, id domain.AdID) (domain.Ad, error) {
	found, err := uc.adDAO.GetAdByID(ctx, id)
	if err != nil {
		return domain.Ad{}, fmt.Errorf("unable to get ad %s: %w", id, err)
	}

	user, err := uc.userClient.GetUserByID(ctx, found.UserID)
	if err != nil {
		return domain.Ad{}, fmt.Errorf("unable to get user %s: %w", found.UserID, err)
	}

	found.Username = user.Nickname

	return found, nil
}

