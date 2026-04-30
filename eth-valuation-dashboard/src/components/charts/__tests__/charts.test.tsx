import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { LineChartComponent } from '../LineChartComponent';
import { BarChartComponent } from '../BarChartComponent';
import { PieChartComponent } from '../PieChartComponent';
import { RadarChartComponent } from '../RadarChartComponent';
import type { TimeSeriesPoint } from '../../../api/types';
import type { PieDataPoint } from '../PieChartComponent';
import type { RadarDataPoint } from '../RadarChartComponent';

// Mock recharts' ResponsiveContainer since it relies on DOM measurements
// that don't work in jsdom (getBoundingClientRect returns 0 width/height)
vi.mock('recharts', async () => {
  const actual = await vi.importActual<typeof import('recharts')>('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="responsive-container">{children}</div>
    ),
  };
});

const sampleTimeSeriesData: TimeSeriesPoint[] = [
  { timestamp: 1700000000, value: 2100 },
  { timestamp: 1700086400, value: 2150 },
  { timestamp: 1700172800, value: 2080 },
];

const emptyTimeSeriesData: TimeSeriesPoint[] = [];

const samplePieData: PieDataPoint[] = [
  { name: 'Lido', value: 32 },
  { name: 'Rocket Pool', value: 8 },
  { name: 'Coinbase', value: 12 },
];

const sampleRadarData: RadarDataPoint[] = [
  { dimension: 'mvrv', score: 65, label: 'MVRV' },
  { dimension: 'nvt', score: 45, label: 'NVT' },
  { dimension: 'pf', score: 70, label: 'P/F Ratio' },
  { dimension: 's2f', score: 55, label: 'S2F' },
  { dimension: 'ethbtc', score: 40, label: 'ETH/BTC' },
];

/**
 * Validates: Requirements 2.2, 6.4, 7.5
 */
describe('LineChartComponent', () => {
  it('renders with title', () => {
    render(<LineChartComponent data={sampleTimeSeriesData} title="Daily Burn" />);
    expect(screen.getByText('Daily Burn')).toBeInTheDocument();
  });

  it('renders without crashing with empty data', () => {
    const { container } = render(
      <LineChartComponent data={emptyTimeSeriesData} title="Empty Chart" />
    );
    expect(container).toBeTruthy();
    expect(screen.getByText('Empty Chart')).toBeInTheDocument();
  });

  it('renders without title when not provided', () => {
    const { container } = render(<LineChartComponent data={sampleTimeSeriesData} />);
    expect(container.querySelector('h4')).toBeNull();
  });

  it('renders the responsive container', () => {
    render(<LineChartComponent data={sampleTimeSeriesData} title="Test" />);
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument();
  });
});

describe('BarChartComponent', () => {
  it('renders with title', () => {
    render(<BarChartComponent data={sampleTimeSeriesData} title="ETF Net Flow" />);
    expect(screen.getByText('ETF Net Flow')).toBeInTheDocument();
  });

  it('renders without crashing with empty data', () => {
    const { container } = render(
      <BarChartComponent data={emptyTimeSeriesData} title="Empty Bar" />
    );
    expect(container).toBeTruthy();
    expect(screen.getByText('Empty Bar')).toBeInTheDocument();
  });

  it('renders without title when not provided', () => {
    const { container } = render(<BarChartComponent data={sampleTimeSeriesData} />);
    expect(container.querySelector('h4')).toBeNull();
  });
});

describe('PieChartComponent', () => {
  it('renders with title', () => {
    render(<PieChartComponent data={samplePieData} title="TVL Distribution" />);
    expect(screen.getByText('TVL Distribution')).toBeInTheDocument();
  });

  it('renders without crashing with empty data', () => {
    const { container } = render(
      <PieChartComponent data={[]} title="Empty Pie" />
    );
    expect(container).toBeTruthy();
    expect(screen.getByText('Empty Pie')).toBeInTheDocument();
  });

  it('renders without title when not provided', () => {
    const { container } = render(<PieChartComponent data={samplePieData} />);
    expect(container.querySelector('h4')).toBeNull();
  });
});

describe('RadarChartComponent', () => {
  it('renders with title', () => {
    render(<RadarChartComponent data={sampleRadarData} title="Valuation Radar" />);
    expect(screen.getByText('Valuation Radar')).toBeInTheDocument();
  });

  it('renders without crashing with empty data', () => {
    const { container } = render(
      <RadarChartComponent data={[]} title="Empty Radar" />
    );
    expect(container).toBeTruthy();
    expect(screen.getByText('Empty Radar')).toBeInTheDocument();
  });

  it('renders without title when not provided', () => {
    const { container } = render(<RadarChartComponent data={sampleRadarData} />);
    expect(container.querySelector('h4')).toBeNull();
  });
});
