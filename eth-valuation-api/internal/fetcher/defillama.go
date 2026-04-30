package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefiLlamaClient is an adapter for the DefiLlama API.
// It provides methods to fetch TVL and protocol data.
type DefiLlamaClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewDefiLlamaClient creates a new DefiLlamaClient with the given configuration.
func NewDefiLlamaClient(baseURL string) *DefiLlamaClient {
	return &DefiLlamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DefiLlamaProtocol holds protocol-level TVL data.
type DefiLlamaProtocol struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Chain    string  `json:"chain"`
	TVL      float64 `json:"tvl"`
	Change1h float64 `json:"change_1h"`
	Change1d float64 `json:"change_1d"`
	Change7d float64 `json:"change_7d"`
	Category string  `json:"category"`
}

// DefiLlamaChainTVL holds chain-level TVL data.
type DefiLlamaChainTVL struct {
	Name string  `json:"name"`
	TVL  float64 `json:"tvl"`
}

// DefiLlamaTVLHistory holds a historical TVL data point.
type DefiLlamaTVLHistory struct {
	Date int64   `json:"date"` // Unix timestamp
	TVL  float64 `json:"tvl"`
}

// DefiLlamaProtocolDetail holds detailed protocol data including historical TVL.
type DefiLlamaProtocolDetail struct {
	ID         string                           `json:"id"`
	Name       string                           `json:"name"`
	Symbol     string                           `json:"symbol"`
	TVL        []DefiLlamaTVLHistory            `json:"tvl"`
	CurrentTVL float64                          `json:"currentChainTvls"`
	Chains     []string                         `json:"chains"`
	ChainTvls  map[string][]DefiLlamaTVLHistory `json:"chainTvls"`
}

// DefiLlamaStablecoin holds stablecoin market data.
type DefiLlamaStablecoin struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Symbol     string  `json:"symbol"`
	CircSupply float64 `json:"circulating"`
	Price      float64 `json:"price"`
}

// DefiLlamaStablecoinsResponse holds the response from the stablecoins endpoint.
type DefiLlamaStablecoinsResponse struct {
	PeggedAssets []DefiLlamaStablecoin `json:"peggedAssets"`
}

// GetProtocols fetches all protocols with their TVL data.
func (c *DefiLlamaClient) GetProtocols(ctx context.Context) ([]DefiLlamaProtocol, error) {
	body, err := c.doRequest(ctx, "/protocols")
	if err != nil {
		return nil, fmt.Errorf("defillama GetProtocols: %w", err)
	}

	var result []DefiLlamaProtocol
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("defillama GetProtocols unmarshal: %w", err)
	}

	return result, nil
}

// GetChainsTVL fetches TVL data for all chains.
func (c *DefiLlamaClient) GetChainsTVL(ctx context.Context) ([]DefiLlamaChainTVL, error) {
	body, err := c.doRequest(ctx, "/v2/chains")
	if err != nil {
		return nil, fmt.Errorf("defillama GetChainsTVL: %w", err)
	}

	var result []DefiLlamaChainTVL
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("defillama GetChainsTVL unmarshal: %w", err)
	}

	return result, nil
}

// GetChainTVLHistory fetches historical TVL data for a specific chain.
func (c *DefiLlamaClient) GetChainTVLHistory(ctx context.Context, chain string) ([]DefiLlamaTVLHistory, error) {
	path := fmt.Sprintf("/v2/historicalChainTvl/%s", chain)
	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("defillama GetChainTVLHistory: %w", err)
	}

	var result []DefiLlamaTVLHistory
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("defillama GetChainTVLHistory unmarshal: %w", err)
	}

	return result, nil
}

// GetTotalTVLHistory fetches historical total TVL across all chains.
func (c *DefiLlamaClient) GetTotalTVLHistory(ctx context.Context) ([]DefiLlamaTVLHistory, error) {
	body, err := c.doRequest(ctx, "/v2/historicalChainTvl")
	if err != nil {
		return nil, fmt.Errorf("defillama GetTotalTVLHistory: %w", err)
	}

	var result []DefiLlamaTVLHistory
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("defillama GetTotalTVLHistory unmarshal: %w", err)
	}

	return result, nil
}

// GetProtocolDetail fetches detailed data for a specific protocol.
func (c *DefiLlamaClient) GetProtocolDetail(ctx context.Context, protocolSlug string) (*DefiLlamaProtocolDetail, error) {
	path := fmt.Sprintf("/protocol/%s", protocolSlug)
	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("defillama GetProtocolDetail: %w", err)
	}

	var result DefiLlamaProtocolDetail
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("defillama GetProtocolDetail unmarshal: %w", err)
	}

	return &result, nil
}

// GetStablecoins fetches stablecoin market data.
func (c *DefiLlamaClient) GetStablecoins(ctx context.Context) (*DefiLlamaStablecoinsResponse, error) {
	body, err := c.doRequest(ctx, "/stablecoins")
	if err != nil {
		return nil, fmt.Errorf("defillama GetStablecoins: %w", err)
	}

	var result DefiLlamaStablecoinsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("defillama GetStablecoins unmarshal: %w", err)
	}

	return &result, nil
}

// doRequest constructs and executes an HTTP request to the DefiLlama API.
func (c *DefiLlamaClient) doRequest(ctx context.Context, path string) ([]byte, error) {
	reqURL := c.baseURL + path

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
