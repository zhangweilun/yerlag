package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/market"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetPriceHistoryHandler returns the price history data handler.
func GetPriceHistoryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeRange := r.URL.Query().Get("range")
		if timeRange == "" {
			timeRange = "1m" // Default to 1 month
		}

		svc := market.NewPriceHistoryService(ctx)
		data, err := svc.GetPriceHistory(r.Context(), timeRange)
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, err.Error()))
			return
		}

		resp := types.SuccessResponse(data, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
			NextRefresh: time.Now().Add(10 * time.Second).Unix(),
		})
		httpx.OkJson(w, resp)
	}
}
