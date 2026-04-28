package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- MovingAverage unit tests ---

func TestMovingAverage(t *testing.T) {
	t.Run("basic 7-day MA", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5, 6, 7}
		r := MovingAverage(values, 7)
		require.NotNil(t, r)
		assert.InDelta(t, 4.0, *r, 1e-10)
	})

	t.Run("more values than window", func(t *testing.T) {
		values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90}
		// Last 3 values: 70, 80, 90 → mean = 80
		r := MovingAverage(values, 3)
		require.NotNil(t, r)
		assert.InDelta(t, 80.0, *r, 1e-10)
	})

	t.Run("window equals length", func(t *testing.T) {
		values := []float64{2, 4, 6}
		r := MovingAverage(values, 3)
		require.NotNil(t, r)
		assert.InDelta(t, 4.0, *r, 1e-10)
	})

	t.Run("window of 1", func(t *testing.T) {
		values := []float64{5, 10, 15}
		r := MovingAverage(values, 1)
		require.NotNil(t, r)
		assert.InDelta(t, 15.0, *r, 1e-10)
	})

	t.Run("insufficient data returns nil", func(t *testing.T) {
		values := []float64{1, 2, 3}
		r := MovingAverage(values, 7)
		assert.Nil(t, r)
	})

	t.Run("empty slice returns nil", func(t *testing.T) {
		r := MovingAverage([]float64{}, 3)
		assert.Nil(t, r)
	})

	t.Run("zero window returns nil", func(t *testing.T) {
		r := MovingAverage([]float64{1, 2, 3}, 0)
		assert.Nil(t, r)
	})

	t.Run("negative window returns nil", func(t *testing.T) {
		r := MovingAverage([]float64{1, 2, 3}, -1)
		assert.Nil(t, r)
	})

	t.Run("all zeros", func(t *testing.T) {
		values := []float64{0, 0, 0, 0, 0, 0, 0}
		r := MovingAverage(values, 7)
		require.NotNil(t, r)
		assert.Equal(t, 0.0, *r)
	})
}

// --- MovingAverageAll unit tests ---

func TestMovingAverageAll(t *testing.T) {
	t.Run("basic sliding window", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5}
		result := MovingAverageAll(values, 3)
		require.Len(t, result, 3)
		assert.InDelta(t, 2.0, result[0], 1e-10) // (1+2+3)/3
		assert.InDelta(t, 3.0, result[1], 1e-10) // (2+3+4)/3
		assert.InDelta(t, 4.0, result[2], 1e-10) // (3+4+5)/3
	})

	t.Run("window equals length", func(t *testing.T) {
		values := []float64{10, 20, 30}
		result := MovingAverageAll(values, 3)
		require.Len(t, result, 1)
		assert.InDelta(t, 20.0, result[0], 1e-10)
	})

	t.Run("window of 1 returns original values", func(t *testing.T) {
		values := []float64{5, 10, 15}
		result := MovingAverageAll(values, 1)
		require.Len(t, result, 3)
		assert.InDelta(t, 5.0, result[0], 1e-10)
		assert.InDelta(t, 10.0, result[1], 1e-10)
		assert.InDelta(t, 15.0, result[2], 1e-10)
	})

	t.Run("insufficient data returns empty", func(t *testing.T) {
		values := []float64{1, 2}
		result := MovingAverageAll(values, 5)
		assert.Empty(t, result)
	})

	t.Run("zero window returns empty", func(t *testing.T) {
		result := MovingAverageAll([]float64{1, 2, 3}, 0)
		assert.Empty(t, result)
	})

	t.Run("negative window returns empty", func(t *testing.T) {
		result := MovingAverageAll([]float64{1, 2, 3}, -1)
		assert.Empty(t, result)
	})

	t.Run("last element matches MovingAverage", func(t *testing.T) {
		values := []float64{10, 20, 30, 40, 50, 60, 70}
		all := MovingAverageAll(values, 7)
		single := MovingAverage(values, 7)
		require.Len(t, all, 1)
		require.NotNil(t, single)
		assert.InDelta(t, *single, all[len(all)-1], 1e-10)
	})
}
