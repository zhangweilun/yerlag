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

// GlassnodeClient is an adapter for the Glassnode API.
// It provides methods to fetch on-chain metrics including MVRV, NVT, and other indicators.
type GlassnodeClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewGlassnodeClient creates a new GlassnodeClient with the given configuration.
func NewGlassnodeClient(baseURL, apiKey string) *GlassnodeClient {
	return &GlassnodeClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GlassnodeDataPoint represents a single time-series data point from Glassnode.
type GlassnodeDataPoint struct {
	Timestamp int64   `json:"t"` // Unix timestamp
	Value     float64 `json:"v"` // Metric value
}

// GlassnodeMetricParams holds parameters for a Glassnode metric query.
type GlassnodeMetricParams struct {
	Asset    string // e.g., "ETH"
	Since    int64  // Start timestamp (Unix)
	Until    int64  // End timestamp (Unix)
	Interval string // "1h", "24h", "10m", "1w", "1month"
	Currency string // "native" or "usd"
}

// GetMVRV fetches the MVRV ratio for ETH.
func (c *GlassnodeClient) GetMVRV(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/market/mvrv", params)
}

// GetNVT fetches the NVT ratio for ETH.
func (c *GlassnodeClient) GetNVT(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/indicators/nvt", params)
}

// GetActiveAddresses fetches the daily active addresses count.
func (c *GlassnodeClient) GetActiveAddresses(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/addresses/active_count", params)
}

// GetNewAddresses fetches the daily new addresses count.
func (c *GlassnodeClient) GetNewAddresses(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/addresses/new_non_zero_count", params)
}

// GetTransactionCount fetches the daily transaction count.
func (c *GlassnodeClient) GetTransactionCount(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/transactions/count", params)
}

// GetExchangeBalance fetches the total ETH balance on exchanges.
func (c *GlassnodeClient) GetExchangeBalance(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/distribution/balance_exchanges", params)
}

// GetExchangeNetFlow fetches the net flow of ETH to/from exchanges.
func (c *GlassnodeClient) GetExchangeNetFlow(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/distribution/exchange_net_position_change", params)
}

// GetRealizedPrice fetches the realized price of ETH.
func (c *GlassnodeClient) GetRealizedPrice(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/market/price_realized_usd", params)
}

// GetMarketCap fetches the market capitalization.
func (c *GlassnodeClient) GetMarketCap(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/market/marketcap_usd", params)
}

// GetTransferVolume fetches the total transfer volume.
func (c *GlassnodeClient) GetTransferVolume(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/transactions/transfers_volume_sum", params)
}

// GetSupply fetches the circulating supply.
func (c *GlassnodeClient) GetSupply(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/supply/current", params)
}

// GetIssuance fetches the daily issuance (new coins minted).
func (c *GlassnodeClient) GetIssuance(ctx context.Context, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	return c.getMetric(ctx, "/metrics/supply/issued", params)
}

// getMetric is the generic method for fetching any Glassnode metric.
func (c *GlassnodeClient) getMetric(ctx context.Context, metricPath string, params GlassnodeMetricParams) ([]GlassnodeDataPoint, error) {
	queryParams := url.Values{}

	asset := params.Asset
	if asset == "" {
		asset = "ETH"
	}
	queryParams.Set("a", asset)

	if params.Since > 0 {
		queryParams.Set("s", fmt.Sprintf("%d", params.Since))
	}
	if params.Until > 0 {
		queryParams.Set("u", fmt.Sprintf("%d", params.Until))
	}
	if params.Interval != "" {
		queryParams.Set("i", params.Interval)
	}
	if params.Currency != "" {
		queryParams.Set("c", params.Currency)
	}

	body, err := c.doRequest(ctx, metricPath, queryParams)
	if err != nil {
		return nil, fmt.Errorf("glassnode %s: %w", metricPath, err)
	}

	var result []GlassnodeDataPoint
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("glassnode %s unmarshal: %w", metricPath, err)
	}

	return result, nil
}

// doRequest constructs and executes an HTTP request to the Glassnode API.
func (c *GlassnodeClient) doRequest(ctx context.Context, path string, params url.Values) ([]byte, error) {
	if c.apiKey != "" {
		params.Set("api_key", c.apiKey)
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
