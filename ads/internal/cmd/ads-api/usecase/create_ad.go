package usecase

import (
	"context"
	"fmt"
	"time"

	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
)

const (
	// maxAdsPerWindow is the number of ads a user can create within quotaWindow.
	maxAdsPerWindow = 5
	// quotaWindow is the rolling period of the quota.
	quotaWindow = 24 * time.Hour
)

type (
	// CreateAdFunc creates an ad owned by the calling user. It fails with QuotaExceededError when the quota is reached.
	CreateAdFunc func(ctx context.Context, req CreateAdRequest) (domain.AdID, error)

	// CreateAdRequest describes the ad to create. OwnerID is the authenticated user: a caller can never create an ad on behalf of someone else.
	CreateAdRequest struct {
		Title      string
		PriceCents int64
		PhotoURL   string
		OwnerID    domain.UserID
	}

	createAd struct {
		adDAO adsdao.DAO
	}
)

// QuotaExceededError is returned when the user reached the ad quota. RetryAfter is the wait before the next ad can be created.
type QuotaExceededError struct {
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e QuotaExceededError) Error() string {
	return fmt.Sprintf("ad quota exceeded: retry in %s", e.RetryAfter.Round(time.Second))
}

// NewCreateAdFunc returns a CreateAdFunc backed by the given DAO.
func NewCreateAdFunc(adDAO adsdao.DAO) CreateAdFunc {
	return createAd{adDAO: adDAO}.createAd
}

func (uc createAd) createAd(ctx context.Context, req CreateAdRequest) (domain.AdID, error) {
	tx, err := uc.adDAO.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to begin ad creation: %w", err)
	}
	// No-op once the transaction is committed.
	defer func() { _ = tx.Rollback() }()

	// Serializes the creations of the owner, so the quota check and the insert cannot interleave.
	if err := tx.LockOwner(ctx, req.OwnerID); err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	count, oldest, err := tx.CountRecentAds(ctx, req.OwnerID, quotaWindow)
	if err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	if count >= maxAdsPerWindow {
		return 0, fmt.Errorf("unable to create ad: %w", QuotaExceededError{
			RetryAfter: max(time.Until(oldest.Add(quotaWindow)), 0),
		})
	}

	adID, err := tx.CreateAd(ctx, adsdao.CreateAdRequest{
		Title:      req.Title,
		PriceCents: req.PriceCents,
		PhotoURL:   req.PhotoURL,
		OwnerID:    req.OwnerID,
	})
	if err != nil {
		return 0, fmt.Errorf("unable to create ad: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("unable to commit ad creation: %w", err)
	}

	return adID, nil
}
