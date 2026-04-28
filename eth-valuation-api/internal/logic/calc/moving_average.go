package calc

// MovingAverage computes the moving average of the last `window` values.
// Returns nil if len(values) < window or window <= 0.
func MovingAverage(values []float64, window int) *float64 {
	if window <= 0 || len(values) < window {
		return nil
	}
	sum := 0.0
	start := len(values) - window
	for i := start; i < len(values); i++ {
		sum += values[i]
	}
	result := sum / float64(window)
	return &result
}

// MovingAverageAll computes moving averages for all positions where enough data
// exists. For each position i (where i >= window-1), the moving average is the
// arithmetic mean of values[i-window+1 .. i]. Returns an empty slice if
// len(values) < window or window <= 0.
func MovingAverageAll(values []float64, window int) []float64 {
	if window <= 0 || len(values) < window {
		return []float64{}
	}
	count := len(values) - window + 1
	result := make([]float64, count)
	w := float64(window)

	for i := 0; i < count; i++ {
		sum := 0.0
		for j := i; j < i+window; j++ {
			sum += values[j]
		}
		result[i] = sum / w
	}
	return result
}
