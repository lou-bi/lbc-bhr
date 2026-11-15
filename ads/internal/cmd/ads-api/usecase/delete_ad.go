package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// DeleteAdFunc deletes a single ad. It fails with ErrNotFound when no ad matches the identifier,
	// or if the ad is not owned by OwnerID
	DeleteAdFunc func(ctx context.Context, req DeleteAdRequest) error

	DeleteAdRequest struct {
		ID 	domain.AdID
		OwnerID domain.UserID
	}

	deleteAd struct {
		adDAO adsdao.DAO
	}
)

// NewDeleteAdFunc returns a NewDeleteAdFunc backed by the given DAO.
func NewDeleteAdFunc(adDAO adsdao.DAO) DeleteAdFunc {
	return deleteAd{adDAO: adDAO}.deleteAd
}

func (uc deleteAd) deleteAd(ctx context.Context, req DeleteAdRequest) error {
	err := uc.adDAO.DeleteAd(ctx, adsdao.DeleteAdRequest{
		ID: req.ID,
		OwnerID: req.OwnerID,
	})
	if err != nil {
		return fmt.Errorf("unable to delete ad %s: %w", req.ID, err)
	}

	return nil
}

