package network

import (
	"context"
	"fmt"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// StakingData holds the staking data response.
type StakingData struct {
	TotalStakedETH      float64                 `json:"totalStakedEth"`
	StakingPercentage   float64                 `json:"stakingPercentage"`
	ActiveValidators    int64                   `json:"activeValidators"`
	StakingYield        float64                 `json:"stakingYield"`
	EntryQueueLength    int64                   `json:"entryQueueLength"`
	ExitQueueLength     int64                   `json:"exitQueueLength"`
	EntryWaitTime       string                  `json:"entryWaitTime"`
	ExitWaitTime        string                  `json:"exitWaitTime"`
	LiquidStakingShares []LiquidStakingShare    `json:"liquidStakingShares"`
	ValidatorHistory    []types.TimeSeriesPoint `json:"validatorHistory"`
	YieldHistory        []types.TimeSeriesPoint `json:"yieldHistory"`
}

// LiquidStakingShare holds data for a single liquid staking protocol.
type LiquidStakingShare struct {
	Protocol  string  `json:"protocol"`
	Share     float64 `json:"share"`
	StakedETH float64 `json:"stakedEth"`
}

// StakingService handles staking data logic.
type StakingService struct {
	svcCtx      *svc.ServiceContext
	beaconchain *fetcher.BeaconChainClient
	etherscan   *fetcher.EtherscanClient
}

// NewStakingService creates a new StakingService.
func NewStakingService(svcCtx *svc.ServiceContext) *StakingService {
	cfg := svcCtx.Config.DataSources
	return &StakingService{
		svcCtx:      svcCtx,
		beaconchain: fetcher.NewBeaconChainClient(cfg.BeaconChain.BaseURL),
		etherscan:   fetcher.NewEtherscanClient(cfg.Etherscan.BaseURL, cfg.Etherscan.APIKey),
	}
}

// GetStakingData fetches and computes staking data.
func (s *StakingService) GetStakingData(ctx context.Context) (*StakingData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.NetworkData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "network:staking", ttl, func() (interface{}, error) {
		return s.fetchStakingData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*StakingData); ok {
		return data, nil
	}

	return s.fetchStakingData(ctx)
}

func (s *StakingService) fetchStakingData(ctx context.Context) (*StakingData, error) {
	// Fetch network stats from Beacon Chain API
	stats, statsErr := s.beaconchain.GetNetworkStats(ctx)

	var totalStakedETH float64
	var activeValidators int64
	var participationRate float64

	if statsErr == nil && stats != nil {
		totalStakedETH = stats.TotalStakedEther
		activeValidators = stats.ActiveValidators
		participationRate = stats.ParticipationRate
	}

	// Fetch total ETH supply for staking percentage calculation
	totalSupply := 0.0
	supply, supplyErr := s.etherscan.GetEthSupply(ctx)
	if supplyErr == nil && supply != nil {
		// Etherscan returns supply in Wei as a string; convert to ETH
		totalSupply = parseEthSupply(supply.EthSupply)
	}

	// Calculate staking percentage
	stakingPct := 0.0
	if r := calc.StakingPercentage(totalStakedETH, totalSupply); r != nil {
		stakingPct = *r
	}

	// Calculate staking yield (approximate: ~3-5% APR based on validator count)
	// In production, this would come from a specialized staking rewards API
	stakingYield := estimateStakingYield(activeValidators)

	// Fetch validator queue data
	var entryQueueLength, exitQueueLength int64
	queue, queueErr := s.beaconchain.GetValidatorQueue(ctx)
	if queueErr == nil && queue != nil {
		entryQueueLength = queue.Entering
		exitQueueLength = queue.Exiting
	}

	// Estimate wait times based on queue lengths
	entryWaitTime := estimateWaitTime(entryQueueLength)
	exitWaitTime := estimateWaitTime(exitQueueLength)

	// Build liquid staking shares
	liquidStakingShares := buildLiquidStakingShares(totalStakedETH)

	// Build validator history from epoch data
	validatorHistory := s.buildValidatorHistory(ctx)

	// Build yield history (placeholder - would come from historical data)
	yieldHistory := []types.TimeSeriesPoint{}

	_ = participationRate // Used in performance module

	return &StakingData{
		TotalStakedETH:      totalStakedETH,
		StakingPercentage:   stakingPct,
		ActiveValidators:    activeValidators,
		StakingYield:        stakingYield,
		EntryQueueLength:    entryQueueLength,
		ExitQueueLength:     exitQueueLength,
		EntryWaitTime:       entryWaitTime,
		ExitWaitTime:        exitWaitTime,
		LiquidStakingShares: liquidStakingShares,
		ValidatorHistory:    validatorHistory,
		YieldHistory:        yieldHistory,
	}, nil
}

// buildValidatorHistory fetches recent epoch data and builds a validator count time series.
func (s *StakingService) buildValidatorHistory(ctx context.Context) []types.TimeSeriesPoint {
	epochs, err := s.beaconchain.GetEpochHistory(ctx, 30)
	if err != nil || len(epochs) == 0 {
		return []types.TimeSeriesPoint{}
	}

	history := make([]types.TimeSeriesPoint, 0, len(epochs))
	for _, epoch := range epochs {
		// Each epoch is ~6.4 minutes; approximate timestamp from epoch number
		ts := epochToTimestamp(epoch.Epoch)
		history = append(history, types.TimeSeriesPoint{
			Timestamp: ts,
			Value:     float64(epoch.ValidatorsCount),
		})
	}

	return history
}

// buildLiquidStakingShares constructs liquid staking protocol shares.
// In production, this data would come from DefiLlama or protocol-specific APIs.
func buildLiquidStakingShares(totalStakedETH float64) []LiquidStakingShare {
	// Known liquid staking protocols and their approximate market shares
	// These would be fetched from real APIs in production
	protocols := []struct {
		Name     string
		SharePct float64 // approximate market share percentage
	}{
		{"Lido", 28.5},
		{"Rocket Pool", 3.2},
		{"Coinbase", 8.5},
		{"Binance", 3.8},
		{"Others", 56.0},
	}

	stakedValues := make([]float64, len(protocols))
	for i, p := range protocols {
		stakedValues[i] = totalStakedETH * p.SharePct / 100.0
	}

	// Use calc.CalculateShares for proper share calculation
	shares := calc.CalculateShares(stakedValues)

	result := make([]LiquidStakingShare, len(protocols))
	for i, p := range protocols {
		stakedETH := 0.0
		if i < len(stakedValues) {
			stakedETH = stakedValues[i]
		}
		share := 0.0
		if i < len(shares) {
			share = shares[i]
		}
		result[i] = LiquidStakingShare{
			Protocol:  p.Name,
			Share:     share,
			StakedETH: stakedETH,
		}
	}

	return result
}

// estimateStakingYield approximates the staking yield based on validator count.
// The yield decreases as more validators join the network.
// Formula approximation: base_reward / sqrt(total_validators) * epochs_per_year
func estimateStakingYield(activeValidators int64) float64 {
	if activeValidators <= 0 {
		return 0
	}
	// Simplified approximation of Ethereum staking APR
	// Real yield depends on many factors; this is a rough estimate
	// ~4.0% at ~900k validators, decreasing with more validators
	baseYield := 4.0
	referenceValidators := 900000.0
	// Yield scales inversely with sqrt of validator count
	ratio := sqrt(referenceValidators / float64(activeValidators))
	return baseYield * ratio
}

// sqrt computes the square root using Newton's method (avoids math import for simplicity).
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 20; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// estimateWaitTime estimates the wait time based on queue length.
// Ethereum processes ~1800 validators per day (churn limit).
func estimateWaitTime(queueLength int64) string {
	if queueLength <= 0 {
		return "~0 days"
	}
	// Churn limit: approximately 1800 validators per day at current network size
	churnPerDay := int64(1800)
	days := queueLength / churnPerDay
	if days == 0 {
		hours := (queueLength * 24) / churnPerDay
		if hours == 0 {
			return "< 1 hour"
		}
		return fmt.Sprintf("~%d hours", hours)
	}
	return fmt.Sprintf("~%d days", days)
}

// epochToTimestamp converts an epoch number to an approximate Unix timestamp.
// Ethereum Beacon Chain genesis: 2020-12-01 12:00:23 UTC (epoch 0)
// Each epoch = 32 slots * 12 seconds = 384 seconds
func epochToTimestamp(epoch int64) int64 {
	genesisTime := int64(1606824023) // Beacon Chain genesis timestamp
	epochDuration := int64(384)      // 32 slots * 12 seconds
	return genesisTime + epoch*epochDuration
}

// parseEthSupply parses the ETH supply string from Etherscan (in Wei) to ETH.
func parseEthSupply(weiStr string) float64 {
	// Etherscan returns supply in Wei as a string
	// Simple conversion: parse as float and divide by 1e18
	val := 0.0
	for _, c := range weiStr {
		if c >= '0' && c <= '9' {
			val = val*10 + float64(c-'0')
		}
	}
	return val / 1e18
}
