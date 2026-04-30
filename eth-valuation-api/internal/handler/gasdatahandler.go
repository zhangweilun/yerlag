package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/onchain"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetGasDataHandler returns the gas fee data handler.
func GetGasDataHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := onchain.NewGasService(ctx)
		data, err := svc.GetGasData(r.Context())
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
