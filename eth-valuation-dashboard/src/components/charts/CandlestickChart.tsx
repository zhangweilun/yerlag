import {
  ComposedChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Cell,
} from 'recharts';
import type { OHLCVPoint } from '../../api/types';

export interface CandlestickChartProps {
  data: OHLCVPoint[];
  title?: string;
  height?: number;
}

interface CandleData {
  timestamp: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  /** Bottom of the candle body (min of open, close) */
  bodyLow: number;
  /** Height of the candle body */
  bodyHeight: number;
  /** Whether the candle is bullish (close >= open) */
  bullish: boolean;
}

function formatTimestamp(ts: number): string {
  return new Date(ts * 1000).toLocaleDateString();
}

function transformData(data: OHLCVPoint[]): CandleData[] {
  return data.map((point) => ({
    ...point,
    bodyLow: Math.min(point.open, point.close),
    bodyHeight: Math.abs(point.close - point.open),
    bullish: point.close >= point.open,
  }));
}

export function CandlestickChart({
  data,
  title,
  height = 300,
}: CandlestickChartProps) {
  const candleData = transformData(data);

  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <ComposedChart data={candleData} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTimestamp}
            stroke="var(--color-text-muted)"
            fontSize={12}
          />
          <YAxis
            domain={['auto', 'auto']}
            stroke="var(--color-text-muted)"
            fontSize={12}
          />
          <Tooltip
            labelFormatter={(label) => formatTimestamp(label as number)}
            formatter={(value, name, props) => {
              const p = props.payload as CandleData;
              if (name === 'bodyHeight') {
                return [`O: ${p.open} H: ${p.high} L: ${p.low} C: ${p.close}`, 'OHLC'];
              }
              return [value, name];
            }}
            contentStyle={{
              backgroundColor: 'var(--color-bg-card)',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
            }}
          />
          {/* Invisible bar to offset the body from the baseline */}
          <Bar dataKey="bodyLow" stackId="candle" fill="transparent" />
          {/* Visible candle body */}
          <Bar dataKey="bodyHeight" stackId="candle" barSize={8}>
            {candleData.map((entry, index) => (
              <Cell
                key={index}
                fill={entry.bullish ? 'var(--color-success)' : 'var(--color-danger)'}
              />
            ))}
          </Bar>
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
