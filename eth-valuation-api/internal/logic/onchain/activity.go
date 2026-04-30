package onchain

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// ActivityData holds the on-chain activity data response.
type ActivityData struct {
	DailyActiveAddresses int64                   `json:"dailyActiveAddresses"`
	DAAMovingAvg7d       float64                 `json:"daaMovingAvg7d"`
	DailyTransactions    int64                   `json:"dailyTransactions"`
	DailyNewAddresses    int64                   `json:"dailyNewAddresses"`
	NVTRatio             float64                 `json:"nvtRatio"`
	NVTHistoricalMedian  float64                 `json:"nvtHistoricalMedian"`
	NVTPercentile        float64                 `json:"nvtPercentile"`
	NVTSignal            string                  `json:"nvtSignal"` // "overvalued" | "undervalued" | "neutral"
	L2Comparison         []L2TransactionData     `json:"l2Comparison"`
	TransactionHistory   []types.TimeSeriesPoint `json:"transactionHistory"`
}

// L2TransactionData holds L2 network transaction comparison data.
type L2TransactionData struct {
	Network           string `json:"network"` // "Arbitrum" | "Optimism" | "Base" | "zkSync"
	DailyTransactions int64  `json:"dailyTransactions"`
}

// ActivityService handles on-chain activity data logic.
type ActivityService struct {
	svcCtx    *svc.ServiceContext
	glassnode *fetcher.GlassnodeClient
	coingecko *fetcher.CoinGeckoClient
	etherscan *fetcher.EtherscanClient
}

// NewActivityService creates a new ActivityService.
func NewActivityService(svcCtx *svc.ServiceContext) *ActivityService {
	cfg := svcCtx.Config.DataSources
	return &ActivityService{
		svcCtx:    svcCtx,
		glassnode: fetcher.NewGlassnodeClient(cfg.Glassnode.BaseURL, cfg.Glassnode.APIKey),
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
		etherscan: fetcher.NewEtherscanClient(cfg.Etherscan.BaseURL, cfg.Etherscan.APIKey),
	}
}

// GetActivityData fetches and computes on-chain activity data.
func (s *ActivityService) GetActivityData(ctx context.Context) (*ActivityData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.OnChainMetrics) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "onchain:activity", ttl, func() (interface{}, error) {
		return s.fetchActivityData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*ActivityData); ok {
		return data, nil
	}

	return s.fetchActivityData(ctx)
}

func (s *ActivityService) fetchActivityData(ctx context.Context) (*ActivityData, error) {
	now := time.Now()
	since := now.AddDate(0, 0, -30).Unix()
	until := now.Unix()

	params := fetcher.GlassnodeMetricParams{
		Asset:    "ETH",
		Since:    since,
		Until:    until,
		Interval: "24h",
	}

	// Fetch daily active addresses
	var dailyActiveAddresses int64
	var daaHistory []float64
	activeAddrData, err := s.glassnode.GetActiveAddresses(ctx, params)
	if err == nil && len(activeAddrData) > 0 {
		dailyActiveAddresses = int64(activeAddrData[len(activeAddrData)-1].Value)
		for _, dp := range activeAddrData {
			daaHistory = append(daaHistory, dp.Value)
		}
	}

	// Calculate 7-day moving average of DAA
	daaMovingAvg7d := 0.0
	if ma := calc.MovingAverage(daaHistory, 7); ma != nil {
		daaMovingAvg7d = *ma
	}

	// Fetch transaction count
	var dailyTransactions int64
	var transactionHistory []types.TimeSeriesPoint
	txData, err := s.glassnode.GetTransactionCount(ctx, params)
	if err == nil && len(txData) > 0 {
		dailyTransactions = int64(txData[len(txData)-1].Value)
		for _, dp := range txData {
			transactionHistory = append(transactionHistory, types.TimeSeriesPoint{
				Timestamp: dp.Timestamp,
				Value:     dp.Value,
			})
		}
	}

	// Fetch new addresses
	var dailyNewAddresses int64
	newAddrData, err := s.glassnode.GetNewAddresses(ctx, params)
	if err == nil && len(newAddrData) > 0 {
		dailyNewAddresses = int64(newAddrData[len(newAddrData)-1].Value)
	}

	// Calculate NVT Ratio
	nvtRatio := 0.0
	nvtPercentile := 0.0
	nvtSignal := "neutral"
	var nvtHistoricalMedian float64

	// Get market cap and transfer volume for NVT
	marketData, mdErr := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if mdErr == nil && len(marketData) > 0 {
		marketCap := marketData[0].MarketCap

		// Fetch transfer volume for NVT calculation
		volumeData, volErr := s.glassnode.GetTransferVolume(ctx, params)
		if volErr == nil && len(volumeData) > 0 {
			dailyVolume := volumeData[len(volumeData)-1].Value
			if r := calc.NVTRatio(marketCap, dailyVolume); r != nil {
				nvtRatio = *r
			}

			// Build NVT history for percentile calculation
			var nvtHistory []float64
			for _, dp := range volumeData {
				if dp.Value > 0 {
					nvtVal := marketCap / dp.Value
					nvtHistory = append(nvtHistory, nvtVal)
				}
			}

			if len(nvtHistory) > 0 {
				nvtPercentile = calc.CalculatePercentile(nvtHistory, nvtRatio)
				nvtSignal = calc.ClassifySignal(nvtPercentile)
				nvtHistoricalMedian = calculateMedian(nvtHistory)
			}
		}
	}

	// L2 comparison data (placeholder values - would come from L2-specific APIs)
	l2Comparison := []L2TransactionData{
		{Network: "Arbitrum", DailyTransactions: 0},
		{Network: "Optimism", DailyTransactions: 0},
		{Network: "Base", DailyTransactions: 0},
		{Network: "zkSync", DailyTransactions: 0},
	}

	return &ActivityData{
		DailyActiveAddresses: dailyActiveAddresses,
		DAAMovingAvg7d:       daaMovingAvg7d,
		DailyTransactions:    dailyTransactions,
		DailyNewAddresses:    dailyNewAddresses,
		NVTRatio:             nvtRatio,
		NVTHistoricalMedian:  nvtHistoricalMedian,
		NVTPercentile:        nvtPercentile,
		NVTSignal:            nvtSignal,
		L2Comparison:         l2Comparison,
		TransactionHistory:   transactionHistory,
	}, nil
}
