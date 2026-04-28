package calc

// NetIssuance computes the net issuance of ETH.
// Formula: newIssuance - burnAmount
func NetIssuance(newIssuance, burnAmount float64) float64 {
	return newIssuance - burnAmount
}

// IsDeflationary returns true if the net issuance is negative,
// indicating that more ETH was burned than issued.
func IsDeflationary(netIssuance float64) bool {
	return netIssuance < 0
}

// AnnualInflationRate computes the annual inflation rate as a percentage.
// Formula: (netIssuance / totalSupply) * 100
// Returns nil when totalSupply is zero.
func AnnualInflationRate(netIssuance, totalSupply float64) *float64 {
	if totalSupply == 0 {
		return nil
	}
	result := (netIssuance / totalSupply) * 100
	return &result
}
