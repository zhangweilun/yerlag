package onchain

import (
	"context"
	"sort"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// TVLData holds the TVL data response.
type TVLData struct {
	TotalTVLUsd         float64                 `json:"totalTvlUsd"`
	TotalTVLEth         float64                 `json:"totalTvlEth"`
	TVLToMarketCapRatio float64                 `json:"tvlToMarketCapRatio"`
	ETHTVLDominance     float64                 `json:"ethTvlDominance"`
	TopProtocols        []ProtocolTVL           `json:"topProtocols"`
	TVLHistory          []types.TimeSeriesPoint `json:"tvlHistory"`
	DominanceHistory    []types.TimeSeriesPoint `json:"dominanceHistory"`
}

// ProtocolTVL holds protocol-level TVL data with share percentage.
type ProtocolTVL struct {
	Name   string  `json:"name"`
	TVLUsd float64 `json:"tvlUsd"`
	Share  float64 `json:"share"` // Percentage share (%)
}

// TVLService handles TVL data logic.
type TVLService struct {
	svcCtx    *svc.ServiceContext
	defillama *fetcher.DefiLlamaClient
	coingecko *fetcher.CoinGeckoClient
}

// NewTVLService creates a new TVLService.
func NewTVLService(svcCtx *svc.ServiceContext) *TVLService {
	cfg := svcCtx.Config.DataSources
	return &TVLService{
		svcCtx:    svcCtx,
		defillama: fetcher.NewDefiLlamaClient(cfg.DefiLlama.BaseURL),
		coingecko: fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey),
	}
}

// GetTVLData fetches and computes TVL data.
func (s *TVLService) GetTVLData(ctx context.Context) (*TVLData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.TVLData) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "onchain:tvl", ttl, func() (interface{}, error) {
		return s.fetchTVLData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*TVLData); ok {
		return data, nil
	}

	return s.fetchTVLData(ctx)
}

func (s *TVLService) fetchTVLData(ctx context.Context) (*TVLData, error) {
	// Fetch chain TVL data
	chainsTVL, err := s.defillama.GetChainsTVL(ctx)
	if err != nil {
		return nil, err
	}

	// Find Ethereum TVL and total TVL across all chains
	var ethTVL float64
	var totalChainsTVL float64
	for _, chain := range chainsTVL {
		totalChainsTVL += chain.TVL
		if chain.Name == "Ethereum" {
			ethTVL = chain.TVL
		}
	}

	// Calculate ETH TVL dominance
	ethTVLDominance := 0.0
	if r := calc.SafeRatio(ethTVL, totalChainsTVL); r != nil {
		ethTVLDominance = *r * 100
	}

	// Fetch ETH price for TVL in ETH conversion
	ethPrice := 0.0
	prices, priceErr := s.coingecko.GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
	if priceErr == nil {
		if ethPrices, ok := prices["ethereum"]; ok {
			ethPrice = ethPrices["usd"]
		}
	}

	totalTVLEth := 0.0
	if ethPrice > 0 {
		totalTVLEth = ethTVL / ethPrice
	}

	// Calculate TVL to market cap ratio
	tvlToMarketCapRatio := 0.0
	marketData, mdErr := s.coingecko.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if mdErr == nil && len(marketData) > 0 {
		marketCap := marketData[0].MarketCap
		if r := calc.TVLToMarketCapRatio(ethTVL, marketCap); r != nil {
			tvlToMarketCapRatio = *r
		}
	}

	// Fetch top protocols on Ethereum
	topProtocols := s.fetchTopProtocols(ctx, ethTVL)

	// Fetch TVL history for Ethereum
	var tvlHistory []types.TimeSeriesPoint
	historyData, histErr := s.defillama.GetChainTVLHistory(ctx, "Ethereum")
	if histErr == nil {
		for _, dp := range historyData {
			tvlHistory = append(tvlHistory, types.TimeSeriesPoint{
				Timestamp: dp.Date,
				Value:     dp.TVL,
			})
		}
	}

	// Calculate dominance history from total TVL history
	var dominanceHistory []types.TimeSeriesPoint
	totalHistory, totalHistErr := s.defillama.GetTotalTVLHistory(ctx)
	if totalHistErr == nil && len(tvlHistory) > 0 {
		dominanceHistory = calculateDominanceHistory(tvlHistory, totalHistory)
	}

	return &TVLData{
		TotalTVLUsd:         ethTVL,
		TotalTVLEth:         totalTVLEth,
		TVLToMarketCapRatio: tvlToMarketCapRatio,
		ETHTVLDominance:     ethTVLDominance,
		TopProtocols:        topProtocols,
		TVLHistory:          tvlHistory,
		DominanceHistory:    dominanceHistory,
	}, nil
}

func (s *TVLService) fetchTopProtocols(ctx context.Context, ethTotalTVL float64) []ProtocolTVL {
	protocols, err := s.defillama.GetProtocols(ctx)
	if err != nil {
		return nil
	}

	// Filter Ethereum protocols and sort by TVL
	var ethProtocols []fetcher.DefiLlamaProtocol
	for _, p := range protocols {
		if p.Chain == "Ethereum" || p.Chain == "Multi-Chain" {
			ethProtocols = append(ethProtocols, p)
		}
	}

	sort.Slice(ethProtocols, func(i, j int) bool {
		return ethProtocols[i].TVL > ethProtocols[j].TVL
	})

	// Take top 10 protocols
	limit := 10
	if len(ethProtocols) < limit {
		limit = len(ethProtocols)
	}
	topProtos := ethProtocols[:limit]

	// Calculate shares
	tvlValues := make([]float64, len(topProtos))
	for i, p := range topProtos {
		tvlValues[i] = p.TVL
	}
	shares := calc.CalculateShares(tvlValues)

	result := make([]ProtocolTVL, len(topProtos))
	for i, p := range topProtos {
		share := 0.0
		if i < len(shares) {
			share = shares[i]
		}
		result[i] = ProtocolTVL{
			Name:   p.Name,
			TVLUsd: p.TVL,
			Share:  share,
		}
	}

	return result
}

// calculateDominanceHistory computes ETH TVL dominance over time.
func calculateDominanceHistory(ethHistory []types.TimeSeriesPoint, totalHistory []fetcher.DefiLlamaTVLHistory) []types.TimeSeriesPoint {
	// Build a map of total TVL by timestamp for quick lookup
	totalMap := make(map[int64]float64, len(totalHistory))
	for _, dp := range totalHistory {
		totalMap[dp.Date] = dp.TVL
	}

	var result []types.TimeSeriesPoint
	for _, ethDP := range ethHistory {
		if totalTVL, ok := totalMap[ethDP.Timestamp]; ok && totalTVL > 0 {
			dominance := (ethDP.Value / totalTVL) * 100
			result = append(result, types.TimeSeriesPoint{
				Timestamp: ethDP.Timestamp,
				Value:     dominance,
			})
		}
	}

	return result
}
