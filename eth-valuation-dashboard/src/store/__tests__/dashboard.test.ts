import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useDashboardStore } from '../dashboard';

vi.mock('../../api/client', () => ({
  forceRefresh: vi.fn().mockResolvedValue({ data: null, meta: { lastUpdated: 0 } }),
  getValuation: vi.fn().mockResolvedValue({ data: { score: 50 }, meta: { lastUpdated: 0 } }),
  createAlertRule: vi.fn().mockResolvedValue({ data: {}, meta: { lastUpdated: 0 } }),
  getMarketData: vi.fn().mockRejectedValue(new Error('mock')),
  getBurnData: vi.fn().mockRejectedValue(new Error('mock')),
  getGasData: vi.fn().mockRejectedValue(new Error('mock')),
  getActivityData: vi.fn().mockRejectedValue(new Error('mock')),
  getTVLData: vi.fn().mockRejectedValue(new Error('mock')),
  getSupplyData: vi.fn().mockRejectedValue(new Error('mock')),
  getETFData: vi.fn().mockRejectedValue(new Error('mock')),
  getGrayscaleData: vi.fn().mockRejectedValue(new Error('mock')),
  getInstitutionalHoldings: vi.fn().mockRejectedValue(new Error('mock')),
  getStakingData: vi.fn().mockRejectedValue(new Error('mock')),
  getNetworkPerformance: vi.fn().mockRejectedValue(new Error('mock')),
  getETHBTCData: vi.fn().mockRejectedValue(new Error('mock')),
  getMacroIndicators: vi.fn().mockRejectedValue(new Error('mock')),
  getActiveAlerts: vi.fn().mockRejectedValue(new Error('mock')),
}));

vi.mock('../../api/polling', () => {
  const PollingService = vi.fn().mockImplementation(() => ({
    start: vi.fn(),
    stop: vi.fn(),
    isRunning: vi.fn().mockReturnValue(false),
  }));
  return { PollingService, POLLING_INTERVALS: { price: 10000 } };
});

describe('DashboardStore', () => {
  beforeEach(() => {
    // Reset store state between tests
    useDashboardStore.setState({
      burnData: null,
      gasData: null,
      activityData: null,
      tvlData: null,
      supplyData: null,
      marketData: null,
      etfData: null,
      grayscaleData: null,
      institutionalHoldings: null,
      stakingData: null,
      networkPerformance: null,
      ethbtcData: null,
      macroIndicators: null,
      valuationScore: null,
      theme: 'dark',
      activeAlerts: [],
      isLoading: {},
      lastUpdated: {},
      errors: {},
    });
  });

  it('has correct initial state', () => {
    const state = useDashboardStore.getState();

    expect(state.burnData).toBeNull();
    expect(state.gasData).toBeNull();
    expect(state.activityData).toBeNull();
    expect(state.tvlData).toBeNull();
    expect(state.supplyData).toBeNull();
    expect(state.marketData).toBeNull();
    expect(state.etfData).toBeNull();
    expect(state.grayscaleData).toBeNull();
    expect(state.institutionalHoldings).toBeNull();
    expect(state.stakingData).toBeNull();
    expect(state.networkPerformance).toBeNull();
    expect(state.ethbtcData).toBeNull();
    expect(state.macroIndicators).toBeNull();
    expect(state.valuationScore).toBeNull();
    expect(state.theme).toBe('dark');
    expect(state.activeAlerts).toEqual([]);
    expect(state.isLoading).toEqual({});
    expect(state.lastUpdated).toEqual({});
    expect(state.errors).toEqual({});
  });

  it('toggleTheme switches from dark to light', () => {
    useDashboardStore.getState().toggleTheme();
    expect(useDashboardStore.getState().theme).toBe('light');
  });

  it('toggleTheme switches from light back to dark', () => {
    useDashboardStore.getState().toggleTheme();
    expect(useDashboardStore.getState().theme).toBe('light');
    useDashboardStore.getState().toggleTheme();
    expect(useDashboardStore.getState().theme).toBe('dark');
  });

  it('state updates correctly via setState', () => {
    useDashboardStore.setState({ theme: 'light' });
    expect(useDashboardStore.getState().theme).toBe('light');

    useDashboardStore.setState({ errors: { price: 'Network error' } });
    expect(useDashboardStore.getState().errors).toEqual({ price: 'Network error' });
  });
});
