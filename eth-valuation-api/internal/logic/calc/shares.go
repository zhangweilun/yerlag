package calc

// CalculateShares takes an array of non-negative component values and returns
// an array of percentage shares that sum to exactly 100%.
//
// Behavior:
//   - If the input slice is empty, returns an empty slice.
//   - If all values are zero, returns equal shares (100/n for each).
//   - If only one non-zero value exists, it gets 100%.
//   - Floating-point rounding is handled by adjusting the largest share
//     so the total sums to exactly 100%.
//
// Each individual share equals (componentValue / totalValue) × 100 before
// the rounding adjustment.
func CalculateShares(values []float64) []float64 {
	n := len(values)
	if n == 0 {
		return []float64{}
	}

	total := 0.0
	for _, v := range values {
		total += v
	}

	shares := make([]float64, n)

	// All values are zero: distribute equally.
	if total == 0 {
		equal := 100.0 / float64(n)
		for i := range shares {
			shares[i] = equal
		}
		return shares
	}

	// Compute raw percentages.
	for i, v := range values {
		shares[i] = (v / total) * 100.0
	}

	// Adjust the largest share so the sum is exactly 100%.
	// Find the index of the largest share (first occurrence).
	sumOthers := 0.0
	largestIdx := 0
	largestVal := shares[0]
	for i, s := range shares {
		if s > largestVal {
			largestVal = s
			largestIdx = i
		}
	}
	for i, s := range shares {
		if i != largestIdx {
			sumOthers += s
		}
	}
	shares[largestIdx] = 100.0 - sumOthers

	return shares
}

// CalculateSharesNamed takes a map of named component values and returns
// a map of percentage shares that sum to 100%.
//
// The same rounding adjustment logic as CalculateShares is applied: the
// component with the largest share absorbs any floating-point remainder.
func CalculateSharesNamed(components map[string]float64) map[string]float64 {
	n := len(components)
	if n == 0 {
		return map[string]float64{}
	}

	total := 0.0
	for _, v := range components {
		total += v
	}

	result := make(map[string]float64, n)

	// All values are zero: distribute equally.
	if total == 0 {
		equal := 100.0 / float64(n)
		for k := range components {
			result[k] = equal
		}
		return result
	}

	// Compute raw percentages and track the largest share.
	largestKey := ""
	largestVal := -1.0
	for k, v := range components {
		share := (v / total) * 100.0
		result[k] = share
		if share > largestVal {
			largestVal = share
			largestKey = k
		}
	}

	// Adjust the largest share so the sum is exactly 100%.
	sumOthers := 0.0
	for k, s := range result {
		if k != largestKey {
			sumOthers += s
		}
	}
	result[largestKey] = 100.0 - sumOthers

	return result
}
