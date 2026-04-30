package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/valuation"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// distributionReq holds the path parameter for the distribution endpoint.
type distributionReq struct {
	Metric string `path:"metric"`
}

// GetDistributionHandler returns the valuation metric distribution handler.
func GetDistributionHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req distributionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid metric parameter"))
			return
		}

		if req.Metric == "" {
			httpx.OkJson(w, types.ErrorResponse(400, "metric parameter is required"))
			return
		}

		svc := valuation.NewValuationService(ctx)
		data, err := svc.GetDistribution(r.Context(), req.Metric)
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
