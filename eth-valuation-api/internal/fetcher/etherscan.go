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

// EtherscanClient is an adapter for the Etherscan API V2.
// V2 uses a unified base URL (https://api.etherscan.io/v2/api) with a chainid parameter.
type EtherscanClient struct {
	baseURL    string
	apiKey     string
	chainID    string // e.g. "1" for Ethereum mainnet
	httpClient *http.Client
}

// NewEtherscanClient creates a new EtherscanClient with the given configuration.
// chainID defaults to "1" (Ethereum mainnet) if empty.
func NewEtherscanClient(baseURL, apiKey string) *EtherscanClient {
	return &EtherscanClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		chainID: "1",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// etherscanResponse is the generic response wrapper from Etherscan API.
type etherscanResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// EtherscanGasOracle holds gas price data from Etherscan.
type EtherscanGasOracle struct {
	SafeGasPrice    string `json:"SafeGasPrice"`
	ProposeGasPrice string `json:"ProposeGasPrice"`
	FastGasPrice    string `json:"FastGasPrice"`
	SuggestBaseFee  string `json:"suggestBaseFee"`
}

// EtherscanTransaction represents a transaction from Etherscan.
type EtherscanTransaction struct {
	Hash        string `json:"hash"`
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:"timeStamp"`
	GasUsed     string `json:"gasUsed"`
	GasPrice    string `json:"gasPrice"`
	Value       string `json:"value"`
}

// EtherscanBlockReward represents block reward data.
type EtherscanBlockReward struct {
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:"timeStamp"`
	BlockReward string `json:"blockReward"`
}

// EtherscanEthSupply holds ETH supply data.
type EtherscanEthSupply struct {
	EthSupply   string `json:"EthSupply"`
	Eth2Staking string `json:"Eth2Staking"`
	BurntFees   string `json:"BurntFees"`
}

// GetGasOracle fetches the current gas oracle data (safe, proposed, fast gas prices).
func (c *EtherscanClient) GetGasOracle(ctx context.Context) (*EtherscanGasOracle, error) {
	params := url.Values{
		"module": {"gastracker"},
		"action": {"gasoracle"},
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("etherscan GetGasOracle: %w", err)
	}

	var result EtherscanGasOracle
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("etherscan GetGasOracle unmarshal: %w", err)
	}

	return &result, nil
}

// GetEthSupply fetches the total ETH supply including staking and burnt fees.
func (c *EtherscanClient) GetEthSupply(ctx context.Context) (*EtherscanEthSupply, error) {
	params := url.Values{
		"module": {"stats"},
		"action": {"ethsupply2"},
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("etherscan GetEthSupply: %w", err)
	}

	var result EtherscanEthSupply
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("etherscan GetEthSupply unmarshal: %w", err)
	}

	return &result, nil
}

// GetBlockCountdown fetches the estimated time for a block to be mined.
func (c *EtherscanClient) GetBlockCountdown(ctx context.Context, blockNo int64) (json.RawMessage, error) {
	params := url.Values{
		"module":  {"block"},
		"action":  {"getblockcountdown"},
		"blockno": {fmt.Sprintf("%d", blockNo)},
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("etherscan GetBlockCountdown: %w", err)
	}

	return body, nil
}

// GetDailyAvgGasPrice fetches the daily average gas price for a date range.
func (c *EtherscanClient) GetDailyAvgGasPrice(ctx context.Context, startDate, endDate string) (json.RawMessage, error) {
	params := url.Values{
		"module":    {"stats"},
		"action":    {"dailyavggasprice"},
		"startdate": {startDate},
		"enddate":   {endDate},
		"sort":      {"asc"},
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("etherscan GetDailyAvgGasPrice: %w", err)
	}

	return body, nil
}

// GetDailyBurntFees fetches the daily total burnt fees for a date range.
func (c *EtherscanClient) GetDailyBurntFees(ctx context.Context, startDate, endDate string) (json.RawMessage, error) {
	params := url.Values{
		"module":    {"stats"},
		"action":    {"dailyburnedeth"},
		"startdate": {startDate},
		"enddate":   {endDate},
		"sort":      {"asc"},
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("etherscan GetDailyBurntFees: %w", err)
	}

	return body, nil
}

// GetDailyTxCount fetches the daily transaction count for a date range.
func (c *EtherscanClient) GetDailyTxCount(ctx context.Context, startDate, endDate string) (json.RawMessage, error) {
	params := url.Values{
		"module":    {"stats"},
		"action":    {"dailytx"},
		"startdate": {startDate},
		"enddate":   {endDate},
		"sort":      {"asc"},
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("etherscan GetDailyTxCount: %w", err)
	}

	return body, nil
}

// doRequest constructs and executes an HTTP request to the Etherscan API V2.
func (c *EtherscanClient) doRequest(ctx context.Context, params url.Values) (json.RawMessage, error) {
	if c.apiKey != "" {
		params.Set("apikey", c.apiKey)
	}
	params.Set("chainid", c.chainID)

	reqURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var apiResp etherscanResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if apiResp.Status != "1" {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	return apiResp.Result, nil
}
