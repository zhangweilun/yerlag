package valuation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// MVRV Score
// ---------------------------------------------------------------------------

func TestCalculateMVRVScore_Basic(t *testing.T) {
	history := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0}
	result := CalculateMVRVScore(MVRVInput{
		MarketValue:   200,
		RealizedValue: 100,
		History:       history,
	})
	assert.Equal(t, 2.0, result.Ratio)
	assert.InDelta(t, 37.5, result.HistoricalPercentile, 0.01) // 3 out of 8 < 2.0
	assert.Equal(t, "neutral", result.Signal)
	assert.InDelta(t, 37.5, result.Score, 0.01)
}

func TestCalculateMVRVScore_ZeroRealizedValue(t *testing.T) {
	result := CalculateMVRVScore(MVRVInput{
		MarketValue:   200,
		RealizedValue: 0,
		History:       []float64{1.0, 2.0},
	})
	assert.Equal(t, 0.0, result.Ratio)
	assert.Equal(t, 0.0, result.Score)
}

// ---------------------------------------------------------------------------
// P/F Ratio Score
// ---------------------------------------------------------------------------

func TestCalculatePriceToFeeScore_Basic(t *testing.T) {
	history := []float64{10, 20, 30, 40, 50}
	result := CalculatePriceToFeeScore(PriceToFeeInput{
		MarketCap:            1000,
		AnnualizedFeeRevenue: 50,
		History:              history,
	})
	assert.Equal(t, 20.0, result.Ratio)
	assert.InDelta(t, 20.0, result.HistoricalPercentile, 0.01) // 1 out of 5 < 20
	assert.Equal(t, "neutral", result.Signal)
}

func TestCalculatePriceToFeeScore_ZeroFee(t *testing.T) {
	result := CalculatePriceToFeeScore(PriceToFeeInput{
		MarketCap:            1000,
		AnnualizedFeeRevenue: 0,
		History:              []float64{10, 20},
	})
	assert.Equal(t, 0.0, result.Ratio)
}

// ---------------------------------------------------------------------------
// DCF Model
// ---------------------------------------------------------------------------

func TestCalculateDCF_Basic(t *testing.T) {
	result := CalculateDCF(DCFInput{
		AnnualCashFlow:     1_000_000,
		CurrentPrice:       2000,
		TotalSupply:        120_000_000,
		DiscountRate:       0.12,
		GrowthRate:         0.10,
		TerminalGrowthRate: 0.03,
		ProjectionYears:    10,
	})

	assert.True(t, result.FairValueLow >= 0, "FairValueLow should be non-negative")
	assert.True(t, result.FairValueMid >= 0, "FairValueMid should be non-negative")
	assert.True(t, result.FairValueHigh >= 0, "FairValueHigh should be non-negative")
	assert.True(t, result.FairValueLow <= result.FairValueMid, "Low ≤ Mid")
	assert.True(t, result.FairValueMid <= result.FairValueHigh, "Mid ≤ High")
	assert.True(t, result.Score >= 0 && result.Score <= 100, "Score in [0,100]")
}

func TestCalculateDCF_ZeroSupply(t *testing.T) {
	result := CalculateDCF(DCFInput{
		AnnualCashFlow:     1_000_000,
		CurrentPrice:       2000,
		TotalSupply:        0,
		DiscountRate:       0.12,
		GrowthRate:         0.10,
		TerminalGrowthRate: 0.03,
		ProjectionYears:    10,
	})
	assert.Equal(t, 0.0, result.FairValueLow)
	assert.Equal(t, 0.0, result.FairValueMid)
	assert.Equal(t, 0.0, result.FairValueHigh)
}

func TestCalculateDCF_DiscountLessThanTerminal(t *testing.T) {
	result := CalculateDCF(DCFInput{
		AnnualCashFlow:     1_000_000,
		CurrentPrice:       2000,
		TotalSupply:        120_000_000,
		DiscountRate:       0.02,
		GrowthRate:         0.10,
		TerminalGrowthRate: 0.03,
		ProjectionYears:    10,
	})
	// Should return zero values when discount rate ≤ terminal growth rate.
	assert.Equal(t, 0.0, result.FairValueLow)
	assert.Equal(t, 0.0, result.FairValueMid)
	assert.Equal(t, 0.0, result.FairValueHigh)
}

func TestCalculateDCF_Ordering(t *testing.T) {
	// With positive cash flow, low < mid < high.
	result := CalculateDCF(DCFInput{
		AnnualCashFlow:     5_000_000,
		CurrentPrice:       3000,
		TotalSupply:        120_000_000,
		DiscountRate:       0.15,
		GrowthRate:         0.08,
		TerminalGrowthRate: 0.02,
		ProjectionYears:    5,
	})
	assert.True(t, result.FairValueLow <= result.FairValueMid)
	assert.True(t, result.FairValueMid <= result.FairValueHigh)
}

// ---------------------------------------------------------------------------
// Stock-to-Flow
// ---------------------------------------------------------------------------

func TestCalculateStockToFlow_Basic(t *testing.T) {
	result := CalculateStockToFlow(StockToFlowInput{
		CurrentStock: 120_000_000,
		AnnualFlow:   1_000_000,
		CurrentPrice: 2000,
	})

	assert.InDelta(t, 120.0, result.Ratio, 0.01)
	assert.True(t, result.ModelPrice > 0, "ModelPrice should be positive")

	expectedDeviation := (2000 - result.ModelPrice) / result.ModelPrice * 100
	assert.InDelta(t, expectedDeviation, result.Deviation, 0.01)
	assert.True(t, result.Score >= 0 && result.Score <= 100)
}

func TestCalculateStockToFlow_ZeroFlow(t *testing.T) {
	result := CalculateStockToFlow(StockToFlowInput{
		CurrentStock: 120_000_000,
		AnnualFlow:   0,
		CurrentPrice: 2000,
	})
	assert.Equal(t, 50.0, result.Score)
	assert.Equal(t, 0.0, result.Ratio)
}

func TestCalculateStockToFlow_NegativeFlow(t *testing.T) {
	result := CalculateStockToFlow(StockToFlowInput{
		CurrentStock: 120_000_000,
		AnnualFlow:   -500_000,
		CurrentPrice: 2000,
	})
	assert.Equal(t, 50.0, result.Score)
}

// ---------------------------------------------------------------------------
// NVT Score
// ---------------------------------------------------------------------------

func TestCalculateNVTScore_Basic(t *testing.T) {
	history := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	result := CalculateNVTScore(NVTInput{
		MarketCap:   500_000,
		DailyVolume: 10_000,
		History:     history,
	})
	assert.InDelta(t, 50.0, result.Ratio, 0.01)
	assert.Equal(t, "neutral", result.Signal)
	assert.True(t, result.Score >= 0 && result.Score <= 100)
}

func TestCalculateNVTScore_HighPercentile(t *testing.T) {
	history := make([]float64, 100)
	for i := range history {
		history[i] = float64(i + 1)
	}
	result := CalculateNVTScore(NVTInput{
		MarketCap:   10_000,
		DailyVolume: 1,
		History:     history,
	})
	// NVT = 10000, which is above all 100 values → percentile = 100
	assert.True(t, result.HistoricalPercentile > 90)
	assert.Equal(t, "overvalued", result.Signal)
}

// ---------------------------------------------------------------------------
// ETH/BTC Score
// ---------------------------------------------------------------------------

func TestCalculateETHBTCScore_Basic(t *testing.T) {
	history := []float64{0.03, 0.04, 0.05, 0.06, 0.07, 0.08}
	result := CalculateETHBTCScore(ETHBTCInput{
		CurrentRatio: 0.05,
		History:      history,
	})
	assert.InDelta(t, 0.05, result.Ratio, 0.001)
	assert.Equal(t, "neutral", result.Signal)
	assert.True(t, result.Score >= 0 && result.Score <= 100)
}

func TestCalculateETHBTCScore_LowPercentile(t *testing.T) {
	history := make([]float64, 100)
	for i := range history {
		history[i] = float64(i+1) * 0.001
	}
	result := CalculateETHBTCScore(ETHBTCInput{
		CurrentRatio: 0.0001,
		History:      history,
	})
	assert.True(t, result.HistoricalPercentile < 10)
	assert.Equal(t, "eth_undervalued", result.Signal)
}

// ---------------------------------------------------------------------------
// Overall Valuation
// ---------------------------------------------------------------------------

func TestCalculateValuation_ScoreRange(t *testing.T) {
	input := ValuationInput{
		MVRV: MVRVInput{
			MarketValue:   200,
			RealizedValue: 100,
			History:       []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0},
		},
		PriceToFee: PriceToFeeInput{
			MarketCap:            1000,
			AnnualizedFeeRevenue: 50,
			History:              []float64{10, 20, 30, 40, 50},
		},
		DCF: DCFInput{
			AnnualCashFlow:     1_000_000,
			CurrentPrice:       2000,
			TotalSupply:        120_000_000,
			DiscountRate:       0.12,
			GrowthRate:         0.10,
			TerminalGrowthRate: 0.03,
			ProjectionYears:    10,
		},
		S2F: StockToFlowInput{
			CurrentStock: 120_000_000,
			AnnualFlow:   1_000_000,
			CurrentPrice: 2000,
		},
		NVT: NVTInput{
			MarketCap:   500_000,
			DailyVolume: 10_000,
			History:     []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		},
		ETHBTC: ETHBTCInput{
			CurrentRatio: 0.05,
			History:      []float64{0.03, 0.04, 0.05, 0.06, 0.07, 0.08},
		},
	}

	result := CalculateValuation(input)

	assert.True(t, result.Overall >= 0, "Overall score >= 0")
	assert.True(t, result.Overall <= 100, "Overall score <= 100")
	assert.Contains(t, []string{"undervalued", "fair", "overvalued"}, result.Status)
}

func TestCalculateValuation_StatusLabels(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{"low score", 10, "undervalued"},
		{"boundary 33", 33, "fair"},
		{"mid score", 50, "fair"},
		{"boundary 66", 66, "fair"},
		{"high score", 80, "overvalued"},
		{"zero", 0, "undervalued"},
		{"hundred", 100, "overvalued"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, classifyOverall(tt.score))
		})
	}
}

func TestCalculateValuation_RadarData(t *testing.T) {
	input := ValuationInput{
		MVRV: MVRVInput{
			MarketValue:   200,
			RealizedValue: 100,
			History:       []float64{1.0, 2.0, 3.0},
		},
		PriceToFee: PriceToFeeInput{
			MarketCap:            1000,
			AnnualizedFeeRevenue: 50,
			History:              []float64{10, 20, 30},
		},
		DCF: DCFInput{
			AnnualCashFlow:     1_000_000,
			CurrentPrice:       2000,
			TotalSupply:        120_000_000,
			DiscountRate:       0.12,
			GrowthRate:         0.10,
			TerminalGrowthRate: 0.03,
			ProjectionYears:    10,
		},
		S2F: StockToFlowInput{
			CurrentStock: 120_000_000,
			AnnualFlow:   1_000_000,
			CurrentPrice: 2000,
		},
		NVT: NVTInput{
			MarketCap:   500_000,
			DailyVolume: 10_000,
			History:     []float64{10, 50, 100},
		},
		ETHBTC: ETHBTCInput{
			CurrentRatio: 0.05,
			History:      []float64{0.03, 0.05, 0.07},
		},
	}

	result := CalculateValuation(input)

	require.Len(t, result.RadarData, 6)

	expectedDimensions := []string{"MVRV", "P/F Ratio", "DCF", "Stock-to-Flow", "NVT", "ETH/BTC"}
	for i, dim := range expectedDimensions {
		assert.Equal(t, dim, result.RadarData[i].Dimension)
		assert.True(t, result.RadarData[i].Score >= 0 && result.RadarData[i].Score <= 100,
			"Radar score for %s should be in [0,100]", dim)
		assert.NotEmpty(t, result.RadarData[i].Label, "Radar label for %s should not be empty", dim)
	}
}

// ---------------------------------------------------------------------------
// Clamp helper
// ---------------------------------------------------------------------------

func TestClamp(t *testing.T) {
	assert.Equal(t, 0.0, clamp(-5, 0, 100))
	assert.Equal(t, 100.0, clamp(150, 0, 100))
	assert.Equal(t, 50.0, clamp(50, 0, 100))
	assert.Equal(t, 0.0, clamp(0, 0, 100))
	assert.Equal(t, 100.0, clamp(100, 0, 100))
}

// ---------------------------------------------------------------------------
// DCF internal helper
// ---------------------------------------------------------------------------

func TestDcfValue_PositiveCashFlow(t *testing.T) {
	val := dcfValue(1_000_000, 0.12, 0.10, 0.03, 10)
	assert.True(t, val > 0, "DCF value should be positive for positive cash flow")
	assert.False(t, math.IsNaN(val), "DCF value should not be NaN")
	assert.False(t, math.IsInf(val, 0), "DCF value should not be Inf")
}

func TestDcfValue_ZeroCashFlow(t *testing.T) {
	val := dcfValue(0, 0.12, 0.10, 0.03, 10)
	assert.Equal(t, 0.0, val)
}

func TestDcfValue_InvalidRates(t *testing.T) {
	val := dcfValue(1_000_000, 0.02, 0.10, 0.03, 10)
	assert.Equal(t, 0.0, val, "Should return 0 when discount rate ≤ terminal growth rate")
}

// ---------------------------------------------------------------------------
// Weights sum to 1.0
// ---------------------------------------------------------------------------

func TestDefaultWeightsSumToOne(t *testing.T) {
	sum := 0.0
	for _, w := range defaultWeights {
		sum += w
	}
	assert.InDelta(t, 1.0, sum, 0.001)
}
