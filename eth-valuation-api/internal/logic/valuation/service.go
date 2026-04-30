package valuation

import (
	"context"
	"math"
	"sort"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
)

// ValuationService provides valuation computation via the service context.
type ValuationService struct {
	svcCtx    *svc.ServiceContext
	coingecko *fetcher.CoinGeckoClient
	glassnode *fetcher.GlassnodeClient
}

// NewValuationService creates a new ValuationService.
func NewValuationService(svcCtx *svc.ServiceContext) *ValuationService {
	cfg := svcCtx.Config.DataSources
	return &ValuationService{
		svcCtx:    svcCtx,
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
		glassnode: fetcher.NewGlassnodeClient(cfg.Glassnode.BaseURL, cfg.Glassnode.APIKey),
	}
}

// GetValuation computes the full valuation score using available data.
func (s *ValuationService) GetValuation(ctx context.Context) (*ValuationScore, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.OnChainMetrics) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "valuation:score", ttl, func() (interface{}, error) {
		return s.computeValuation(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*ValuationScore); ok {
		return data, nil
	}

	return s.computeValuation(ctx)
}

// GetDCFValuation computes the DCF valuation with custom assumptions.
func (s *ValuationService) GetDCFValuation(ctx context.Context, assumptions DCFAssumptions) (*DCFResult, error) {
	// Fetch current market data for DCF input
	markets, err := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if err != nil {
		return nil, err
	}

	currentPrice := 0.0
	totalSupply := 120_000_000.0 // default
	if len(markets) > 0 {
		currentPrice = markets[0].CurrentPrice
		if markets[0].TotalSupply > 0 {
			totalSupply = markets[0].TotalSupply
		}
	}

	// Estimate annual cash flow from fee revenue (simplified)
	annualCashFlow := totalSupply * 0.005 * currentPrice // ~0.5% of supply value as proxy

	input := DCFInput{
		AnnualCashFlow:     annualCashFlow,
		CurrentPrice:       currentPrice,
		TotalSupply:        totalSupply,
		DiscountRate:       assumptions.DiscountRate,
		GrowthRate:         assumptions.GrowthRate,
		TerminalGrowthRate: assumptions.TerminalGrowthRate,
		ProjectionYears:    assumptions.ProjectionYears,
	}

	result := CalculateDCF(input)
	return &result, nil
}

// DistributionData holds the historical distribution of a metric.
type DistributionData struct {
	Values            []float64       `json:"values"`
	Percentiles       map[int]float64 `json:"percentiles"`
	CurrentValue      float64         `json:"currentValue"`
	CurrentPercentile float64         `json:"currentPercentile"`
	MetricName        string          `json:"metricName"`
}

// GetDistribution computes the historical distribution for a given metric.
func (s *ValuationService) GetDistribution(ctx context.Context, metric string) (*DistributionData, error) {
	// Generate synthetic historical distribution based on metric type
	// In production, this would fetch real historical data from Glassnode/other sources
	history := s.generateMetricHistory(metric)
	currentValue := s.getCurrentMetricValue(metric, history)

	percentile := calc.CalculatePercentile(history, currentValue)

	// Calculate key percentiles
	percentiles := calculatePercentiles(history, []int{10, 25, 50, 75, 90})

	return &DistributionData{
		Values:            history,
		Percentiles:       percentiles,
		CurrentValue:      currentValue,
		CurrentPercentile: percentile,
		MetricName:        metric,
	}, nil
}

// computeValuation builds the full valuation input and runs the engine.
func (s *ValuationService) computeValuation(ctx context.Context) (*ValuationScore, error) {
	// Fetch market data
	markets, err := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if err != nil {
		return nil, err
	}

	currentPrice := 0.0
	marketCap := 0.0
	totalSupply := 120_000_000.0
	if len(markets) > 0 {
		currentPrice = markets[0].CurrentPrice
		marketCap = markets[0].MarketCap
		if markets[0].TotalSupply > 0 {
			totalSupply = markets[0].TotalSupply
		}
	}

	// Build valuation input with available data
	// Use reasonable defaults/estimates where real data isn't available
	input := ValuationInput{
		MVRV: MVRVInput{
			MarketValue:   marketCap,
			RealizedValue: marketCap * 0.7, // estimate
			History:       generateHistoricalRatios(1.0, 3.0, 100),
		},
		PriceToFee: PriceToFeeInput{
			MarketCap:            marketCap,
			AnnualizedFeeRevenue: marketCap * 0.02, // estimate ~2% fee yield
			History:              generateHistoricalRatios(20, 100, 100),
		},
		DCF: DCFInput{
			AnnualCashFlow:     totalSupply * 0.005 * currentPrice,
			CurrentPrice:       currentPrice,
			TotalSupply:        totalSupply,
			DiscountRate:       0.12,
			GrowthRate:         0.15,
			TerminalGrowthRate: 0.03,
			ProjectionYears:    10,
		},
		S2F: StockToFlowInput{
			CurrentStock: totalSupply,
			AnnualFlow:   totalSupply * 0.005, // ~0.5% net issuance
			CurrentPrice: currentPrice,
		},
		NVT: NVTInput{
			MarketCap:   marketCap,
			DailyVolume: marketCap * 0.01, // estimate ~1% daily volume
			History:     generateHistoricalRatios(30, 150, 100),
		},
		ETHBTC: ETHBTCInput{
			CurrentRatio: 0.05, // default ETH/BTC ratio
			History:      generateHistoricalRatios(0.02, 0.08, 100),
		},
	}

	score := CalculateValuation(input)
	return &score, nil
}

// generateMetricHistory generates synthetic historical data for a metric.
func (s *ValuationService) generateMetricHistory(metric string) []float64 {
	switch metric {
	case "mvrv":
		return generateHistoricalRatios(0.5, 4.0, 200)
	case "nvt":
		return generateHistoricalRatios(20, 200, 200)
	case "pf", "price_to_fee":
		return generateHistoricalRatios(15, 120, 200)
	case "s2f", "stock_to_flow":
		return generateHistoricalRatios(50, 500, 200)
	case "ethbtc", "eth_btc":
		return generateHistoricalRatios(0.01, 0.1, 200)
	default:
		return generateHistoricalRatios(0, 100, 200)
	}
}

// getCurrentMetricValue returns the current value for a metric from history.
func (s *ValuationService) getCurrentMetricValue(metric string, history []float64) float64 {
	if len(history) == 0 {
		return 0
	}
	// Use the median as a reasonable "current" value proxy
	sorted := make([]float64, len(history))
	copy(sorted, history)
	sort.Float64s(sorted)
	return sorted[len(sorted)/2]
}

// generateHistoricalRatios generates a slice of values uniformly distributed
// between min and max.
func generateHistoricalRatios(min, max float64, count int) []float64 {
	if count <= 0 {
		return nil
	}
	values := make([]float64, count)
	step := (max - min) / float64(count-1)
	for i := range values {
		values[i] = min + step*float64(i)
	}
	return values
}

// calculatePercentiles computes the values at given percentile positions.
func calculatePercentiles(data []float64, pcts []int) map[int]float64 {
	if len(data) == 0 {
		return map[int]float64{}
	}

	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	result := make(map[int]float64, len(pcts))
	n := float64(len(sorted))
	for _, p := range pcts {
		idx := math.Min(float64(p)/100.0*n, n-1)
		result[p] = sorted[int(idx)]
	}
	return result
}
