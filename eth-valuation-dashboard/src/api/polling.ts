import {
  getMarketData,
  getBurnData,
  getGasData,
  getActivityData,
  getTVLData,
  getSupplyData,
  getETFData,
  getGrayscaleData,
  getInstitutionalHoldings,
  getStakingData,
  getNetworkPerformance,
  getETHBTCData,
  getMacroIndicators,
  getActiveAlerts,
} from './client';
import type {
  MarketData,
  BurnData,
  GasData,
  ActivityData,
  TVLData,
  SupplyData,
  ETFData,
  GrayscaleData,
  InstitutionalHoldings,
  StakingData,
  NetworkPerformance,
  ETHBTCData,
  MacroIndicators,
  Alert,
} from './types';

/** Polling intervals in milliseconds */
export const POLLING_INTERVALS = {
  price: 10_000,
  onChain: 300_000,
  institutional: 3_600_000,
  network: 300_000,
  macro: 3_600_000,
  alerts: 30_000,
} as const;

export type PollingGroup = keyof typeof POLLING_INTERVALS;

export interface PollingCallbacks {
  onMarketData?: (data: MarketData, lastUpdated: number) => void;
  onBurnData?: (data: BurnData, lastUpdated: number) => void;
  onGasData?: (data: GasData, lastUpdated: number) => void;
  onActivityData?: (data: ActivityData, lastUpdated: number) => void;
  onTVLData?: (data: TVLData, lastUpdated: number) => void;
  onSupplyData?: (data: SupplyData, lastUpdated: number) => void;
  onETFData?: (data: ETFData, lastUpdated: number) => void;
  onGrayscaleData?: (data: GrayscaleData, lastUpdated: number) => void;
  onInstitutionalHoldings?: (data: InstitutionalHoldings, lastUpdated: number) => void;
  onStakingData?: (data: StakingData, lastUpdated: number) => void;
  onNetworkPerformance?: (data: NetworkPerformance, lastUpdated: number) => void;
  onETHBTCData?: (data: ETHBTCData, lastUpdated: number) => void;
  onMacroIndicators?: (data: MacroIndicators, lastUpdated: number) => void;
  onAlerts?: (data: Alert[], lastUpdated: number) => void;
  onError?: (group: PollingGroup, error: unknown) => void;
}

/**
 * PollingService manages multiple polling intervals for different data groups.
 * On failure, it retains the last successful data and records the lastUpdated timestamp.
 */
export class PollingService {
  private timers: Map<PollingGroup, ReturnType<typeof setInterval>> = new Map();
  private callbacks: PollingCallbacks = {};
  private running = false;

  /** Start polling all data groups with the provided callbacks. */
  start(callbacks: PollingCallbacks): void {
    if (this.running) {
      this.stop();
    }
    this.callbacks = callbacks;
    this.running = true;

    // Immediately fetch all groups, then set up intervals
    this.pollPrice();
    this.pollOnChain();
    this.pollInstitutional();
    this.pollNetwork();
    this.pollMacro();
    this.pollAlerts();

    this.timers.set('price', setInterval(() => this.pollPrice(), POLLING_INTERVALS.price));
    this.timers.set('onChain', setInterval(() => this.pollOnChain(), POLLING_INTERVALS.onChain));
    this.timers.set('institutional', setInterval(() => this.pollInstitutional(), POLLING_INTERVALS.institutional));
    this.timers.set('network', setInterval(() => this.pollNetwork(), POLLING_INTERVALS.network));
    this.timers.set('macro', setInterval(() => this.pollMacro(), POLLING_INTERVALS.macro));
    this.timers.set('alerts', setInterval(() => this.pollAlerts(), POLLING_INTERVALS.alerts));
  }

  /** Stop all polling intervals. */
  stop(): void {
    for (const timer of this.timers.values()) {
      clearInterval(timer);
    }
    this.timers.clear();
    this.running = false;
  }

  /** Check if the service is currently running. */
  isRunning(): boolean {
    return this.running;
  }

  // --- Private polling methods ---

  private async pollPrice(): Promise<void> {
    try {
      const response = await getMarketData();
      this.callbacks.onMarketData?.(response.data, response.meta.lastUpdated);
    } catch (error) {
      this.callbacks.onError?.('price', error);
    }
  }

  private async pollOnChain(): Promise<void> {
    const group: PollingGroup = 'onChain';
    try {
      const [burn, gas, activity, tvl, supply] = await Promise.allSettled([
        getBurnData(),
        getGasData(),
        getActivityData(),
        getTVLData(),
        getSupplyData(),
      ]);

      if (burn.status === 'fulfilled') {
        this.callbacks.onBurnData?.(burn.value.data, burn.value.meta.lastUpdated);
      }
      if (gas.status === 'fulfilled') {
        this.callbacks.onGasData?.(gas.value.data, gas.value.meta.lastUpdated);
      }
      if (activity.status === 'fulfilled') {
        this.callbacks.onActivityData?.(activity.value.data, activity.value.meta.lastUpdated);
      }
      if (tvl.status === 'fulfilled') {
        this.callbacks.onTVLData?.(tvl.value.data, tvl.value.meta.lastUpdated);
      }
      if (supply.status === 'fulfilled') {
        this.callbacks.onSupplyData?.(supply.value.data, supply.value.meta.lastUpdated);
      }

      // Report errors for any failed requests
      const results = [burn, gas, activity, tvl, supply];
      const hasFailure = results.some((r) => r.status === 'rejected');
      if (hasFailure) {
        const firstError = results.find((r) => r.status === 'rejected') as PromiseRejectedResult;
        this.callbacks.onError?.(group, firstError.reason);
      }
    } catch (error) {
      this.callbacks.onError?.(group, error);
    }
  }

  private async pollInstitutional(): Promise<void> {
    const group: PollingGroup = 'institutional';
    try {
      const [etf, grayscale, holdings] = await Promise.allSettled([
        getETFData(),
        getGrayscaleData(),
        getInstitutionalHoldings(),
      ]);

      if (etf.status === 'fulfilled') {
        this.callbacks.onETFData?.(etf.value.data, etf.value.meta.lastUpdated);
      }
      if (grayscale.status === 'fulfilled') {
        this.callbacks.onGrayscaleData?.(grayscale.value.data, grayscale.value.meta.lastUpdated);
      }
      if (holdings.status === 'fulfilled') {
        this.callbacks.onInstitutionalHoldings?.(holdings.value.data, holdings.value.meta.lastUpdated);
      }

      const results = [etf, grayscale, holdings];
      const hasFailure = results.some((r) => r.status === 'rejected');
      if (hasFailure) {
        const firstError = results.find((r) => r.status === 'rejected') as PromiseRejectedResult;
        this.callbacks.onError?.(group, firstError.reason);
      }
    } catch (error) {
      this.callbacks.onError?.(group, error);
    }
  }

  private async pollNetwork(): Promise<void> {
    const group: PollingGroup = 'network';
    try {
      const [staking, performance] = await Promise.allSettled([
        getStakingData(),
        getNetworkPerformance(),
      ]);

      if (staking.status === 'fulfilled') {
        this.callbacks.onStakingData?.(staking.value.data, staking.value.meta.lastUpdated);
      }
      if (performance.status === 'fulfilled') {
        this.callbacks.onNetworkPerformance?.(performance.value.data, performance.value.meta.lastUpdated);
      }

      const results = [staking, performance];
      const hasFailure = results.some((r) => r.status === 'rejected');
      if (hasFailure) {
        const firstError = results.find((r) => r.status === 'rejected') as PromiseRejectedResult;
        this.callbacks.onError?.(group, firstError.reason);
      }
    } catch (error) {
      this.callbacks.onError?.(group, error);
    }
  }

  private async pollMacro(): Promise<void> {
    const group: PollingGroup = 'macro';
    try {
      const [ethbtc, indicators] = await Promise.allSettled([
        getETHBTCData(),
        getMacroIndicators(),
      ]);

      if (ethbtc.status === 'fulfilled') {
        this.callbacks.onETHBTCData?.(ethbtc.value.data, ethbtc.value.meta.lastUpdated);
      }
      if (indicators.status === 'fulfilled') {
        this.callbacks.onMacroIndicators?.(indicators.value.data, indicators.value.meta.lastUpdated);
      }

      const results = [ethbtc, indicators];
      const hasFailure = results.some((r) => r.status === 'rejected');
      if (hasFailure) {
        const firstError = results.find((r) => r.status === 'rejected') as PromiseRejectedResult;
        this.callbacks.onError?.(group, firstError.reason);
      }
    } catch (error) {
      this.callbacks.onError?.(group, error);
    }
  }

  private async pollAlerts(): Promise<void> {
    try {
      const response = await getActiveAlerts();
      this.callbacks.onAlerts?.(response.data, response.meta.lastUpdated);
    } catch (error) {
      this.callbacks.onError?.('alerts', error);
    }
  }
}
