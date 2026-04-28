package calc

// CalculateExchangeSpreads computes the spread for each exchange relative to the
// average price across all exchanges.
// Formula: spread = (exchangePrice - averagePrice) / averagePrice * 100
// Returns an empty map if fewer than 2 exchanges are provided.
func CalculateExchangeSpreads(prices map[string]float64) map[string]float64 {
	if len(prices) < 2 {
		return map[string]float64{}
	}

	// Calculate the arithmetic mean of all exchange prices.
	var sum float64
	for _, price := range prices {
		sum += price
	}
	avg := sum / float64(len(prices))

	// If the average is zero, all prices are zero and spreads are undefined.
	if avg == 0 {
		return map[string]float64{}
	}

	spreads := make(map[string]float64, len(prices))
	for exchange, price := range prices {
		spreads[exchange] = (price - avg) / avg * 100
	}
	return spreads
}
