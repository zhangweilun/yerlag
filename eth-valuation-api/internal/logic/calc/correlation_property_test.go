package calc

// Feature: eth-valuation-dashboard, Property 8: 滚动相关系数范围不变量
//
// For any two non-constant price series of equal length (≥ 2 data points), the
// Pearson correlation coefficient SHALL be in the range [-1, 1]. For identical
// series, the correlation SHALL be 1. For a series and its negation, the
// correlation SHALL be -1.
//
// **Validates: Requirements 12.2, 13.3**

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// genNonConstantPriceSeries generates a non-constant float64 slice of the given
// length. Values are drawn from a reasonable price range and the generator
// guarantees at least two distinct values so that the standard deviation is
// non-zero.
func genNonConstantPriceSeries(length int) *rapid.Generator[[]float64] {
	return rapid.Custom[[]float64](func(t *rapid.T) []float64 {
		series := make([]float64, length)
		for i := range series {
			series[i] = rapid.Float64Range(0.01, 1e6).Draw(t, "price")
		}
		// Ensure non-constant: if all values are the same, perturb the last one.
		allSame := true
		for i := 1; i < length; i++ {
			if series[i] != series[0] {
				allSame = false
				break
			}
		}
		if allSame {
			series[length-1] = series[0] + 1.0
		}
		return series
	})
}

// genNonConstantPriceSeriesPair generates two non-constant price series of the
// same randomly chosen length (between 2 and maxLen).
func genNonConstantPriceSeriesPair(maxLen int) *rapid.Generator[[2][]float64] {
	return rapid.Custom[[2][]float64](func(t *rapid.T) [2][]float64 {
		n := rapid.IntRange(2, maxLen).Draw(t, "seriesLength")
		x := genNonConstantPriceSeries(n).Draw(t, "seriesX")
		y := genNonConstantPriceSeries(n).Draw(t, "seriesY")
		return [2][]float64{x, y}
	})
}

// TestProperty8_PearsonCorrelation_RangeInvariant verifies that for any two
// non-constant price series of equal length (≥ 2), the Pearson correlation
// coefficient is in [-1, 1].
func TestProperty8_PearsonCorrelation_RangeInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		pair := genNonConstantPriceSeriesPair(200).Draw(t, "pair")
		x, y := pair[0], pair[1]

		r := PearsonCorrelation(x, y)
		require.NotNil(t, r, "PearsonCorrelation should not return nil for non-constant equal-length series")

		assert.GreaterOrEqual(t, *r, -1.0, "correlation should be >= -1")
		assert.LessOrEqual(t, *r, 1.0, "correlation should be <= 1")
	})
}

// TestProperty8_PearsonCorrelation_IdenticalSeries verifies that for any
// non-constant series, the correlation with itself is 1.
func TestProperty8_PearsonCorrelation_IdenticalSeries(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(2, 200).Draw(t, "length")
		x := genNonConstantPriceSeries(n).Draw(t, "series")

		r := PearsonCorrelation(x, x)
		require.NotNil(t, r, "PearsonCorrelation should not return nil for identical non-constant series")
		assert.InDelta(t, 1.0, *r, 1e-9, "correlation of identical series should be 1")
	})
}

// TestProperty8_PearsonCorrelation_NegatedSeries verifies that for any
// non-constant series, the correlation with its negation is -1.
func TestProperty8_PearsonCorrelation_NegatedSeries(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(2, 200).Draw(t, "length")
		x := genNonConstantPriceSeries(n).Draw(t, "series")

		negX := make([]float64, n)
		for i := range x {
			negX[i] = -x[i]
		}

		r := PearsonCorrelation(x, negX)
		require.NotNil(t, r, "PearsonCorrelation should not return nil for series and its negation")
		assert.InDelta(t, -1.0, *r, 1e-9, "correlation of series with its negation should be -1")
	})
}

// TestProperty8_RollingCorrelation_RangeInvariant verifies that every value
// produced by RollingCorrelation is in [-1, 1] for non-constant input series.
func TestProperty8_RollingCorrelation_RangeInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(4, 100).Draw(t, "seriesLength")
		window := rapid.IntRange(2, n).Draw(t, "window")
		x := genNonConstantPriceSeries(n).Draw(t, "seriesX")
		y := genNonConstantPriceSeries(n).Draw(t, "seriesY")

		result := RollingCorrelation(x, y, window)
		expectedLen := n - window + 1
		require.Len(t, result, expectedLen)

		for i, v := range result {
			assert.GreaterOrEqual(t, v, -1.0, "rolling correlation[%d] should be >= -1", i)
			assert.LessOrEqual(t, v, 1.0, "rolling correlation[%d] should be <= 1", i)
		}
	})
}

// TestProperty8_PearsonCorrelation_Symmetry verifies that corr(x, y) == corr(y, x).
func TestProperty8_PearsonCorrelation_Symmetry(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		pair := genNonConstantPriceSeriesPair(200).Draw(t, "pair")
		x, y := pair[0], pair[1]

		rXY := PearsonCorrelation(x, y)
		rYX := PearsonCorrelation(y, x)

		require.NotNil(t, rXY)
		require.NotNil(t, rYX)
		assert.InDelta(t, *rXY, *rYX, 1e-12, "Pearson correlation should be symmetric")
	})
}
