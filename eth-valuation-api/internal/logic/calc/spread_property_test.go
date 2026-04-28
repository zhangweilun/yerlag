package calc

// Feature: eth-valuation-dashboard, Property 12: 交易所价差计算
//
// For any set of exchange prices (at least 2 exchanges), the spread for each exchange
// SHALL equal (exchangePrice - averagePrice) / averagePrice × 100, where averagePrice
// is the arithmetic mean of all exchange prices.
//
// **Validates: Requirements 6.5**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// genExchangeNames generates a slice of unique exchange names with at least 2 entries.
func genExchangeNames() *rapid.Generator[[]string] {
	allExchanges := []string{
		"Binance", "Coinbase", "OKX", "Kraken", "Bybit",
		"Bitfinex", "Huobi", "KuCoin", "Gate", "Gemini",
	}
	return rapid.Custom(func(t *rapid.T) []string {
		count := rapid.IntRange(2, len(allExchanges)).Draw(t, "exchangeCount")
		return allExchanges[:count]
	})
}

// genPositivePrice generates a random positive price.
func genPositivePrice() *rapid.Generator[float64] {
	return rapid.Float64Range(0.01, 1e6)
}

// TestProperty12_SpreadFormula_Correctness verifies that for any set of at least 2
// exchange prices, the spread for each exchange equals
// (exchangePrice - averagePrice) / averagePrice * 100.
func TestProperty12_SpreadFormula_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genExchangeNames().Draw(t, "exchanges")
		prices := make(map[string]float64, len(names))
		var sum float64
		for _, name := range names {
			p := genPositivePrice().Draw(t, "price_"+name)
			prices[name] = p
			sum += p
		}
		avg := sum / float64(len(names))

		spreads := CalculateExchangeSpreads(prices)

		assert.Len(t, spreads, len(prices), "spreads map should have same length as prices")

		for exchange, price := range prices {
			expected := (price - avg) / avg * 100
			assert.InDelta(t, expected, spreads[exchange],
				math.Abs(expected)*1e-10+1e-12,
				"spread for %s: price=%v, avg=%v", exchange, price, avg)
		}
	})
}

// TestProperty12_SpreadSum_NearZero verifies that the sum of all spreads is
// approximately zero (since spreads are deviations from the mean).
func TestProperty12_SpreadSum_NearZero(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genExchangeNames().Draw(t, "exchanges")
		prices := make(map[string]float64, len(names))
		for _, name := range names {
			prices[name] = genPositivePrice().Draw(t, "price_"+name)
		}

		spreads := CalculateExchangeSpreads(prices)

		var spreadSum float64
		for _, s := range spreads {
			spreadSum += s
		}

		assert.InDelta(t, 0.0, spreadSum, 1e-8,
			"sum of all spreads should be approximately zero, got %v", spreadSum)
	})
}

// TestProperty12_FewerThan2Exchanges verifies that fewer than 2 exchanges
// returns an empty map.
func TestProperty12_FewerThan2Exchanges(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		price := genPositivePrice().Draw(t, "price")

		// Single exchange
		spreads := CalculateExchangeSpreads(map[string]float64{"Binance": price})
		assert.Empty(t, spreads, "single exchange should return empty map")

		// Empty map
		spreads = CalculateExchangeSpreads(map[string]float64{})
		assert.Empty(t, spreads, "empty map should return empty map")
	})
}
