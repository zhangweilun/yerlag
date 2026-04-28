package calc

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sumFloat64 returns the sum of a float64 slice.
func sumFloat64(s []float64) float64 {
	total := 0.0
	for _, v := range s {
		total += v
	}
	return total
}

// sumMapValues returns the sum of all values in a map.
func sumMapValues(m map[string]float64) float64 {
	total := 0.0
	for _, v := range m {
		total += v
	}
	return total
}

// --- CalculateShares tests ---

func TestCalculateShares_EmptyInput(t *testing.T) {
	result := CalculateShares([]float64{})
	assert.Empty(t, result)
}

func TestCalculateShares_SingleElement(t *testing.T) {
	result := CalculateShares([]float64{42.0})
	assert.Len(t, result, 1)
	assert.Equal(t, 100.0, result[0])
}

func TestCalculateShares_AllZeros(t *testing.T) {
	result := CalculateShares([]float64{0, 0, 0, 0})
	assert.Len(t, result, 4)
	for _, s := range result {
		assert.InDelta(t, 25.0, s, 1e-10)
	}
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_EqualValues(t *testing.T) {
	result := CalculateShares([]float64{10, 10, 10, 10})
	assert.Len(t, result, 4)
	for _, s := range result {
		assert.InDelta(t, 25.0, s, 1e-10)
	}
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_KnownDistribution(t *testing.T) {
	// 50, 30, 20 → 50%, 30%, 20%
	result := CalculateShares([]float64{50, 30, 20})
	assert.Len(t, result, 3)
	assert.InDelta(t, 50.0, result[0], 0.01)
	assert.InDelta(t, 30.0, result[1], 0.01)
	assert.InDelta(t, 20.0, result[2], 0.01)
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_OneNonZero(t *testing.T) {
	result := CalculateShares([]float64{0, 0, 100, 0})
	assert.Len(t, result, 4)
	assert.Equal(t, 0.0, result[0])
	assert.Equal(t, 0.0, result[1])
	assert.Equal(t, 100.0, result[2])
	assert.Equal(t, 0.0, result[3])
}

func TestCalculateShares_SumExactly100(t *testing.T) {
	// Values that cause floating-point rounding issues: thirds
	result := CalculateShares([]float64{1, 1, 1})
	assert.Len(t, result, 3)
	assert.Equal(t, 100.0, sumFloat64(result), "shares must sum to exactly 100")
}

func TestCalculateShares_LargeValues(t *testing.T) {
	result := CalculateShares([]float64{1e12, 5e11, 3e11, 2e11})
	assert.Len(t, result, 4)
	assert.InDelta(t, 50.0, result[0], 0.01)
	assert.InDelta(t, 25.0, result[1], 0.01)
	assert.InDelta(t, 15.0, result[2], 0.01)
	assert.InDelta(t, 10.0, result[3], 0.01)
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_SmallValues(t *testing.T) {
	result := CalculateShares([]float64{0.001, 0.002, 0.007})
	assert.Len(t, result, 3)
	assert.InDelta(t, 10.0, result[0], 0.01)
	assert.InDelta(t, 20.0, result[1], 0.01)
	assert.InDelta(t, 70.0, result[2], 0.01)
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_IndividualShareFormula(t *testing.T) {
	values := []float64{100, 200, 300, 400}
	total := 1000.0
	result := CalculateShares(values)
	for i, v := range values {
		expected := (v / total) * 100.0
		assert.InDelta(t, expected, result[i], 0.01,
			"share[%d] should be (%.2f / %.2f) * 100", i, v, total)
	}
}

func TestCalculateShares_NonNegativeShares(t *testing.T) {
	result := CalculateShares([]float64{0, 0, 5, 0, 0})
	for i, s := range result {
		assert.True(t, s >= 0, "share[%d] should be non-negative, got %f", i, s)
	}
}

func TestCalculateShares_ProtocolTVLShares(t *testing.T) {
	// Simulates protocol TVL distribution (Req 3.4, 5.3)
	protocols := []float64{
		25_000_000_000, // Lido
		15_000_000_000, // Aave
		10_000_000_000, // MakerDAO
		8_000_000_000,  // Uniswap
		5_000_000_000,  // Compound
	}
	result := CalculateShares(protocols)
	assert.Len(t, result, 5)
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
	// Lido should have the largest share
	assert.True(t, result[0] > result[1])
}

func TestCalculateShares_SupplyDistribution(t *testing.T) {
	// Simulates supply distribution (Req 17.2)
	// staked, defi, exchange, other
	supply := []float64{30_000_000, 20_000_000, 15_000_000, 55_000_000}
	result := CalculateShares(supply)
	assert.Len(t, result, 4)
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
	assert.InDelta(t, 25.0, result[0], 0.01)  // staked
	assert.InDelta(t, 45.83, result[3], 0.01) // other (largest)
}

// --- CalculateSharesNamed tests ---

func TestCalculateSharesNamed_EmptyInput(t *testing.T) {
	result := CalculateSharesNamed(map[string]float64{})
	assert.Empty(t, result)
}

func TestCalculateSharesNamed_SingleElement(t *testing.T) {
	result := CalculateSharesNamed(map[string]float64{"only": 42.0})
	assert.Len(t, result, 1)
	assert.Equal(t, 100.0, result["only"])
}

func TestCalculateSharesNamed_AllZeros(t *testing.T) {
	result := CalculateSharesNamed(map[string]float64{
		"a": 0, "b": 0, "c": 0,
	})
	assert.Len(t, result, 3)
	for _, s := range result {
		assert.InDelta(t, 100.0/3.0, s, 1e-10)
	}
	assert.InDelta(t, 100.0, sumMapValues(result), 0.01)
}

func TestCalculateSharesNamed_KnownDistribution(t *testing.T) {
	result := CalculateSharesNamed(map[string]float64{
		"Lido":       60,
		"RocketPool": 25,
		"Coinbase":   15,
	})
	assert.Len(t, result, 3)
	assert.InDelta(t, 60.0, result["Lido"], 0.01)
	assert.InDelta(t, 25.0, result["RocketPool"], 0.01)
	assert.InDelta(t, 15.0, result["Coinbase"], 0.01)
	assert.InDelta(t, 100.0, sumMapValues(result), 0.01)
}

func TestCalculateSharesNamed_SumExactly100(t *testing.T) {
	// Thirds cause rounding issues
	result := CalculateSharesNamed(map[string]float64{
		"a": 1, "b": 1, "c": 1,
	})
	assert.Equal(t, 100.0, sumMapValues(result), "named shares must sum to exactly 100")
}

func TestCalculateSharesNamed_ETFMarketShares(t *testing.T) {
	// Simulates ETF market share distribution (Req 8.7)
	result := CalculateSharesNamed(map[string]float64{
		"BlackRock": 500_000,
		"Fidelity":  200_000,
		"Grayscale": 150_000,
		"Others":    50_000,
	})
	assert.Len(t, result, 4)
	assert.InDelta(t, 100.0, sumMapValues(result), 0.01)
	// BlackRock should have the largest share
	for k, v := range result {
		if k != "BlackRock" {
			assert.True(t, result["BlackRock"] > v,
				"BlackRock share should be largest, but %s has %.2f", k, v)
		}
	}
}

func TestCalculateSharesNamed_ClientDiversity(t *testing.T) {
	// Simulates client diversity shares (Req 11.5)
	result := CalculateSharesNamed(map[string]float64{
		"Geth":       45,
		"Prysm":      30,
		"Lighthouse": 15,
		"Teku":       7,
		"Nimbus":     3,
	})
	assert.Len(t, result, 5)
	assert.InDelta(t, 100.0, sumMapValues(result), 0.01)
	assert.InDelta(t, 45.0, result["Geth"], 0.01)
}

func TestCalculateSharesNamed_IndividualShareFormula(t *testing.T) {
	components := map[string]float64{
		"a": 100,
		"b": 200,
		"c": 300,
	}
	total := 600.0
	result := CalculateSharesNamed(components)
	for k, v := range components {
		expected := (v / total) * 100.0
		assert.InDelta(t, expected, result[k], 0.01,
			"share[%s] should be (%.2f / %.2f) * 100", k, v, total)
	}
}

func TestCalculateSharesNamed_NonNegativeShares(t *testing.T) {
	result := CalculateSharesNamed(map[string]float64{
		"a": 0, "b": 5, "c": 0,
	})
	for k, s := range result {
		assert.True(t, s >= 0, "share[%s] should be non-negative, got %f", k, s)
	}
}

func TestCalculateShares_ManyEqualSmallValues(t *testing.T) {
	// 10 equal values → each should be 10%
	values := make([]float64, 10)
	for i := range values {
		values[i] = 7.0
	}
	result := CalculateShares(values)
	assert.Len(t, result, 10)
	for _, s := range result {
		assert.InDelta(t, 10.0, s, 0.01)
	}
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_VerySmallAndLarge(t *testing.T) {
	// One dominant value and one tiny value
	result := CalculateShares([]float64{1e-10, 1e10})
	assert.Len(t, result, 2)
	assert.True(t, result[0] < 0.01, "tiny value should have near-zero share")
	assert.InDelta(t, 100.0, result[1], 0.01)
	assert.InDelta(t, 100.0, sumFloat64(result), 0.01)
}

func TestCalculateShares_NaNNotProduced(t *testing.T) {
	result := CalculateShares([]float64{0, 0})
	for i, s := range result {
		assert.False(t, math.IsNaN(s), "share[%d] should not be NaN", i)
		assert.False(t, math.IsInf(s, 0), "share[%d] should not be Inf", i)
	}
}
