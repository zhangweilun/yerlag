import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Legend,
} from 'recharts';

export interface StackedDataset {
  name: string;
  color: string;
  data: { timestamp: number; value: number }[];
}

export interface StackedAreaChartProps {
  datasets: StackedDataset[];
  title?: string;
  height?: number;
}

function formatTimestamp(ts: number): string {
  return new Date(ts * 1000).toLocaleDateString();
}

/**
 * Merges multiple datasets into a single array keyed by timestamp.
 * Each entry has { timestamp, [name1]: value1, [name2]: value2, ... }
 */
function mergeDatasets(datasets: StackedDataset[]): Record<string, number>[] {
  const map = new Map<number, Record<string, number>>();

  for (const ds of datasets) {
    for (const point of ds.data) {
      if (!map.has(point.timestamp)) {
        map.set(point.timestamp, { timestamp: point.timestamp });
      }
      map.get(point.timestamp)![ds.name] = point.value;
    }
  }

  return Array.from(map.values()).sort((a, b) => a.timestamp - b.timestamp);
}

export function StackedAreaChartComponent({
  datasets,
  title,
  height = 300,
}: StackedAreaChartProps) {
  const mergedData = mergeDatasets(datasets);

  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={mergedData} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
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
          <Legend />
          {datasets.map((ds) => (
            <Area
              key={ds.name}
              type="monotone"
              dataKey={ds.name}
              stackId="1"
              stroke={ds.color}
              fill={ds.color}
              fillOpacity={0.4}
            />
          ))}
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
