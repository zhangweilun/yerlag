package macro

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// ETHBTCData holds the ETH/BTC correlation and relative valuation data.
type ETHBTCData struct {
	ETHBTCPrice         float64                 `json:"ethBtcPrice"`
	ETHBTCHistory       []types.TimeSeriesPoint `json:"ethBtcHistory"`
	Correlation30d      float64                 `json:"correlation30d"`
	Correlation90d      float64                 `json:"correlation90d"`
	ETHDominance        float64                 `json:"ethDominance"`
	ETHDominanceHistory []types.TimeSeriesPoint `json:"ethDominanceHistory"`
	ETHBTCPercentile    float64                 `json:"ethBtcPercentile"`
	ETHBTCSignal        string                  `json:"ethBtcSignal"` // "eth_undervalued" | "eth_overvalued" | "neutral"
}

// ETHBTCService handles ETH/BTC correlation data logic.
type ETHBTCService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
}

// NewETHBTCService creates a new ETHBTCService.
func NewETHBTCService(svcCtx *svc.ServiceContext) *ETHBTCService {
	cfg := svcCtx.Config.DataSources
	return &ETHBTCService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetETHBTCData fetches and computes ETH/BTC correlation data.
func (s *ETHBTCService) GetETHBTCData(ctx context.Context) (*ETHBTCData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.MacroData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "macro:ethbtc", ttl, func() (interface{}, error) {
		return s.fetchETHBTCData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*ETHBTCData); ok {
		return data, nil
	}

	return s.fetchETHBTCData(ctx)
}

func (s *ETHBTCService) fetchETHBTCData(ctx context.Context) (*ETHBTCData, error) {
	// Fetch ETH and BTC prices for ETH/BTC ratio
	prices, priceErr := s.coingecko.GetSimplePrice(ctx, []string{"ethereum", "bitcoin"}, []string{"usd", "btc"})

	var ethBTCPrice float64
	if priceErr == nil {
		if ethPrices, ok := prices["ethereum"]; ok {
			ethBTCPrice = ethPrices["btc"]
		}
	}

	// Fetch historical ETH price data (365 days) for correlation calculation
	ethChart, ethChartErr := s.coingecko.GetMarketChart(ctx, "ethereum", "usd", 365)
	btcChart, btcChartErr := s.coingecko.GetMarketChart(ctx, "bitcoin", "usd", 365)

	// Build ETH/BTC price history
	ethBTCHistory := s.buildETHBTCHistory(ethChart, btcChart, ethChartErr, btcChartErr)

	// Calculate 30-day and 90-day rolling correlations
	correlation30d := s.calculateCorrelation(ethChart, btcChart, ethChartErr, btcChartErr, 30)
	correlation90d := s.calculateCorrelation(ethChart, btcChart, ethChartErr, btcChartErr, 90)

	// Fetch global data for ETH Dominance
	ethDominance := 0.0
	globalData, globalErr := s.coingecko.GetGlobalData(ctx)
	if globalErr == nil && globalData != nil {
		if ethPct, ok := globalData.Data.MarketCapPercentage["eth"]; ok {
			ethDominance = ethPct
		}
	}

	// Build ETH dominance history from market cap data
	ethDominanceHistory := s.buildDominanceHistory(ethChart, ethChartErr)

	// Calculate ETH/BTC percentile from historical data
	ethBTCPercentile := s.calculateETHBTCPercentile(ethBTCHistory, ethBTCPrice)

	// Classify signal based on percentile
	ethBTCSignal := calc.ClassifyETHBTCSignal(ethBTCPercentile)

	return &ETHBTCData{
		ETHBTCPrice:         ethBTCPrice,
		ETHBTCHistory:       ethBTCHistory,
		Correlation30d:      correlation30d,
		Correlation90d:      correlation90d,
		ETHDominance:        ethDominance,
		ETHDominanceHistory: ethDominanceHistory,
		ETHBTCPercentile:    ethBTCPercentile,
		ETHBTCSignal:        ethBTCSignal,
	}, nil
}

// buildETHBTCHistory constructs the ETH/BTC price history from individual price charts.
func (s *ETHBTCService) buildETHBTCHistory(ethChart, btcChart *fetcher.CoinGeckoMarketChart, ethErr, btcErr error) []types.TimeSeriesPoint {
	if ethErr != nil || btcErr != nil || ethChart == nil || btcChart == nil {
		return []types.TimeSeriesPoint{}
	}

	// Build a map of BTC prices by timestamp (rounded to nearest hour)
	btcPriceMap := make(map[int64]float64)
	for _, point := range btcChart.Prices {
		if len(point) < 2 {
			continue
		}
		ts := int64(point[0]) / 3600000 * 3600000 // Round to nearest hour
		btcPriceMap[ts] = point[1]
	}

	history := make([]types.TimeSeriesPoint, 0)
	for _, point := range ethChart.Prices {
		if len(point) < 2 {
			continue
		}
		ts := int64(point[0]) / 3600000 * 3600000
		ethPrice := point[1]
		if btcPrice, ok := btcPriceMap[ts]; ok && btcPrice > 0 {
			history = append(history, types.TimeSeriesPoint{
				Timestamp: int64(point[0]) / 1000, // Convert ms to seconds
				Value:     ethPrice / btcPrice,
			})
		}
	}

	return history
}

// calculateCorrelation computes the Pearson correlation between ETH and BTC prices
// over the specified window (in days).
func (s *ETHBTCService) calculateCorrelation(ethChart, btcChart *fetcher.CoinGeckoMarketChart, ethErr, btcErr error, windowDays int) float64 {
	if ethErr != nil || btcErr != nil || ethChart == nil || btcChart == nil {
		return 0
	}

	// Extract daily price series (use the last windowDays data points)
	ethPrices := extractPriceSeries(ethChart.Prices)
	btcPrices := extractPriceSeries(btcChart.Prices)

	// Align series to same length
	minLen := len(ethPrices)
	if len(btcPrices) < minLen {
		minLen = len(btcPrices)
	}

	if minLen < windowDays {
		// Not enough data for the requested window
		if minLen < 2 {
			return 0
		}
		windowDays = minLen
	}

	// Use the most recent windowDays data points
	ethSlice := ethPrices[len(ethPrices)-windowDays:]
	btcSlice := btcPrices[len(btcPrices)-windowDays:]

	r := calc.PearsonCorrelation(ethSlice, btcSlice)
	if r == nil {
		return 0
	}
	return *r
}

// buildDominanceHistory builds an approximate ETH dominance history.
// In production, this would come from a dedicated API endpoint.
func (s *ETHBTCService) buildDominanceHistory(ethChart *fetcher.CoinGeckoMarketChart, ethErr error) []types.TimeSeriesPoint {
	if ethErr != nil || ethChart == nil {
		return []types.TimeSeriesPoint{}
	}

	// Use market cap data as a proxy for dominance history
	// In production, we'd fetch total crypto market cap history and compute the ratio
	history := make([]types.TimeSeriesPoint, 0)
	for _, point := range ethChart.MarketCaps {
		if len(point) < 2 {
			continue
		}
		history = append(history, types.TimeSeriesPoint{
			Timestamp: int64(point[0]) / 1000,
			Value:     point[1],
		})
	}

	return history
}

// calculateETHBTCPercentile computes the historical percentile of the current ETH/BTC ratio.
func (s *ETHBTCService) calculateETHBTCPercentile(history []types.TimeSeriesPoint, currentRatio float64) float64 {
	if len(history) == 0 || currentRatio == 0 {
		return 50 // Default to neutral if no data
	}

	values := make([]float64, len(history))
	for i, point := range history {
		values[i] = point.Value
	}

	return calc.CalculatePercentile(values, currentRatio)
}

// extractPriceSeries extracts the price values from a CoinGecko market chart price array.
func extractPriceSeries(prices [][]float64) []float64 {
	result := make([]float64, 0, len(prices))
	for _, point := range prices {
		if len(point) >= 2 {
			result = append(result, point[1])
		}
	}
	return result
}
