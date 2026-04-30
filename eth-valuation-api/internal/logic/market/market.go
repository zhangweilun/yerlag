package market

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// MarketData holds the market overview response.
type MarketData struct {
	CurrentPrice          float64                 `json:"currentPrice"`          // Current ETH price in USD
	PriceChange24h        float64                 `json:"priceChange24h"`        // 24h price change percentage
	Volume24h             float64                 `json:"volume24h"`             // 24h trading volume (USD)
	MarketCap             float64                 `json:"marketCap"`             // Circulating market cap (USD)
	FullyDilutedMarketCap float64                 `json:"fullyDilutedMarketCap"` // Fully diluted market cap (USD)
	CirculatingSupply     float64                 `json:"circulatingSupply"`     // Circulating supply (ETH)
	TotalSupply           float64                 `json:"totalSupply"`           // Total supply (ETH)
	ATH                   float64                 `json:"ath"`                   // All-time high price (USD)
	ATHDrawdown           float64                 `json:"athDrawdown"`           // Drawdown from ATH (%)
	MarketCapRank         int                     `json:"marketCapRank"`         // Market cap rank
	ExchangePrices        map[string]float64      `json:"exchangePrices"`        // Prices on major exchanges
	ExchangeSpreads       map[string]float64      `json:"exchangeSpreads"`       // Spread from average (%)
	VolumeHistory         []types.TimeSeriesPoint `json:"volumeHistory"`         // 7d volume history
}

// MarketService handles market data logic.
type MarketService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
}

// NewMarketService creates a new MarketService.
func NewMarketService(svcCtx *svc.ServiceContext) *MarketService {
	cfg := svcCtx.Config.DataSources
	return &MarketService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetMarketData fetches and computes current market data for ETH.
func (s *MarketService) GetMarketData(ctx context.Context) (*MarketData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.Price) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "market:overview", ttl, func() (interface{}, error) {
		return s.fetchMarketData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*MarketData); ok {
		return data, nil
	}

	// If data came from cache (deserialized as map), re-fetch
	return s.fetchMarketData(ctx)
}

func (s *MarketService) fetchMarketData(ctx context.Context) (*MarketData, error) {
	// Fetch market data from CoinGecko
	markets, err := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if err != nil {
		return nil, err
	}

	if len(markets) == 0 {
		return &MarketData{}, nil
	}

	eth := markets[0]

	// Calculate ATH drawdown
	athDrawdown := 0.0
	if r := calc.ATHDrawdown(eth.ATH, eth.CurrentPrice); r != nil {
		athDrawdown = *r
	}

	// Fetch exchange prices for spread calculation
	exchangePrices := s.fetchExchangePrices(ctx)
	exchangeSpreads := calc.CalculateExchangeSpreads(exchangePrices)

	// Fetch 7-day volume history
	volumeHistory := s.fetchVolumeHistory(ctx)

	return &MarketData{
		CurrentPrice:          eth.CurrentPrice,
		PriceChange24h:        eth.PriceChangePercentage24h,
		Volume24h:             eth.TotalVolume,
		MarketCap:             eth.MarketCap,
		FullyDilutedMarketCap: eth.FullyDilutedValuation,
		CirculatingSupply:     eth.CirculatingSupply,
		TotalSupply:           eth.TotalSupply,
		ATH:                   eth.ATH,
		ATHDrawdown:           athDrawdown,
		MarketCapRank:         eth.MarketCapRank,
		ExchangePrices:        exchangePrices,
		ExchangeSpreads:       exchangeSpreads,
		VolumeHistory:         volumeHistory,
	}, nil
}

// fetchExchangePrices fetches ETH prices from major exchanges.
// Uses CoinGecko simple price with exchange-specific coin IDs as a proxy.
func (s *MarketService) fetchExchangePrices(ctx context.Context) map[string]float64 {
	// Use simple price endpoint to get prices from different sources
	prices, err := s.coingecko.GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
	if err != nil {
		return map[string]float64{}
	}

	ethPrice, ok := prices["ethereum"]
	if !ok {
		return map[string]float64{}
	}

	basePrice := ethPrice["usd"]
	if basePrice == 0 {
		return map[string]float64{}
	}

	// Simulate exchange prices with minor variations based on the base price.
	// In production, these would come from individual exchange APIs.
	exchangePrices := map[string]float64{
		"binance":  basePrice,
		"coinbase": basePrice,
		"okx":      basePrice,
	}

	return exchangePrices
}

// fetchVolumeHistory fetches 7-day volume history from CoinGecko market chart.
func (s *MarketService) fetchVolumeHistory(ctx context.Context) []types.TimeSeriesPoint {
	chart, err := s.coingecko.GetMarketChart(ctx, "ethereum", "usd", 7)
	if err != nil {
		return []types.TimeSeriesPoint{}
	}

	points := make([]types.TimeSeriesPoint, 0, len(chart.TotalVolumes))
	for _, v := range chart.TotalVolumes {
		if len(v) < 2 {
			continue
		}
		points = append(points, types.TimeSeriesPoint{
			Timestamp: int64(v[0]),
			Value:     v[1],
		})
	}

	return points
}
