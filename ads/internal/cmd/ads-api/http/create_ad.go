package http

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

type createAdBody struct {
	Title      string `json:"title"`
	PriceCents int64  `json:"price_cents"`
	PhotoURL   string `json:"photo_url"`
}

type createAdResponse struct {
	ID int64 `json:"id"`
}

// retryAfterSeconds returns d in whole seconds, at least one.
func retryAfterSeconds(d time.Duration) int {
	return max(1, int(math.Ceil(d.Seconds())))
}

// handlerCreateAd serves POST /ads.
func handlerCreateAd(createAd usecase.CreateAdFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var body createAdBody
		if err := decodeBody(r, &body); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		adID, err := createAd(r.Context(), usecase.CreateAdRequest{
			Title:      body.Title,
			PriceCents: body.PriceCents,
			PhotoURL:   body.PhotoURL,
			OwnerID:    middleware.UserIDFromHeader(r),
		})
		if err != nil {
			if errors.Is(err, usecase.ErrUnknownOwner) {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			var quotaErr usecase.QuotaExceededError
			if errors.As(err, &quotaErr) {
				rw.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(quotaErr.RetryAfter)))
				rw.WriteHeader(http.StatusTooManyRequests)
				return
			}

			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rw).Encode(createAdResponse{ID: int64(adID)})
	}
}
