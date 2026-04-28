package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePercentile(t *testing.T) {
	t.Run("empty history returns 0", func(t *testing.T) {
		result := CalculatePercentile([]float64{}, 50)
		assert.Equal(t, 0.0, result)
	})

	t.Run("value below all returns 0", func(t *testing.T) {
		history := []float64{10, 20, 30, 40, 50}
		result := CalculatePercentile(history, 5)
		assert.Equal(t, 0.0, result)
	})

	t.Run("value above all returns 100", func(t *testing.T) {
		history := []float64{10, 20, 30, 40, 50}
		result := CalculatePercentile(history, 60)
		assert.Equal(t, 100.0, result)
	})

	t.Run("value in the middle", func(t *testing.T) {
		history := []float64{10, 20, 30, 40, 50}
		// 30 is greater than 10, 20 → 2 out of 5 → 40th percentile
		result := CalculatePercentile(history, 30)
		assert.Equal(t, 40.0, result)
	})

	t.Run("value equal to median of even distribution", func(t *testing.T) {
		// 1..100
		history := make([]float64, 100)
		for i := 0; i < 100; i++ {
			history[i] = float64(i + 1)
		}
		// 50 is greater than 1..49 → 49 out of 100 → 49th percentile
		result := CalculatePercentile(history, 50)
		assert.Equal(t, 49.0, result)
	})

	t.Run("single element equal to current", func(t *testing.T) {
		result := CalculatePercentile([]float64{42}, 42)
		assert.Equal(t, 0.0, result)
	})

	t.Run("single element less than current", func(t *testing.T) {
		result := CalculatePercentile([]float64{10}, 42)
		assert.Equal(t, 100.0, result)
	})

	t.Run("duplicate values in history", func(t *testing.T) {
		history := []float64{10, 10, 20, 20, 30}
		// 20 is greater than 10, 10 → 2 out of 5 → 40th percentile
		result := CalculatePercentile(history, 20)
		assert.Equal(t, 40.0, result)
	})

	t.Run("all same values equal to current", func(t *testing.T) {
		history := []float64{50, 50, 50, 50, 50}
		result := CalculatePercentile(history, 50)
		assert.Equal(t, 0.0, result)
	})

	t.Run("all same values less than current", func(t *testing.T) {
		history := []float64{50, 50, 50, 50, 50}
		result := CalculatePercentile(history, 51)
		assert.Equal(t, 100.0, result)
	})

	t.Run("high percentile value", func(t *testing.T) {
		// 1..100
		history := make([]float64, 100)
		for i := 0; i < 100; i++ {
			history[i] = float64(i + 1)
		}
		// 95 is greater than 1..94 → 94 out of 100 → 94th percentile
		result := CalculatePercentile(history, 95)
		assert.Equal(t, 94.0, result)
	})

	t.Run("low percentile value", func(t *testing.T) {
		// 1..100
		history := make([]float64, 100)
		for i := 0; i < 100; i++ {
			history[i] = float64(i + 1)
		}
		// 5 is greater than 1..4 → 4 out of 100 → 4th percentile
		result := CalculatePercentile(history, 5)
		assert.Equal(t, 4.0, result)
	})
}

func TestClassifySignal(t *testing.T) {
	t.Run("overvalued when percentile > 90", func(t *testing.T) {
		assert.Equal(t, "overvalued", ClassifySignal(91))
		assert.Equal(t, "overvalued", ClassifySignal(95))
		assert.Equal(t, "overvalued", ClassifySignal(100))
		assert.Equal(t, "overvalued", ClassifySignal(90.01))
	})

	t.Run("undervalued when percentile < 10", func(t *testing.T) {
		assert.Equal(t, "undervalued", ClassifySignal(9))
		assert.Equal(t, "undervalued", ClassifySignal(5))
		assert.Equal(t, "undervalued", ClassifySignal(0))
		assert.Equal(t, "undervalued", ClassifySignal(9.99))
	})

	t.Run("neutral when percentile between 10 and 90 inclusive", func(t *testing.T) {
		assert.Equal(t, "neutral", ClassifySignal(10))
		assert.Equal(t, "neutral", ClassifySignal(50))
		assert.Equal(t, "neutral", ClassifySignal(90))
		assert.Equal(t, "neutral", ClassifySignal(45.5))
	})

	t.Run("boundary at exactly 90", func(t *testing.T) {
		assert.Equal(t, "neutral", ClassifySignal(90))
	})

	t.Run("boundary at exactly 10", func(t *testing.T) {
		assert.Equal(t, "neutral", ClassifySignal(10))
	})
}

func TestClassifyETHBTCSignal(t *testing.T) {
	t.Run("eth_overvalued when percentile > 90", func(t *testing.T) {
		assert.Equal(t, "eth_overvalued", ClassifyETHBTCSignal(91))
		assert.Equal(t, "eth_overvalued", ClassifyETHBTCSignal(95))
		assert.Equal(t, "eth_overvalued", ClassifyETHBTCSignal(100))
		assert.Equal(t, "eth_overvalued", ClassifyETHBTCSignal(90.01))
	})

	t.Run("eth_undervalued when percentile < 10", func(t *testing.T) {
		assert.Equal(t, "eth_undervalued", ClassifyETHBTCSignal(9))
		assert.Equal(t, "eth_undervalued", ClassifyETHBTCSignal(5))
		assert.Equal(t, "eth_undervalued", ClassifyETHBTCSignal(0))
		assert.Equal(t, "eth_undervalued", ClassifyETHBTCSignal(9.99))
	})

	t.Run("neutral when percentile between 10 and 90 inclusive", func(t *testing.T) {
		assert.Equal(t, "neutral", ClassifyETHBTCSignal(10))
		assert.Equal(t, "neutral", ClassifyETHBTCSignal(50))
		assert.Equal(t, "neutral", ClassifyETHBTCSignal(90))
		assert.Equal(t, "neutral", ClassifyETHBTCSignal(45.5))
	})

	t.Run("boundary at exactly 90", func(t *testing.T) {
		assert.Equal(t, "neutral", ClassifyETHBTCSignal(90))
	})

	t.Run("boundary at exactly 10", func(t *testing.T) {
		assert.Equal(t, "neutral", ClassifyETHBTCSignal(10))
	})
}

func TestSortedPercentile(t *testing.T) {
	t.Run("empty history returns 0", func(t *testing.T) {
		result := SortedPercentile([]float64{}, 50)
		assert.Equal(t, 0.0, result)
	})

	t.Run("matches CalculatePercentile for basic cases", func(t *testing.T) {
		history := []float64{10, 20, 30, 40, 50}

		assert.Equal(t,
			CalculatePercentile(history, 5),
			SortedPercentile(history, 5),
		)
		assert.Equal(t,
			CalculatePercentile(history, 30),
			SortedPercentile(history, 30),
		)
		assert.Equal(t,
			CalculatePercentile(history, 60),
			SortedPercentile(history, 60),
		)
	})

	t.Run("does not mutate original history", func(t *testing.T) {
		history := []float64{50, 10, 40, 20, 30}
		original := make([]float64, len(history))
		copy(original, history)

		SortedPercentile(history, 25)

		assert.Equal(t, original, history)
	})
}
