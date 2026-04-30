package handler

import (
	"net/http"
	"strconv"
	"time"

	"eth-valuation-api/internal/logic/valuation"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetDCFValuationHandler returns the DCF valuation model handler.
func GetDCFValuationHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse DCF assumptions from query parameters with defaults
		discountRate := parseFloatParam(r, "discountRate", 0.12)
		growthRate := parseFloatParam(r, "growthRate", 0.15)
		terminalGrowthRate := parseFloatParam(r, "terminalGrowthRate", 0.03)
		projectionYears := parseIntParam(r, "projectionYears", 10)

		assumptions := valuation.DCFAssumptions{
			DiscountRate:       discountRate,
			GrowthRate:         growthRate,
			TerminalGrowthRate: terminalGrowthRate,
			ProjectionYears:    projectionYears,
		}

		svc := valuation.NewValuationService(ctx)
		data, err := svc.GetDCFValuation(r.Context(), assumptions)
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

// parseFloatParam extracts a float64 query parameter with a default value.
func parseFloatParam(r *http.Request, key string, defaultVal float64) float64 {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return f
}

// parseIntParam extracts an int query parameter with a default value.
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
