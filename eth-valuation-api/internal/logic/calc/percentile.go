package calc

import "sort"

// CalculatePercentile computes the percentile (0-100) of currentValue within
// the given historical distribution. It counts how many values in history are
// strictly less than currentValue, then divides by the total count.
// Returns 0 if history is empty.
func CalculatePercentile(history []float64, currentValue float64) float64 {
	n := len(history)
	if n == 0 {
		return 0
	}

	count := 0
	for _, v := range history {
		if v < currentValue {
			count++
		}
	}

	return float64(count) / float64(n) * 100
}

// ClassifySignal classifies a metric based on its historical percentile.
// Returns "overvalued" if percentile > 90, "undervalued" if percentile < 10,
// "neutral" otherwise.
// Applied to NVT Ratio signals.
func ClassifySignal(percentile float64) string {
	if percentile > 90 {
		return "overvalued"
	}
	if percentile < 10 {
		return "undervalued"
	}
	return "neutral"
}

// ClassifyETHBTCSignal classifies the ETH/BTC ratio based on its historical percentile.
// Returns "eth_overvalued" if percentile > 90, "eth_undervalued" if percentile < 10,
// "neutral" otherwise.
// Applied to ETH/BTC ratio signals.
func ClassifyETHBTCSignal(percentile float64) string {
	if percentile > 90 {
		return "eth_overvalued"
	}
	if percentile < 10 {
		return "eth_undervalued"
	}
	return "neutral"
}

// SortedPercentile computes the percentile using a sorted copy of history
// for potentially better performance on large datasets. Behavior is identical
// to CalculatePercentile.
func SortedPercentile(history []float64, currentValue float64) float64 {
	n := len(history)
	if n == 0 {
		return 0
	}

	sorted := make([]float64, n)
	copy(sorted, history)
	sort.Float64s(sorted)

	// Use sort.SearchFloat64s to find the insertion point
	idx := sort.SearchFloat64s(sorted, currentValue)

	return float64(idx) / float64(n) * 100
}
