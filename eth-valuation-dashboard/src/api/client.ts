import axios from 'axios';
import type {
  APIResponse,
  OverviewData,
  BurnData,
  GasData,
  ActivityData,
  TVLData,
  SupplyData,
  MarketData,
  OHLCVPoint,
  ValuationScore,
  DCFResult,
  DCFAssumptions,
  DistributionData,
  ETFData,
  GrayscaleData,
  InstitutionalHoldings,
  StakingData,
  NetworkPerformance,
  ETHBTCData,
  MacroIndicators,
  Alert,
  AlertRule,
  ExportChartRequest,
  ExportCSVRequest,
  ShareRequest,
  ExportResponse,
  ShareResponse,
  TimeRange,
} from './types';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8888/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// --- Overview ---

export async function getOverview(): Promise<APIResponse<OverviewData>> {
  const { data } = await apiClient.get<APIResponse<OverviewData>>('/overview');
  return data;
}

// --- On-Chain Module ---

export async function getBurnData(): Promise<APIResponse<BurnData>> {
  const { data } = await apiClient.get<APIResponse<BurnData>>('/onchain/burn');
  return data;
}

export async function getGasData(): Promise<APIResponse<GasData>> {
  const { data } = await apiClient.get<APIResponse<GasData>>('/onchain/gas');
  return data;
}

export async function getActivityData(): Promise<APIResponse<ActivityData>> {
  const { data } = await apiClient.get<APIResponse<ActivityData>>('/onchain/activity');
  return data;
}

export async function getTVLData(): Promise<APIResponse<TVLData>> {
  const { data } = await apiClient.get<APIResponse<TVLData>>('/onchain/tvl');
  return data;
}

export async function getSupplyData(): Promise<APIResponse<SupplyData>> {
  const { data } = await apiClient.get<APIResponse<SupplyData>>('/onchain/supply');
  return data;
}

// --- Market Module ---

export async function getMarketData(): Promise<APIResponse<MarketData>> {
  const { data } = await apiClient.get<APIResponse<MarketData>>('/market');
  return data;
}

export async function getPriceHistory(timeRange: TimeRange): Promise<APIResponse<OHLCVPoint[]>> {
  const { data } = await apiClient.get<APIResponse<OHLCVPoint[]>>('/market/price-history', {
    params: { timeRange },
  });
  return data;
}

// --- Valuation Module ---

export async function getValuation(): Promise<APIResponse<ValuationScore>> {
  const { data } = await apiClient.get<APIResponse<ValuationScore>>('/valuation');
  return data;
}

export async function getDCFValuation(params: Partial<DCFAssumptions>): Promise<APIResponse<DCFResult>> {
  const { data } = await apiClient.get<APIResponse<DCFResult>>('/valuation/dcf', {
    params,
  });
  return data;
}

export async function getDistribution(metric: string): Promise<APIResponse<DistributionData>> {
  const { data } = await apiClient.get<APIResponse<DistributionData>>(`/valuation/distribution/${metric}`);
  return data;
}

// --- Institutional Module ---

export async function getETFData(): Promise<APIResponse<ETFData>> {
  const { data } = await apiClient.get<APIResponse<ETFData>>('/institutional/etf');
  return data;
}

export async function getGrayscaleData(): Promise<APIResponse<GrayscaleData>> {
  const { data } = await apiClient.get<APIResponse<GrayscaleData>>('/institutional/grayscale');
  return data;
}

export async function getInstitutionalHoldings(): Promise<APIResponse<InstitutionalHoldings>> {
  const { data } = await apiClient.get<APIResponse<InstitutionalHoldings>>('/institutional/holdings');
  return data;
}

// --- Network Module ---

export async function getStakingData(): Promise<APIResponse<StakingData>> {
  const { data } = await apiClient.get<APIResponse<StakingData>>('/network/staking');
  return data;
}

export async function getNetworkPerformance(): Promise<APIResponse<NetworkPerformance>> {
  const { data } = await apiClient.get<APIResponse<NetworkPerformance>>('/network/performance');
  return data;
}

// --- Macro Module ---

export async function getETHBTCData(): Promise<APIResponse<ETHBTCData>> {
  const { data } = await apiClient.get<APIResponse<ETHBTCData>>('/macro/ethbtc');
  return data;
}

export async function getMacroIndicators(): Promise<APIResponse<MacroIndicators>> {
  const { data } = await apiClient.get<APIResponse<MacroIndicators>>('/macro/indicators');
  return data;
}

// --- Alerts Module ---

export async function getActiveAlerts(): Promise<APIResponse<Alert[]>> {
  const { data } = await apiClient.get<APIResponse<Alert[]>>('/alerts');
  return data;
}

export async function getAlertHistory(days: number): Promise<APIResponse<Alert[]>> {
  const { data } = await apiClient.get<APIResponse<Alert[]>>('/alerts/history', {
    params: { days },
  });
  return data;
}

export async function createAlertRule(rule: Omit<AlertRule, 'id'>): Promise<APIResponse<AlertRule>> {
  const { data } = await apiClient.post<APIResponse<AlertRule>>('/alerts/rules', rule);
  return data;
}

export async function updateAlertRule(id: string, rule: Partial<AlertRule>): Promise<APIResponse<AlertRule>> {
  const { data } = await apiClient.put<APIResponse<AlertRule>>(`/alerts/rules/${id}`, rule);
  return data;
}

export async function deleteAlertRule(id: string): Promise<APIResponse<null>> {
  const { data } = await apiClient.delete<APIResponse<null>>(`/alerts/rules/${id}`);
  return data;
}

// --- Export & Share ---

export async function exportCSV(request: ExportCSVRequest): Promise<APIResponse<ExportResponse>> {
  const { data } = await apiClient.post<APIResponse<ExportResponse>>('/export/csv', request);
  return data;
}

export async function exportChart(request: ExportChartRequest): Promise<APIResponse<ExportResponse>> {
  const { data } = await apiClient.post<APIResponse<ExportResponse>>('/export/chart', request);
  return data;
}

export async function generateShareLink(request: ShareRequest): Promise<APIResponse<ShareResponse>> {
  const { data } = await apiClient.post<APIResponse<ShareResponse>>('/share', request);
  return data;
}

// --- Utility ---

export async function forceRefresh(): Promise<APIResponse<null>> {
  const { data } = await apiClient.post<APIResponse<null>>('/refresh');
  return data;
}

export { apiClient };
