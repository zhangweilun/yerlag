import { create } from 'zustand';
import type {
  BurnData,
  GasData,
  ActivityData,
  TVLData,
  SupplyData,
  MarketData,
  ETFData,
  GrayscaleData,
  InstitutionalHoldings,
  StakingData,
  NetworkPerformance,
  ETHBTCData,
  MacroIndicators,
  ValuationScore,
  Alert,
  AlertRule,
} from '../api/types';
import { forceRefresh, getValuation, createAlertRule } from '../api/client';
import { PollingService } from '../api/polling';
import type { PollingGroup } from '../api/polling';

export type Theme = 'light' | 'dark';

export interface DashboardStore {
  // Data state
  burnData: BurnData | null;
  gasData: GasData | null;
  activityData: ActivityData | null;
  tvlData: TVLData | null;
  supplyData: SupplyData | null;
  marketData: MarketData | null;
  etfData: ETFData | null;
  grayscaleData: GrayscaleData | null;
  institutionalHoldings: InstitutionalHoldings | null;
  stakingData: StakingData | null;
  networkPerformance: NetworkPerformance | null;
  ethbtcData: ETHBTCData | null;
  macroIndicators: MacroIndicators | null;
  valuationScore: ValuationScore | null;

  // UI state
  theme: Theme;
  activeAlerts: Alert[];
  isLoading: Record<string, boolean>;
  lastUpdated: Record<string, number>;
  errors: Record<string, string>;

  // Actions
  refreshAll: () => Promise<void>;
  refreshModule: (module: PollingGroup) => Promise<void>;
  toggleTheme: () => void;
  setAlertRule: (rule: Omit<AlertRule, 'id'>) => Promise<void>;
  startPolling: () => void;
  stopPolling: () => void;
}

let pollingService: PollingService | null = null;

export const useDashboardStore = create<DashboardStore>((set, get) => ({
  // Data state initial values
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

  // UI state initial values
  theme: 'dark',
  activeAlerts: [],
  isLoading: {},
  lastUpdated: {},
  errors: {},

  // Actions
  refreshAll: async () => {
    set((state) => ({ isLoading: { ...state.isLoading, all: true } }));
    try {
      await forceRefresh();
      // After force refresh, restart polling to get fresh data
      const { stopPolling, startPolling } = get();
      stopPolling();
      startPolling();
      set((state) => {
        const errors = { ...state.errors };
        delete errors['all'];
        return { isLoading: { ...state.isLoading, all: false }, errors };
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to refresh';
      set((state) => ({
        isLoading: { ...state.isLoading, all: false },
        errors: { ...state.errors, all: message },
      }));
    }
  },

  refreshModule: async (module: PollingGroup) => {
    set((state) => ({ isLoading: { ...state.isLoading, [module]: true } }));
    try {
      // Stop and restart polling to trigger immediate re-fetch
      const { stopPolling, startPolling } = get();
      stopPolling();
      startPolling();
      set((state) => {
        const errors = { ...state.errors };
        delete errors[module];
        return { isLoading: { ...state.isLoading, [module]: false }, errors };
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : `Failed to refresh ${module}`;
      set((state) => ({
        isLoading: { ...state.isLoading, [module]: false },
        errors: { ...state.errors, [module]: message },
      }));
    }
  },

  toggleTheme: () => {
    set((state) => ({ theme: state.theme === 'light' ? 'dark' : 'light' }));
  },

  setAlertRule: async (rule: Omit<AlertRule, 'id'>) => {
    set((state) => ({ isLoading: { ...state.isLoading, alertRule: true } }));
    try {
      await createAlertRule(rule);
      set((state) => {
        const errors = { ...state.errors };
        delete errors['alertRule'];
        return { isLoading: { ...state.isLoading, alertRule: false }, errors };
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create alert rule';
      set((state) => ({
        isLoading: { ...state.isLoading, alertRule: false },
        errors: { ...state.errors, alertRule: message },
      }));
    }
  },

  startPolling: () => {
    if (pollingService?.isRunning()) {
      return;
    }
    pollingService = new PollingService();
    pollingService.start({
      onMarketData: (data, lastUpdated) => {
        set((state) => ({
          marketData: data,
          lastUpdated: { ...state.lastUpdated, market: lastUpdated },
        }));
      },
      onBurnData: (data, lastUpdated) => {
        set((state) => ({
          burnData: data,
          lastUpdated: { ...state.lastUpdated, burn: lastUpdated },
        }));
      },
      onGasData: (data, lastUpdated) => {
        set((state) => ({
          gasData: data,
          lastUpdated: { ...state.lastUpdated, gas: lastUpdated },
        }));
      },
      onActivityData: (data, lastUpdated) => {
        set((state) => ({
          activityData: data,
          lastUpdated: { ...state.lastUpdated, activity: lastUpdated },
        }));
      },
      onTVLData: (data, lastUpdated) => {
        set((state) => ({
          tvlData: data,
          lastUpdated: { ...state.lastUpdated, tvl: lastUpdated },
        }));
      },
      onSupplyData: (data, lastUpdated) => {
        set((state) => ({
          supplyData: data,
          lastUpdated: { ...state.lastUpdated, supply: lastUpdated },
        }));
      },
      onETFData: (data, lastUpdated) => {
        set((state) => ({
          etfData: data,
          lastUpdated: { ...state.lastUpdated, etf: lastUpdated },
        }));
      },
      onGrayscaleData: (data, lastUpdated) => {
        set((state) => ({
          grayscaleData: data,
          lastUpdated: { ...state.lastUpdated, grayscale: lastUpdated },
        }));
      },
      onInstitutionalHoldings: (data, lastUpdated) => {
        set((state) => ({
          institutionalHoldings: data,
          lastUpdated: { ...state.lastUpdated, institutionalHoldings: lastUpdated },
        }));
      },
      onStakingData: (data, lastUpdated) => {
        set((state) => ({
          stakingData: data,
          lastUpdated: { ...state.lastUpdated, staking: lastUpdated },
        }));
      },
      onNetworkPerformance: (data, lastUpdated) => {
        set((state) => ({
          networkPerformance: data,
          lastUpdated: { ...state.lastUpdated, networkPerformance: lastUpdated },
        }));
      },
      onETHBTCData: (data, lastUpdated) => {
        set((state) => ({
          ethbtcData: data,
          lastUpdated: { ...state.lastUpdated, ethbtc: lastUpdated },
        }));
      },
      onMacroIndicators: (data, lastUpdated) => {
        set((state) => ({
          macroIndicators: data,
          lastUpdated: { ...state.lastUpdated, macroIndicators: lastUpdated },
        }));
      },
      onAlerts: (data, lastUpdated) => {
        set((state) => ({
          activeAlerts: data,
          lastUpdated: { ...state.lastUpdated, alerts: lastUpdated },
        }));
      },
      onError: (group: PollingGroup, error: unknown) => {
        const message = error instanceof Error ? error.message : `Polling error: ${group}`;
        set((state) => ({
          errors: { ...state.errors, [group]: message },
        }));
      },
    });

    // Also fetch valuation score on start
    void getValuation().then((response) => {
      set((state) => ({
        valuationScore: response.data,
        lastUpdated: { ...state.lastUpdated, valuation: response.meta.lastUpdated },
      }));
    }).catch((error: unknown) => {
      const message = error instanceof Error ? error.message : 'Failed to fetch valuation';
      set((state) => ({
        errors: { ...state.errors, valuation: message },
      }));
    });
  },

  stopPolling: () => {
    if (pollingService) {
      pollingService.stop();
      pollingService = null;
    }
  },
}));
