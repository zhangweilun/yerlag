package handler

import (
	"context"
	"net/http"
	"time"

	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// allCacheKeys lists all known cache keys used by the data modules.
var allCacheKeys = []string{
	"onchain:burn",
	"onchain:gas",
	"onchain:activity",
	"onchain:tvl",
	"onchain:supply",
	"market:overview",
	"network:staking",
	"network:performance",
	"macro:ethbtc",
	"macro:indicators",
	"institutional:etf",
	"institutional:grayscale",
	"institutional:holdings",
	"valuation:score",
}

// forceRefreshResp is the response for force refresh.
type forceRefreshResp struct {
	// Invalidated is the number of cache keys that were invalidated.
	Invalidated int `json:"invalidated"`
	// Total is the total number of cache keys attempted.
	Total int `json:"total"`
	// Errors lists any keys that failed to invalidate.
	Errors []string `json:"errors,omitempty"`
}

// ForceRefreshHandler returns the force refresh handler.
// It invalidates all module caches via DataFetcher.InvalidateCache() and returns success.
func ForceRefreshHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bgCtx := context.Background()
		invalidated := 0
		var errors []string

		for _, key := range allCacheKeys {
			if err := ctx.DataFetcher.InvalidateCache(bgCtx, key); err != nil {
				errors = append(errors, key+": "+err.Error())
			} else {
				invalidated++
			}
		}

		resp := types.SuccessResponse(forceRefreshResp{
			Invalidated: invalidated,
			Total:       len(allCacheKeys),
			Errors:      errors,
		}, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
		})
		httpx.OkJson(w, resp)
	}
}
