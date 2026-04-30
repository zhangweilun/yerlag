import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts';
import type { TimeSeriesPoint } from '../../api/types';

export interface LineChartProps {
  data: TimeSeriesPoint[];
  title?: string;
  color?: string;
  height?: number;
}

function formatTimestamp(ts: number): string {
  return new Date(ts * 1000).toLocaleDateString();
}

export function LineChartComponent({
  data,
  title,
  color = 'var(--color-accent)',
  height = 300,
}: LineChartProps) {
  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
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
          <Line
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
