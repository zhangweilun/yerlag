package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- PearsonCorrelation unit tests ---

func TestPearsonCorrelation(t *testing.T) {
	t.Run("identical series returns 1", func(t *testing.T) {
		x := []float64{1, 2, 3, 4, 5}
		r := PearsonCorrelation(x, x)
		require.NotNil(t, r)
		assert.InDelta(t, 1.0, *r, 1e-10)
	})

	t.Run("negated series returns -1", func(t *testing.T) {
		x := []float64{1, 2, 3, 4, 5}
		y := []float64{-1, -2, -3, -4, -5}
		r := PearsonCorrelation(x, y)
		require.NotNil(t, r)
		assert.InDelta(t, -1.0, *r, 1e-10)
	})

	t.Run("perfectly positive linear relationship", func(t *testing.T) {
		x := []float64{10, 20, 30, 40, 50}
		y := []float64{2, 4, 6, 8, 10} // y = x/5
		r := PearsonCorrelation(x, y)
		require.NotNil(t, r)
		assert.InDelta(t, 1.0, *r, 1e-10)
	})

	t.Run("perfectly negative linear relationship", func(t *testing.T) {
		x := []float64{1, 2, 3, 4, 5}
		y := []float64{10, 8, 6, 4, 2} // y = 12 - 2x
		r := PearsonCorrelation(x, y)
		require.NotNil(t, r)
		assert.InDelta(t, -1.0, *r, 1e-10)
	})

	t.Run("two points", func(t *testing.T) {
		x := []float64{1, 3}
		y := []float64{2, 6}
		r := PearsonCorrelation(x, y)
		require.NotNil(t, r)
		assert.InDelta(t, 1.0, *r, 1e-10)
	})

	t.Run("known correlation value", func(t *testing.T) {
		// x = [1,2,3,4,5], y = [2,4,5,4,5]
		// Manual calculation: r ≈ 0.7745966692
		x := []float64{1, 2, 3, 4, 5}
		y := []float64{2, 4, 5, 4, 5}
		r := PearsonCorrelation(x, y)
		require.NotNil(t, r)
		assert.InDelta(t, 0.7745966692, *r, 1e-6)
	})

	t.Run("different lengths returns nil", func(t *testing.T) {
		x := []float64{1, 2, 3}
		y := []float64{1, 2}
		r := PearsonCorrelation(x, y)
		assert.Nil(t, r)
	})

	t.Run("fewer than 2 points returns nil", func(t *testing.T) {
		x := []float64{1}
		y := []float64{2}
		r := PearsonCorrelation(x, y)
		assert.Nil(t, r)
	})

	t.Run("empty slices returns nil", func(t *testing.T) {
		r := PearsonCorrelation([]float64{}, []float64{})
		assert.Nil(t, r)
	})

	t.Run("constant x series returns nil", func(t *testing.T) {
		x := []float64{5, 5, 5, 5}
		y := []float64{1, 2, 3, 4}
		r := PearsonCorrelation(x, y)
		assert.Nil(t, r)
	})

	t.Run("constant y series returns nil", func(t *testing.T) {
		x := []float64{1, 2, 3, 4}
		y := []float64{7, 7, 7, 7}
		r := PearsonCorrelation(x, y)
		assert.Nil(t, r)
	})

	t.Run("both constant returns nil", func(t *testing.T) {
		x := []float64{3, 3, 3}
		y := []float64{3, 3, 3}
		r := PearsonCorrelation(x, y)
		assert.Nil(t, r)
	})
}

// --- RollingCorrelation unit tests ---

func TestRollingCorrelation(t *testing.T) {
	t.Run("basic rolling window", func(t *testing.T) {
		x := []float64{1, 2, 3, 4, 5}
		y := []float64{2, 4, 6, 8, 10}
		result := RollingCorrelation(x, y, 3)
		require.Len(t, result, 3)
		// Each window has perfect positive correlation.
		for i, v := range result {
			assert.InDelta(t, 1.0, v, 1e-10, "window %d", i)
		}
	})

	t.Run("window equals length", func(t *testing.T) {
		x := []float64{1, 2, 3, 4}
		y := []float64{4, 3, 2, 1}
		result := RollingCorrelation(x, y, 4)
		require.Len(t, result, 1)
		assert.InDelta(t, -1.0, result[0], 1e-10)
	})

	t.Run("constant sub-series produces 0", func(t *testing.T) {
		x := []float64{5, 5, 5, 1, 2}
		y := []float64{1, 2, 3, 4, 5}
		result := RollingCorrelation(x, y, 3)
		require.Len(t, result, 3)
		// First window: x=[5,5,5] is constant → 0
		assert.Equal(t, 0.0, result[0])
	})

	t.Run("different lengths returns empty", func(t *testing.T) {
		x := []float64{1, 2, 3}
		y := []float64{1, 2}
		result := RollingCorrelation(x, y, 2)
		assert.Empty(t, result)
	})

	t.Run("window larger than data returns empty", func(t *testing.T) {
		x := []float64{1, 2}
		y := []float64{3, 4}
		result := RollingCorrelation(x, y, 5)
		assert.Empty(t, result)
	})

	t.Run("window less than 2 returns empty", func(t *testing.T) {
		x := []float64{1, 2, 3}
		y := []float64{4, 5, 6}
		result := RollingCorrelation(x, y, 1)
		assert.Empty(t, result)
	})

	t.Run("window of 0 returns empty", func(t *testing.T) {
		result := RollingCorrelation([]float64{1, 2}, []float64{3, 4}, 0)
		assert.Empty(t, result)
	})

	t.Run("negative window returns empty", func(t *testing.T) {
		result := RollingCorrelation([]float64{1, 2}, []float64{3, 4}, -1)
		assert.Empty(t, result)
	})
}
