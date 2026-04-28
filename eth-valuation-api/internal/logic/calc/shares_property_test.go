package calc

// Feature: eth-valuation-dashboard, Property 4: 份额/占比计算求和不变量
//
// For any set of non-negative component values (e.g., protocol TVL shares, ETF market
// shares, liquid staking shares, client diversity shares, supply distribution categories),
// the calculated percentage shares SHALL sum to 100% (within floating-point tolerance of
// ±0.01%), and each individual share SHALL equal (componentValue / totalValue) × 100.
//
// **Validates: Requirements 3.4, 5.3, 8.7, 10.5, 11.5, 17.2**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// genNonNegativeValues generates a slice of 1-20 non-negative float64 values.
func genNonNegativeValues() *rapid.Generator[[]float64] {
	return rapid.SliceOfN(
		rapid.Float64Range(0, 1e12),
		1, 20,
	)
}

// genAllZeroValues generates a slice of 1-20 zero values.
func genAllZeroValues() *rapid.Generator[[]float64] {
	return rapid.Custom(func(t *rapid.T) []float64 {
		n := rapid.IntRange(1, 20).Draw(t, "length")
		return make([]float64, n)
	})
}

// genNamedComponents generates a map of 1-20 named non-negative float64 values.
func genNamedComponents() *rapid.Generator[map[string]float64] {
	return rapid.Custom(func(t *rapid.T) map[string]float64 {
		n := rapid.IntRange(1, 20).Draw(t, "numComponents")
		result := make(map[string]float64, n)
		for i := 0; i < n; i++ {
			key := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]{0,9}`).Draw(t, "key")
			val := rapid.Float64Range(0, 1e12).Draw(t, "value")
			result[key] = val
		}
		// Ensure at least 1 entry (map keys may collide, but n>=1 guarantees at least 1)
		if len(result) == 0 {
			result["fallback"] = rapid.Float64Range(0, 1e12).Draw(t, "fallbackValue")
		}
		return result
	})
}

// --- Property 4.1: CalculateShares sum invariant ---

// TestProperty4_CalculateShares_SumTo100 verifies that for any random non-negative
// float64 array (1-20 elements), CalculateShares returns shares summing to 100% (±0.01%).
func TestProperty4_CalculateShares_SumTo100(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeValues().Draw(t, "values")

		shares := CalculateShares(values)

		assert.Len(t, shares, len(values), "output length must match input length")

		sum := 0.0
		for _, s := range shares {
			sum += s
		}
		assert.InDelta(t, 100.0, sum, 0.01,
			"shares must sum to 100%% (±0.01%%), got %v for input %v", sum, values)
	})
}

// --- Property 4.2: Individual share formula correctness ---

// TestProperty4_CalculateShares_IndividualShareFormula verifies that each individual
// share equals (componentValue / totalValue) × 100 within tolerance.
func TestProperty4_CalculateShares_IndividualShareFormula(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeValues().Draw(t, "values")

		total := 0.0
		for _, v := range values {
			total += v
		}

		// Skip all-zero case (handled separately)
		if total == 0 {
			return
		}

		shares := CalculateShares(values)

		for i, v := range values {
			expected := (v / total) * 100.0
			// The rounding adjustment on the largest share can shift it slightly,
			// so we allow a tolerance of 0.01%.
			assert.InDelta(t, expected, shares[i], 0.01,
				"share[%d] should be (%.6f / %.6f) × 100 = %.6f, got %.6f",
				i, v, total, expected, shares[i])
		}
	})
}

// --- Property 4.3: All shares are non-negative ---

// TestProperty4_CalculateShares_NonNegative verifies that all shares returned by
// CalculateShares are non-negative.
func TestProperty4_CalculateShares_NonNegative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genNonNegativeValues().Draw(t, "values")

		shares := CalculateShares(values)

		for i, s := range shares {
			assert.True(t, s >= 0,
				"share[%d] should be non-negative, got %f for input %v", i, s, values)
			assert.False(t, math.IsNaN(s),
				"share[%d] should not be NaN for input %v", i, values)
			assert.False(t, math.IsInf(s, 0),
				"share[%d] should not be Inf for input %v", i, values)
		}
	})
}

// --- Property 4.4: CalculateSharesNamed same properties hold ---

// TestProperty4_CalculateSharesNamed_SumTo100 verifies that for CalculateSharesNamed
// with random named components, shares sum to 100% (±0.01%).
func TestProperty4_CalculateSharesNamed_SumTo100(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		components := genNamedComponents().Draw(t, "components")

		shares := CalculateSharesNamed(components)

		assert.Len(t, shares, len(components),
			"output map size must match input map size")

		sum := 0.0
		for _, s := range shares {
			sum += s
		}
		assert.InDelta(t, 100.0, sum, 0.01,
			"named shares must sum to 100%% (±0.01%%), got %v", sum)
	})
}

// TestProperty4_CalculateSharesNamed_IndividualShareFormula verifies that each named
// share equals (componentValue / totalValue) × 100 within tolerance.
func TestProperty4_CalculateSharesNamed_IndividualShareFormula(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		components := genNamedComponents().Draw(t, "components")

		total := 0.0
		for _, v := range components {
			total += v
		}

		// Skip all-zero case
		if total == 0 {
			return
		}

		shares := CalculateSharesNamed(components)

		for k, v := range components {
			expected := (v / total) * 100.0
			assert.InDelta(t, expected, shares[k], 0.01,
				"share[%s] should be (%.6f / %.6f) × 100 = %.6f, got %.6f",
				k, v, total, expected, shares[k])
		}
	})
}

// TestProperty4_CalculateSharesNamed_NonNegative verifies that all named shares
// are non-negative and finite.
func TestProperty4_CalculateSharesNamed_NonNegative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		components := genNamedComponents().Draw(t, "components")

		shares := CalculateSharesNamed(components)

		for k, s := range shares {
			assert.True(t, s >= 0,
				"share[%s] should be non-negative, got %f", k, s)
			assert.False(t, math.IsNaN(s),
				"share[%s] should not be NaN", k)
			assert.False(t, math.IsInf(s, 0),
				"share[%s] should not be Inf", k)
		}
	})
}

// --- Property 4.5: All-zero values produce equal shares ---

// TestProperty4_CalculateShares_AllZerosEqualShares verifies that when all values
// are zero, shares are equal (100/n for each element).
func TestProperty4_CalculateShares_AllZerosEqualShares(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		values := genAllZeroValues().Draw(t, "zeroValues")

		shares := CalculateShares(values)
		n := len(values)
		expected := 100.0 / float64(n)

		for i, s := range shares {
			assert.InDelta(t, expected, s, 1e-10,
				"share[%d] should be 100/%d = %.6f when all values are zero, got %.6f",
				i, n, expected, s)
		}

		sum := 0.0
		for _, s := range shares {
			sum += s
		}
		assert.InDelta(t, 100.0, sum, 0.01,
			"all-zero shares must still sum to 100%%, got %v", sum)
	})
}

// TestProperty4_CalculateSharesNamed_AllZerosEqualShares verifies that when all
// named component values are zero, shares are equal (100/n for each).
func TestProperty4_CalculateSharesNamed_AllZerosEqualShares(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 20).Draw(t, "numComponents")
		components := make(map[string]float64, n)
		for i := 0; i < n; i++ {
			key := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]{0,9}`).Draw(t, "key")
			components[key] = 0
		}
		if len(components) == 0 {
			components["fallback"] = 0
		}

		shares := CalculateSharesNamed(components)
		expected := 100.0 / float64(len(components))

		for k, s := range shares {
			assert.InDelta(t, expected, s, 1e-10,
				"share[%s] should be 100/%d = %.6f when all values are zero, got %.6f",
				k, len(components), expected, s)
		}

		sum := 0.0
		for _, s := range shares {
			sum += s
		}
		assert.InDelta(t, 100.0, sum, 0.01,
			"all-zero named shares must still sum to 100%%, got %v", sum)
	})
}
