package usecase

import (
	"context"
	"fmt"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

type (
	// ImportAdsFunc creates all the given ads in a single transaction.
	// It returns the identifiers of the created ads.
	ImportAdsFunc func(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error)

	// ImportAdsRequest describes the ads to import.
	ImportAdsRequest struct {
		Ads     []ImportAdItem
		OwnerID domain.UserID
	}

	// ImportAdItem describes one ad of an import.
	ImportAdItem struct {
		Title      string
		PriceCents int64
		PhotoURL   string
	}

	importAds struct {
		adDAO adsdao.DAO
	}
)

// NewImportAdsFunc returns an ImportAdsFunc backed by the given DAO.
func NewImportAdsFunc(adDAO adsdao.DAO) ImportAdsFunc {
	return importAds{adDAO: adDAO}.importAds
}

func (uc importAds) importAds(ctx context.Context, req ImportAdsRequest) ([]domain.AdID, error) {
	ids := make([]domain.AdID, 0, len(req.Ads))
	if len(req.Ads) == 0 {
		return ids, nil
	}

	tx, err := uc.adDAO.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin import: %w", err)
	}
	// No-op once the transaction is committed.
	defer func() { _ = tx.Rollback() }()

	for i, item := range req.Ads {
		adID, err := tx.CreateAd(ctx, adsdao.CreateAdRequest{
			Title:      item.Title,
			PriceCents: item.PriceCents,
			PhotoURL:   item.PhotoURL,
			OwnerID:    req.OwnerID,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to import ad %d: %w", i, err)
		}

		ids = append(ids, adID)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to commit import: %w", err)
	}

	return ids, nil
}
