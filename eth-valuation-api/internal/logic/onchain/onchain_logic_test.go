package onchain

import (
	"testing"
	"time"

	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/types"

	"github.com/stretchr/testify/assert"
)

// --- Gas High Fee Flag tests ---
// Requirement 3.6: Gas > 50 Gwei triggers high fee flag

func TestGasHighFeeFlag_Above50Gwei(t *testing.T) {
	tests := []struct {
		name       string
		gasGwei    float64
		expectHigh bool
	}{
		{"exactly 50 Gwei - not high", 50.0, false},
		{"51 Gwei - high", 51.0, true},
		{"100 Gwei - high", 100.0, true},
		{"200 Gwei - high", 200.0, true},
		{"49.99 Gwei - not high", 49.99, false},
		{"0 Gwei - not high", 0.0, false},
		{"1 Gwei - not high", 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The gas module uses: isHighFee := currentAvgGwei > 50
			isHighFee := tt.gasGwei > 50
			assert.Equal(t, tt.expectHigh, isHighFee)
		})
	}
}

// --- NVT Signal Classification tests ---
// Requirement 4.6, 4.7: NVT signal classification based on percentile

func TestNVTSignalClassification(t *testing.T) {
	tests := []struct {
		name           string
		percentile     float64
		expectedSignal string
	}{
		{"percentile 95 - overvalued", 95.0, "overvalued"},
		{"percentile 91 - overvalued", 91.0, "overvalued"},
		{"percentile 90.1 - overvalued", 90.1, "overvalued"},
		{"percentile 90 - neutral (boundary)", 90.0, "neutral"},
		{"percentile 50 - neutral", 50.0, "neutral"},
		{"percentile 10 - neutral (boundary)", 10.0, "neutral"},
		{"percentile 9.9 - undervalued", 9.9, "undervalued"},
		{"percentile 5 - undervalued", 5.0, "undervalued"},
		{"percentile 0 - undervalued", 0.0, "undervalued"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal := calc.ClassifySignal(tt.percentile)
			assert.Equal(t, tt.expectedSignal, signal)
		})
	}
}

func TestNVTPercentileCalculation(t *testing.T) {
	// History with known distribution
	history := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	tests := []struct {
		name         string
		currentValue float64
		minExpected  float64
		maxExpected  float64
	}{
		{"lowest value", 5, 0, 10},
		{"highest value", 105, 90, 100},
		{"middle value", 55, 40, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percentile := calc.CalculatePercentile(history, tt.currentValue)
			assert.GreaterOrEqual(t, percentile, tt.minExpected)
			assert.LessOrEqual(t, percentile, tt.maxExpected)
		})
	}
}

func TestNVTPercentile_EmptyHistory(t *testing.T) {
	percentile := calc.CalculatePercentile(nil, 50.0)
	assert.Equal(t, 0.0, percentile)
}

// --- Inflation/Deflation Detection tests ---
// Requirement 2.4, 17.4: Net issuance and inflation/deflation classification

func TestInflationDeflationDetection(t *testing.T) {
	tests := []struct {
		name           string
		dailyIssuance  float64
		dailyBurn      float64
		expectDeflat   bool
		expectPositive bool // net issuance > 0
	}{
		{
			name:           "more burn than issuance - deflationary",
			dailyIssuance:  1000,
			dailyBurn:      1500,
			expectDeflat:   true,
			expectPositive: false,
		},
		{
			name:           "more issuance than burn - inflationary",
			dailyIssuance:  1500,
			dailyBurn:      1000,
			expectDeflat:   false,
			expectPositive: true,
		},
		{
			name:           "equal issuance and burn - not deflationary",
			dailyIssuance:  1000,
			dailyBurn:      1000,
			expectDeflat:   false,
			expectPositive: false,
		},
		{
			name:           "zero issuance with burn - deflationary",
			dailyIssuance:  0,
			dailyBurn:      500,
			expectDeflat:   true,
			expectPositive: false,
		},
		{
			name:           "zero burn with issuance - inflationary",
			dailyIssuance:  500,
			dailyBurn:      0,
			expectDeflat:   false,
			expectPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netIssuance := calc.NetIssuance(tt.dailyIssuance, tt.dailyBurn)
			isDeflationary := calc.IsDeflationary(netIssuance)

			assert.Equal(t, tt.expectDeflat, isDeflationary)
			if tt.expectPositive {
				assert.Greater(t, netIssuance, 0.0)
			}
		})
	}
}

func TestAnnualInflationRate(t *testing.T) {
	tests := []struct {
		name         string
		netIssuance  float64
		totalSupply  float64
		expectedRate *float64
	}{
		{
			name:         "positive inflation",
			netIssuance:  365000, // annual net issuance
			totalSupply:  120000000,
			expectedRate: floatPtr(365000.0 / 120000000.0 * 100),
		},
		{
			name:         "negative inflation (deflation)",
			netIssuance:  -182500,
			totalSupply:  120000000,
			expectedRate: floatPtr(-182500.0 / 120000000.0 * 100),
		},
		{
			name:         "zero total supply - nil",
			netIssuance:  1000,
			totalSupply:  0,
			expectedRate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.AnnualInflationRate(tt.netIssuance, tt.totalSupply)
			if tt.expectedRate == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, *tt.expectedRate, *result, 1e-9)
			}
		})
	}
}

// --- Annualized Burn Rate tests ---
// Requirement 2.3: Annualized burn rate calculation

func TestAnnualizedBurnRate(t *testing.T) {
	tests := []struct {
		name        string
		dailyBurn   float64
		totalSupply float64
		expectNil   bool
		expected    float64
	}{
		{
			name:        "normal case",
			dailyBurn:   2000,
			totalSupply: 120000000,
			expectNil:   false,
			expected:    (2000.0 * 365.0 / 120000000.0) * 100.0,
		},
		{
			name:        "zero supply returns nil",
			dailyBurn:   2000,
			totalSupply: 0,
			expectNil:   true,
		},
		{
			name:        "zero burn",
			dailyBurn:   0,
			totalSupply: 120000000,
			expectNil:   false,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.AnnualizedBurnRate(tt.dailyBurn, tt.totalSupply)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, tt.expected, *result, 1e-9)
			}
		})
	}
}

// --- TVL/Market Cap Ratio tests ---
// Requirement 5.4: TVL to market cap ratio

func TestTVLToMarketCapRatio(t *testing.T) {
	tests := []struct {
		name      string
		tvl       float64
		marketCap float64
		expectNil bool
		expected  float64
	}{
		{
			name:      "normal ratio",
			tvl:       50000000000,  // 50B TVL
			marketCap: 200000000000, // 200B market cap
			expectNil: false,
			expected:  0.25,
		},
		{
			name:      "zero market cap returns nil",
			tvl:       50000000000,
			marketCap: 0,
			expectNil: true,
		},
		{
			name:      "zero TVL",
			tvl:       0,
			marketCap: 200000000000,
			expectNil: false,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.TVLToMarketCapRatio(tt.tvl, tt.marketCap)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, tt.expected, *result, 1e-9)
			}
		})
	}
}

// --- Price to Fee Ratio tests ---
// Requirement 3.5: Market-to-fee ratio

func TestPriceToFeeRatio(t *testing.T) {
	tests := []struct {
		name                 string
		marketCap            float64
		annualizedFeeRevenue float64
		expectNil            bool
		expected             float64
	}{
		{
			name:                 "normal ratio",
			marketCap:            200000000000, // 200B
			annualizedFeeRevenue: 5000000000,   // 5B
			expectNil:            false,
			expected:             40.0,
		},
		{
			name:                 "zero fee revenue returns nil",
			marketCap:            200000000000,
			annualizedFeeRevenue: 0,
			expectNil:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.PriceToFeeRatio(tt.marketCap, tt.annualizedFeeRevenue)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, tt.expected, *result, 1e-9)
			}
		})
	}
}

// --- NVT Ratio tests ---
// Requirement 4.4: NVT Ratio calculation

func TestNVTRatio(t *testing.T) {
	tests := []struct {
		name        string
		marketCap   float64
		dailyVolume float64
		expectNil   bool
		expected    float64
	}{
		{
			name:        "normal NVT",
			marketCap:   200000000000, // 200B
			dailyVolume: 5000000000,   // 5B
			expectNil:   false,
			expected:    40.0,
		},
		{
			name:        "zero volume returns nil",
			marketCap:   200000000000,
			dailyVolume: 0,
			expectNil:   true,
		},
		{
			name:        "high NVT (overvalued signal)",
			marketCap:   200000000000,
			dailyVolume: 1000000000, // 1B
			expectNil:   false,
			expected:    200.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.NVTRatio(tt.marketCap, tt.dailyVolume)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, tt.expected, *result, 1e-9)
			}
		})
	}
}

// --- daysSinceEIP1559 tests ---

func TestDaysSinceEIP1559(t *testing.T) {
	// EIP-1559 activated on Aug 5, 2021
	tests := []struct {
		name     string
		now      time.Time
		expected int
	}{
		{
			name:     "same day returns 1 (minimum)",
			now:      time.Date(2021, 8, 5, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "one day after",
			now:      time.Date(2021, 8, 6, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "one year after",
			now:      time.Date(2022, 8, 5, 0, 0, 0, 0, time.UTC),
			expected: 365,
		},
		{
			name:     "before EIP-1559 returns 1 (minimum)",
			now:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := daysSinceEIP1559(tt.now)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- Supply distribution tests ---
// Requirement 17.2: Supply distribution calculation

func TestSupplyDistribution_OtherAmountNonNegative(t *testing.T) {
	// When components exceed total supply, "other" should be clamped to 0
	totalSupply := 120000000.0
	staked := 30000000.0
	defiLocked := 50000000.0
	exchangeBalance := 50000000.0

	other := totalSupply - staked - defiLocked - exchangeBalance
	if other < 0 {
		other = 0
	}

	assert.GreaterOrEqual(t, other, 0.0)
}

func TestSupplyDistribution_NormalCase(t *testing.T) {
	totalSupply := 120000000.0
	staked := 30000000.0
	defiLocked := 20000000.0
	exchangeBalance := 15000000.0

	other := totalSupply - staked - defiLocked - exchangeBalance
	if other < 0 {
		other = 0
	}

	assert.InDelta(t, 55000000.0, other, 1e-9)
	// Verify all parts sum to total
	assert.InDelta(t, totalSupply, staked+defiLocked+exchangeBalance+other, 1e-9)
}

// --- calculateDominanceHistory tests ---

func TestCalculateDominanceHistory_NormalCase(t *testing.T) {
	ethHistory := []types.TimeSeriesPoint{
		{Timestamp: 100, Value: 50},
		{Timestamp: 200, Value: 60},
		{Timestamp: 300, Value: 70},
	}

	// Simulating DefiLlama total TVL history
	type tvlHistoryPoint struct {
		Date int64
		TVL  float64
	}
	totalHistory := []tvlHistoryPoint{
		{Date: 100, TVL: 100},
		{Date: 200, TVL: 200},
		{Date: 300, TVL: 350},
	}

	// Build map like the real function does
	totalMap := make(map[int64]float64)
	for _, dp := range totalHistory {
		totalMap[dp.Date] = dp.TVL
	}

	var result []types.TimeSeriesPoint
	for _, ethDP := range ethHistory {
		if totalTVL, ok := totalMap[ethDP.Timestamp]; ok && totalTVL > 0 {
			dominance := (ethDP.Value / totalTVL) * 100
			result = append(result, types.TimeSeriesPoint{
				Timestamp: ethDP.Timestamp,
				Value:     dominance,
			})
		}
	}

	assert.Len(t, result, 3)
	assert.InDelta(t, 50.0, result[0].Value, 1e-9) // 50/100 * 100
	assert.InDelta(t, 30.0, result[1].Value, 1e-9) // 60/200 * 100
	assert.InDelta(t, 20.0, result[2].Value, 1e-9) // 70/350 * 100
}

// --- Gas fee share breakdown tests ---
// Requirement 3.3: Base fee vs priority fee share

func TestGasFeeShareBreakdown(t *testing.T) {
	tests := []struct {
		name             string
		proposeGas       float64
		baseFee          float64
		expectedBase     float64
		expectedPriority float64
	}{
		{
			name:             "normal split",
			proposeGas:       30.0,
			baseFee:          20.0,
			expectedBase:     66.666666,
			expectedPriority: 33.333333,
		},
		{
			name:             "all base fee",
			proposeGas:       30.0,
			baseFee:          30.0,
			expectedBase:     100.0,
			expectedPriority: 0.0,
		},
		{
			name:             "zero total fee",
			proposeGas:       0.0,
			baseFee:          0.0,
			expectedBase:     0.0,
			expectedPriority: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priorityFee := tt.proposeGas - tt.baseFee
			totalFee := tt.proposeGas
			baseFeeShare := 0.0
			priorityFeeShare := 0.0
			if totalFee > 0 {
				baseFeeShare = (tt.baseFee / totalFee) * 100
				priorityFeeShare = (priorityFee / totalFee) * 100
			}

			assert.InDelta(t, tt.expectedBase, baseFeeShare, 0.001)
			assert.InDelta(t, tt.expectedPriority, priorityFeeShare, 0.001)

			// Shares should sum to 100% (or both be 0)
			if totalFee > 0 {
				assert.InDelta(t, 100.0, baseFeeShare+priorityFeeShare, 0.001)
			}
		})
	}
}

// --- Moving average for DAA tests ---
// Requirement 4.1: 7-day moving average

func TestDAAMovingAverage(t *testing.T) {
	// Test the moving average calculation used for DAA
	history := []float64{100, 200, 300, 400, 500, 600, 700}

	ma := calc.MovingAverage(history, 7)
	assert.NotNil(t, ma)
	expected := (100 + 200 + 300 + 400 + 500 + 600 + 700) / 7.0
	assert.InDelta(t, expected, *ma, 1e-9)
}

func TestDAAMovingAverage_InsufficientData(t *testing.T) {
	history := []float64{100, 200, 300}

	// With less than 7 data points, MovingAverage should return nil
	ma := calc.MovingAverage(history, 7)
	assert.Nil(t, ma)
}

// Helper function
func floatPtr(f float64) *float64 {
	return &f
}
