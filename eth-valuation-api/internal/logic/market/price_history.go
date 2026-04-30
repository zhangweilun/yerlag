package market

import (
	"context"
	"fmt"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/svc"
)

// OHLCVPoint represents a single OHLCV candlestick data point.
type OHLCVPoint struct {
	Timestamp int64   `json:"timestamp"` // Unix timestamp (ms)
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
}

// PriceHistoryData holds the price history response.
type PriceHistoryData struct {
	TimeRange string       `json:"timeRange"` // "1d", "1w", "1m", "3m", "1y", "all"
	OHLCV     []OHLCVPoint `json:"ohlcv"`     // OHLCV candlestick data
	StartTime int64        `json:"startTime"` // Start timestamp (ms)
	EndTime   int64        `json:"endTime"`   // End timestamp (ms)
}

// PriceHistoryService handles historical price data logic.
type PriceHistoryService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
}

// NewPriceHistoryService creates a new PriceHistoryService.
func NewPriceHistoryService(svcCtx *svc.ServiceContext) *PriceHistoryService {
	cfg := svcCtx.Config.DataSources
	return &PriceHistoryService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetPriceHistory fetches OHLCV price history for the given time range.
// Supported ranges: "1d", "1w", "1m", "3m", "1y", "all"
func (s *PriceHistoryService) GetPriceHistory(ctx context.Context, timeRange string) (*PriceHistoryData, error) {
	days, err := timeRangeToDays(timeRange)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("market:price_history:%s", timeRange)
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.Price) * time.Second

	result, fetchErr := s.svcCtx.DataFetcher.Fetch(ctx, cacheKey, ttl, func() (interface{}, error) {
		return s.fetchPriceHistory(ctx, timeRange, days)
	})
	if fetchErr != nil {
		return nil, fetchErr
	}

	if data, ok := result.Data.(*PriceHistoryData); ok {
		return data, nil
	}

	// If data came from cache (deserialized as map), re-fetch
	return s.fetchPriceHistory(ctx, timeRange, days)
}

func (s *PriceHistoryService) fetchPriceHistory(ctx context.Context, timeRange string, days int) (*PriceHistoryData, error) {
	ohlcvRaw, err := s.coingecko.GetOHLCV(ctx, "ethereum", "usd", days)
	if err != nil {
		return nil, err
	}

	ohlcv := make([]OHLCVPoint, 0, len(ohlcvRaw))
	for _, point := range ohlcvRaw {
		ohlcv = append(ohlcv, OHLCVPoint{
			Timestamp: point.Timestamp,
			Open:      point.Open,
			High:      point.High,
			Low:       point.Low,
			Close:     point.Close,
		})
	}

	now := time.Now()
	startTime := now.AddDate(0, 0, -days).UnixMilli()
	endTime := now.UnixMilli()

	return &PriceHistoryData{
		TimeRange: timeRange,
		OHLCV:     ohlcv,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// timeRangeToDays converts a time range string to the number of days.
func timeRangeToDays(timeRange string) (int, error) {
	switch timeRange {
	case "1d":
		return 1, nil
	case "1w":
		return 7, nil
	case "1m":
		return 30, nil
	case "3m":
		return 90, nil
	case "1y":
		return 365, nil
	case "all":
		return 1825, nil // ~5 years max
	default:
		return 0, fmt.Errorf("invalid time range: %s (supported: 1d, 1w, 1m, 3m, 1y, all)", timeRange)
	}
}
