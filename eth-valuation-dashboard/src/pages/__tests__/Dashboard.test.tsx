import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { useDashboardStore } from '../../store/dashboard';
import { Dashboard } from '../Dashboard';

// Mock recharts ResponsiveContainer (same pattern as other tests)
vi.mock('recharts', async () => {
  const actual = await vi.importActual<typeof import('recharts')>('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="responsive-container">{children}</div>
    ),
  };
});

// Mock API client to prevent real network calls
vi.mock('../../api/client', () => ({
  forceRefresh: vi.fn().mockResolvedValue({ data: null, meta: { lastUpdated: 0 } }),
  getValuation: vi.fn().mockResolvedValue({ data: null, meta: { lastUpdated: 0 } }),
  getPriceHistory: vi.fn().mockResolvedValue({ data: [], meta: { lastUpdated: 0 } }),
  createAlertRule: vi.fn(),
  apiClient: { get: vi.fn(), post: vi.fn() },
}));

// Mock polling service to prevent real intervals
vi.mock('../../api/polling', () => {
  return {
    PollingService: class MockPollingService {
      start = vi.fn();
      stop = vi.fn();
      isRunning = vi.fn().mockReturnValue(false);
    },
  };
});

beforeEach(() => {
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
    activeAlerts: [],
    isLoading: {},
    errors: {},
  });
});

/**
 * Validates: Requirements 1.4, 14.3, 14.4
 */
describe('Dashboard Integration', () => {
  it('renders the title and summary bar', () => {
    render(<Dashboard />);

    expect(screen.getByText('ETH Valuation Dashboard')).toBeInTheDocument();
    expect(screen.getByText('ETH Price')).toBeInTheDocument();
    expect(screen.getByText('24h Change')).toBeInTheDocument();
    expect(screen.getByText('Market Cap Rank')).toBeInTheDocument();
    expect(screen.getByText('Valuation Score')).toBeInTheDocument();
  });

  it('renders all section titles', () => {
    render(<Dashboard />);

    expect(screen.getByText('On-Chain Data')).toBeInTheDocument();
    expect(screen.getByText('Market Data')).toBeInTheDocument();
    expect(screen.getByText('Valuation')).toBeInTheDocument();
    expect(screen.getByText('Institutional Data')).toBeInTheDocument();
    expect(screen.getByText('Network Health')).toBeInTheDocument();
    expect(screen.getByText('Macro Economy')).toBeInTheDocument();
  });

  it('renders the refresh button', () => {
    render(<Dashboard />);

    const refreshButton = screen.getByRole('button', { name: 'Refresh all data' });
    expect(refreshButton).toBeInTheDocument();
    expect(refreshButton).toHaveTextContent('↻ Refresh');
  });

  it('clicking refresh button triggers refreshAll', async () => {
    const { forceRefresh } = await import('../../api/client');

    render(<Dashboard />);

    const refreshButton = screen.getByRole('button', { name: 'Refresh all data' });
    fireEvent.click(refreshButton);

    expect(forceRefresh).toHaveBeenCalled();
  });
});
