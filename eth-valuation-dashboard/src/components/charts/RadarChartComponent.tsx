import {
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

export interface RadarDataPoint {
  dimension: string;
  score: number;
  label: string;
}

export interface RadarChartProps {
  data: RadarDataPoint[];
  title?: string;
  height?: number;
}

export function RadarChartComponent({
  data,
  title,
  height = 300,
}: RadarChartProps) {
  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <RadarChart data={data} cx="50%" cy="50%" outerRadius="70%">
          <PolarGrid stroke="var(--color-border)" />
          <PolarAngleAxis
            dataKey="label"
            stroke="var(--color-text-muted)"
            fontSize={12}
          />
          <PolarRadiusAxis
            domain={[0, 100]}
            stroke="var(--color-text-muted)"
            fontSize={10}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-bg-card)',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
            }}
          />
          <Radar
            name="Score"
            dataKey="score"
            stroke="var(--color-accent)"
            fill="var(--color-accent)"
            fillOpacity={0.3}
            strokeWidth={2}
          />
        </RadarChart>
      </ResponsiveContainer>
    </div>
  );
}
