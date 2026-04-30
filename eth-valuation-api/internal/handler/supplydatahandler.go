package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/onchain"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetSupplyDataHandler returns the ETH supply data handler.
func GetSupplyDataHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := onchain.NewSupplyService(ctx)
		data, err := svc.GetSupplyData(r.Context())
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		resp := types.SuccessResponse(data, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
			NextRefresh: time.Now().Add(5 * time.Minute).Unix(),
		})
		httpx.OkJson(w, resp)
	}
}
