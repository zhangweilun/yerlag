package handler

import (
	"net/http"
	"time"

	"eth-valuation-api/internal/logic/alert"
	"eth-valuation-api/internal/logic/market"
	"eth-valuation-api/internal/logic/valuation"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// OverviewData holds the aggregated dashboard overview data.
type OverviewData struct {
	CurrentPrice    float64 `json:"currentPrice"`
	PriceChange24h  float64 `json:"priceChange24h"`
	MarketCap       float64 `json:"marketCap"`
	MarketCapRank   int     `json:"marketCapRank"`
	ValuationScore  float64 `json:"valuationScore"`
	ValuationStatus string  `json:"valuationStatus"`
	ActiveAlerts    int     `json:"activeAlerts"`
}

// GetOverviewHandler returns the dashboard overview data handler.
func GetOverviewHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overview := &OverviewData{}

		// Fetch market data for price, market cap, rank
		marketSvc := market.NewMarketService(ctx)
		marketData, err := marketSvc.GetMarketData(r.Context())
		if err == nil && marketData != nil {
			overview.CurrentPrice = marketData.CurrentPrice
			overview.PriceChange24h = marketData.PriceChange24h
			overview.MarketCap = marketData.MarketCap
			overview.MarketCapRank = marketData.MarketCapRank
		}

		// Fetch valuation score
		valuationSvc := valuation.NewValuationService(ctx)
		valuationData, err := valuationSvc.GetValuation(r.Context())
		if err == nil && valuationData != nil {
			overview.ValuationScore = valuationData.Overall
			overview.ValuationStatus = valuationData.Status
		}

		// Fetch active alerts count
		alertSvc := alert.NewAlertService(ctx.DB)
		activeAlerts, err := alertSvc.GetActiveAlerts()
		if err == nil {
			overview.ActiveAlerts = len(activeAlerts)
		}

		resp := types.SuccessResponse(overview, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
			NextRefresh: time.Now().Add(10 * time.Second).Unix(),
		})
		httpx.OkJson(w, resp)
	}
}
