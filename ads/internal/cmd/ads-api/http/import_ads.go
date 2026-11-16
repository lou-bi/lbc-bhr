package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

type importAdsBody struct {
	Ads []createAdBody `json:"ads"`
}

type importAdsResponse struct {
	IDs []int64 `json:"ids"`
}

// handlerImportAds serves POST /ads/import. The import is atomic: either every ad is created, or none is.
func handlerImportAds(importAds usecase.ImportAdsFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var body importAdsBody
		if err := decodeBody(r, &body); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		items := make([]usecase.ImportAdItem, 0, len(body.Ads))
		for _, item := range body.Ads {
			items = append(items, usecase.ImportAdItem{
				Title:      item.Title,
				PriceCents: item.PriceCents,
				PhotoURL:   item.PhotoURL,
			})
		}

		adIDs, err := importAds(r.Context(), usecase.ImportAdsRequest{
			Ads:     items,
			OwnerID: domain.UserID(middleware.UserIDFromRequest(r)),
		})
		if err != nil {
			if errors.Is(err, usecase.ErrUnknownOwner) {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		ids := make([]int64, 0, len(adIDs))
		for _, adID := range adIDs {
			ids = append(ids, int64(adID))
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rw).Encode(importAdsResponse{IDs: ids})
	}
}
