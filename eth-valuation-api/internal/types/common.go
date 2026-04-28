package types

// TimeSeriesPoint represents a single data point in a time series.
type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// APIResponse is the generic API response wrapper.
type APIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    Meta   `json:"meta"`
}

// Meta contains metadata about the API response data freshness.
type Meta struct {
	LastUpdated int64  `json:"lastUpdated"`
	Source      string `json:"source"` // "live" | "cache"
	NextRefresh int64  `json:"nextRefresh"`
}

// SuccessResponse creates a successful API response.
func SuccessResponse[T any](data T, meta Meta) APIResponse[T] {
	return APIResponse[T]{
		Code:    0,
		Message: "ok",
		Data:    data,
		Meta:    meta,
	}
}

// ErrorResponse creates an error API response with nil data.
func ErrorResponse(code int, message string) APIResponse[any] {
	return APIResponse[any]{
		Code:    code,
		Message: message,
		Data:    nil,
		Meta:    Meta{},
	}
}
