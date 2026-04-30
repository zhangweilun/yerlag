import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { useDashboardStore } from '../../store/dashboard';
import { AlertBanner } from '../alert/AlertBanner';
import { ExportButton } from '../common/ExportButton';
import { generateCSVContent } from '../../utils/export';
import type { Alert } from '../../api/types';

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
    activeAlerts: [],
    isLoading: {},
    errors: {},
  });
});

/**
 * Validates: Requirements 15.1-15.5
 */
describe('AlertBanner', () => {
  it('renders nothing when there are no alerts', () => {
    const { container } = render(<AlertBanner />);
    expect(container.firstChild).toBeNull();
  });

  it('renders alerts sorted by severity (high first)', () => {
    const alerts: Alert[] = [
      {
        id: 'a1',
        ruleId: 'r1',
        triggeredAt: Math.floor(Date.now() / 1000) - 60,
        severity: 'low',
        title: 'Low Alert',
        message: 'Low severity message',
        metricKey: 'gas_avg_gwei',
        currentValue: 30,
        thresholdValue: 25,
        acknowledged: false,
      },
      {
        id: 'a2',
        ruleId: 'r2',
        triggeredAt: Math.floor(Date.now() / 1000) - 120,
        severity: 'high',
        title: 'High Alert',
        message: 'High severity message',
        metricKey: 'burn_daily',
        currentValue: 5000,
        thresholdValue: 3000,
        acknowledged: false,
      },
      {
        id: 'a3',
        ruleId: 'r3',
        triggeredAt: Math.floor(Date.now() / 1000) - 180,
        severity: 'medium',
        title: 'Medium Alert',
        message: 'Medium severity message',
        metricKey: 'etf_net_flow',
        currentValue: 1000000,
        thresholdValue: 500000,
        acknowledged: false,
      },
    ];

    useDashboardStore.setState({ activeAlerts: alerts });

    render(<AlertBanner />);

    // Verify all alerts are rendered
    expect(screen.getByText(/High Alert/)).toBeInTheDocument();
    expect(screen.getByText(/Medium Alert/)).toBeInTheDocument();
    expect(screen.getByText(/Low Alert/)).toBeInTheDocument();

    // Verify sorting: high should appear before medium, medium before low
    const alertItems = screen.getAllByText(/Alert:/);
    // The text content includes "title: message" format
    const alertTexts = alertItems.map((el) => el.textContent);
    expect(alertTexts[0]).toContain('High Alert');
    expect(alertTexts[1]).toContain('Medium Alert');
    expect(alertTexts[2]).toContain('Low Alert');
  });

  it('shows alert count badge', () => {
    const alerts: Alert[] = [
      {
        id: 'a1',
        ruleId: 'r1',
        triggeredAt: Math.floor(Date.now() / 1000) - 60,
        severity: 'high',
        title: 'Alert 1',
        message: 'Message 1',
        metricKey: 'burn_daily',
        currentValue: 5000,
        thresholdValue: 3000,
        acknowledged: false,
      },
      {
        id: 'a2',
        ruleId: 'r2',
        triggeredAt: Math.floor(Date.now() / 1000) - 120,
        severity: 'medium',
        title: 'Alert 2',
        message: 'Message 2',
        metricKey: 'gas_avg_gwei',
        currentValue: 60,
        thresholdValue: 50,
        acknowledged: false,
      },
    ];

    useDashboardStore.setState({ activeAlerts: alerts });

    render(<AlertBanner />);
    expect(screen.getByText('2')).toBeInTheDocument();
  });
});

/**
 * Validates: Requirements 16.1-16.4
 */
describe('ExportButton', () => {
  it('renders with format label', () => {
    const onClick = vi.fn();
    render(<ExportButton onClick={onClick} format="CSV" />);

    expect(screen.getByText('CSV')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Export as CSV' })).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const onClick = vi.fn();
    render(<ExportButton onClick={onClick} format="PNG" />);

    const button = screen.getByRole('button', { name: 'Export as PNG' });
    fireEvent.click(button);

    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('renders with custom title', () => {
    const onClick = vi.fn();
    render(<ExportButton onClick={onClick} format="SVG" title="Download chart as SVG" />);

    expect(screen.getByRole('button', { name: 'Download chart as SVG' })).toBeInTheDocument();
  });
});

/**
 * Validates: Requirements 16.2
 */
describe('generateCSVContent', () => {
  it('produces valid CSV output with headers and data rows', () => {
    const headers = ['name', 'price', 'change'];
    const records = [
      { name: 'ETH', price: 3250.5, change: 4.25 },
      { name: 'BTC', price: 65000, change: -1.2 },
    ];

    const csv = generateCSVContent(headers, records);

    // Should contain header line
    expect(csv).toContain('name,price,change');
    // Should contain data rows
    expect(csv).toContain('ETH,3250.5,4.25');
    expect(csv).toContain('BTC,65000,-1.2');
    // Should use CRLF line endings per RFC 4180
    expect(csv).toContain('\r\n');
  });

  it('escapes fields containing commas', () => {
    const headers = ['description', 'value'];
    const records = [{ description: 'ETH, the token', value: 100 }];

    const csv = generateCSVContent(headers, records);

    expect(csv).toContain('"ETH, the token"');
  });

  it('handles null and undefined values as empty strings', () => {
    const headers = ['name', 'optional'];
    const records = [{ name: 'ETH', optional: null }];

    const csv = generateCSVContent(headers, records);

    // The row should have name followed by empty value
    expect(csv).toContain('ETH,');
  });
});
