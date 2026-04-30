package institutional

import (
	"context"
	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
	"time"
)

// ETFData holds the ETF holdings and flow data response.
type ETFData struct {
	ETFs                    []ETFHolding            `json:"etfs"`
	TotalHoldingsETH        float64                 `json:"totalHoldingsEth"`
	TotalHoldingsUSD        float64                 `json:"totalHoldingsUsd"`
	HoldingsPercentOfSupply float64                 `json:"holdingsPercentOfSupply"`
	CumulativeNetFlow       float64                 `json:"cumulativeNetFlow"`
	NetFlowHistory          []types.TimeSeriesPoint `json:"netFlowHistory"`
}

// ETFHolding holds data for a single ETF issuer.
type ETFHolding struct {
	Issuer          string                  `json:"issuer"`
	Ticker          string                  `json:"ticker"`
	HoldingsETH     float64                 `json:"holdingsEth"`
	HoldingsUSD     float64                 `json:"holdingsUsd"`
	DailyNetFlowUSD float64                 `json:"dailyNetFlowUsd"`
	MarketShare     float64                 `json:"marketShare"`
	FlowHistory     []types.TimeSeriesPoint `json:"flowHistory"`
}

// ETFService handles ETF data logic.
type ETFService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
}

// NewETFService creates a new ETFService.
func NewETFService(svcCtx *svc.ServiceContext) *ETFService {
	cfg := svcCtx.Config.DataSources
	return &ETFService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetETFData fetches and computes ETF holdings and flow data.
func (s *ETFService) GetETFData(ctx context.Context) (*ETFData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.InstitutionalData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "institutional:etf", ttl, func() (interface{}, error) {
		return s.fetchETFData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*ETFData); ok {
		return data, nil
	}

	return s.fetchETFData(ctx)
}

func (s *ETFService) fetchETFData(ctx context.Context) (*ETFData, error) {
	// Fetch current ETH price for USD conversion
	ethPrice := 0.0
	prices, priceErr := s.coingecko.GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
	if priceErr == nil {
		if ethPrices, ok := prices["ethereum"]; ok {
			ethPrice = ethPrices["usd"]
		}
	}

	// Fetch ETH total supply for holdings percentage
	totalSupply := 0.0
	marketData, mdErr := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if mdErr == nil && len(marketData) > 0 {
		totalSupply = marketData[0].CirculatingSupply
	}

	// Known ETH spot ETF issuers (data would come from specialized ETF data APIs in production)
	etfIssuers := []struct {
		Issuer string
		Ticker string
	}{
		{"BlackRock", "ETHA"},
		{"Fidelity", "FETH"},
		{"Grayscale", "ETHE"},
		{"Bitwise", "ETHW"},
		{"VanEck", "ETHV"},
		{"Invesco", "QETH"},
		{"21Shares", "CETH"},
		{"Franklin Templeton", "EZET"},
	}

	var etfs []ETFHolding
	var totalHoldingsETH float64
	var cumulativeNetFlow float64

	holdingsValues := make([]float64, len(etfIssuers))
	for i, issuer := range etfIssuers {
		// In production, these values would come from ETF data providers
		// Using placeholder structure that would be populated by real API data
		holding := ETFHolding{
			Issuer:          issuer.Issuer,
			Ticker:          issuer.Ticker,
			HoldingsETH:     0,
			HoldingsUSD:     0,
			DailyNetFlowUSD: 0,
			MarketShare:     0,
			FlowHistory:     []types.TimeSeriesPoint{},
		}
		etfs = append(etfs, holding)
		totalHoldingsETH += holding.HoldingsETH
		cumulativeNetFlow += holding.DailyNetFlowUSD
		holdingsValues[i] = holding.HoldingsETH
	}

	// Calculate market shares
	shares := calc.CalculateShares(holdingsValues)
	for i := range etfs {
		if i < len(shares) {
			etfs[i].MarketShare = shares[i]
		}
	}

	// Calculate total holdings in USD
	totalHoldingsUSD := totalHoldingsETH * ethPrice

	// Calculate holdings as percentage of supply
	holdingsPercent := 0.0
	if r := calc.ETFHoldingsPercentage(totalHoldingsETH, totalSupply); r != nil {
		holdingsPercent = *r
	}

	return &ETFData{
		ETFs:                    etfs,
		TotalHoldingsETH:        totalHoldingsETH,
		TotalHoldingsUSD:        totalHoldingsUSD,
		HoldingsPercentOfSupply: holdingsPercent,
		CumulativeNetFlow:       cumulativeNetFlow,
		NetFlowHistory:          []types.TimeSeriesPoint{},
	}, nil
}
