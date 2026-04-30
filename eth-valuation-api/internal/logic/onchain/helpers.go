package onchain

import (
	"encoding/json"
	"math/big"
	"sort"

	"eth-valuation-api/internal/types"
)

// parseWeiToEth converts a string representation of wei to ETH (float64).
func parseWeiToEth(weiStr string) float64 {
	if weiStr == "" {
		return 0
	}

	wei, _, err := big.ParseFloat(weiStr, 10, 256, big.ToNearestEven)
	if err != nil || wei == nil {
		return 0
	}

	ethDivisor := new(big.Float).SetFloat64(1e18)
	eth := new(big.Float).Quo(wei, ethDivisor)

	result, _ := eth.Float64()
	return result
}

// dailyDataPoint represents a single daily data point from Etherscan.
type dailyDataPoint struct {
	UTCDate   string `json:"UTCDate"`
	UnixTS    string `json:"unixTimeStamp"`
	Value     string `json:"value"`
	DailyBurn string `json:"dailyBurntFees,omitempty"`
}

// parseDailyTimeSeries parses raw JSON daily time series data from Etherscan
// into a slice of TimeSeriesPoint.
func parseDailyTimeSeries(raw json.RawMessage) []types.TimeSeriesPoint {
	if raw == nil {
		return nil
	}

	var points []dailyDataPoint
	if err := json.Unmarshal(raw, &points); err != nil {
		return nil
	}

	result := make([]types.TimeSeriesPoint, 0, len(points))
	for _, p := range points {
		ts := parseTimestamp(p.UnixTS)
		val := parseFloat(p.Value)
		if val == 0 && p.DailyBurn != "" {
			val = parseFloat(p.DailyBurn)
		}
		// Convert wei values to ETH if they look like wei (very large numbers)
		if val > 1e15 {
			val = val / 1e18
		}
		result = append(result, types.TimeSeriesPoint{
			Timestamp: ts,
			Value:     val,
		})
	}

	return result
}

// parseTimestamp parses a string timestamp to int64.
func parseTimestamp(s string) int64 {
	if s == "" {
		return 0
	}
	var ts int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			ts = ts*10 + int64(c-'0')
		}
	}
	return ts
}

// parseFloat parses a string to float64.
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var result float64
	var decimal float64
	var isDecimal bool
	var decimalPlace float64 = 0.1
	var negative bool

	for i, c := range s {
		if c == '-' && i == 0 {
			negative = true
			continue
		}
		if c == '.' {
			isDecimal = true
			continue
		}
		if c >= '0' && c <= '9' {
			if isDecimal {
				decimal += float64(c-'0') * decimalPlace
				decimalPlace *= 0.1
			} else {
				result = result*10 + float64(c-'0')
			}
		}
	}

	result += decimal
	if negative {
		result = -result
	}
	return result
}

// sumLastNDays sums the values of the last N data points in a time series.
func sumLastNDays(history []types.TimeSeriesPoint, n int) float64 {
	if len(history) == 0 {
		return 0
	}

	start := len(history) - n
	if start < 0 {
		start = 0
	}

	var sum float64
	for i := start; i < len(history); i++ {
		sum += history[i].Value
	}
	return sum
}

// calculateMedian computes the median of a float64 slice.
func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}
