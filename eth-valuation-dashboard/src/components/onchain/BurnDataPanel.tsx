import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import { StackedAreaChartComponent } from '../charts/StackedAreaChartComponent';
import styles from './OnChainPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

export function BurnDataPanel() {
  const burnData = useDashboardStore((s) => s.burnData);
  const supplyData = useDashboardStore((s) => s.supplyData);

  if (!burnData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>EIP-1559 销毁数据</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const supplyDatasets = supplyData
    ? [
        { name: '质押', color: '#3b82f6', data: supplyData.supplyHistory },
        { name: '交易所', color: '#f59e0b', data: supplyData.exchangeBalanceHistory },
      ]
    : [];

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>EIP-1559 销毁数据</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>24h 销毁量</p>
          <p className={styles.statValue}>{formatNumber(burnData.daily)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>7d 销毁量</p>
          <p className={styles.statValue}>{formatNumber(burnData.weekly)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>30d 销毁量</p>
          <p className={styles.statValue}>{formatNumber(burnData.monthly)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>累计销毁量</p>
          <p className={styles.statValue}>{formatNumber(burnData.cumulative)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>年化销毁率</p>
          <p className={styles.statValue}>{(burnData.annualizedBurnRate * 100).toFixed(2)}%</p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={burnData.dailyHistory}
          title="每日销毁趋势"
          color="#ef4444"
          height={220}
        />
      </div>

      {supplyDatasets.length > 0 && (
        <div className={styles.chartSection}>
          <StackedAreaChartComponent
            datasets={supplyDatasets}
            title="供应量分布"
            height={220}
          />
        </div>
      )}
    </div>
  );
}
