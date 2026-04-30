package macro

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// MacroIndicators holds the macro economic indicators data.
type MacroIndicators struct {
	DXYIndex             float64                 `json:"dxyIndex"`
	DXYHistory           []types.TimeSeriesPoint `json:"dxyHistory"`
	Treasury10Y          float64                 `json:"treasury10y"`
	Treasury10YHistory   []types.TimeSeriesPoint `json:"treasury10yHistory"`
	NasdaqCorrelation30d float64                 `json:"nasdaqCorrelation30d"`
	NasdaqCorrelation90d float64                 `json:"nasdaqCorrelation90d"`
	FedFundsRate         float64                 `json:"fedFundsRate"`
	RateExpectations     []RateExpectation       `json:"rateExpectations"`
	FearGreedIndex       int                     `json:"fearGreedIndex"`
	FearGreedHistory     []types.TimeSeriesPoint `json:"fearGreedHistory"`
	StablecoinMarketCap  float64                 `json:"stablecoinMarketCap"`
	StablecoinHistory    []types.TimeSeriesPoint `json:"stablecoinHistory"`
}

// RateExpectation holds a future interest rate expectation.
type RateExpectation struct {
	Date         string  `json:"date"`
	ExpectedRate float64 `json:"expectedRate"`
	Probability  float64 `json:"probability"`
}

// IndicatorsService handles macro economic indicators logic.
type IndicatorsService struct {
	svcCtx    *svc.ServiceContext
	tradfi    *fetcher.TradFiClient
	coingecko *fetcher.CoinGeckoClient
}

// NewIndicatorsService creates a new IndicatorsService.
func NewIndicatorsService(svcCtx *svc.ServiceContext) *IndicatorsService {
	cfg := svcCtx.Config.DataSources
	return &IndicatorsService{
		svcCtx:    svcCtx,
		tradfi:    fetcher.NewTradFiClient(cfg.TradFi.BaseURL, cfg.TradFi.APIKey),
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetMacroIndicators fetches and computes macro economic indicators.
func (s *IndicatorsService) GetMacroIndicators(ctx context.Context) (*MacroIndicators, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.MacroData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "macro:indicators", ttl, func() (interface{}, error) {
		return s.fetchMacroIndicators(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*MacroIndicators); ok {
		return data, nil
	}

	return s.fetchMacroIndicators(ctx)
}

func (s *IndicatorsService) fetchMacroIndicators(ctx context.Context) (*MacroIndicators, error) {
	// Fetch DXY data
	dxyIndex, dxyHistory := s.fetchDXYData(ctx)

	// Fetch Treasury yield data
	treasury10Y, treasury10YHistory := s.fetchTreasuryData(ctx)

	// Fetch Nasdaq correlation with ETH
	nasdaqCorrelation30d, nasdaqCorrelation90d := s.fetchNasdaqCorrelation(ctx)

	// Fetch Federal Funds Rate and expectations
	fedFundsRate, rateExpectations := s.fetchFedRateData(ctx)

	// Fetch Fear & Greed Index
	fearGreedIndex, fearGreedHistory := s.fetchFearGreedData(ctx)

	// Fetch Stablecoin market cap
	stablecoinMarketCap, stablecoinHistory := s.fetchStablecoinData(ctx)

	return &MacroIndicators{
		DXYIndex:             dxyIndex,
		DXYHistory:           dxyHistory,
		Treasury10Y:          treasury10Y,
		Treasury10YHistory:   treasury10YHistory,
		NasdaqCorrelation30d: nasdaqCorrelation30d,
		NasdaqCorrelation90d: nasdaqCorrelation90d,
		FedFundsRate:         fedFundsRate,
		RateExpectations:     rateExpectations,
		FearGreedIndex:       fearGreedIndex,
		FearGreedHistory:     fearGreedHistory,
		StablecoinMarketCap:  stablecoinMarketCap,
		StablecoinHistory:    stablecoinHistory,
	}, nil
}

// fetchDXYData fetches the US Dollar Index data.
func (s *IndicatorsService) fetchDXYData(ctx context.Context) (float64, []types.TimeSeriesPoint) {
	dxyData, err := s.tradfi.GetDXY(ctx, 365)
	if err != nil || dxyData == nil {
		return 0, []types.TimeSeriesPoint{}
	}

	history := tradFiDataPointsToTimeSeries(dxyData.History)
	return dxyData.Current, history
}

// fetchTreasuryData fetches the US 10-year Treasury yield data.
func (s *IndicatorsService) fetchTreasuryData(ctx context.Context) (float64, []types.TimeSeriesPoint) {
	treasuryData, err := s.tradfi.GetTreasuryYields(ctx, 365)
	if err != nil || treasuryData == nil {
		return 0, []types.TimeSeriesPoint{}
	}

	history := tradFiDataPointsToTimeSeries(treasuryData.History)
	return treasuryData.Yield10Y, history
}

// fetchNasdaqCorrelation computes the 30-day and 90-day correlation between ETH and Nasdaq.
func (s *IndicatorsService) fetchNasdaqCorrelation(ctx context.Context) (float64, float64) {
	// Fetch Nasdaq historical data
	nasdaqData, nasdaqErr := s.tradfi.GetNasdaq(ctx, 365)
	if nasdaqErr != nil || nasdaqData == nil || len(nasdaqData.History) == 0 {
		return 0, 0
	}

	// Fetch ETH historical price data
	ethChart, ethErr := s.coingecko.GetMarketChart(ctx, "ethereum", "usd", 365)
	if ethErr != nil || ethChart == nil || len(ethChart.Prices) == 0 {
		return 0, 0
	}

	// Extract ETH daily prices
	ethPrices := extractPriceSeries(ethChart.Prices)

	// Extract Nasdaq daily values
	nasdaqPrices := make([]float64, 0, len(nasdaqData.History))
	for _, point := range nasdaqData.History {
		nasdaqPrices = append(nasdaqPrices, point.Value)
	}

	// Align series to same length
	minLen := len(ethPrices)
	if len(nasdaqPrices) < minLen {
		minLen = len(nasdaqPrices)
	}

	// Calculate 30-day correlation
	corr30d := 0.0
	if minLen >= 30 {
		ethSlice := ethPrices[len(ethPrices)-30:]
		nasdaqSlice := nasdaqPrices[len(nasdaqPrices)-30:]
		if r := calc.PearsonCorrelation(ethSlice, nasdaqSlice); r != nil {
			corr30d = *r
		}
	}

	// Calculate 90-day correlation
	corr90d := 0.0
	if minLen >= 90 {
		ethSlice := ethPrices[len(ethPrices)-90:]
		nasdaqSlice := nasdaqPrices[len(nasdaqPrices)-90:]
		if r := calc.PearsonCorrelation(ethSlice, nasdaqSlice); r != nil {
			corr90d = *r
		}
	}

	return corr30d, corr90d
}

// fetchFedRateData fetches the Federal Funds Rate and rate expectations.
func (s *IndicatorsService) fetchFedRateData(ctx context.Context) (float64, []RateExpectation) {
	fedData, err := s.tradfi.GetFedRate(ctx)
	if err != nil || fedData == nil {
		return 0, []RateExpectation{}
	}

	expectations := make([]RateExpectation, 0, len(fedData.RateExpectations))
	for _, exp := range fedData.RateExpectations {
		expectations = append(expectations, RateExpectation{
			Date:         exp.Date,
			ExpectedRate: exp.ExpectedRate,
			Probability:  exp.Probability,
		})
	}

	return fedData.CurrentRate, expectations
}

// fetchFearGreedData fetches the crypto Fear & Greed Index.
func (s *IndicatorsService) fetchFearGreedData(ctx context.Context) (int, []types.TimeSeriesPoint) {
	fgData, err := s.tradfi.GetFearGreedIndex(ctx, 365)
	if err != nil || fgData == nil {
		return 0, []types.TimeSeriesPoint{}
	}

	history := tradFiDataPointsToTimeSeries(fgData.History)
	return fgData.Value, history
}

// fetchStablecoinData fetches stablecoin market cap data.
// Uses CoinGecko to get USDT and USDC market cap data.
func (s *IndicatorsService) fetchStablecoinData(ctx context.Context) (float64, []types.TimeSeriesPoint) {
	// Fetch current stablecoin market data
	stablecoins := []string{"tether", "usd-coin"}
	marketData, err := s.coingecko.GetMarketData(ctx, stablecoins, "usd")
	if err != nil || len(marketData) == 0 {
		return 0, []types.TimeSeriesPoint{}
	}

	// Sum up stablecoin market caps
	var totalMarketCap float64
	for _, coin := range marketData {
		totalMarketCap += coin.MarketCap
	}

	// Fetch historical market cap for USDT (as proxy for stablecoin market cap trend)
	usdtChart, chartErr := s.coingecko.GetMarketChart(ctx, "tether", "usd", 365)
	history := []types.TimeSeriesPoint{}
	if chartErr == nil && usdtChart != nil {
		for _, point := range usdtChart.MarketCaps {
			if len(point) < 2 {
				continue
			}
			history = append(history, types.TimeSeriesPoint{
				Timestamp: int64(point[0]) / 1000,
				Value:     point[1],
			})
		}
	}

	return totalMarketCap, history
}

// tradFiDataPointsToTimeSeries converts TradFi data points to TimeSeriesPoint slice.
func tradFiDataPointsToTimeSeries(points []fetcher.TradFiDataPoint) []types.TimeSeriesPoint {
	result := make([]types.TimeSeriesPoint, 0, len(points))
	for _, point := range points {
		ts := parseDateToTimestamp(point.Date)
		result = append(result, types.TimeSeriesPoint{
			Timestamp: ts,
			Value:     point.Value,
		})
	}
	return result
}

// parseDateToTimestamp parses a YYYY-MM-DD date string to a Unix timestamp.
func parseDateToTimestamp(dateStr string) int64 {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0
	}
	return t.Unix()
}
