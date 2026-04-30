package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CoinGeckoClient is an adapter for the CoinGecko API.
// It provides methods to fetch price, market cap, volume, and OHLCV data.
type CoinGeckoClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewCoinGeckoClient creates a new CoinGeckoClient with the given configuration.
func NewCoinGeckoClient(baseURL, apiKey string) *CoinGeckoClient {
	return &CoinGeckoClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CoinGeckoPrice holds current price and market data for a coin.
type CoinGeckoPrice struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	MarketCapRank            int     `json:"market_cap_rank"`
	FullyDilutedValuation    float64 `json:"fully_diluted_valuation"`
	TotalVolume              float64 `json:"total_volume"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	CirculatingSupply        float64 `json:"circulating_supply"`
	TotalSupply              float64 `json:"total_supply"`
	ATH                      float64 `json:"ath"`
	ATHChangePercentage      float64 `json:"ath_change_percentage"`
}

// CoinGeckoMarketChart holds historical market chart data.
type CoinGeckoMarketChart struct {
	Prices       [][]float64 `json:"prices"`        // [[timestamp, price], ...]
	MarketCaps   [][]float64 `json:"market_caps"`   // [[timestamp, market_cap], ...]
	TotalVolumes [][]float64 `json:"total_volumes"` // [[timestamp, volume], ...]
}

// CoinGeckoOHLCV represents a single OHLCV data point.
type CoinGeckoOHLCV struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
}

// CoinGeckoCoinData holds detailed coin data from the /coins/{id} endpoint.
type CoinGeckoCoinData struct {
	ID         string `json:"id"`
	Symbol     string `json:"symbol"`
	MarketData struct {
		CurrentPrice          map[string]float64 `json:"current_price"`
		MarketCap             map[string]float64 `json:"market_cap"`
		TotalVolume           map[string]float64 `json:"total_volume"`
		PriceChangePercent24h float64            `json:"price_change_percentage_24h"`
		ATH                   map[string]float64 `json:"ath"`
		ATHChangePercentage   map[string]float64 `json:"ath_change_percentage"`
		CirculatingSupply     float64            `json:"circulating_supply"`
		TotalSupply           float64            `json:"total_supply"`
	} `json:"market_data"`
}

// CoinGeckoGlobalData holds global cryptocurrency market data.
type CoinGeckoGlobalData struct {
	Data struct {
		TotalMarketCap         map[string]float64 `json:"total_market_cap"`
		MarketCapPercentage    map[string]float64 `json:"market_cap_percentage"`
		TotalVolume            map[string]float64 `json:"total_volume"`
		MarketCapChangePercent float64            `json:"market_cap_change_percentage_24h_usd"`
	} `json:"data"`
}

// GetMarketData fetches current market data for the specified coins.
func (c *CoinGeckoClient) GetMarketData(ctx context.Context, coinIDs []string, vsCurrency string) ([]CoinGeckoPrice, error) {
	params := url.Values{
		"vs_currency": {vsCurrency},
		"ids":         {strings.Join(coinIDs, ",")},
		"order":       {"market_cap_desc"},
		"sparkline":   {"false"},
	}

	body, err := c.doRequest(ctx, "/coins/markets", params)
	if err != nil {
		return nil, fmt.Errorf("coingecko GetMarketData: %w", err)
	}

	var result []CoinGeckoPrice
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("coingecko GetMarketData unmarshal: %w", err)
	}

	return result, nil
}

// GetMarketChart fetches historical market chart data for a coin.
func (c *CoinGeckoClient) GetMarketChart(ctx context.Context, coinID, vsCurrency string, days int) (*CoinGeckoMarketChart, error) {
	params := url.Values{
		"vs_currency": {vsCurrency},
		"days":        {fmt.Sprintf("%d", days)},
	}

	path := fmt.Sprintf("/coins/%s/market_chart", coinID)
	body, err := c.doRequest(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("coingecko GetMarketChart: %w", err)
	}

	var result CoinGeckoMarketChart
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("coingecko GetMarketChart unmarshal: %w", err)
	}

	return &result, nil
}

// GetOHLCV fetches OHLCV (candlestick) data for a coin.
// Days can be 1, 7, 14, 30, 90, 180, 365, or "max".
func (c *CoinGeckoClient) GetOHLCV(ctx context.Context, coinID, vsCurrency string, days int) ([]CoinGeckoOHLCV, error) {
	params := url.Values{
		"vs_currency": {vsCurrency},
		"days":        {fmt.Sprintf("%d", days)},
	}

	path := fmt.Sprintf("/coins/%s/ohlc", coinID)
	body, err := c.doRequest(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("coingecko GetOHLCV: %w", err)
	}

	// CoinGecko returns OHLCV as [[timestamp, open, high, low, close], ...]
	var raw [][]float64
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("coingecko GetOHLCV unmarshal: %w", err)
	}

	result := make([]CoinGeckoOHLCV, 0, len(raw))
	for _, point := range raw {
		if len(point) < 5 {
			continue
		}
		result = append(result, CoinGeckoOHLCV{
			Timestamp: int64(point[0]),
			Open:      point[1],
			High:      point[2],
			Low:       point[3],
			Close:     point[4],
		})
	}

	return result, nil
}

// GetCoinData fetches detailed data for a specific coin.
func (c *CoinGeckoClient) GetCoinData(ctx context.Context, coinID string) (*CoinGeckoCoinData, error) {
	params := url.Values{
		"localization":   {"false"},
		"tickers":        {"false"},
		"community_data": {"false"},
		"developer_data": {"false"},
	}

	path := fmt.Sprintf("/coins/%s", coinID)
	body, err := c.doRequest(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("coingecko GetCoinData: %w", err)
	}

	var result CoinGeckoCoinData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("coingecko GetCoinData unmarshal: %w", err)
	}

	return &result, nil
}

// GetGlobalData fetches global cryptocurrency market data.
func (c *CoinGeckoClient) GetGlobalData(ctx context.Context) (*CoinGeckoGlobalData, error) {
	body, err := c.doRequest(ctx, "/global", nil)
	if err != nil {
		return nil, fmt.Errorf("coingecko GetGlobalData: %w", err)
	}

	var result CoinGeckoGlobalData
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("coingecko GetGlobalData unmarshal: %w", err)
	}

	return &result, nil
}

// GetSimplePrice fetches simple price data for coins (lightweight endpoint).
func (c *CoinGeckoClient) GetSimplePrice(ctx context.Context, coinIDs []string, vsCurrencies []string) (map[string]map[string]float64, error) {
	params := url.Values{
		"ids":           {strings.Join(coinIDs, ",")},
		"vs_currencies": {strings.Join(vsCurrencies, ",")},
	}

	body, err := c.doRequest(ctx, "/simple/price", params)
	if err != nil {
		return nil, fmt.Errorf("coingecko GetSimplePrice: %w", err)
	}

	var result map[string]map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("coingecko GetSimplePrice unmarshal: %w", err)
	}

	return result, nil
}

// doRequest constructs and executes an HTTP request to the CoinGecko API.
func (c *CoinGeckoClient) doRequest(ctx context.Context, path string, params url.Values) ([]byte, error) {
	reqURL := c.baseURL + path
	if params != nil {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

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
