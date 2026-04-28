package valuation

// Feature: eth-valuation-dashboard, Property 1: 综合估值评分范围不变量
//
// For any set of input metrics provided to the Valuation_Engine, the calculated
// overall score SHALL always be in the range [0, 100], and the status label SHALL
// be "undervalued" when score < 33, "fair" when 33 ≤ score ≤ 66, and "overvalued"
// when score > 66. All RadarData scores SHALL be in [0, 100] and RadarData SHALL
// have exactly 6 entries.
//
// **Validates: Requirements 1.3**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// --- Generators ---

// genPositiveFloat generates positive floats in a realistic range.
func genPositiveFloat() *rapid.Generator[float64] {
	return rapid.Float64Range(1e-6, 1e12)
}

// genPositiveHistory generates a slice of positive floats with at least 2 elements,
// suitable for historical data arrays.
func genPositiveHistory(minLen, maxLen int) *rapid.Generator[[]float64] {
	return rapid.SliceOfN(rapid.Float64Range(1e-6, 1e9), minLen, maxLen)
}

// genSmallPositiveFloat generates small positive floats for ratios like ETH/BTC.
func genSmallPositiveFloat() *rapid.Generator[float64] {
	return rapid.Float64Range(0.001, 1.0)
}

// genDiscountAndTerminalRates generates a discount rate that is strictly greater
// than the terminal growth rate, both positive.
func genDiscountAndTerminalRates(t *rapid.T) (discountRate, terminalGrowthRate float64) {
	terminalGrowthRate = rapid.Float64Range(0.001, 0.05).Draw(t, "terminalGrowthRate")
	// Ensure discount rate > terminal growth rate
	discountRate = rapid.Float64Range(terminalGrowthRate+0.01, 0.30).Draw(t, "discountRate")
	return
}

// genValuationInput generates a random ValuationInput with realistic ranges.
func genValuationInput(t *rapid.T) ValuationInput {
	// MVRV input
	mvrvMarketValue := genPositiveFloat().Draw(t, "mvrvMarketValue")
	mvrvRealizedValue := genPositiveFloat().Draw(t, "mvrvRealizedValue")
	mvrvHistory := genPositiveHistory(2, 50).Draw(t, "mvrvHistory")

	// PriceToFee input
	pfMarketCap := genPositiveFloat().Draw(t, "pfMarketCap")
	pfFeeRevenue := genPositiveFloat().Draw(t, "pfFeeRevenue")
	pfHistory := genPositiveHistory(2, 50).Draw(t, "pfHistory")

	// DCF input
	dcfCashFlow := genPositiveFloat().Draw(t, "dcfCashFlow")
	dcfPrice := genPositiveFloat().Draw(t, "dcfPrice")
	dcfSupply := genPositiveFloat().Draw(t, "dcfSupply")
	discountRate, terminalGrowthRate := genDiscountAndTerminalRates(t)
	growthRate := rapid.Float64Range(0.01, 0.25).Draw(t, "growthRate")
	projectionYears := rapid.IntRange(1, 30).Draw(t, "projectionYears")

	// Stock-to-Flow input
	s2fStock := genPositiveFloat().Draw(t, "s2fStock")
	s2fFlow := genPositiveFloat().Draw(t, "s2fFlow")
	s2fPrice := genPositiveFloat().Draw(t, "s2fPrice")

	// NVT input
	nvtMarketCap := genPositiveFloat().Draw(t, "nvtMarketCap")
	nvtVolume := genPositiveFloat().Draw(t, "nvtVolume")
	nvtHistory := genPositiveHistory(2, 50).Draw(t, "nvtHistory")

	// ETH/BTC input
	ethbtcRatio := genSmallPositiveFloat().Draw(t, "ethbtcRatio")
	ethbtcHistory := rapid.SliceOfN(genSmallPositiveFloat(), 2, 50).Draw(t, "ethbtcHistory")

	return ValuationInput{
		MVRV: MVRVInput{
			MarketValue:   mvrvMarketValue,
			RealizedValue: mvrvRealizedValue,
			History:       mvrvHistory,
		},
		PriceToFee: PriceToFeeInput{
			MarketCap:            pfMarketCap,
			AnnualizedFeeRevenue: pfFeeRevenue,
			History:              pfHistory,
		},
		DCF: DCFInput{
			AnnualCashFlow:     dcfCashFlow,
			CurrentPrice:       dcfPrice,
			TotalSupply:        dcfSupply,
			DiscountRate:       discountRate,
			GrowthRate:         growthRate,
			TerminalGrowthRate: terminalGrowthRate,
			ProjectionYears:    projectionYears,
		},
		S2F: StockToFlowInput{
			CurrentStock: s2fStock,
			AnnualFlow:   s2fFlow,
			CurrentPrice: s2fPrice,
		},
		NVT: NVTInput{
			MarketCap:   nvtMarketCap,
			DailyVolume: nvtVolume,
			History:     nvtHistory,
		},
		ETHBTC: ETHBTCInput{
			CurrentRatio: ethbtcRatio,
			History:      ethbtcHistory,
		},
	}
}

// --- Property Tests ---

// Feature: eth-valuation-dashboard, Property 10: DCF 估值范围有序性
//
// For any valid DCF assumptions (positive discount rate > terminal growth rate,
// positive projection years) and non-negative cash flow inputs, the DCF model
// SHALL produce fairValueLow ≤ fairValueMid ≤ fairValueHigh, and all three
// values SHALL be non-negative. The Score SHALL be in [0, 100].
//
// **Validates: Requirements 7.3**

// TestProperty10_DCFValuationRangeOrdering verifies that for any randomly
// generated valid DCFInput, the resulting fair values are ordered
// low ≤ mid ≤ high, all non-negative, and the score is in [0, 100].
func TestProperty10_DCFValuationRangeOrdering(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate terminal growth rate first (small positive)
		terminalGrowthRate := rapid.Float64Range(0.001, 0.05).Draw(t, "terminalGrowthRate")
		// Discount rate must be strictly greater than terminal growth rate
		discountRate := rapid.Float64Range(terminalGrowthRate+0.01, 0.30).Draw(t, "discountRate")
		growthRate := rapid.Float64Range(0.01, 0.25).Draw(t, "growthRate")
		projectionYears := rapid.IntRange(1, 30).Draw(t, "projectionYears")

		input := DCFInput{
			AnnualCashFlow:     genPositiveFloat().Draw(t, "annualCashFlow"),
			CurrentPrice:       genPositiveFloat().Draw(t, "currentPrice"),
			TotalSupply:        genPositiveFloat().Draw(t, "totalSupply"),
			DiscountRate:       discountRate,
			GrowthRate:         growthRate,
			TerminalGrowthRate: terminalGrowthRate,
			ProjectionYears:    projectionYears,
		}

		result := CalculateDCF(input)

		// All fair values must be non-negative
		assert.GreaterOrEqual(t, result.FairValueLow, 0.0,
			"FairValueLow should be >= 0, got %v", result.FairValueLow)
		assert.GreaterOrEqual(t, result.FairValueMid, 0.0,
			"FairValueMid should be >= 0, got %v", result.FairValueMid)
		assert.GreaterOrEqual(t, result.FairValueHigh, 0.0,
			"FairValueHigh should be >= 0, got %v", result.FairValueHigh)

		// Ordering: low ≤ mid ≤ high
		assert.LessOrEqual(t, result.FairValueLow, result.FairValueMid,
			"FairValueLow (%v) should be <= FairValueMid (%v)", result.FairValueLow, result.FairValueMid)
		assert.LessOrEqual(t, result.FairValueMid, result.FairValueHigh,
			"FairValueMid (%v) should be <= FairValueHigh (%v)", result.FairValueMid, result.FairValueHigh)

		// Score must be in [0, 100]
		assert.GreaterOrEqual(t, result.Score, 0.0,
			"Score should be >= 0, got %v", result.Score)
		assert.LessOrEqual(t, result.Score, 100.0,
			"Score should be <= 100, got %v", result.Score)
	})
}

// TestProperty1_OverallScoreRange verifies that the overall valuation score is
// always in [0, 100] for any valid random input.
func TestProperty1_OverallScoreRange(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := genValuationInput(t)
		result := CalculateValuation(input)

		assert.GreaterOrEqual(t, result.Overall, 0.0,
			"Overall score should be >= 0, got %v", result.Overall)
		assert.LessOrEqual(t, result.Overall, 100.0,
			"Overall score should be <= 100, got %v", result.Overall)
	})
}

// TestProperty1_StatusLabelCorrectness verifies that the status label matches
// the overall score: "undervalued" when < 33, "fair" when 33-66, "overvalued" when > 66.
func TestProperty1_StatusLabelCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := genValuationInput(t)
		result := CalculateValuation(input)

		switch {
		case result.Overall < 33:
			assert.Equal(t, "undervalued", result.Status,
				"Score %.2f should map to 'undervalued'", result.Overall)
		case result.Overall > 66:
			assert.Equal(t, "overvalued", result.Status,
				"Score %.2f should map to 'overvalued'", result.Overall)
		default:
			assert.Equal(t, "fair", result.Status,
				"Score %.2f should map to 'fair'", result.Overall)
		}
	})
}

// TestProperty1_RadarDataInvariants verifies that RadarData has exactly 6 entries
// and all scores are in [0, 100].
func TestProperty1_RadarDataInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := genValuationInput(t)
		result := CalculateValuation(input)

		require.Len(t, result.RadarData, 6,
			"RadarData should have exactly 6 entries")

		for _, point := range result.RadarData {
			assert.GreaterOrEqual(t, point.Score, 0.0,
				"RadarData score for %s should be >= 0, got %v", point.Dimension, point.Score)
			assert.LessOrEqual(t, point.Score, 100.0,
				"RadarData score for %s should be <= 100, got %v", point.Dimension, point.Score)
		}
	})
}

// Feature: eth-valuation-dashboard, Property 11: Stock-to-Flow 模型计算
//
// For any positive current stock and positive annual flow, the Stock-to-Flow
// ratio SHALL equal currentStock / annualFlow, and the model predicted price
// SHALL be a deterministic function of the S2F ratio. The deviation SHALL equal
// (currentPrice - modelPrice) / modelPrice × 100.
//
// **Validates: Requirements 7.4**

// TestProperty11_StockToFlowModelCalculation verifies that for any randomly
// generated positive stock, positive flow, and positive price, the S2F ratio,
// model price, deviation, and score are computed correctly.
func TestProperty11_StockToFlowModelCalculation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		stock := genPositiveFloat().Draw(t, "stock")
		flow := genPositiveFloat().Draw(t, "flow")
		price := genPositiveFloat().Draw(t, "price")

		input := StockToFlowInput{
			CurrentStock: stock,
			AnnualFlow:   flow,
			CurrentPrice: price,
		}

		result := CalculateStockToFlow(input)

		// S2F ratio = stock / flow
		expectedRatio := stock / flow
		assert.InDelta(t, expectedRatio, result.Ratio, 1e-9,
			"S2F ratio should equal stock/flow: expected %v, got %v", expectedRatio, result.Ratio)

		// Model price must be positive for positive S2F ratio
		assert.Greater(t, result.ModelPrice, 0.0,
			"ModelPrice should be > 0 for positive S2F ratio, got %v", result.ModelPrice)

		// Model price = exp(3.0 * ln(s2fRatio) + 1.0)
		expectedModelPrice := math.Exp(3.0*math.Log(expectedRatio) + 1.0)
		assert.InDelta(t, expectedModelPrice, result.ModelPrice, 1e-6,
			"ModelPrice should equal exp(3*ln(ratio)+1): expected %v, got %v", expectedModelPrice, result.ModelPrice)

		// Deviation = (currentPrice - modelPrice) / modelPrice * 100
		expectedDeviation := (price - result.ModelPrice) / result.ModelPrice * 100
		assert.InDelta(t, expectedDeviation, result.Deviation, 1e-6,
			"Deviation should equal (price-modelPrice)/modelPrice*100: expected %v, got %v", expectedDeviation, result.Deviation)

		// Score must be in [0, 100]
		assert.GreaterOrEqual(t, result.Score, 0.0,
			"Score should be >= 0, got %v", result.Score)
		assert.LessOrEqual(t, result.Score, 100.0,
			"Score should be <= 100, got %v", result.Score)
	})
}
