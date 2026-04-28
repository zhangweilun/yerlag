package calc

// SafeRatio computes numerator / denominator with divide-by-zero protection.
// Returns nil when the denominator is zero.
func SafeRatio(numerator, denominator float64) *float64 {
	if denominator == 0 {
		return nil
	}
	result := numerator / denominator
	return &result
}

// NVTRatio computes the Network Value to Transactions Ratio.
// Formula: marketCap / dailyVolume
// Returns nil when dailyVolume is zero.
func NVTRatio(marketCap, dailyVolume float64) *float64 {
	return SafeRatio(marketCap, dailyVolume)
}

// MVRVRatio computes the Market Value to Realized Value Ratio.
// Formula: marketValue / realizedValue
// Returns nil when realizedValue is zero.
func MVRVRatio(marketValue, realizedValue float64) *float64 {
	return SafeRatio(marketValue, realizedValue)
}

// PriceToFeeRatio computes the Price-to-Fee Ratio (P/F Ratio).
// Formula: marketCap / annualizedFeeRevenue
// Returns nil when annualizedFeeRevenue is zero.
func PriceToFeeRatio(marketCap, annualizedFeeRevenue float64) *float64 {
	return SafeRatio(marketCap, annualizedFeeRevenue)
}

// TVLToMarketCapRatio computes the TVL to Market Cap ratio.
// Formula: tvl / marketCap
// Returns nil when marketCap is zero.
func TVLToMarketCapRatio(tvl, marketCap float64) *float64 {
	return SafeRatio(tvl, marketCap)
}

// TPSRatio computes the TPS utilization ratio.
// Formula: currentTps / maxTps
// Returns nil when maxTps is zero.
func TPSRatio(currentTps, maxTps float64) *float64 {
	return SafeRatio(currentTps, maxTps)
}

// StakingPercentage computes the staking percentage.
// Formula: (stakedEth / totalSupply) * 100
// Returns nil when totalSupply is zero.
func StakingPercentage(stakedEth, totalSupply float64) *float64 {
	r := SafeRatio(stakedEth, totalSupply)
	if r == nil {
		return nil
	}
	result := *r * 100
	return &result
}

// ETFHoldingsPercentage computes the ETF holdings as a percentage of circulating supply.
// Formula: (totalEtfHoldings / circulatingSupply) * 100
// Returns nil when circulatingSupply is zero.
func ETFHoldingsPercentage(totalEtfHoldings, circulatingSupply float64) *float64 {
	r := SafeRatio(totalEtfHoldings, circulatingSupply)
	if r == nil {
		return nil
	}
	result := *r * 100
	return &result
}

// ETHDominance computes ETH's market cap dominance percentage.
// Formula: (ethMarketCap / totalMarketCap) * 100
// Returns nil when totalMarketCap is zero.
func ETHDominance(ethMarketCap, totalMarketCap float64) *float64 {
	r := SafeRatio(ethMarketCap, totalMarketCap)
	if r == nil {
		return nil
	}
	result := *r * 100
	return &result
}

// AnnualizedBurnRate computes the annualized burn rate as a percentage of total supply.
// Formula: (dailyBurn * 365 / totalSupply) * 100
// Returns nil when totalSupply is zero.
func AnnualizedBurnRate(dailyBurn, totalSupply float64) *float64 {
	if totalSupply == 0 {
		return nil
	}
	result := (dailyBurn * 365 / totalSupply) * 100
	return &result
}

// ATHDrawdown computes the percentage drawdown from the all-time high.
// Formula: ((ath - current) / ath) * 100
// Returns nil when ath is zero.
func ATHDrawdown(ath, current float64) *float64 {
	if ath == 0 {
		return nil
	}
	result := ((ath - current) / ath) * 100
	return &result
}
