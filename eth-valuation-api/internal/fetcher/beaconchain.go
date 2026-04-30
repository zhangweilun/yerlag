package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BeaconChainClient is an adapter for the Beacon Chain API (beaconcha.in).
// It provides methods to fetch staking and validator data.
type BeaconChainClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBeaconChainClient creates a new BeaconChainClient with the given configuration.
func NewBeaconChainClient(baseURL string) *BeaconChainClient {
	return &BeaconChainClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// beaconChainResponse is the generic response wrapper from the Beacon Chain API.
type beaconChainResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// BeaconChainEpoch holds epoch-level data.
type BeaconChainEpoch struct {
	Epoch                   int64   `json:"epoch"`
	ValidatorsCount         int64   `json:"validatorscount"`
	AverageValidatorBalance float64 `json:"averagevalidatorbalance"`
	TotalValidatorBalance   float64 `json:"totalvalidatorbalance"`
	EligibleEther           float64 `json:"eligibleether"`
	GlobalParticipationRate float64 `json:"globalparticipationrate"`
	MissedBlocks            int64   `json:"missedblocks"`
	ProposedBlocks          int64   `json:"proposedblocks"`
	ScheduledBlocks         int64   `json:"scheduledblocks"`
	Finalized               bool    `json:"finalized"`
}

// BeaconChainValidatorQueue holds validator entry/exit queue data.
type BeaconChainValidatorQueue struct {
	Entering int64 `json:"beaconchain_entering"`
	Exiting  int64 `json:"beaconchain_exiting"`
}

// BeaconChainNetworkStats holds overall network statistics.
type BeaconChainNetworkStats struct {
	CurrentEpoch      int64   `json:"currentEpoch"`
	CurrentSlot       int64   `json:"currentSlot"`
	ActiveValidators  int64   `json:"activeValidators"`
	TotalValidators   int64   `json:"totalValidators"`
	TotalStakedEther  float64 `json:"totalStakedEther"`
	AverageBalance    float64 `json:"averageBalance"`
	ParticipationRate float64 `json:"participationRate"`
}

// BeaconChainValidatorPerformance holds validator performance data.
type BeaconChainValidatorPerformance struct {
	Balance          int64 `json:"balance"`
	EffectiveBalance int64 `json:"effectivebalance"`
	Performance1d    int64 `json:"performance1d"`
	Performance7d    int64 `json:"performance7d"`
	Performance31d   int64 `json:"performance31d"`
	Performance365d  int64 `json:"performance365d"`
}

// GetLatestEpoch fetches the latest epoch data.
func (c *BeaconChainClient) GetLatestEpoch(ctx context.Context) (*BeaconChainEpoch, error) {
	body, err := c.doRequest(ctx, "/epoch/latest")
	if err != nil {
		return nil, fmt.Errorf("beaconchain GetLatestEpoch: %w", err)
	}

	var result BeaconChainEpoch
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("beaconchain GetLatestEpoch unmarshal: %w", err)
	}

	return &result, nil
}

// GetEpoch fetches data for a specific epoch.
func (c *BeaconChainClient) GetEpoch(ctx context.Context, epoch int64) (*BeaconChainEpoch, error) {
	path := fmt.Sprintf("/epoch/%d", epoch)
	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("beaconchain GetEpoch: %w", err)
	}

	var result BeaconChainEpoch
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("beaconchain GetEpoch unmarshal: %w", err)
	}

	return &result, nil
}

// GetValidatorQueue fetches the current validator entry/exit queue.
func (c *BeaconChainClient) GetValidatorQueue(ctx context.Context) (*BeaconChainValidatorQueue, error) {
	body, err := c.doRequest(ctx, "/validators/queue")
	if err != nil {
		return nil, fmt.Errorf("beaconchain GetValidatorQueue: %w", err)
	}

	var result BeaconChainValidatorQueue
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("beaconchain GetValidatorQueue unmarshal: %w", err)
	}

	return &result, nil
}

// GetEpochHistory fetches epoch data for a range of recent epochs.
func (c *BeaconChainClient) GetEpochHistory(ctx context.Context, limit int) ([]BeaconChainEpoch, error) {
	path := fmt.Sprintf("/epochs?limit=%d", limit)
	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("beaconchain GetEpochHistory: %w", err)
	}

	var result []BeaconChainEpoch
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("beaconchain GetEpochHistory unmarshal: %w", err)
	}

	return result, nil
}

// GetNetworkStats fetches overall network statistics.
// This is a convenience method that aggregates data from the latest epoch.
func (c *BeaconChainClient) GetNetworkStats(ctx context.Context) (*BeaconChainNetworkStats, error) {
	epoch, err := c.GetLatestEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("beaconchain GetNetworkStats: %w", err)
	}

	stats := &BeaconChainNetworkStats{
		CurrentEpoch:      epoch.Epoch,
		ActiveValidators:  epoch.ValidatorsCount,
		TotalStakedEther:  epoch.EligibleEther,
		AverageBalance:    epoch.AverageValidatorBalance,
		ParticipationRate: epoch.GlobalParticipationRate,
	}

	return stats, nil
}

// GetMissedSlots calculates the missed slots rate from recent epochs.
func (c *BeaconChainClient) GetMissedSlots(ctx context.Context, epochCount int) (missed int64, total int64, err error) {
	epochs, err := c.GetEpochHistory(ctx, epochCount)
	if err != nil {
		return 0, 0, fmt.Errorf("beaconchain GetMissedSlots: %w", err)
	}

	for _, e := range epochs {
		missed += e.MissedBlocks
		total += e.ScheduledBlocks
	}

	return missed, total, nil
}

// doRequest constructs and executes an HTTP request to the Beacon Chain API.
func (c *BeaconChainClient) doRequest(ctx context.Context, path string) (json.RawMessage, error) {
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

	// Try to parse as the standard beaconcha.in response format
	var apiResp beaconChainResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err == nil && apiResp.Data != nil {
		return apiResp.Data, nil
	}

	// If not in standard format, return raw body
	return bodyBytes, nil
}
