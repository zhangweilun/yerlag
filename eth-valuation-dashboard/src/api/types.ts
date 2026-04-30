// Common types matching backend API response structures

export interface TimeSeriesPoint {
  timestamp: number;
  value: number;
}

export interface Meta {
  lastUpdated: number;
  source: 'live' | 'cache';
  nextRefresh: number;
}

export interface APIResponse<T> {
  code: number;
  message: string;
  data: T;
  meta: Meta;
}

// On-Chain Module Types

export interface BurnData {
  daily: number;
  weekly: number;
  monthly: number;
  cumulative: number;
  annualizedBurnRate: number;
  dailyHistory: TimeSeriesPoint[];
}

export interface GasData {
  currentAvgGwei: number;
  currentAvgUsd: number;
  dailyFeeRevenueEth: number;
  dailyFeeRevenueUsd: number;
  priorityFeeShare: number;
  baseFeeShare: number;
  annualizedRevenue: number;
  priceToFeeRatio: number;
  gasHistory: TimeSeriesPoint[];
  isHighFee: boolean;
}

export interface L2TransactionData {
  network: string;
  dailyTransactions: number;
}

export interface ActivityData {
  dailyActiveAddresses: number;
  daaMovingAvg7d: number;
  dailyTransactions: number;
  dailyNewAddresses: number;
  nvtRatio: number;
  nvtHistoricalMedian: number;
  nvtPercentile: number;
  nvtSignal: 'overvalued' | 'undervalued' | 'neutral';
  l2Comparison: L2TransactionData[];
  transactionHistory: TimeSeriesPoint[];
}

export interface ProtocolTVL {
  name: string;
  tvlUsd: number;
  share: number;
}

export interface TVLData {
  totalTvlUsd: number;
  totalTvlEth: number;
  tvlToMarketCapRatio: number;
  ethTvlDominance: number;
  topProtocols: ProtocolTVL[];
  tvlHistory: TimeSeriesPoint[];
  dominanceHistory: TimeSeriesPoint[];
}

export interface SupplyData {
  totalSupply: number;
  stakedAmount: number;
  defiLocked: number;
  exchangeBalance: number;
  otherAmount: number;
  netIssuance: number;
  isDeflationary: boolean;
  annualInflationRate: number;
  supplyHistory: TimeSeriesPoint[];
  exchangeBalanceHistory: TimeSeriesPoint[];
}

// Market Module Types

export interface ExchangePrice {
  exchange: string;
  price: number;
  spread: number;
}

export interface MarketData {
  currentPrice: number;
  priceChange24h: number;
  volume24h: number;
  volume7d: number;
  circulatingMarketCap: number;
  fullyDilutedMarketCap: number;
  marketCapRank: number;
  athPrice: number;
  athDrawdown: number;
  exchangePrices: ExchangePrice[];
}

export interface OHLCVPoint {
  timestamp: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export type TimeRange = '1d' | '1w' | '1m' | '3m' | '1y' | 'all';

// Valuation Module Types

export interface MVRVResult {
  ratio: number;
  historicalPercentile: number;
  signal: 'overvalued' | 'undervalued' | 'neutral';
  history: TimeSeriesPoint[];
}

export interface PriceToFeeResult {
  ratio: number;
  historicalPercentile: number;
  signal: string;
}

export interface DCFAssumptions {
  discountRate: number;
  growthRate: number;
  terminalGrowthRate: number;
  projectionYears: number;
}

export interface DCFResult {
  fairValueLow: number;
  fairValueMid: number;
  fairValueHigh: number;
  assumptions: DCFAssumptions;
}

export interface StockToFlowResult {
  ratio: number;
  modelPrice: number;
  deviation: number;
}

export interface NVTResult {
  ratio: number;
  historicalPercentile: number;
  signal: string;
}

export interface ETHBTCResult {
  ratio: number;
  historicalPercentile: number;
  signal: string;
}

export interface ValuationBreakdown {
  mvrvScore: MVRVResult;
  priceToFeeScore: PriceToFeeResult;
  dcfValuation: DCFResult;
  stockToFlowScore: StockToFlowResult;
  nvtScore: NVTResult;
  ethBtcScore: ETHBTCResult;
}

export interface RadarDataPoint {
  dimension: string;
  score: number;
  label: string;
}

export interface ValuationScore {
  overall: number;
  status: 'undervalued' | 'fair' | 'overvalued';
  breakdown: ValuationBreakdown;
  radarData: RadarDataPoint[];
}

export interface DistributionData {
  values: number[];
  percentiles: Record<number, number>;
  currentValue: number;
  currentPercentile: number;
}

// Institutional Module Types

export interface ETFHolding {
  issuer: string;
  ticker: string;
  holdingsEth: number;
  holdingsUsd: number;
  dailyNetFlowUsd: number;
  marketShare: number;
  flowHistory: TimeSeriesPoint[];
}

export interface ETFData {
  etfs: ETFHolding[];
  totalHoldingsEth: number;
  totalHoldingsUsd: number;
  holdingsPercentOfSupply: number;
  cumulativeNetFlow: number;
  netFlowHistory: TimeSeriesPoint[];
}

export interface GrayscaleData {
  holdingsEth: number;
  nav: number;
  premiumDiscount: number;
  premiumHistory: TimeSeriesPoint[];
}

export interface InstitutionEntry {
  name: string;
  holdingsEth: number;
  holdingsUsd: number;
}

export interface InstitutionalHoldings {
  institutions: InstitutionEntry[];
  cmeFuturesOI: number;
  cmeFuturesHistory: TimeSeriesPoint[];
}

// Network Module Types

export interface LiquidStakingShare {
  protocol: string;
  share: number;
  stakedEth: number;
}

export interface StakingData {
  totalStakedEth: number;
  stakingPercentage: number;
  activeValidators: number;
  stakingYield: number;
  entryQueueLength: number;
  exitQueueLength: number;
  entryWaitTime: string;
  exitWaitTime: string;
  liquidStakingShares: LiquidStakingShare[];
  validatorHistory: TimeSeriesPoint[];
  yieldHistory: TimeSeriesPoint[];
}

export interface ClientShare {
  client: string;
  share: number;
}

export interface NetworkPerformance {
  avgBlockTime: number;
  blockUtilization: number;
  currentTps: number;
  maxTps: number;
  tpsRatio: number;
  missedSlots24h: number;
  missedSlotsRate: number;
  attestationRate: number;
  clientDiversity: ClientShare[];
}

// Macro Module Types

export interface ETHBTCData {
  ethBtcPrice: number;
  ethBtcHistory: TimeSeriesPoint[];
  correlation30d: number;
  correlation90d: number;
  ethDominance: number;
  ethDominanceHistory: TimeSeriesPoint[];
  ethBtcPercentile: number;
  ethBtcSignal: 'eth_undervalued' | 'eth_overvalued' | 'neutral';
}

export interface RateExpectation {
  date: string;
  expectedRate: number;
  probability: number;
}

export interface MacroIndicators {
  dxyIndex: number;
  dxyHistory: TimeSeriesPoint[];
  treasury10y: number;
  treasury10yHistory: TimeSeriesPoint[];
  nasdaqCorrelation30d: number;
  nasdaqCorrelation90d: number;
  fedFundsRate: number;
  rateExpectations: RateExpectation[];
  fearGreedIndex: number;
  fearGreedHistory: TimeSeriesPoint[];
  stablecoinMarketCap: number;
  stablecoinHistory: TimeSeriesPoint[];
}

// Alert Module Types

export interface AlertCondition {
  type: 'gt' | 'lt' | 'gt_percent_change' | 'lt_percent_change';
  referenceValue?: number;
  referencePeriodDays?: number;
}

export interface AlertRule {
  id: string;
  metricKey: string;
  condition: AlertCondition;
  threshold: number;
  severity: 'high' | 'medium' | 'low';
  enabled: boolean;
  message: string;
}

export interface Alert {
  id: string;
  ruleId: string;
  triggeredAt: number;
  severity: 'high' | 'medium' | 'low';
  title: string;
  message: string;
  metricKey: string;
  currentValue: number;
  thresholdValue: number;
  acknowledged: boolean;
}

// Export & Share Types

export interface ExportChartRequest {
  chartId: string;
  format: 'png' | 'svg';
  width?: number;
  height?: number;
}

export interface ExportCSVRequest {
  module: string;
  timeRange?: TimeRange;
}

export interface ShareRequest {
  dashboardState: string;
  snapshotData?: string;
}

export interface ExportResponse {
  url: string;
  filename: string;
}

export interface ShareResponse {
  shareId: string;
  url: string;
  expiresAt: number;
}

// Overview Type

export interface OverviewData {
  currentPrice: number;
  priceChange24h: number;
  marketCapRank: number;
  valuationScore: number;
  valuationStatus: 'undervalued' | 'fair' | 'overvalued';
  activeAlerts: number;
}
