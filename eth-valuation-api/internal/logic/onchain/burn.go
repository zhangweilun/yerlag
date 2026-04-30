package onchain

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// BurnData holds the EIP-1559 burn data response.
type BurnData struct {
	Daily              float64                 `json:"daily"`              // 24h burn amount (ETH)
	Weekly             float64                 `json:"weekly"`             // 7d burn amount
	Monthly            float64                 `json:"monthly"`            // 30d burn amount
	Cumulative         float64                 `json:"cumulative"`         // Cumulative burn since EIP-1559
	AnnualizedBurnRate float64                 `json:"annualizedBurnRate"` // Annualized burn rate (%)
	DailyHistory       []types.TimeSeriesPoint `json:"dailyHistory"`       // Daily burn history
}

// BurnService handles EIP-1559 burn data logic.
type BurnService struct {
	svcCtx    *svc.ServiceContext
	etherscan *fetcher.EtherscanClient
	glassnode *fetcher.GlassnodeClient
}

// NewBurnService creates a new BurnService.
func NewBurnService(svcCtx *svc.ServiceContext) *BurnService {
	cfg := svcCtx.Config.DataSources
	return &BurnService{
		svcCtx:    svcCtx,
		etherscan: fetcher.NewEtherscanClient(cfg.Etherscan.BaseURL, cfg.Etherscan.APIKey),
		glassnode: fetcher.NewGlassnodeClient(cfg.Glassnode.BaseURL, cfg.Glassnode.APIKey),
	}
}

// GetBurnData fetches and computes EIP-1559 burn data.
func (s *BurnService) GetBurnData(ctx context.Context) (*BurnData, error) {
	now := time.Now()
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.OnChainMetrics) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "onchain:burn", ttl, func() (interface{}, error) {
		return s.fetchBurnData(ctx, now)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*BurnData); ok {
		return data, nil
	}

	// If data came from cache (deserialized as map), re-fetch
	return s.fetchBurnData(ctx, now)
}

func (s *BurnService) fetchBurnData(ctx context.Context, now time.Time) (*BurnData, error) {
	// Fetch ETH supply data which includes burnt fees
	supplyData, err := s.etherscan.GetEthSupply(ctx)
	if err != nil {
		return nil, err
	}

	// Parse cumulative burn from supply data (BurntFees is in wei)
	cumulativeBurn := parseWeiToEth(supplyData.BurntFees)

	// Fetch daily burn history for the last 30 days
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -30).Format("2006-01-02")

	dailyBurnRaw, err := s.etherscan.GetDailyBurntFees(ctx, startDate, endDate)
	if err != nil {
		// Non-fatal: continue with partial data
		dailyBurnRaw = nil
	}

	dailyHistory := parseDailyTimeSeries(dailyBurnRaw)

	// Calculate period burns from history
	daily := sumLastNDays(dailyHistory, 1)
	weekly := sumLastNDays(dailyHistory, 7)
	monthly := sumLastNDays(dailyHistory, 30)

	// Get total supply for annualized burn rate calculation
	totalSupply := parseWeiToEth(supplyData.EthSupply)

	// Calculate annualized burn rate
	annualizedBurnRate := 0.0
	if r := calc.AnnualizedBurnRate(daily, totalSupply); r != nil {
		annualizedBurnRate = *r
	}

	return &BurnData{
		Daily:              daily,
		Weekly:             weekly,
		Monthly:            monthly,
		Cumulative:         cumulativeBurn,
		AnnualizedBurnRate: annualizedBurnRate,
		DailyHistory:       dailyHistory,
	}, nil
}
