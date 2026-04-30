import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import styles from './OnChainPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

export function GasDataPanel() {
  const gasData = useDashboardStore((s) => s.gasData);

  if (!gasData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>Gas 费用</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>
        Gas 费用
        {gasData.isHighFee && (
          <span className={`${styles.badge} ${styles.badgeWarning}`}>高费用</span>
        )}
      </h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>当前均价</p>
          <p className={styles.statValue}>{gasData.currentAvgGwei.toFixed(1)} Gwei</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>日费用收入 (ETH)</p>
          <p className={styles.statValue}>{formatNumber(gasData.dailyFeeRevenueEth)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>日费用收入 (USD)</p>
          <p className={styles.statValue}>${formatNumber(gasData.dailyFeeRevenueUsd)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>市费率 (P/F)</p>
          <p className={styles.statValue}>{gasData.priceToFeeRatio.toFixed(1)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>年化收入</p>
          <p className={styles.statValue}>${formatNumber(gasData.annualizedRevenue)}</p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={gasData.gasHistory}
          title="Gas 价格历史趋势"
          color="#f59e0b"
          height={220}
        />
      </div>
    </div>
  );
}
