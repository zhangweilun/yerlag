package calc

// Feature: eth-valuation-dashboard, Property 6: 百分位信号分类正确性
//
// For any metric value and its corresponding historical distribution, the percentile
// calculation SHALL correctly position the value within the distribution, and the signal
// classification SHALL be "overvalued" (or "eth_overvalued") when percentile > 90,
// "undervalued" (or "eth_undervalued") when percentile < 10, and "neutral" otherwise.
// This applies to NVT Ratio signals and ETH/BTC ratio signals.
//
// **Validates: Requirements 4.6, 4.7, 7.6, 12.4, 12.5, 12.6**

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// genHistorySlice generates a non-empty slice of float64 values representing a
// historical distribution. Length is between 1 and 500.
func genHistorySlice() *rapid.Generator[[]float64] {
	return rapid.SliceOfN(rapid.Float64Range(-1e12, 1e12), 1, 500)
}

// genCurrentValue generates a random current metric value.
func genCurrentValue() *rapid.Generator[float64] {
	return rapid.Float64Range(-1e12, 1e12)
}

// genPercentile generates a random percentile value in [0, 100].
func genPercentile() *rapid.Generator[float64] {
	return rapid.Float64Range(0, 100)
}

// --- Property 1: Percentile is always in [0, 100] range ---

// TestProperty6_PercentileAlwaysInRange verifies that for any non-empty history
// and any current value, CalculatePercentile returns a value in [0, 100].
func TestProperty6_PercentileAlwaysInRange(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		history := genHistorySlice().Draw(t, "history")
		current := genCurrentValue().Draw(t, "currentValue")

		percentile := CalculatePercentile(history, current)

		assert.GreaterOrEqual(t, percentile, 0.0,
			"Percentile should be >= 0, got %v for history=%v, current=%v", percentile, history, current)
		assert.LessOrEqual(t, percentile, 100.0,
			"Percentile should be <= 100, got %v for history=%v, current=%v", percentile, history, current)
	})
}

// --- Property 2: ClassifySignal returns correct label based on percentile thresholds ---

// TestProperty6_ClassifySignal_Correctness verifies that ClassifySignal returns
// "overvalued" when percentile > 90, "undervalued" when percentile < 10, and
// "neutral" otherwise.
func TestProperty6_ClassifySignal_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		percentile := genPercentile().Draw(t, "percentile")

		signal := ClassifySignal(percentile)

		if percentile > 90 {
			assert.Equal(t, "overvalued", signal,
				"ClassifySignal(%v) should be 'overvalued' when percentile > 90", percentile)
		} else if percentile < 10 {
			assert.Equal(t, "undervalued", signal,
				"ClassifySignal(%v) should be 'undervalued' when percentile < 10", percentile)
		} else {
			assert.Equal(t, "neutral", signal,
				"ClassifySignal(%v) should be 'neutral' when 10 <= percentile <= 90", percentile)
		}
	})
}

// --- Property 3: ClassifyETHBTCSignal returns correct label based on percentile thresholds ---

// TestProperty6_ClassifyETHBTCSignal_Correctness verifies that ClassifyETHBTCSignal
// returns "eth_overvalued" when percentile > 90, "eth_undervalued" when percentile < 10,
// and "neutral" otherwise.
func TestProperty6_ClassifyETHBTCSignal_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		percentile := genPercentile().Draw(t, "percentile")

		signal := ClassifyETHBTCSignal(percentile)

		if percentile > 90 {
			assert.Equal(t, "eth_overvalued", signal,
				"ClassifyETHBTCSignal(%v) should be 'eth_overvalued' when percentile > 90", percentile)
		} else if percentile < 10 {
			assert.Equal(t, "eth_undervalued", signal,
				"ClassifyETHBTCSignal(%v) should be 'eth_undervalued' when percentile < 10", percentile)
		} else {
			assert.Equal(t, "neutral", signal,
				"ClassifyETHBTCSignal(%v) should be 'neutral' when 10 <= percentile <= 90", percentile)
		}
	})
}

// --- Property 4: Percentile of a value above all history values is 100 ---

// TestProperty6_PercentileAboveAllIs100 verifies that when the current value is
// strictly greater than all values in the history, the percentile is 100.
func TestProperty6_PercentileAboveAllIs100(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		history := rapid.SliceOfN(rapid.Float64Range(-1e12, 1e12), 1, 500).Draw(t, "history")

		// Find the maximum value in history
		maxVal := history[0]
		for _, v := range history[1:] {
			if v > maxVal {
				maxVal = v
			}
		}

		// Current value is strictly above all history values
		current := maxVal + 1.0

		percentile := CalculatePercentile(history, current)

		assert.Equal(t, 100.0, percentile,
			"Percentile should be 100 when current (%v) > all history values (max=%v)", current, maxVal)
	})
}

// --- Property 5: Percentile of a value below all history values is 0 ---

// TestProperty6_PercentileBelowAllIs0 verifies that when the current value is
// strictly less than or equal to all values in the history (i.e., no history value
// is strictly less than current), the percentile is 0.
func TestProperty6_PercentileBelowAllIs0(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		history := rapid.SliceOfN(rapid.Float64Range(-1e12, 1e12), 1, 500).Draw(t, "history")

		// Find the minimum value in history
		minVal := history[0]
		for _, v := range history[1:] {
			if v < minVal {
				minVal = v
			}
		}

		// Current value is strictly below all history values
		current := minVal - 1.0

		percentile := CalculatePercentile(history, current)

		assert.Equal(t, 0.0, percentile,
			"Percentile should be 0 when current (%v) < all history values (min=%v)", current, minVal)
	})
}
