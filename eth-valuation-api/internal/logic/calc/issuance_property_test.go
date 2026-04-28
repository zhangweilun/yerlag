package calc

// Feature: eth-valuation-dashboard, Property 7: 净发行量与通胀/通缩分类
//
// For any positive new issuance amount and non-negative burn amount, the net issuance
// SHALL equal (newIssuance - burnAmount), the IsDeflationary flag SHALL be true if and
// only if netIssuance < 0, and the annual inflation rate SHALL equal
// (netIssuance / totalSupply) × 100.
//
// **Validates: Requirements 2.4, 17.4**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// genPositiveIssuance generates random positive new issuance amounts.
func genPositiveIssuance() *rapid.Generator[float64] {
	return rapid.Float64Range(1e-9, 1e12)
}

// genNonNegativeBurn generates random non-negative burn amounts.
func genNonNegativeBurn() *rapid.Generator[float64] {
	return rapid.Float64Range(0, 1e12)
}

// genPositiveTotalSupply generates random positive total supply values.
func genPositiveTotalSupply() *rapid.Generator[float64] {
	return rapid.Float64Range(1e-9, 1e15)
}

// TestProperty7_NetIssuance_Correctness verifies that for any positive new issuance
// and non-negative burn amount, NetIssuance returns newIssuance - burnAmount.
func TestProperty7_NetIssuance_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		newIssuance := genPositiveIssuance().Draw(t, "newIssuance")
		burnAmount := genNonNegativeBurn().Draw(t, "burnAmount")

		result := NetIssuance(newIssuance, burnAmount)
		expected := newIssuance - burnAmount

		assert.InDelta(t, expected, result, math.Abs(expected)*1e-10+1e-15,
			"NetIssuance(%v, %v) should equal %v", newIssuance, burnAmount, expected)
	})
}

// TestProperty7_IsDeflationary_Correctness verifies that IsDeflationary returns true
// if and only if netIssuance < 0.
func TestProperty7_IsDeflationary_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		newIssuance := genPositiveIssuance().Draw(t, "newIssuance")
		burnAmount := genNonNegativeBurn().Draw(t, "burnAmount")

		netIssuance := NetIssuance(newIssuance, burnAmount)
		isDefl := IsDeflationary(netIssuance)

		if netIssuance < 0 {
			assert.True(t, isDefl,
				"IsDeflationary should be true when netIssuance=%v < 0", netIssuance)
		} else {
			assert.False(t, isDefl,
				"IsDeflationary should be false when netIssuance=%v >= 0", netIssuance)
		}
	})
}

// TestProperty7_AnnualInflationRate_Correctness verifies that for any net issuance
// and positive total supply, AnnualInflationRate returns (netIssuance / totalSupply) * 100.
func TestProperty7_AnnualInflationRate_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		newIssuance := genPositiveIssuance().Draw(t, "newIssuance")
		burnAmount := genNonNegativeBurn().Draw(t, "burnAmount")
		totalSupply := genPositiveTotalSupply().Draw(t, "totalSupply")

		netIssuance := NetIssuance(newIssuance, burnAmount)
		result := AnnualInflationRate(netIssuance, totalSupply)

		require.NotNil(t, result, "AnnualInflationRate should not return nil for positive totalSupply")

		expected := (netIssuance / totalSupply) * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"AnnualInflationRate(%v, %v) should equal %v", netIssuance, totalSupply, expected)
	})
}

// TestProperty7_AnnualInflationRate_ZeroSupply verifies that AnnualInflationRate
// returns nil when totalSupply is zero.
func TestProperty7_AnnualInflationRate_ZeroSupply(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		newIssuance := genPositiveIssuance().Draw(t, "newIssuance")
		burnAmount := genNonNegativeBurn().Draw(t, "burnAmount")

		netIssuance := NetIssuance(newIssuance, burnAmount)
		result := AnnualInflationRate(netIssuance, 0)

		assert.Nil(t, result, "AnnualInflationRate should return nil when totalSupply is 0")
	})
}
