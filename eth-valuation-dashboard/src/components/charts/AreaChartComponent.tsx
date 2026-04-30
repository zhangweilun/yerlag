import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts';
import type { TimeSeriesPoint } from '../../api/types';

export interface AreaChartProps {
  data: TimeSeriesPoint[];
  title?: string;
  color?: string;
  height?: number;
}

function formatTimestamp(ts: number): string {
  return new Date(ts * 1000).toLocaleDateString();
}

export function AreaChartComponent({
  data,
  title,
  color = 'var(--color-accent)',
  height = 300,
}: AreaChartProps) {
  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTimestamp}
            stroke="var(--color-text-muted)"
            fontSize={12}
          />
          <YAxis stroke="var(--color-text-muted)" fontSize={12} />
          <Tooltip
            labelFormatter={(label) => formatTimestamp(label as number)}
            contentStyle={{
              backgroundColor: 'var(--color-bg-card)',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
            }}
          />
          <Area
            type="monotone"
            dataKey="value"
            stroke={color}
            fill={color}
            fillOpacity={0.2}
            strokeWidth={2}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
