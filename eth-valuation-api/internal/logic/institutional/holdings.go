package institutional

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// InstitutionalHoldings holds the institutional holdings data response.
type InstitutionalHoldings struct {
	Institutions      []InstitutionEntry      `json:"institutions"`
	CMEFuturesOI      float64                 `json:"cmeFuturesOI"`
	CMEFuturesHistory []types.TimeSeriesPoint `json:"cmeFuturesHistory"`
}

// InstitutionEntry holds data for a single institution.
type InstitutionEntry struct {
	Name        string  `json:"name"`
	HoldingsETH float64 `json:"holdingsEth"`
	HoldingsUSD float64 `json:"holdingsUsd"`
}

// HoldingsService handles institutional holdings data logic.
type HoldingsService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
}

// NewHoldingsService creates a new HoldingsService.
func NewHoldingsService(svcCtx *svc.ServiceContext) *HoldingsService {
	cfg := svcCtx.Config.DataSources
	return &HoldingsService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetInstitutionalHoldings fetches and computes institutional holdings data.
func (s *HoldingsService) GetInstitutionalHoldings(ctx context.Context) (*InstitutionalHoldings, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.InstitutionalData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "institutional:holdings", ttl, func() (interface{}, error) {
		return s.fetchHoldingsData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*InstitutionalHoldings); ok {
		return data, nil
	}

	return s.fetchHoldingsData(ctx)
}

func (s *HoldingsService) fetchHoldingsData(ctx context.Context) (*InstitutionalHoldings, error) {
	// Fetch current ETH price for USD conversion
	ethPrice := 0.0
	prices, priceErr := s.coingecko.GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
	if priceErr == nil {
		if ethPrices, ok := prices["ethereum"]; ok {
			ethPrice = ethPrices["usd"]
		}
	}

	// Known institutional holders (data would come from SEC filings, on-chain analysis, etc.)
	// Placeholder structure for real data integration
	knownInstitutions := []struct {
		Name        string
		HoldingsETH float64
	}{
		{"Ethereum Foundation", 0},
		{"Paradigm", 0},
		{"a16z Crypto", 0},
		{"Galaxy Digital", 0},
		{"Coinbase Institutional", 0},
	}

	institutions := make([]InstitutionEntry, 0, len(knownInstitutions))
	for _, inst := range knownInstitutions {
		institutions = append(institutions, InstitutionEntry{
			Name:        inst.Name,
			HoldingsETH: inst.HoldingsETH,
			HoldingsUSD: inst.HoldingsETH * ethPrice,
		})
	}

	// CME Ethereum Futures Open Interest
	// In production, this would come from CME data feeds or financial data APIs
	cmeFuturesOI := 0.0

	return &InstitutionalHoldings{
		Institutions:      institutions,
		CMEFuturesOI:      cmeFuturesOI,
		CMEFuturesHistory: []types.TimeSeriesPoint{},
	}, nil
}
