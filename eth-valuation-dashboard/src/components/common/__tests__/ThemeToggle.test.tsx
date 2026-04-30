import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeToggle } from '../ThemeToggle';
import { useDashboardStore } from '../../../store/dashboard';

vi.mock('../../../api/client', () => ({
  forceRefresh: vi.fn(),
  getValuation: vi.fn().mockResolvedValue({ data: { score: 50 }, meta: { lastUpdated: 0 } }),
  createAlertRule: vi.fn(),
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

vi.mock('../../../api/polling', () => {
  const PollingService = vi.fn().mockImplementation(() => ({
    start: vi.fn(),
    stop: vi.fn(),
    isRunning: vi.fn().mockReturnValue(false),
  }));
  return { PollingService, POLLING_INTERVALS: { price: 10000 } };
});

describe('ThemeToggle', () => {
  beforeEach(() => {
    useDashboardStore.setState({ theme: 'dark' });
  });

  it('renders a button', () => {
    render(<ThemeToggle />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('clicking toggles the theme from dark to light', () => {
    render(<ThemeToggle />);
    const button = screen.getByRole('button');

    expect(useDashboardStore.getState().theme).toBe('dark');
    fireEvent.click(button);
    expect(useDashboardStore.getState().theme).toBe('light');
  });

  it('displays correct aria-label based on current theme', () => {
    render(<ThemeToggle />);
    expect(screen.getByLabelText('Switch to light mode')).toBeInTheDocument();
  });
});
