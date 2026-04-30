import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';

export interface PieDataPoint {
  name: string;
  value: number;
}

export interface PieChartProps {
  data: PieDataPoint[];
  title?: string;
  height?: number;
}

const DEFAULT_COLORS = [
  '#3b82f6',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#06b6d4',
  '#ec4899',
  '#84cc16',
];

export function PieChartComponent({
  data,
  title,
  height = 300,
}: PieChartProps) {
  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            outerRadius="70%"
            label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(1)}%`}
            labelLine={false}
          >
            {data.map((_entry, index) => (
              <Cell
                key={index}
                fill={DEFAULT_COLORS[index % DEFAULT_COLORS.length]}
              />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-bg-card)',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
            }}
          />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
