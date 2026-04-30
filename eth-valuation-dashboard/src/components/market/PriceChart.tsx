import { useState, useEffect, useCallback } from 'react';
import { CandlestickChart } from '../charts/CandlestickChart';
import { getPriceHistory } from '../../api/client';
import type { OHLCVPoint, TimeRange } from '../../api/types';
import styles from './MarketPanel.module.css';

const TIME_RANGES: { label: string; value: TimeRange }[] = [
  { label: '1D', value: '1d' },
  { label: '1W', value: '1w' },
  { label: '1M', value: '1m' },
  { label: '3M', value: '3m' },
  { label: '1Y', value: '1y' },
  { label: 'ALL', value: 'all' },
];

export function PriceChart() {
  const [timeRange, setTimeRange] = useState<TimeRange>('1m');
  const [priceData, setPriceData] = useState<OHLCVPoint[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchData = useCallback(async (range: TimeRange) => {
    setLoading(true);
    try {
      const response = await getPriceHistory(range);
      setPriceData(response.data);
    } catch {
      // Keep previous data on error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchData(timeRange);
  }, [timeRange, fetchData]);

  const handleRangeChange = (range: TimeRange) => {
    setTimeRange(range);
  };

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>价格走势 (K 线图)</h3>

      <div className={styles.timeRangeSelector}>
        {TIME_RANGES.map(({ label, value }) => (
          <button
            key={value}
            className={`${styles.timeRangeBtn} ${timeRange === value ? styles.timeRangeBtnActive : ''}`}
            onClick={() => handleRangeChange(value)}
            type="button"
          >
            {label}
          </button>
        ))}
      </div>

      {loading && priceData.length === 0 ? (
        <div className={styles.loading}>加载中...</div>
      ) : (
        <div className={styles.chartSection}>
          <CandlestickChart data={priceData} height={320} />
        </div>
      )}
    </div>
  );
}
