import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { PollingService, POLLING_INTERVALS } from '../polling';

vi.mock('../client', () => ({
  getMarketData: vi.fn().mockRejectedValue(new Error('not called')),
  getBurnData: vi.fn().mockRejectedValue(new Error('not called')),
  getGasData: vi.fn().mockRejectedValue(new Error('not called')),
  getActivityData: vi.fn().mockRejectedValue(new Error('not called')),
  getTVLData: vi.fn().mockRejectedValue(new Error('not called')),
  getSupplyData: vi.fn().mockRejectedValue(new Error('not called')),
  getETFData: vi.fn().mockRejectedValue(new Error('not called')),
  getGrayscaleData: vi.fn().mockRejectedValue(new Error('not called')),
  getInstitutionalHoldings: vi.fn().mockRejectedValue(new Error('not called')),
  getStakingData: vi.fn().mockRejectedValue(new Error('not called')),
  getNetworkPerformance: vi.fn().mockRejectedValue(new Error('not called')),
  getETHBTCData: vi.fn().mockRejectedValue(new Error('not called')),
  getMacroIndicators: vi.fn().mockRejectedValue(new Error('not called')),
  getActiveAlerts: vi.fn().mockRejectedValue(new Error('not called')),
}));

describe('POLLING_INTERVALS', () => {
  it('has correct interval values', () => {
    expect(POLLING_INTERVALS.price).toBe(10_000);
    expect(POLLING_INTERVALS.onChain).toBe(300_000);
    expect(POLLING_INTERVALS.institutional).toBe(3_600_000);
    expect(POLLING_INTERVALS.network).toBe(300_000);
    expect(POLLING_INTERVALS.macro).toBe(3_600_000);
    expect(POLLING_INTERVALS.alerts).toBe(30_000);
  });
});

describe('PollingService', () => {
  let service: PollingService;

  beforeEach(() => {
    vi.useFakeTimers();
    service = new PollingService();
  });

  afterEach(() => {
    service.stop();
    vi.useRealTimers();
  });

  it('isRunning() returns false before start', () => {
    expect(service.isRunning()).toBe(false);
  });

  it('isRunning() returns true after start', () => {
    service.start({});
    expect(service.isRunning()).toBe(true);
  });

  it('isRunning() returns false after stop', () => {
    service.start({});
    service.stop();
    expect(service.isRunning()).toBe(false);
  });

  it('start() creates interval timers', () => {
    const setIntervalSpy = vi.spyOn(globalThis, 'setInterval');
    service.start({});
    // 6 groups: price, onChain, institutional, network, macro, alerts
    expect(setIntervalSpy.mock.calls.length).toBe(6);
    setIntervalSpy.mockRestore();
  });

  it('stop() clears all interval timers', () => {
    const clearIntervalSpy = vi.spyOn(globalThis, 'clearInterval');
    service.start({});
    service.stop();
    expect(clearIntervalSpy.mock.calls.length).toBe(6);
    clearIntervalSpy.mockRestore();
  });

  it('calls onError callback when polling fails', async () => {
    const onError = vi.fn();
    service.start({ onError });

    // Flush microtasks so the immediate poll calls resolve
    await vi.advanceTimersByTimeAsync(0);

    expect(onError).toHaveBeenCalled();
  });
});
