package institutional

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// GrayscaleData holds the Grayscale trust data response.
type GrayscaleData struct {
	HoldingsETH     float64                 `json:"holdingsEth"`
	NAV             float64                 `json:"nav"`
	PremiumDiscount float64                 `json:"premiumDiscount"` // Premium/discount rate (%)
	PremiumHistory  []types.TimeSeriesPoint `json:"premiumHistory"`
}

// GrayscaleService handles Grayscale trust data logic.
type GrayscaleService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
}

// NewGrayscaleService creates a new GrayscaleService.
func NewGrayscaleService(svcCtx *svc.ServiceContext) *GrayscaleService {
	cfg := svcCtx.Config.DataSources
	return &GrayscaleService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetGrayscaleData fetches and computes Grayscale trust data.
func (s *GrayscaleService) GetGrayscaleData(ctx context.Context) (*GrayscaleData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.InstitutionalData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "institutional:grayscale", ttl, func() (interface{}, error) {
		return s.fetchGrayscaleData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*GrayscaleData); ok {
		return data, nil
	}

	return s.fetchGrayscaleData(ctx)
}

func (s *GrayscaleService) fetchGrayscaleData(ctx context.Context) (*GrayscaleData, error) {
	// In production, Grayscale data would come from specialized financial data APIs
	// (e.g., Grayscale's own API, Bloomberg, or financial data providers)
	// Using CoinGecko for ETH price as a baseline

	ethPrice := 0.0
	prices, priceErr := s.coingecko.GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
	if priceErr == nil {
		if ethPrices, ok := prices["ethereum"]; ok {
			ethPrice = ethPrices["usd"]
		}
	}

	// Placeholder values - would be populated from Grayscale/financial data APIs
	holdingsETH := 0.0
	nav := ethPrice * holdingsETH
	marketPrice := nav // In reality, market price differs from NAV

	// Calculate premium/discount using the calc utility
	premiumDiscount := 0.0
	if r := calc.PremiumDiscountRate(nav, marketPrice); r != nil {
		premiumDiscount = *r
	}

	return &GrayscaleData{
		HoldingsETH:     holdingsETH,
		NAV:             nav,
		PremiumDiscount: premiumDiscount,
		PremiumHistory:  []types.TimeSeriesPoint{},
	}, nil
}
