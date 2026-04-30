package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/market"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetMarketDataHandler returns the market data handler.
func GetMarketDataHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := market.NewMarketService(ctx)
		data, err := svc.GetMarketData(r.Context())
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
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
