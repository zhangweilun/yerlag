package calc

// Feature: eth-valuation-dashboard, Property 15: Grayscale 溢价/折价率计算
//
// For any positive NAV and positive market price, the premium/discount rate SHALL
// equal (marketPrice - nav) / nav × 100. A positive value indicates premium,
// a negative value indicates discount.
//
// **Validates: Requirements 9.1**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// genPositiveNAV generates random positive NAV values.
func genPositiveNAV() *rapid.Generator[float64] {
	return rapid.Float64Range(0.01, 1e6)
}

// genPositiveMarketPrice generates random positive market price values.
func genPositiveMarketPrice() *rapid.Generator[float64] {
	return rapid.Float64Range(0.01, 1e6)
}

// TestProperty15_PremiumDiscountRate_Correctness verifies that for any positive NAV
// and positive market price, PremiumDiscountRate returns (marketPrice - nav) / nav * 100.
func TestProperty15_PremiumDiscountRate_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nav := genPositiveNAV().Draw(t, "nav")
		marketPrice := genPositiveMarketPrice().Draw(t, "marketPrice")

		result := PremiumDiscountRate(nav, marketPrice)
		require.NotNil(t, result, "PremiumDiscountRate should not return nil for positive NAV")

		expected := (marketPrice - nav) / nav * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-12,
			"PremiumDiscountRate(%v, %v) should equal %v", nav, marketPrice, expected)
	})
}

// TestProperty15_PremiumSign verifies that the premium/discount rate is positive
// when marketPrice > NAV and negative when marketPrice < NAV.
func TestProperty15_PremiumSign(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nav := genPositiveNAV().Draw(t, "nav")
		marketPrice := genPositiveMarketPrice().Draw(t, "marketPrice")

		result := PremiumDiscountRate(nav, marketPrice)
		require.NotNil(t, result)

		if marketPrice > nav {
			assert.Greater(t, *result, 0.0,
				"premium should be positive when marketPrice(%v) > nav(%v)", marketPrice, nav)
		} else if marketPrice < nav {
			assert.Less(t, *result, 0.0,
				"discount should be negative when marketPrice(%v) < nav(%v)", marketPrice, nav)
		} else {
			assert.Equal(t, 0.0, *result,
				"rate should be zero when marketPrice equals nav")
		}
	})
}

// TestProperty15_ZeroNAV verifies that PremiumDiscountRate returns nil when NAV is zero.
func TestProperty15_ZeroNAV(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		marketPrice := genPositiveMarketPrice().Draw(t, "marketPrice")

		result := PremiumDiscountRate(0, marketPrice)
		assert.Nil(t, result, "PremiumDiscountRate should return nil when NAV is 0")
	})
}
