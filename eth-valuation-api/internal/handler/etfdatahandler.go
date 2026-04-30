package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/institutional"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetETFDataHandler returns the ETF holdings data handler.
func GetETFDataHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := institutional.NewETFService(ctx)
		data, err := svc.GetETFData(r.Context())
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		resp := types.SuccessResponse(data, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
			NextRefresh: time.Now().Add(1 * time.Hour).Unix(),
		})
		httpx.OkJson(w, resp)
	}
}
