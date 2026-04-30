package onchain

import (
	"context"
	"time"

	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/logic/calc"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"
)

// SupplyData holds the ETH supply data response.
type SupplyData struct {
	TotalSupply            float64                 `json:"totalSupply"`
	StakedAmount           float64                 `json:"stakedAmount"`
	DeFiLocked             float64                 `json:"defiLocked"`
	ExchangeBalance        float64                 `json:"exchangeBalance"`
	OtherAmount            float64                 `json:"otherAmount"`
	NetIssuance            float64                 `json:"netIssuance"`
	IsDeflationary         bool                    `json:"isDeflationary"`
	AnnualInflationRate    float64                 `json:"annualInflationRate"`
	SupplyHistory          []types.TimeSeriesPoint `json:"supplyHistory"`
	ExchangeBalanceHistory []types.TimeSeriesPoint `json:"exchangeBalanceHistory"`
}

// SupplyService handles ETH supply data logic.
type SupplyService struct {
	svcCtx    *svc.ServiceContext
	etherscan *fetcher.EtherscanClient
	glassnode *fetcher.GlassnodeClient
	defillama *fetcher.DefiLlamaClient
}

// NewSupplyService creates a new SupplyService.
func NewSupplyService(svcCtx *svc.ServiceContext) *SupplyService {
	cfg := svcCtx.Config.DataSources
	return &SupplyService{
		svcCtx:    svcCtx,
		etherscan: fetcher.NewEtherscanClient(cfg.Etherscan.BaseURL, cfg.Etherscan.APIKey),
		glassnode: fetcher.NewGlassnodeClient(cfg.Glassnode.BaseURL, cfg.Glassnode.APIKey),
		defillama: fetcher.NewDefiLlamaClient(cfg.DefiLlama.BaseURL),
	}
}

// GetSupplyData fetches and computes ETH supply data.
func (s *SupplyService) GetSupplyData(ctx context.Context) (*SupplyData, error) {
	ttl := time.Duration(s.svcCtx.Config.CacheTTL.OnChainMetrics) * time.Second

	result, err := s.svcCtx.DataFetcher.Fetch(ctx, "onchain:supply", ttl, func() (interface{}, error) {
		return s.fetchSupplyData(ctx)
	})
	if err != nil {
		return nil, err
	}

	if data, ok := result.Data.(*SupplyData); ok {
		return data, nil
	}

	return s.fetchSupplyData(ctx)
}

func (s *SupplyService) fetchSupplyData(ctx context.Context) (*SupplyData, error) {
	// Fetch ETH supply data from Etherscan
	supplyData, err := s.etherscan.GetEthSupply(ctx)
	if err != nil {
		return nil, err
	}

	totalSupply := parseWeiToEth(supplyData.EthSupply)
	stakedAmount := parseWeiToEth(supplyData.Eth2Staking)
	burntFees := parseWeiToEth(supplyData.BurntFees)

	// Fetch exchange balance from Glassnode
	now := time.Now()
	since := now.AddDate(0, 0, -90).Unix()
	until := now.Unix()

	params := fetcher.GlassnodeMetricParams{
		Asset:    "ETH",
		Since:    since,
		Until:    until,
		Interval: "24h",
	}

	var exchangeBalance float64
	var exchangeBalanceHistory []types.TimeSeriesPoint
	exchangeData, exErr := s.glassnode.GetExchangeBalance(ctx, params)
	if exErr == nil && len(exchangeData) > 0 {
		exchangeBalance = exchangeData[len(exchangeData)-1].Value
		for _, dp := range exchangeData {
			exchangeBalanceHistory = append(exchangeBalanceHistory, types.TimeSeriesPoint{
				Timestamp: dp.Timestamp,
				Value:     dp.Value,
			})
		}
	}

	// Fetch DeFi TVL on Ethereum (as DeFi locked amount)
	var defiLocked float64
	chainsTVL, tvlErr := s.defillama.GetChainsTVL(ctx)
	if tvlErr == nil {
		for _, chain := range chainsTVL {
			if chain.Name == "Ethereum" {
				// Convert USD TVL to ETH (approximate)
				// Get ETH price for conversion
				prices, priceErr := s.coingecko().GetSimplePrice(ctx, []string{"ethereum"}, []string{"usd"})
				if priceErr == nil {
					if ethPrices, ok := prices["ethereum"]; ok && ethPrices["usd"] > 0 {
						defiLocked = chain.TVL / ethPrices["usd"]
					}
				}
				break
			}
		}
	}

	// Calculate "other" amount
	otherAmount := totalSupply - stakedAmount - defiLocked - exchangeBalance
	if otherAmount < 0 {
		otherAmount = 0
	}

	// Fetch issuance data for net issuance calculation
	var dailyIssuance float64
	issuanceData, issErr := s.glassnode.GetIssuance(ctx, params)
	if issErr == nil && len(issuanceData) > 0 {
		dailyIssuance = issuanceData[len(issuanceData)-1].Value
	}

	// Calculate daily burn (approximate from cumulative burn / days since EIP-1559)
	// Use a simpler approach: annualized burn from the burn rate
	dailyBurn := burntFees / float64(daysSinceEIP1559(now))
	if dailyBurn < 0 {
		dailyBurn = 0
	}

	// Net issuance = daily issuance - daily burn
	netIssuance := calc.NetIssuance(dailyIssuance, dailyBurn)
	isDeflationary := calc.IsDeflationary(netIssuance)

	// Annual inflation rate
	annualInflationRate := 0.0
	annualNetIssuance := netIssuance * 365
	if r := calc.AnnualInflationRate(annualNetIssuance, totalSupply); r != nil {
		annualInflationRate = *r
	}

	// Fetch supply history
	var supplyHistory []types.TimeSeriesPoint
	supplyHistData, supHistErr := s.glassnode.GetSupply(ctx, params)
	if supHistErr == nil {
		for _, dp := range supplyHistData {
			supplyHistory = append(supplyHistory, types.TimeSeriesPoint{
				Timestamp: dp.Timestamp,
				Value:     dp.Value,
			})
		}
	}

	return &SupplyData{
		TotalSupply:            totalSupply,
		StakedAmount:           stakedAmount,
		DeFiLocked:             defiLocked,
		ExchangeBalance:        exchangeBalance,
		OtherAmount:            otherAmount,
		NetIssuance:            netIssuance,
		IsDeflationary:         isDeflationary,
		AnnualInflationRate:    annualInflationRate,
		SupplyHistory:          supplyHistory,
		ExchangeBalanceHistory: exchangeBalanceHistory,
	}, nil
}

// coingecko returns a CoinGecko client (helper to avoid storing it as a field
// since it's only used in one place).
func (s *SupplyService) coingecko() *fetcher.CoinGeckoClient {
	cfg := s.svcCtx.Config.DataSources
	return fetcher.NewCoinGeckoClient(cfg.CoinGecko.BaseURL, cfg.CoinGecko.APIKey)
}

// daysSinceEIP1559 returns the number of days since EIP-1559 activation (Aug 5, 2021).
func daysSinceEIP1559(now time.Time) int {
	eip1559Date := time.Date(2021, 8, 5, 0, 0, 0, 0, time.UTC)
	days := int(now.Sub(eip1559Date).Hours() / 24)
	if days < 1 {
		return 1
	}
	return days
}
