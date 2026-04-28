package calc

import "math"

// PearsonCorrelation computes the Pearson correlation coefficient between two
// equal-length series. Returns nil if the series have different lengths, fewer
// than 2 points, or either series is constant (zero standard deviation).
func PearsonCorrelation(x, y []float64) *float64 {
	n := len(x)
	if n != len(y) || n < 2 {
		return nil
	}

	// Compute means.
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Compute covariance and standard deviations.
	var cov, varX, varY float64
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		cov += dx * dy
		varX += dx * dx
		varY += dy * dy
	}

	// Guard against constant series (zero variance).
	if varX == 0 || varY == 0 {
		return nil
	}

	r := cov / (math.Sqrt(varX) * math.Sqrt(varY))

	// Clamp to [-1, 1] to handle floating-point rounding.
	if r > 1 {
		r = 1
	} else if r < -1 {
		r = -1
	}

	return &r
}

// RollingCorrelation computes the Pearson correlation coefficient over a sliding
// window of the given size. For each position i (where i >= window-1), the
// correlation is computed over x[i-window+1..i] and y[i-window+1..i].
// Returns an empty slice if the series have different lengths, window < 2, or
// len(x) < window. Positions where either sub-series is constant produce 0 in
// the output (no correlation measurable).
func RollingCorrelation(x, y []float64, window int) []float64 {
	n := len(x)
	if n != len(y) || window < 2 || n < window {
		return []float64{}
	}

	count := n - window + 1
	result := make([]float64, count)

	for i := 0; i < count; i++ {
		r := PearsonCorrelation(x[i:i+window], y[i:i+window])
		if r != nil {
			result[i] = *r
		}
		// If r is nil (constant sub-series), result[i] stays 0.
	}

	return result
}
