import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  ReferenceLine,
} from 'recharts';
import type { DistributionData } from '../../api/types';
import styles from './ValuationPanel.module.css';

export interface DistributionChartProps {
  data: DistributionData;
  title?: string;
  height?: number;
  color?: string;
}

interface BucketData {
  range: string;
  count: number;
  isCurrent: boolean;
}

function buildHistogram(values: number[], currentValue: number, bucketCount = 20): BucketData[] {
  if (values.length === 0) return [];

  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min;

  if (range === 0) {
    return [{ range: min.toFixed(2), count: values.length, isCurrent: true }];
  }

  const bucketSize = range / bucketCount;
  const buckets: BucketData[] = [];

  for (let i = 0; i < bucketCount; i++) {
    const bucketMin = min + i * bucketSize;
    const bucketMax = min + (i + 1) * bucketSize;
    const count = values.filter(
      (v) => (i === bucketCount - 1 ? v >= bucketMin && v <= bucketMax : v >= bucketMin && v < bucketMax)
    ).length;
    const isCurrent = currentValue >= bucketMin && (i === bucketCount - 1 ? currentValue <= bucketMax : currentValue < bucketMax);

    buckets.push({
      range: `${bucketMin.toFixed(2)}`,
      count,
      isCurrent,
    });
  }

  return buckets;
}

export function DistributionChart({
  data,
  title = '历史分布',
  height = 250,
  color = 'var(--color-accent)',
}: DistributionChartProps) {
  const histogram = buildHistogram(data.values, data.currentValue);

  if (histogram.length === 0) {
    return (
      <div className={styles.panel}>
        {title && <h3 className={styles.panelTitle}>{title}</h3>}
        <div className={styles.loading}>暂无分布数据</div>
      </div>
    );
  }

  return (
    <div>
      {title && <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <BarChart data={histogram} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
          <XAxis
            dataKey="range"
            stroke="var(--color-text-muted)"
            fontSize={10}
            interval="preserveStartEnd"
          />
          <YAxis stroke="var(--color-text-muted)" fontSize={12} />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-bg-card)',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
            }}
            formatter={(value) => [`${value} 次`, '频次']}
            labelFormatter={(label) => `区间: ${label}`}
          />
          <ReferenceLine
            x={histogram.find((b) => b.isCurrent)?.range}
            stroke="#ef4444"
            strokeWidth={2}
            strokeDasharray="4 4"
            label={{ value: '当前值', position: 'top', fontSize: 11, fill: '#ef4444' }}
          />
          <Bar
            dataKey="count"
            fill={color}
            radius={[4, 4, 0, 0]}
          />
        </BarChart>
      </ResponsiveContainer>
      <div className={styles.distributionMarker}>
        <p className={styles.markerLabel}>
          当前值: {data.currentValue.toFixed(4)} (第 {data.currentPercentile.toFixed(1)} 百分位)
        </p>
      </div>
    </div>
  );
}
