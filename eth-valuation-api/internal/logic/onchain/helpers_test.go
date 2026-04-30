package onchain

import (
	"encoding/json"
	"math"
	"testing"

	"eth-valuation-api/internal/types"

	"github.com/stretchr/testify/assert"
)

// --- parseWeiToEth tests ---

func TestParseWeiToEth_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		weiStr   string
		expected float64
	}{
		{
			name:     "1 ETH in wei",
			weiStr:   "1000000000000000000",
			expected: 1.0,
		},
		{
			name:     "0.5 ETH in wei",
			weiStr:   "500000000000000000",
			expected: 0.5,
		},
		{
			name:     "10 ETH in wei",
			weiStr:   "10000000000000000000",
			expected: 10.0,
		},
		{
			name:     "very small amount",
			weiStr:   "1000000000",
			expected: 0.000000001,
		},
		{
			name:     "large amount (1M ETH)",
			weiStr:   "1000000000000000000000000",
			expected: 1000000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWeiToEth(tt.weiStr)
			assert.InDelta(t, tt.expected, result, 1e-9)
		})
	}
}

func TestParseWeiToEth_EmptyString(t *testing.T) {
	result := parseWeiToEth("")
	assert.Equal(t, 0.0, result)
}

func TestParseWeiToEth_InvalidString(t *testing.T) {
	result := parseWeiToEth("not_a_number")
	assert.Equal(t, 0.0, result)
}

func TestParseWeiToEth_Zero(t *testing.T) {
	result := parseWeiToEth("0")
	assert.Equal(t, 0.0, result)
}

// --- parseTimestamp tests ---

func TestParseTimestamp_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"unix epoch", "0", 0},
		{"typical timestamp", "1700000000", 1700000000},
		{"another timestamp", "1625097600", 1625097600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimestamp(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTimestamp_EmptyString(t *testing.T) {
	result := parseTimestamp("")
	assert.Equal(t, int64(0), result)
}

func TestParseTimestamp_NonNumericIgnored(t *testing.T) {
	// parseTimestamp only extracts digits
	result := parseTimestamp("abc123def456")
	assert.Equal(t, int64(123456), result)
}

// --- parseFloat tests ---

func TestParseFloat_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"integer", "42", 42.0},
		{"decimal", "3.14", 3.14},
		{"negative", "-5.5", -5.5},
		{"zero", "0", 0.0},
		{"large number", "1234567890", 1234567890.0},
		{"small decimal", "0.001", 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloat(tt.input)
			assert.InDelta(t, tt.expected, result, 1e-9)
		})
	}
}

func TestParseFloat_EmptyString(t *testing.T) {
	result := parseFloat("")
	assert.Equal(t, 0.0, result)
}

func TestParseFloat_NegativeZero(t *testing.T) {
	result := parseFloat("-0")
	assert.Equal(t, 0.0, math.Abs(result))
}

// --- sumLastNDays tests ---

func TestSumLastNDays_NormalCase(t *testing.T) {
	history := []types.TimeSeriesPoint{
		{Timestamp: 1, Value: 10},
		{Timestamp: 2, Value: 20},
		{Timestamp: 3, Value: 30},
		{Timestamp: 4, Value: 40},
		{Timestamp: 5, Value: 50},
	}

	// Sum last 3 days
	result := sumLastNDays(history, 3)
	assert.InDelta(t, 120.0, result, 1e-9) // 30 + 40 + 50

	// Sum last 1 day
	result = sumLastNDays(history, 1)
	assert.InDelta(t, 50.0, result, 1e-9)

	// Sum all days
	result = sumLastNDays(history, 5)
	assert.InDelta(t, 150.0, result, 1e-9)
}

func TestSumLastNDays_NGreaterThanLength(t *testing.T) {
	history := []types.TimeSeriesPoint{
		{Timestamp: 1, Value: 10},
		{Timestamp: 2, Value: 20},
	}

	// N > len(history) should sum all available
	result := sumLastNDays(history, 100)
	assert.InDelta(t, 30.0, result, 1e-9)
}

func TestSumLastNDays_EmptyHistory(t *testing.T) {
	result := sumLastNDays(nil, 7)
	assert.Equal(t, 0.0, result)

	result = sumLastNDays([]types.TimeSeriesPoint{}, 7)
	assert.Equal(t, 0.0, result)
}

func TestSumLastNDays_ZeroDays(t *testing.T) {
	history := []types.TimeSeriesPoint{
		{Timestamp: 1, Value: 10},
		{Timestamp: 2, Value: 20},
	}

	result := sumLastNDays(history, 0)
	assert.Equal(t, 0.0, result)
}

// --- calculateMedian tests ---

func TestCalculateMedian_OddCount(t *testing.T) {
	values := []float64{3, 1, 2}
	result := calculateMedian(values)
	assert.InDelta(t, 2.0, result, 1e-9)
}

func TestCalculateMedian_EvenCount(t *testing.T) {
	values := []float64{4, 1, 3, 2}
	result := calculateMedian(values)
	assert.InDelta(t, 2.5, result, 1e-9)
}

func TestCalculateMedian_SingleElement(t *testing.T) {
	values := []float64{42}
	result := calculateMedian(values)
	assert.InDelta(t, 42.0, result, 1e-9)
}

func TestCalculateMedian_EmptySlice(t *testing.T) {
	result := calculateMedian(nil)
	assert.Equal(t, 0.0, result)

	result = calculateMedian([]float64{})
	assert.Equal(t, 0.0, result)
}

func TestCalculateMedian_DoesNotMutateInput(t *testing.T) {
	values := []float64{5, 3, 1, 4, 2}
	original := make([]float64, len(values))
	copy(original, values)

	calculateMedian(values)
	assert.Equal(t, original, values)
}

// --- parseDailyTimeSeries tests ---

func TestParseDailyTimeSeries_ValidData(t *testing.T) {
	data := []dailyDataPoint{
		{UTCDate: "2024-01-01", UnixTS: "1704067200", Value: "100.5"},
		{UTCDate: "2024-01-02", UnixTS: "1704153600", Value: "200.75"},
	}
	raw, _ := json.Marshal(data)

	result := parseDailyTimeSeries(json.RawMessage(raw))
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1704067200), result[0].Timestamp)
	assert.InDelta(t, 100.5, result[0].Value, 1e-9)
	assert.Equal(t, int64(1704153600), result[1].Timestamp)
	assert.InDelta(t, 200.75, result[1].Value, 1e-9)
}

func TestParseDailyTimeSeries_WeiConversion(t *testing.T) {
	// Values > 1e15 should be converted from wei to ETH
	data := []dailyDataPoint{
		{UTCDate: "2024-01-01", UnixTS: "1704067200", Value: "2000000000000000000"}, // 2 ETH in wei
	}
	raw, _ := json.Marshal(data)

	result := parseDailyTimeSeries(json.RawMessage(raw))
	assert.Len(t, result, 1)
	assert.InDelta(t, 2.0, result[0].Value, 1e-9)
}

func TestParseDailyTimeSeries_DailyBurnFallback(t *testing.T) {
	// When Value is "0" but DailyBurn is set, use DailyBurn
	data := []dailyDataPoint{
		{UTCDate: "2024-01-01", UnixTS: "1704067200", Value: "0", DailyBurn: "500.25"},
	}
	raw, _ := json.Marshal(data)

	result := parseDailyTimeSeries(json.RawMessage(raw))
	assert.Len(t, result, 1)
	assert.InDelta(t, 500.25, result[0].Value, 1e-9)
}

func TestParseDailyTimeSeries_NilInput(t *testing.T) {
	result := parseDailyTimeSeries(nil)
	assert.Nil(t, result)
}

func TestParseDailyTimeSeries_InvalidJSON(t *testing.T) {
	result := parseDailyTimeSeries(json.RawMessage([]byte("invalid json")))
	assert.Nil(t, result)
}

func TestParseDailyTimeSeries_EmptyArray(t *testing.T) {
	raw, _ := json.Marshal([]dailyDataPoint{})
	result := parseDailyTimeSeries(json.RawMessage(raw))
	assert.Empty(t, result)
}
