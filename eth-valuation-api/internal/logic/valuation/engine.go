package valuation

import (
	"eth-valuation-api/internal/logic/calc"
	"math"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// ValuationScore is the top-level result returned by the valuation engine.
type ValuationScore struct {
	Overall   float64            `json:"overall"` // 0-100
	Status    string             `json:"status"`  // "undervalued" | "fair" | "overvalued"
	Breakdown ValuationBreakdown `json:"breakdown"`
	RadarData []RadarDataPoint   `json:"radarData"`
}

// ValuationBreakdown holds the individual model results.
type ValuationBreakdown struct {
	MVRVScore        MVRVResult        `json:"mvrvScore"`
	PriceToFeeScore  PriceToFeeResult  `json:"priceToFeeScore"`
	DCFValuation     DCFResult         `json:"dcfValuation"`
	StockToFlowScore StockToFlowResult `json:"stockToFlowScore"`
	NVTScore         NVTResult         `json:"nvtScore"`
	ETHBTCScore      ETHBTCResult      `json:"ethBtcScore"`
}

// RadarDataPoint represents one axis of the radar chart.
type RadarDataPoint struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"` // 0-100
	Label     string  `json:"label"`
}

// MVRVResult holds the MVRV Ratio evaluation.
type MVRVResult struct {
	Ratio                float64 `json:"ratio"`
	HistoricalPercentile float64 `json:"historicalPercentile"`
	Signal               string  `json:"signal"` // "overvalued" | "undervalued" | "neutral"
	Score                float64 `json:"score"`  // 0-100
}

// PriceToFeeResult holds the P/F Ratio evaluation.
type PriceToFeeResult struct {
	Ratio                float64 `json:"ratio"`
	HistoricalPercentile float64 `json:"historicalPercentile"`
	Signal               string  `json:"signal"`
	Score                float64 `json:"score"` // 0-100
}

// DCFResult holds the DCF model output.
type DCFResult struct {
	FairValueLow  float64        `json:"fairValueLow"`
	FairValueMid  float64        `json:"fairValueMid"`
	FairValueHigh float64        `json:"fairValueHigh"`
	Assumptions   DCFAssumptions `json:"assumptions"`
	Score         float64        `json:"score"` // 0-100
}

// DCFAssumptions are the parameters fed into the DCF model.
type DCFAssumptions struct {
	DiscountRate       float64 `json:"discountRate"`
	GrowthRate         float64 `json:"growthRate"`
	TerminalGrowthRate float64 `json:"terminalGrowthRate"`
	ProjectionYears    int     `json:"projectionYears"`
}

// StockToFlowResult holds the S2F model output.
type StockToFlowResult struct {
	Ratio      float64 `json:"ratio"`
	ModelPrice float64 `json:"modelPrice"`
	Deviation  float64 `json:"deviation"` // (currentPrice - modelPrice) / modelPrice * 100
	Score      float64 `json:"score"`     // 0-100
}

// NVTResult holds the NVT Ratio evaluation.
type NVTResult struct {
	Ratio                float64 `json:"ratio"`
	HistoricalPercentile float64 `json:"historicalPercentile"`
	Signal               string  `json:"signal"`
	Score                float64 `json:"score"` // 0-100
}

// ETHBTCResult holds the ETH/BTC relative valuation evaluation.
type ETHBTCResult struct {
	Ratio                float64 `json:"ratio"`
	HistoricalPercentile float64 `json:"historicalPercentile"`
	Signal               string  `json:"signal"`
	Score                float64 `json:"score"` // 0-100
}

// ---------------------------------------------------------------------------
// Input types
// ---------------------------------------------------------------------------

// MVRVInput contains the data needed to score the MVRV metric.
type MVRVInput struct {
	MarketValue   float64
	RealizedValue float64
	History       []float64 // historical MVRV ratios
}

// PriceToFeeInput contains the data needed to score the P/F metric.
type PriceToFeeInput struct {
	MarketCap            float64
	AnnualizedFeeRevenue float64
	History              []float64 // historical P/F ratios
}

// DCFInput contains the data needed for the DCF model.
type DCFInput struct {
	AnnualCashFlow     float64 // current annual cash flow (fee revenue + burn value)
	CurrentPrice       float64
	TotalSupply        float64
	DiscountRate       float64 // e.g. 0.12 for 12%
	GrowthRate         float64 // e.g. 0.15 for 15%
	TerminalGrowthRate float64 // e.g. 0.03 for 3%
	ProjectionYears    int
}

// StockToFlowInput contains the data needed for the S2F model.
type StockToFlowInput struct {
	CurrentStock float64 // total supply
	AnnualFlow   float64 // annual net issuance (can be negative for deflationary)
	CurrentPrice float64
}

// NVTInput contains the data needed to score the NVT metric.
type NVTInput struct {
	MarketCap   float64
	DailyVolume float64
	History     []float64 // historical NVT ratios
}

// ETHBTCInput contains the data needed to score the ETH/BTC metric.
type ETHBTCInput struct {
	CurrentRatio float64
	History      []float64 // historical ETH/BTC ratios
}

// ValuationInput aggregates all inputs for the valuation engine.
type ValuationInput struct {
	MVRV       MVRVInput
	PriceToFee PriceToFeeInput
	DCF        DCFInput
	S2F        StockToFlowInput
	NVT        NVTInput
	ETHBTC     ETHBTCInput
}

// ---------------------------------------------------------------------------
// Individual model calculations
// ---------------------------------------------------------------------------

// CalculateMVRVScore computes the MVRV score.
// Higher percentile → more overvalued → higher score.
func CalculateMVRVScore(input MVRVInput) MVRVResult {
	ratio := 0.0
	if r := calc.MVRVRatio(input.MarketValue, input.RealizedValue); r != nil {
		ratio = *r
	}

	percentile := calc.CalculatePercentile(input.History, ratio)
	signal := calc.ClassifySignal(percentile)

	return MVRVResult{
		Ratio:                ratio,
		HistoricalPercentile: percentile,
		Signal:               signal,
		Score:                clamp(percentile, 0, 100),
	}
}

// CalculatePriceToFeeScore computes the P/F Ratio score.
// Higher percentile → more overvalued → higher score.
func CalculatePriceToFeeScore(input PriceToFeeInput) PriceToFeeResult {
	ratio := 0.0
	if r := calc.PriceToFeeRatio(input.MarketCap, input.AnnualizedFeeRevenue); r != nil {
		ratio = *r
	}

	percentile := calc.CalculatePercentile(input.History, ratio)
	signal := calc.ClassifySignal(percentile)

	return PriceToFeeResult{
		Ratio:                ratio,
		HistoricalPercentile: percentile,
		Signal:               signal,
		Score:                clamp(percentile, 0, 100),
	}
}

// CalculateDCF computes the DCF fair value range.
//
// The model projects annual cash flows forward using the growth rate, then
// discounts them back at the discount rate. A terminal value is computed using
// the Gordon Growth Model. Three scenarios are produced:
//   - Low:  uses 80% of the base cash flow
//   - Mid:  uses 100% of the base cash flow
//   - High: uses 120% of the base cash flow
//
// The resulting per-token fair values are always non-negative and ordered
// low ≤ mid ≤ high.
func CalculateDCF(input DCFInput) DCFResult {
	assumptions := DCFAssumptions{
		DiscountRate:       input.DiscountRate,
		GrowthRate:         input.GrowthRate,
		TerminalGrowthRate: input.TerminalGrowthRate,
		ProjectionYears:    input.ProjectionYears,
	}

	if input.TotalSupply <= 0 || input.ProjectionYears <= 0 || input.DiscountRate <= input.TerminalGrowthRate {
		return DCFResult{Assumptions: assumptions}
	}

	lowCF := input.AnnualCashFlow * 0.8
	midCF := input.AnnualCashFlow
	highCF := input.AnnualCashFlow * 1.2

	lowVal := dcfValue(lowCF, input.DiscountRate, input.GrowthRate, input.TerminalGrowthRate, input.ProjectionYears)
	midVal := dcfValue(midCF, input.DiscountRate, input.GrowthRate, input.TerminalGrowthRate, input.ProjectionYears)
	highVal := dcfValue(highCF, input.DiscountRate, input.GrowthRate, input.TerminalGrowthRate, input.ProjectionYears)

	// Per-token values, clamped to non-negative.
	fairLow := math.Max(0, lowVal/input.TotalSupply)
	fairMid := math.Max(0, midVal/input.TotalSupply)
	fairHigh := math.Max(0, highVal/input.TotalSupply)

	// Ensure ordering: low ≤ mid ≤ high.
	if fairLow > fairMid {
		fairLow = fairMid
	}
	if fairMid > fairHigh {
		fairMid = fairHigh
	}
	if fairLow > fairMid {
		fairLow = fairMid
	}

	// Score: how the current price compares to the mid fair value.
	// If currentPrice < fairValueMid → undervalued (lower score).
	// If currentPrice > fairValueMid → overvalued (higher score).
	score := 50.0
	if fairMid > 0 && input.CurrentPrice > 0 {
		ratio := input.CurrentPrice / fairMid
		// Map ratio to 0-100: ratio=0.5 → 25, ratio=1.0 → 50, ratio=2.0 → 100
		score = clamp(ratio*50, 0, 100)
	}

	return DCFResult{
		FairValueLow:  fairLow,
		FairValueMid:  fairMid,
		FairValueHigh: fairHigh,
		Assumptions:   assumptions,
		Score:         score,
	}
}

// dcfValue computes the total present value of projected cash flows plus
// terminal value.
func dcfValue(baseCF, discountRate, growthRate, terminalGrowthRate float64, years int) float64 {
	if discountRate <= terminalGrowthRate {
		return 0
	}

	var totalPV float64
	cf := baseCF
	for i := 1; i <= years; i++ {
		cf *= (1 + growthRate)
		pv := cf / math.Pow(1+discountRate, float64(i))
		totalPV += pv
	}

	// Terminal value using Gordon Growth Model.
	terminalCF := cf * (1 + terminalGrowthRate)
	terminalValue := terminalCF / (discountRate - terminalGrowthRate)
	pvTerminal := terminalValue / math.Pow(1+discountRate, float64(years))
	totalPV += pvTerminal

	return totalPV
}

// CalculateStockToFlow computes the Stock-to-Flow model result.
//
// S2F ratio = stock / flow (with zero protection).
// Model price is derived as: modelPrice = exp(3.0 * ln(s2fRatio) + 1.0)
// This is a simplified S2F regression model.
// Deviation = (currentPrice - modelPrice) / modelPrice * 100
func CalculateStockToFlow(input StockToFlowInput) StockToFlowResult {
	if input.AnnualFlow <= 0 {
		// If flow is zero or negative (deflationary), S2F is infinite/undefined.
		// Return a neutral result.
		return StockToFlowResult{
			Score: 50,
		}
	}

	s2fRatio := input.CurrentStock / input.AnnualFlow

	// Model price: simplified S2F power-law regression.
	// modelPrice = exp(3.0 * ln(s2fRatio) + 1.0)
	var modelPrice float64
	if s2fRatio > 0 {
		modelPrice = math.Exp(3.0*math.Log(s2fRatio) + 1.0)
	}

	deviation := 0.0
	if modelPrice > 0 {
		deviation = (input.CurrentPrice - modelPrice) / modelPrice * 100
	}

	// Score: deviation maps to 0-100.
	// Large positive deviation → overvalued → higher score.
	// Large negative deviation → undervalued → lower score.
	// deviation = 0 → score = 50
	// We use a sigmoid-like mapping: score = 50 + deviation/4, clamped.
	score := clamp(50+deviation/4, 0, 100)

	return StockToFlowResult{
		Ratio:      s2fRatio,
		ModelPrice: modelPrice,
		Deviation:  deviation,
		Score:      score,
	}
}

// CalculateNVTScore computes the NVT Ratio score.
// Higher percentile → more overvalued → higher score.
func CalculateNVTScore(input NVTInput) NVTResult {
	ratio := 0.0
	if r := calc.NVTRatio(input.MarketCap, input.DailyVolume); r != nil {
		ratio = *r
	}

	percentile := calc.CalculatePercentile(input.History, ratio)
	signal := calc.ClassifySignal(percentile)

	return NVTResult{
		Ratio:                ratio,
		HistoricalPercentile: percentile,
		Signal:               signal,
		Score:                clamp(percentile, 0, 100),
	}
}

// CalculateETHBTCScore computes the ETH/BTC relative valuation score.
// Higher percentile → ETH relatively more expensive → higher score.
func CalculateETHBTCScore(input ETHBTCInput) ETHBTCResult {
	percentile := calc.CalculatePercentile(input.History, input.CurrentRatio)
	signal := calc.ClassifyETHBTCSignal(percentile)

	return ETHBTCResult{
		Ratio:                input.CurrentRatio,
		HistoricalPercentile: percentile,
		Signal:               signal,
		Score:                clamp(percentile, 0, 100),
	}
}

// ---------------------------------------------------------------------------
// Aggregation
// ---------------------------------------------------------------------------

// defaultWeights defines the weight of each model in the overall score.
// All weights sum to 1.0.
var defaultWeights = map[string]float64{
	"mvrv":   0.20,
	"pf":     0.15,
	"dcf":    0.20,
	"s2f":    0.15,
	"nvt":    0.15,
	"ethbtc": 0.15,
}

// CalculateValuation runs all models and produces the aggregate ValuationScore.
func CalculateValuation(input ValuationInput) ValuationScore {
	mvrv := CalculateMVRVScore(input.MVRV)
	pf := CalculatePriceToFeeScore(input.PriceToFee)
	dcf := CalculateDCF(input.DCF)
	s2f := CalculateStockToFlow(input.S2F)
	nvt := CalculateNVTScore(input.NVT)
	ethbtc := CalculateETHBTCScore(input.ETHBTC)

	// Weighted average.
	overall := mvrv.Score*defaultWeights["mvrv"] +
		pf.Score*defaultWeights["pf"] +
		dcf.Score*defaultWeights["dcf"] +
		s2f.Score*defaultWeights["s2f"] +
		nvt.Score*defaultWeights["nvt"] +
		ethbtc.Score*defaultWeights["ethbtc"]

	overall = clamp(overall, 0, 100)
	status := classifyOverall(overall)

	breakdown := ValuationBreakdown{
		MVRVScore:        mvrv,
		PriceToFeeScore:  pf,
		DCFValuation:     dcf,
		StockToFlowScore: s2f,
		NVTScore:         nvt,
		ETHBTCScore:      ethbtc,
	}

	radarData := buildRadarData(mvrv, pf, dcf, s2f, nvt, ethbtc)

	return ValuationScore{
		Overall:   overall,
		Status:    status,
		Breakdown: breakdown,
		RadarData: radarData,
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clamp restricts v to the range [lo, hi].
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// classifyOverall maps the overall score to a status label.
//   - score < 33  → "undervalued"
//   - 33 ≤ score ≤ 66 → "fair"
//   - score > 66  → "overvalued"
func classifyOverall(score float64) string {
	if score < 33 {
		return "undervalued"
	}
	if score > 66 {
		return "overvalued"
	}
	return "fair"
}

// buildRadarData produces the radar chart data points from individual model
// results.
func buildRadarData(
	mvrv MVRVResult,
	pf PriceToFeeResult,
	dcf DCFResult,
	s2f StockToFlowResult,
	nvt NVTResult,
	ethbtc ETHBTCResult,
) []RadarDataPoint {
	return []RadarDataPoint{
		{Dimension: "MVRV", Score: clamp(mvrv.Score, 0, 100), Label: mvrv.Signal},
		{Dimension: "P/F Ratio", Score: clamp(pf.Score, 0, 100), Label: pf.Signal},
		{Dimension: "DCF", Score: clamp(dcf.Score, 0, 100), Label: dcfLabel(dcf)},
		{Dimension: "Stock-to-Flow", Score: clamp(s2f.Score, 0, 100), Label: s2fLabel(s2f)},
		{Dimension: "NVT", Score: clamp(nvt.Score, 0, 100), Label: nvt.Signal},
		{Dimension: "ETH/BTC", Score: clamp(ethbtc.Score, 0, 100), Label: ethbtc.Signal},
	}
}

// dcfLabel produces a human-readable label for the DCF result.
func dcfLabel(dcf DCFResult) string {
	if dcf.Score < 33 {
		return "undervalued"
	}
	if dcf.Score > 66 {
		return "overvalued"
	}
	return "fair"
}

// s2fLabel produces a human-readable label for the S2F result.
func s2fLabel(s2f StockToFlowResult) string {
	if s2f.Score < 33 {
		return "undervalued"
	}
	if s2f.Score > 66 {
		return "overvalued"
	}
	return "fair"
}
