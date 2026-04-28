package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateExchangeSpreads(t *testing.T) {
	t.Run("two exchanges with equal prices", func(t *testing.T) {
		prices := map[string]float64{
			"Binance":  2000,
			"Coinbase": 2000,
		}
		spreads := CalculateExchangeSpreads(prices)
		assert.Len(t, spreads, 2)
		assert.Equal(t, 0.0, spreads["Binance"])
		assert.Equal(t, 0.0, spreads["Coinbase"])
	})

	t.Run("two exchanges with different prices", func(t *testing.T) {
		prices := map[string]float64{
			"Binance":  2010,
			"Coinbase": 1990,
		}
		// avg = 2000
		// Binance spread = (2010 - 2000) / 2000 * 100 = 0.5
		// Coinbase spread = (1990 - 2000) / 2000 * 100 = -0.5
		spreads := CalculateExchangeSpreads(prices)
		assert.Len(t, spreads, 2)
		assert.InDelta(t, 0.5, spreads["Binance"], 1e-10)
		assert.InDelta(t, -0.5, spreads["Coinbase"], 1e-10)
	})

	t.Run("three exchanges", func(t *testing.T) {
		prices := map[string]float64{
			"Binance":  3000,
			"Coinbase": 3030,
			"OKX":      2970,
		}
		// avg = 3000
		spreads := CalculateExchangeSpreads(prices)
		assert.Len(t, spreads, 3)
		assert.InDelta(t, 0.0, spreads["Binance"], 1e-10)
		assert.InDelta(t, 1.0, spreads["Coinbase"], 1e-10)
		assert.InDelta(t, -1.0, spreads["OKX"], 1e-10)
	})

	t.Run("fewer than 2 exchanges returns empty map", func(t *testing.T) {
		prices := map[string]float64{
			"Binance": 2000,
		}
		spreads := CalculateExchangeSpreads(prices)
		assert.Empty(t, spreads)
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		spreads := CalculateExchangeSpreads(map[string]float64{})
		assert.Empty(t, spreads)
	})

	t.Run("nil map returns empty map", func(t *testing.T) {
		spreads := CalculateExchangeSpreads(nil)
		assert.Empty(t, spreads)
	})
}
