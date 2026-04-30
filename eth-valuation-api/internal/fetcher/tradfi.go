package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// TradFiClient is an adapter for traditional finance data APIs.
// It provides methods to fetch DXY, treasury yields, and Nasdaq data.
type TradFiClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewTradFiClient creates a new TradFiClient with the given configuration.
func NewTradFiClient(baseURL, apiKey string) *TradFiClient {
	return &TradFiClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TradFiDataPoint represents a single time-series data point for traditional finance data.
type TradFiDataPoint struct {
	Date  string  `json:"date"` // YYYY-MM-DD format
	Value float64 `json:"value"`
}

// TradFiDXYData holds US Dollar Index (DXY) data.
type TradFiDXYData struct {
	Current float64           `json:"current"`
	History []TradFiDataPoint `json:"history"`
}

// TradFiTreasuryData holds US Treasury yield data.
type TradFiTreasuryData struct {
	Yield10Y float64           `json:"yield10y"`
	Yield2Y  float64           `json:"yield2y"`
	Yield30Y float64           `json:"yield30y"`
	History  []TradFiDataPoint `json:"history"`
}

// TradFiNasdaqData holds Nasdaq index data.
type TradFiNasdaqData struct {
	Current float64           `json:"current"`
	History []TradFiDataPoint `json:"history"`
}

// TradFiFedRateData holds Federal Reserve interest rate data.
type TradFiFedRateData struct {
	CurrentRate      float64              `json:"currentRate"`
	RateExpectations []TradFiRateExpected `json:"rateExpectations"`
}

// TradFiRateExpected holds a future rate expectation.
type TradFiRateExpected struct {
	Date         string  `json:"date"`
	ExpectedRate float64 `json:"expectedRate"`
	Probability  float64 `json:"probability"`
}

// TradFiFearGreedData holds the crypto Fear & Greed Index data.
type TradFiFearGreedData struct {
	Value   int               `json:"value"` // 0-100
	Label   string            `json:"label"` // "Extreme Fear", "Fear", "Neutral", "Greed", "Extreme Greed"
	History []TradFiDataPoint `json:"history"`
}

// GetDXY fetches the US Dollar Index (DXY) current value and history.
func (c *TradFiClient) GetDXY(ctx context.Context, days int) (*TradFiDXYData, error) {
	params := url.Values{
		"indicator": {"dxy"},
		"days":      {fmt.Sprintf("%d", days)},
	}

	body, err := c.doRequest(ctx, "/indicators/dxy", params)
	if err != nil {
		return nil, fmt.Errorf("tradfi GetDXY: %w", err)
	}

	var result TradFiDXYData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tradfi GetDXY unmarshal: %w", err)
	}

	return &result, nil
}

// GetTreasuryYields fetches US Treasury yield data.
func (c *TradFiClient) GetTreasuryYields(ctx context.Context, days int) (*TradFiTreasuryData, error) {
	params := url.Values{
		"indicator": {"treasury"},
		"days":      {fmt.Sprintf("%d", days)},
	}

	body, err := c.doRequest(ctx, "/indicators/treasury", params)
	if err != nil {
		return nil, fmt.Errorf("tradfi GetTreasuryYields: %w", err)
	}

	var result TradFiTreasuryData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tradfi GetTreasuryYields unmarshal: %w", err)
	}

	return &result, nil
}

// GetNasdaq fetches Nasdaq index data.
func (c *TradFiClient) GetNasdaq(ctx context.Context, days int) (*TradFiNasdaqData, error) {
	params := url.Values{
		"indicator": {"nasdaq"},
		"days":      {fmt.Sprintf("%d", days)},
	}

	body, err := c.doRequest(ctx, "/indicators/nasdaq", params)
	if err != nil {
		return nil, fmt.Errorf("tradfi GetNasdaq: %w", err)
	}

	var result TradFiNasdaqData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tradfi GetNasdaq unmarshal: %w", err)
	}

	return &result, nil
}

// GetFedRate fetches the Federal Reserve interest rate and expectations.
func (c *TradFiClient) GetFedRate(ctx context.Context) (*TradFiFedRateData, error) {
	params := url.Values{
		"indicator": {"fed_rate"},
	}

	body, err := c.doRequest(ctx, "/indicators/fed-rate", params)
	if err != nil {
		return nil, fmt.Errorf("tradfi GetFedRate: %w", err)
	}

	var result TradFiFedRateData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tradfi GetFedRate unmarshal: %w", err)
	}

	return &result, nil
}

// GetFearGreedIndex fetches the crypto Fear & Greed Index.
func (c *TradFiClient) GetFearGreedIndex(ctx context.Context, days int) (*TradFiFearGreedData, error) {
	params := url.Values{
		"indicator": {"fear_greed"},
		"days":      {fmt.Sprintf("%d", days)},
	}

	body, err := c.doRequest(ctx, "/indicators/fear-greed", params)
	if err != nil {
		return nil, fmt.Errorf("tradfi GetFearGreedIndex: %w", err)
	}

	var result TradFiFearGreedData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tradfi GetFearGreedIndex unmarshal: %w", err)
	}

	return &result, nil
}

// doRequest constructs and executes an HTTP request to the TradFi API.
func (c *TradFiClient) doRequest(ctx context.Context, path string, params url.Values) ([]byte, error) {
	if c.apiKey != "" {
		params.Set("apikey", c.apiKey)
	}

	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: invalid API key")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429): retry after cooldown")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return bodyBytes, nil
}
