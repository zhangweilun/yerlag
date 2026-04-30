package network

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
)

// NetworkPerformance holds the network performance data response.
type NetworkPerformance struct {
	AvgBlockTime     float64       `json:"avgBlockTime"`
	BlockUtilization float64       `json:"blockUtilization"`
	CurrentTPS       float64       `json:"currentTps"`
	MaxTPS           float64       `json:"maxTps"`
	TPSRatio         float64       `json:"tpsRatio"`
	MissedSlots24h   int64         `json:"missedSlots24h"`
	MissedSlotsRate  float64       `json:"missedSlotsRate"`
	AttestationRate  float64       `json:"attestationRate"`
	ClientDiversity  []ClientShare `json:"clientDiversity"`
}

// ClientShare holds data for a single client's share of the network.
type ClientShare struct {
	Client string  `json:"client"`
	Share  float64 `json:"share"`
}

// PerformanceService handles network performance data logic.
type PerformanceService struct {
	svcCtx      *svc.ServiceContext
	beaconchain *fetcher.BeaconChainClient
}

// NewPerformanceService creates a new PerformanceService.
func NewPerformanceService(svcCtx *svc.ServiceContext) *PerformanceService {
	cfg := svcCtx.Config.DataSources
	return &PerformanceService{
		svcCtx:      svcCtx,
		beaconchain: fetcher.NewBeaconChainClient(cfg.BeaconChain.BaseURL),
	}
}

// GetNetworkPerformance fetches and computes network performance data.
func (s *PerformanceService) GetNetworkPerformance(ctx context.Context) (*NetworkPerformance, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.NetworkData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "network:performance", ttl, func() (interface{}, error) {
		return s.fetchPerformanceData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*NetworkPerformance); ok {
		return data, nil
	}

	return s.fetchPerformanceData(ctx)
}

func (s *PerformanceService) fetchPerformanceData(ctx context.Context) (*NetworkPerformance, error) {
	// Fetch latest epoch data for participation rate and block info
	epoch, epochErr := s.beaconchain.GetLatestEpoch(ctx)

	var attestationRate float64
	if epochErr == nil && epoch != nil {
		attestationRate = epoch.GlobalParticipationRate
	}

	// Ethereum PoS block time is fixed at 12 seconds per slot
	avgBlockTime := 12.0

	// Calculate block utilization from proposed vs scheduled blocks
	blockUtilization := 0.0
	if epochErr == nil && epoch != nil && epoch.ScheduledBlocks > 0 {
		blockUtilization = float64(epoch.ProposedBlocks) / float64(epoch.ScheduledBlocks) * 100.0
	}

	// Fetch missed slots data (last ~225 epochs ≈ 24 hours)
	// 24 hours / 6.4 min per epoch ≈ 225 epochs
	epochsIn24h := 225
	missedSlots24h, totalSlots24h, missedErr := s.beaconchain.GetMissedSlots(ctx, epochsIn24h)
	if missedErr != nil {
		missedSlots24h = 0
		totalSlots24h = 0
	}

	// Calculate missed slots rate
	missedSlotsRate := 0.0
	if totalSlots24h > 0 {
		missedSlotsRate = float64(missedSlots24h) / float64(totalSlots24h) * 100.0
	}

	// Estimate current TPS from recent transaction data
	// Ethereum mainnet processes ~12-15 TPS on average
	currentTPS := estimateCurrentTPS(epoch)

	// Theoretical max TPS for Ethereum mainnet (with current gas limit)
	maxTPS := 30.0

	// Calculate TPS ratio using calc package
	tpsRatio := 0.0
	if r := calc.TPSRatio(currentTPS, maxTPS); r != nil {
		tpsRatio = *r
	}

	// Build client diversity data
	clientDiversity := buildClientDiversity()

	return &NetworkPerformance{
		AvgBlockTime:     avgBlockTime,
		BlockUtilization: blockUtilization,
		CurrentTPS:       currentTPS,
		MaxTPS:           maxTPS,
		TPSRatio:         tpsRatio,
		MissedSlots24h:   missedSlots24h,
		MissedSlotsRate:  missedSlotsRate,
		AttestationRate:  attestationRate,
		ClientDiversity:  clientDiversity,
	}, nil
}

// estimateCurrentTPS estimates the current TPS from epoch data.
// In production, this would come from a dedicated transaction count API.
func estimateCurrentTPS(epoch *fetcher.BeaconChainEpoch) float64 {
	if epoch == nil {
		return 0
	}
	// Approximate: if we have proposed blocks info, estimate TPS
	// Each block can contain ~150-200 transactions on average
	// Epoch has 32 slots, each slot is 12 seconds = 384 seconds per epoch
	if epoch.ProposedBlocks > 0 {
		avgTxPerBlock := 150.0
		epochDuration := 384.0 // seconds
		totalTx := float64(epoch.ProposedBlocks) * avgTxPerBlock
		return totalTx / epochDuration
	}
	// Default estimate
	return 12.0
}

// buildClientDiversity constructs the client diversity distribution.
// In production, this data would come from clientdiversity.org or similar sources.
func buildClientDiversity() []ClientShare {
	// Approximate execution client diversity (as of 2024)
	// These would be fetched from real monitoring APIs in production
	clients := []struct {
		Name  string
		Value float64
	}{
		{"Geth", 45.0},
		{"Nethermind", 22.0},
		{"Besu", 18.0},
		{"Erigon", 10.0},
		{"Others", 5.0},
	}

	values := make([]float64, len(clients))
	for i, c := range clients {
		values[i] = c.Value
	}

	// Use calc.CalculateShares for proper share calculation
	shares := calc.CalculateShares(values)

	result := make([]ClientShare, len(clients))
	for i, c := range clients {
		share := 0.0
		if i < len(shares) {
			share = shares[i]
		}
		result[i] = ClientShare{
			Client: c.Name,
			Share:  share,
		}
	}

	return result
}
