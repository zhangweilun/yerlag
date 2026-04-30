package onchain

import (
	"context"
	"strconv"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// GasData holds the gas fee data response.
type GasData struct {
	CurrentAvgGwei     float64                 `json:"currentAvgGwei"`
	CurrentAvgUsd      float64                 `json:"currentAvgUsd"`
	DailyFeeRevenueEth float64                 `json:"dailyFeeRevenueEth"`
	DailyFeeRevenueUsd float64                 `json:"dailyFeeRevenueUsd"`
	PriorityFeeShare   float64                 `json:"priorityFeeShare"`  // Validator tip share (%)
	BaseFeeShare       float64                 `json:"baseFeeShare"`      // Base fee burn share (%)
	AnnualizedRevenue  float64                 `json:"annualizedRevenue"` // Annualized fee revenue (USD)
	PriceToFeeRatio    float64                 `json:"priceToFeeRatio"`   // Market-to-fee ratio
	GasHistory         []types.TimeSeriesPoint `json:"gasHistory"`
	IsHighFee          bool                    `json:"isHighFee"` // Gas > 50 Gwei
}

// GasService handles gas fee data logic.
type GasService struct {
	svcCtx    *svc.ServiceContext
	etherscan *fetcher.EtherscanClient
	coingecko *fetcher.CoinGeckoClient
}

// NewGasService creates a new GasService.
func NewGasService(svcCtx *svc.ServiceContext) *GasService {
	cfg := svcCtx.Config.DataSources
	return &GasService{
		svcCtx:    svcCtx,
		etherscan: fetcher.NewEtherscanClient(cfg.Etherscan.BaseURL, cfg.Etherscan.APIKey),
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetGasData fetches and computes gas fee data.
func (s *GasService) GetGasData(ctx context.Context) (*GasData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.GasData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "onchain:gas", ttl, func() (interface{}, error) {
		return s.fetchGasData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*GasData); ok {
		return data, nil
	}

	return s.fetchGasData(ctx)
}

func (s *GasService) fetchGasData(ctx context.Context) (*GasData, error) {
	// Fetch current gas oracle
	gasOracle, err := s.etherscan.GetGasOracle(ctx)
	if err != nil {
		return nil, err
	}

	// Parse gas prices
	proposeGas, _ := strconv.ParseFloat(gasOracle.ProposeGasPrice, 64)
	baseFee, _ := strconv.ParseFloat(gasOracle.SuggestBaseFee, 64)

	// Fetch current ETH price for USD conversion
	ethPrice := 0.0
	prices, priceErr := s.coingecko.GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
	if priceErr == nil {
		if ethPrices, ok := prices["ethereum"]; ok {
			ethPrice = ethPrices["usd"]
		}
	}

	// Current average gas in Gwei
	currentAvgGwei := proposeGas

	// Convert to USD (gas price in Gwei * 21000 gas units * ETH price / 1e9)
	currentAvgUsd := 0.0
	if ethPrice > 0 {
		currentAvgUsd = currentAvgGwei * 21000 * ethPrice / 1e9
	}

	// Estimate daily fee revenue
	// Approximate: average gas price * average daily gas used
	// Using a rough estimate of 15M gas per block * 7200 blocks/day
	avgDailyGasUsed := 15_000_000.0 * 7200.0
	dailyFeeRevenueEth := currentAvgGwei * avgDailyGasUsed / 1e9
	dailyFeeRevenueUsd := dailyFeeRevenueEth * ethPrice

	// Annualized revenue
	annualizedRevenue := dailyFeeRevenueUsd * 365

	// Fee share breakdown: base fee vs priority fee
	priorityFee := proposeGas - baseFee
	totalFee := proposeGas
	baseFeeShare := 0.0
	priorityFeeShare := 0.0
	if totalFee > 0 {
		baseFeeShare = (baseFee / totalFee) * 100
		priorityFeeShare = (priorityFee / totalFee) * 100
	}

	// Price-to-fee ratio (market cap / annualized fee revenue)
	priceToFeeRatio := 0.0
	marketData, mdErr := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if mdErr == nil && len(marketData) > 0 {
		marketCap := marketData[0].MarketCap
		if r := calc.PriceToFeeRatio(marketCap, annualizedRevenue); r != nil {
			priceToFeeRatio = *r
		}
	}

	// Fetch gas history for the last 30 days
	now := time.Now()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -30).Format("2006-01-02")

	var gasHistory []types.TimeSeriesPoint
	historyRaw, histErr := s.etherscan.GetDailyAvgGasPrice(ctx, startDate, endDate)
	if histErr == nil {
		gasHistory = parseDailyTimeSeries(historyRaw)
	}

	// High fee flag: Gas > 50 Gwei
	isHighFee := currentAvgGwei > 50

	return &GasData{
		CurrentAvgGwei:     currentAvgGwei,
		CurrentAvgUsd:      currentAvgUsd,
		DailyFeeRevenueEth: dailyFeeRevenueEth,
		DailyFeeRevenueUsd: dailyFeeRevenueUsd,
		PriorityFeeShare:   priorityFeeShare,
		BaseFeeShare:       baseFeeShare,
		AnnualizedRevenue:  annualizedRevenue,
		PriceToFeeRatio:    priceToFeeRatio,
		GasHistory:         gasHistory,
		IsHighFee:          isHighFee,
	}, nil
}
