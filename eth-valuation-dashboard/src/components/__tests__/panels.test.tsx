import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { useDashboardStore } from '../../store/dashboard';
import { BurnDataPanel } from '../onchain/BurnDataPanel';
import { MarketOverview } from '../market/MarketOverview';
import { ValuationScore } from '../valuation/ValuationScore';
import { StakingPanel } from '../network/StakingPanel';
import type {
  BurnData,
  MarketData,
  ValuationScore as ValuationScoreType,
  StakingData,
} from '../../api/types';

// Mock recharts ResponsiveContainer (same pattern as charts.test.tsx)
vi.mock('recharts', async () => {
  const actual = await vi.importActual<typeof import('recharts')>('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="responsive-container">{children}</div>
    ),
  };
});

// Mock API client and polling to prevent real network calls
vi.mock('../../api/client', () => ({
  forceRefresh: vi.fn(),
  getValuation: vi.fn().mockResolvedValue({ data: null, meta: { lastUpdated: 0 } }),
  createAlertRule: vi.fn(),
  apiClient: { get: vi.fn(), post: vi.fn() },
}));

vi.mock('../../api/polling', () => ({
  PollingService: vi.fn().mockImplementation(() => ({
    start: vi.fn(),
    stop: vi.fn(),
    isRunning: vi.fn().mockReturnValue(false),
  })),
}));

beforeEach(() => {
  useDashboardStore.setState({
    burnData: null,
    supplyData: null,
    marketData: null,
    valuationScore: null,
    stakingData: null,
  });
});

/**
 * Validates: Requirements 1.1, 1.2, 6.4
 */
describe('BurnDataPanel', () => {
  it('renders loading state when data is null', () => {
    render(<BurnDataPanel />);
    expect(screen.getByText('加载中...')).toBeInTheDocument();
    expect(screen.getByText('EIP-1559 销毁数据')).toBeInTheDocument();
  });

  it('renders stat values when data is present', () => {
    const burnData: BurnData = {
      daily: 1500,
      weekly: 10500,
      monthly: 45000,
      cumulative: 4200000,
      annualizedBurnRate: 0.0312,
      dailyHistory: [{ timestamp: 1700000000, value: 1500 }],
    };
    useDashboardStore.setState({ burnData });

    render(<BurnDataPanel />);
    expect(screen.getByText('1.50K ETH')).toBeInTheDocument();
    expect(screen.getByText('10.50K ETH')).toBeInTheDocument();
    expect(screen.getByText('45.00K ETH')).toBeInTheDocument();
    expect(screen.getByText('4.20M ETH')).toBeInTheDocument();
    expect(screen.getByText('3.12%')).toBeInTheDocument();
  });
});

describe('MarketOverview', () => {
  it('renders loading state when data is null', () => {
    render(<MarketOverview />);
    expect(screen.getByText('加载中...')).toBeInTheDocument();
    expect(screen.getByText('市场总览')).toBeInTheDocument();
  });

  it('renders price and change when data is present', () => {
    const marketData: MarketData = {
      currentPrice: 3250.5,
      priceChange24h: 4.25,
      volume24h: 12500000000,
      volume7d: 85000000000,
      circulatingMarketCap: 390000000000,
      fullyDilutedMarketCap: 395000000000,
      marketCapRank: 2,
      athPrice: 4878,
      athDrawdown: 33.4,
      exchangePrices: [],
    };
    useDashboardStore.setState({ marketData });

    render(<MarketOverview />);
    expect(screen.getByText('$3.25K')).toBeInTheDocument();
    expect(screen.getByText('+4.25%')).toBeInTheDocument();
    expect(screen.getByText('$12.50B')).toBeInTheDocument();
  });
});

describe('ValuationScore', () => {
  it('renders loading state when data is null', () => {
    render(<ValuationScore />);
    expect(screen.getByText('加载中...')).toBeInTheDocument();
    expect(screen.getByText('综合估值评分')).toBeInTheDocument();
  });

  it('renders score when data is present', () => {
    const valuationScore: ValuationScoreType = {
      overall: 62.5,
      status: 'undervalued',
      breakdown: {
        mvrvScore: { ratio: 1.2, historicalPercentile: 35, signal: 'undervalued', history: [] },
        priceToFeeScore: { ratio: 50, historicalPercentile: 40, signal: 'undervalued' },
        dcfValuation: {
          fairValueLow: 3000,
          fairValueMid: 4500,
          fairValueHigh: 6000,
          assumptions: { discountRate: 0.1, growthRate: 0.15, terminalGrowthRate: 0.03, projectionYears: 5 },
        },
        stockToFlowScore: { ratio: 20, modelPrice: 5000, deviation: -0.3 },
        nvtScore: { ratio: 45, historicalPercentile: 30, signal: 'undervalued' },
        ethBtcScore: { ratio: 0.05, historicalPercentile: 25, signal: 'undervalued' },
      },
      radarData: [
        { dimension: 'mvrv', score: 65, label: 'MVRV' },
        { dimension: 'nvt', score: 70, label: 'NVT' },
      ],
    };
    useDashboardStore.setState({ valuationScore });

    render(<ValuationScore />);
    expect(screen.getByText('63')).toBeInTheDocument();
    expect(screen.getByText('低估')).toBeInTheDocument();
  });
});

describe('StakingPanel', () => {
  it('renders loading state when data is null', () => {
    render(<StakingPanel />);
    expect(screen.getByText('加载中...')).toBeInTheDocument();
    expect(screen.getByText('质押数据')).toBeInTheDocument();
  });

  it('renders staking stats when data is present', () => {
    const stakingData: StakingData = {
      totalStakedEth: 32000000,
      stakingPercentage: 26.5,
      activeValidators: 950000,
      stakingYield: 3.8,
      entryQueueLength: 1200,
      exitQueueLength: 450,
      entryWaitTime: '~2d',
      exitWaitTime: '~1d',
      liquidStakingShares: [
        { protocol: 'Lido', share: 32, stakedEth: 10000000 },
        { protocol: 'Rocket Pool', share: 8, stakedEth: 2500000 },
      ],
      validatorHistory: [{ timestamp: 1700000000, value: 900000 }],
      yieldHistory: [{ timestamp: 1700000000, value: 3.8 }],
    };
    useDashboardStore.setState({ stakingData });

    render(<StakingPanel />);
    expect(screen.getByText('32.00M ETH')).toBeInTheDocument();
    expect(screen.getByText('26.50%')).toBeInTheDocument();
    expect(screen.getByText('3.80%')).toBeInTheDocument();
  });
});
