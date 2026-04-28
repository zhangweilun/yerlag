package calc

// Feature: eth-valuation-dashboard, Property 5: 比率与百分比计算正确性
//
// For any valid numerator and positive denominator, ratio calculations SHALL produce
// the correct result: ratio = numerator / denominator. This applies to NVT Ratio
// (marketCap / dailyVolume), MVRV Ratio (marketValue / realizedValue), P/F Ratio
// (marketCap / annualizedFeeRevenue), TVL/MarketCap ratio, TPS ratio (currentTps / maxTps),
// staking percentage (stakedEth / totalSupply × 100), ETF holdings percentage
// (totalEtfHoldings / circulatingSupply × 100), ETH dominance (ethMarketCap / totalMarketCap × 100),
// annualized burn rate (dailyBurn × 365 / totalSupply × 100), and ATH drawdown
// ((ath - current) / ath × 100).
//
// **Validates: Requirements 2.3, 3.5, 4.4, 5.4, 5.5, 6.6, 7.1, 7.2, 8.3, 10.1, 11.2, 12.3**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// genNumerator generates random numerator values (any finite float64).
func genNumerator() *rapid.Generator[float64] {
	return rapid.Float64Range(-1e15, 1e15)
}

// genPositiveDenominator generates random positive denominator values (never zero).
func genPositiveDenominator() *rapid.Generator[float64] {
	return rapid.Float64Range(1e-9, 1e15)
}

// genNonNegative generates random non-negative float64 values.
func genNonNegative() *rapid.Generator[float64] {
	return rapid.Float64Range(0, 1e15)
}

// --- SafeRatio property tests ---

// TestProperty5_SafeRatio_PositiveDenominator verifies that for any numerator and
// positive denominator, SafeRatio returns numerator/denominator.
func TestProperty5_SafeRatio_PositiveDenominator(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		num := genNumerator().Draw(t, "numerator")
		den := genPositiveDenominator().Draw(t, "denominator")

		result := SafeRatio(num, den)
		require.NotNil(t, result, "SafeRatio should not return nil for positive denominator")

		expected := num / den
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"SafeRatio(%v, %v) should equal %v", num, den, expected)
	})
}

// TestProperty5_SafeRatio_ZeroDenominator verifies that for any numerator and
// zero denominator, SafeRatio returns nil.
func TestProperty5_SafeRatio_ZeroDenominator(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		num := genNumerator().Draw(t, "numerator")

		result := SafeRatio(num, 0)
		assert.Nil(t, result, "SafeRatio(%v, 0) should return nil", num)
	})
}

// --- Specific ratio function property tests (positive denominator) ---

// TestProperty5_NVTRatio_Correctness verifies NVT Ratio = marketCap / dailyVolume.
func TestProperty5_NVTRatio_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		marketCap := genNonNegative().Draw(t, "marketCap")
		dailyVolume := genPositiveDenominator().Draw(t, "dailyVolume")

		result := NVTRatio(marketCap, dailyVolume)
		require.NotNil(t, result)

		expected := marketCap / dailyVolume
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"NVTRatio(%v, %v) should equal %v", marketCap, dailyVolume, expected)
	})
}

// TestProperty5_MVRVRatio_Correctness verifies MVRV Ratio = marketValue / realizedValue.
func TestProperty5_MVRVRatio_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		marketValue := genNonNegative().Draw(t, "marketValue")
		realizedValue := genPositiveDenominator().Draw(t, "realizedValue")

		result := MVRVRatio(marketValue, realizedValue)
		require.NotNil(t, result)

		expected := marketValue / realizedValue
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"MVRVRatio(%v, %v) should equal %v", marketValue, realizedValue, expected)
	})
}

// TestProperty5_PriceToFeeRatio_Correctness verifies P/F Ratio = marketCap / annualizedFeeRevenue.
func TestProperty5_PriceToFeeRatio_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		marketCap := genNonNegative().Draw(t, "marketCap")
		feeRevenue := genPositiveDenominator().Draw(t, "annualizedFeeRevenue")

		result := PriceToFeeRatio(marketCap, feeRevenue)
		require.NotNil(t, result)

		expected := marketCap / feeRevenue
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"PriceToFeeRatio(%v, %v) should equal %v", marketCap, feeRevenue, expected)
	})
}

// TestProperty5_TVLToMarketCapRatio_Correctness verifies TVL/MarketCap ratio.
func TestProperty5_TVLToMarketCapRatio_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tvl := genNonNegative().Draw(t, "tvl")
		marketCap := genPositiveDenominator().Draw(t, "marketCap")

		result := TVLToMarketCapRatio(tvl, marketCap)
		require.NotNil(t, result)

		expected := tvl / marketCap
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"TVLToMarketCapRatio(%v, %v) should equal %v", tvl, marketCap, expected)
	})
}

// TestProperty5_TPSRatio_Correctness verifies TPS Ratio = currentTps / maxTps.
func TestProperty5_TPSRatio_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		currentTps := genNonNegative().Draw(t, "currentTps")
		maxTps := genPositiveDenominator().Draw(t, "maxTps")

		result := TPSRatio(currentTps, maxTps)
		require.NotNil(t, result)

		expected := currentTps / maxTps
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"TPSRatio(%v, %v) should equal %v", currentTps, maxTps, expected)
	})
}

// --- Percentage function property tests ---

// TestProperty5_StakingPercentage_Correctness verifies StakingPercentage = (stakedEth / totalSupply) * 100.
func TestProperty5_StakingPercentage_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		stakedEth := genNonNegative().Draw(t, "stakedEth")
		totalSupply := genPositiveDenominator().Draw(t, "totalSupply")

		result := StakingPercentage(stakedEth, totalSupply)
		require.NotNil(t, result)

		expected := (stakedEth / totalSupply) * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"StakingPercentage(%v, %v) should equal %v", stakedEth, totalSupply, expected)
	})
}

// TestProperty5_ETFHoldingsPercentage_Correctness verifies ETFHoldingsPercentage = (holdings / supply) * 100.
func TestProperty5_ETFHoldingsPercentage_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		holdings := genNonNegative().Draw(t, "totalEtfHoldings")
		supply := genPositiveDenominator().Draw(t, "circulatingSupply")

		result := ETFHoldingsPercentage(holdings, supply)
		require.NotNil(t, result)

		expected := (holdings / supply) * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"ETFHoldingsPercentage(%v, %v) should equal %v", holdings, supply, expected)
	})
}

// TestProperty5_ETHDominance_Correctness verifies ETHDominance = (ethMarketCap / totalMarketCap) * 100.
func TestProperty5_ETHDominance_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ethMC := genNonNegative().Draw(t, "ethMarketCap")
		totalMC := genPositiveDenominator().Draw(t, "totalMarketCap")

		result := ETHDominance(ethMC, totalMC)
		require.NotNil(t, result)

		expected := (ethMC / totalMC) * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"ETHDominance(%v, %v) should equal %v", ethMC, totalMC, expected)
	})
}

// TestProperty5_AnnualizedBurnRate_Correctness verifies AnnualizedBurnRate = (dailyBurn * 365 / totalSupply) * 100.
func TestProperty5_AnnualizedBurnRate_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		dailyBurn := genNonNegative().Draw(t, "dailyBurn")
		totalSupply := genPositiveDenominator().Draw(t, "totalSupply")

		result := AnnualizedBurnRate(dailyBurn, totalSupply)
		require.NotNil(t, result)

		expected := (dailyBurn * 365 / totalSupply) * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"AnnualizedBurnRate(%v, %v) should equal %v", dailyBurn, totalSupply, expected)
	})
}

// TestProperty5_ATHDrawdown_Correctness verifies ATHDrawdown = ((ath - current) / ath) * 100.
func TestProperty5_ATHDrawdown_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ath := genPositiveDenominator().Draw(t, "ath")
		current := genNonNegative().Draw(t, "current")

		result := ATHDrawdown(ath, current)
		require.NotNil(t, result)

		expected := ((ath - current) / ath) * 100
		assert.InDelta(t, expected, *result, math.Abs(expected)*1e-10+1e-15,
			"ATHDrawdown(%v, %v) should equal %v", ath, current, expected)
	})
}

// --- Zero denominator property tests for all functions ---

// TestProperty5_AllRatios_ZeroDenominator verifies that all ratio and percentage
// functions return nil when the denominator is zero.
func TestProperty5_AllRatios_ZeroDenominator(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		num := genNumerator().Draw(t, "numerator")

		assert.Nil(t, NVTRatio(num, 0), "NVTRatio with zero denominator should return nil")
		assert.Nil(t, MVRVRatio(num, 0), "MVRVRatio with zero denominator should return nil")
		assert.Nil(t, PriceToFeeRatio(num, 0), "PriceToFeeRatio with zero denominator should return nil")
		assert.Nil(t, TVLToMarketCapRatio(num, 0), "TVLToMarketCapRatio with zero denominator should return nil")
		assert.Nil(t, TPSRatio(num, 0), "TPSRatio with zero denominator should return nil")
		assert.Nil(t, StakingPercentage(num, 0), "StakingPercentage with zero denominator should return nil")
		assert.Nil(t, ETFHoldingsPercentage(num, 0), "ETFHoldingsPercentage with zero denominator should return nil")
		assert.Nil(t, ETHDominance(num, 0), "ETHDominance with zero denominator should return nil")
		assert.Nil(t, AnnualizedBurnRate(num, 0), "AnnualizedBurnRate with zero denominator should return nil")
		assert.Nil(t, ATHDrawdown(0, num), "ATHDrawdown with zero ATH should return nil")
	})
}
