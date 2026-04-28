package calc

// Feature: eth-valuation-dashboard, Property 9: 移动平均计算正确性
//
// For any array of at least 7 non-negative daily values, the 7-day moving average
// for the last day SHALL equal the arithmetic mean of the last 7 values:
// MA7 = (sum of last 7 values) / 7.
//
// **Validates: Requirements 4.1**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// genNonNegativeValue generates a non-negative float64 suitable for daily values
// (e.g., daily active addresses, transaction counts). The range is constrained
// to avoid extreme magnitude differences that cause floating-point drift in
// sliding window computations.
func genNonNegativeValue() *rapid.Generator[float64] {
	return rapid.Float64Range(0, 1e9)
}

// genNonNegativeSlice generates a slice of at least `minLen` non-negative float64 values.
func genNonNegativeSlice(minLen, maxLen int) *rapid.Generator[[]float64] {
	return rapid.Custom[[]float64](func(t *rapid.T) []float64 {
		n := rapid.IntRange(minLen, maxLen).Draw(t, "length")
		values := make([]float64, n)
		for i := range values {
			values[i] = genNonNegativeValue().Draw(t, "value")
		}
		return values
	})
}

// TestProperty9_MovingAverage_7Day verifies that for any array of at least 7
// non-negative daily values, the 7-day moving average equals the arithmetic
// mean of the last 7 values.
func TestProperty9_MovingAverage_7Day(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeSlice(7, 200).Draw(t, "values")

		result := MovingAverage(values, 7)
		require.NotNil(t, result, "MovingAverage should not return nil for len >= 7")

		// Compute expected: arithmetic mean of last 7 values.
		sum := 0.0
		for i := len(values) - 7; i < len(values); i++ {
			sum += values[i]
		}
		expected := sum / 7.0

		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-9+1e-12,
			"MA7 should equal arithmetic mean of last 7 values")
	})
}

// TestProperty9_MovingAverage_InsufficientData verifies that MovingAverage
// returns nil when the input has fewer values than the window size.
func TestProperty9_MovingAverage_InsufficientData(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 6).Draw(t, "length")
		values := make([]float64, n)
		for i := range values {
			values[i] = genNonNegativeValue().Draw(t, "value")
		}

		result := MovingAverage(values, 7)
		assert.Nil(t, result, "MovingAverage should return nil when len(values) < window")
	})
}

// TestProperty9_MovingAverageAll_LastEqualsMovingAverage verifies that the last
// element of MovingAverageAll equals MovingAverage for the same window.
func TestProperty9_MovingAverageAll_LastEqualsMovingAverage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeSlice(7, 200).Draw(t, "values")

		all := MovingAverageAll(values, 7)
		single := MovingAverage(values, 7)

		require.NotEmpty(t, all)
		require.NotNil(t, single)

		assert.Equal(t, *single, all[len(all)-1],
			"Last element of MovingAverageAll should equal MovingAverage")
	})
}

// TestProperty9_MovingAverageAll_EachPositionCorrect verifies that each element
// in MovingAverageAll equals the arithmetic mean of the corresponding window.
func TestProperty9_MovingAverageAll_EachPositionCorrect(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeSlice(7, 100).Draw(t, "values")

		all := MovingAverageAll(values, 7)
		expectedLen := len(values) - 7 + 1
		require.Len(t, all, expectedLen)

		for i := 0; i < expectedLen; i++ {
			sum := 0.0
			for j := i; j < i+7; j++ {
				sum += values[j]
			}
			expected := sum / 7.0
			assert.Equal(t, expected, all[i],
				"MovingAverageAll[%d] should equal mean of values[%d:%d]", i, i, i+7)
		}
	})
}

// TestProperty9_MovingAverage_NonNegativeResult verifies that for non-negative
// input values, the moving average is also non-negative.
func TestProperty9_MovingAverage_NonNegativeResult(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeSlice(7, 200).Draw(t, "values")

		result := MovingAverage(values, 7)
		require.NotNil(t, result)
		assert.GreaterOrEqual(t, *result, 0.0,
			"Moving average of non-negative values should be non-negative")
	})
}
