package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"eth-valuation-api/internal/logic/alert"
	"eth-valuation-api/internal/logic/institutional"
	"eth-valuation-api/internal/logic/macro"
	"eth-valuation-api/internal/logic/market"
	"eth-valuation-api/internal/logic/network"
	"eth-valuation-api/internal/logic/onchain"
	"eth-valuation-api/internal/logic/valuation"
	"eth-valuation-api/internal/svc"
)

// Refresh intervals for each data module.
const (
	PriceInterval         = 10 * time.Second
	OnChainInterval       = 5 * time.Minute
	InstitutionalInterval = 1 * time.Hour
	NetworkInterval       = 5 * time.Minute
	MacroInterval         = 1 * time.Hour
)

// Scheduler manages periodic data refresh for all modules and triggers
// alert evaluation after each refresh cycle.
type Scheduler struct {
	svcCtx *svc.ServiceContext

	// Services
	marketSvc      *market.MarketService
	burnSvc        *onchain.BurnService
	gasSvc         *onchain.GasService
	activitySvc    *onchain.ActivityService
	tvlSvc         *onchain.TVLService
	supplySvc      *onchain.SupplyService
	etfSvc         *institutional.ETFService
	grayscaleSvc   *institutional.GrayscaleService
	holdingsSvc    *institutional.HoldingsService
	stakingSvc     *network.StakingService
	performanceSvc *network.PerformanceService
	ethbtcSvc      *macro.ETHBTCService
	indicatorsSvc  *macro.IndicatorsService
	valuationSvc   *valuation.ValuationService
	alertSvc       alert.AlertService

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a new Scheduler with all service dependencies.
func NewScheduler(svcCtx *svc.ServiceContext) *Scheduler {
	return &Scheduler{
		svcCtx:         svcCtx,
		marketSvc:      market.NewMarketService(svcCtx),
		burnSvc:        onchain.NewBurnService(svcCtx),
		gasSvc:         onchain.NewGasService(svcCtx),
		activitySvc:    onchain.NewActivityService(svcCtx),
		tvlSvc:         onchain.NewTVLService(svcCtx),
		supplySvc:      onchain.NewSupplyService(svcCtx),
		etfSvc:         institutional.NewETFService(svcCtx),
		grayscaleSvc:   institutional.NewGrayscaleService(svcCtx),
		holdingsSvc:    institutional.NewHoldingsService(svcCtx),
		stakingSvc:     network.NewStakingService(svcCtx),
		performanceSvc: network.NewPerformanceService(svcCtx),
		ethbtcSvc:      macro.NewETHBTCService(svcCtx),
		indicatorsSvc:  macro.NewIndicatorsService(svcCtx),
		valuationSvc:   valuation.NewValuationService(svcCtx),
		alertSvc:       alert.NewAlertService(svcCtx.DB),
	}
}

// Start begins all scheduled refresh goroutines. It is non-blocking.
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.startTicker(ctx, "price", PriceInterval, s.refreshPrice)
	s.startTicker(ctx, "onchain", OnChainInterval, s.refreshOnChain)
	s.startTicker(ctx, "institutional", InstitutionalInterval, s.refreshInstitutional)
	s.startTicker(ctx, "network", NetworkInterval, s.refreshNetwork)
	s.startTicker(ctx, "macro", MacroInterval, s.refreshMacro)

	log.Println("[scheduler] started all refresh jobs")
}

// Stop gracefully shuts down all scheduled goroutines and waits for them to finish.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	log.Println("[scheduler] stopped")
}

// startTicker launches a goroutine that calls fn at the given interval until ctx is cancelled.
func (s *Scheduler) startTicker(ctx context.Context, name string, interval time.Duration, fn func(context.Context)) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start.
		fn(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fn(ctx)
			}
		}
	}()
}

// ---------------------------------------------------------------------------
// Refresh functions
// ---------------------------------------------------------------------------

// refreshPrice refreshes market price data and triggers alert evaluation.
func (s *Scheduler) refreshPrice(ctx context.Context) {
	data, err := s.marketSvc.GetMarketData(ctx)
	if err != nil {
		log.Printf("[scheduler] price refresh error: %v", err)
		return
	}

	metrics := map[string]float64{
		"eth_price":        data.CurrentPrice,
		"eth_volume_24h":   data.Volume24h,
		"eth_market_cap":   data.MarketCap,
		"eth_ath_drawdown": data.ATHDrawdown,
	}
	s.evaluateAlerts(metrics)
}

// refreshOnChain refreshes all on-chain data modules and triggers alert evaluation.
func (s *Scheduler) refreshOnChain(ctx context.Context) {
	metrics := make(map[string]float64)

	// Burn data
	if burnData, err := s.burnSvc.GetBurnData(ctx); err == nil {
		metrics["burn_daily"] = burnData.Daily
		metrics["burn_annualized_rate"] = burnData.AnnualizedBurnRate
	} else {
		log.Printf("[scheduler] burn refresh error: %v", err)
	}

	// Gas data
	if gasData, err := s.gasSvc.GetGasData(ctx); err == nil {
		metrics["gas_avg_gwei"] = gasData.CurrentAvgGwei
	} else {
		log.Printf("[scheduler] gas refresh error: %v", err)
	}

	// Activity data
	if actData, err := s.activitySvc.GetActivityData(ctx); err == nil {
		metrics["daily_active_addresses"] = float64(actData.DailyActiveAddresses)
		metrics["daily_transactions"] = float64(actData.DailyTransactions)
	} else {
		log.Printf("[scheduler] activity refresh error: %v", err)
	}

	// TVL data
	if tvlData, err := s.tvlSvc.GetTVLData(ctx); err == nil {
		metrics["tvl_total"] = tvlData.TotalTVLUsd
		metrics["tvl_dominance_weekly_change"] = tvlData.ETHTVLDominance
	} else {
		log.Printf("[scheduler] tvl refresh error: %v", err)
	}

	// Supply data
	if supplyData, err := s.supplySvc.GetSupplyData(ctx); err == nil {
		metrics["total_supply"] = supplyData.TotalSupply
		metrics["exchange_balance"] = supplyData.ExchangeBalance
	} else {
		log.Printf("[scheduler] supply refresh error: %v", err)
	}

	s.evaluateAlerts(metrics)
}

// refreshInstitutional refreshes institutional data and triggers alert evaluation.
func (s *Scheduler) refreshInstitutional(ctx context.Context) {
	metrics := make(map[string]float64)

	// ETF data
	if etfData, err := s.etfSvc.GetETFData(ctx); err == nil {
		metrics["etf_total_holdings_eth"] = etfData.TotalHoldingsETH
		metrics["etf_cumulative_net_flow"] = etfData.CumulativeNetFlow
	} else {
		log.Printf("[scheduler] etf refresh error: %v", err)
	}

	// Grayscale data
	if gsData, err := s.grayscaleSvc.GetGrayscaleData(ctx); err == nil {
		metrics["grayscale_discount_pct"] = gsData.PremiumDiscount
	} else {
		log.Printf("[scheduler] grayscale refresh error: %v", err)
	}

	// Holdings data
	if _, err := s.holdingsSvc.GetInstitutionalHoldings(ctx); err != nil {
		log.Printf("[scheduler] holdings refresh error: %v", err)
	}

	s.evaluateAlerts(metrics)
}

// refreshNetwork refreshes network health data and triggers alert evaluation.
func (s *Scheduler) refreshNetwork(ctx context.Context) {
	metrics := make(map[string]float64)

	// Staking data
	if stakingData, err := s.stakingSvc.GetStakingData(ctx); err == nil {
		metrics["total_staked_eth"] = stakingData.TotalStakedETH
		metrics["active_validators"] = float64(stakingData.ActiveValidators)
		metrics["staking_yield"] = stakingData.StakingYield
	} else {
		log.Printf("[scheduler] staking refresh error: %v", err)
	}

	// Performance data
	if perfData, err := s.performanceSvc.GetNetworkPerformance(ctx); err == nil {
		metrics["missed_slots_rate_pct"] = perfData.MissedSlotsRate
		metrics["current_tps"] = perfData.CurrentTPS
	} else {
		log.Printf("[scheduler] performance refresh error: %v", err)
	}

	s.evaluateAlerts(metrics)
}

// refreshMacro refreshes macro economic data and triggers alert evaluation.
func (s *Scheduler) refreshMacro(ctx context.Context) {
	metrics := make(map[string]float64)

	// ETH/BTC data
	if ethbtcData, err := s.ethbtcSvc.GetETHBTCData(ctx); err == nil {
		metrics["ethbtc_price"] = ethbtcData.ETHBTCPrice
		metrics["eth_dominance"] = ethbtcData.ETHDominance
	} else {
		log.Printf("[scheduler] ethbtc refresh error: %v", err)
	}

	// Macro indicators
	if macroData, err := s.indicatorsSvc.GetMacroIndicators(ctx); err == nil {
		metrics["dxy_index"] = macroData.DXYIndex
		metrics["treasury_10y"] = macroData.Treasury10Y
		metrics["fear_greed_index"] = float64(macroData.FearGreedIndex)
		metrics["stablecoin_market_cap"] = macroData.StablecoinMarketCap
	} else {
		log.Printf("[scheduler] macro indicators refresh error: %v", err)
	}

	s.evaluateAlerts(metrics)
}

// evaluateAlerts runs the alert service against the collected metrics.
func (s *Scheduler) evaluateAlerts(metrics map[string]float64) {
	if len(metrics) == 0 {
		return
	}
	triggered := s.alertSvc.EvaluateAlerts(metrics)
	if len(triggered) > 0 {
		log.Printf("[scheduler] %d alert(s) triggered", len(triggered))
	}
}
