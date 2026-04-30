import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import { BarChartComponent } from '../charts/BarChartComponent';
import styles from './OnChainPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

function getSignalBadge(signal: string) {
  switch (signal) {
    case 'overvalued':
      return <span className={`${styles.badge} ${styles.badgeDanger}`}>高估</span>;
    case 'undervalued':
      return <span className={`${styles.badge} ${styles.badgeSuccess}`}>低估</span>;
    default:
      return <span className={`${styles.badge} ${styles.badgeNeutral}`}>中性</span>;
  }
}

export function ActivityPanel() {
  const activityData = useDashboardStore((s) => s.activityData);

  if (!activityData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>网络活跃度</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const l2ChartData = activityData.l2Comparison.map((item, index) => ({
    timestamp: index,
    value: item.dailyTransactions,
  }));

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>网络活跃度</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>日活跃地址 (DAA)</p>
          <p className={styles.statValue}>{formatNumber(activityData.dailyActiveAddresses)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>7d DAA 均值</p>
          <p className={styles.statValue}>{formatNumber(activityData.daaMovingAvg7d)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>日交易笔数</p>
          <p className={styles.statValue}>{formatNumber(activityData.dailyTransactions)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>日新增地址</p>
          <p className={styles.statValue}>{formatNumber(activityData.dailyNewAddresses)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>NVT Ratio</p>
          <p className={styles.statValue}>
            {activityData.nvtRatio.toFixed(1)}
            {getSignalBadge(activityData.nvtSignal)}
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={activityData.transactionHistory}
          title="交易量趋势"
          color="#8b5cf6"
          height={220}
        />
      </div>

      {l2ChartData.length > 0 && (
        <div className={styles.chartSection}>
          <BarChartComponent
            data={l2ChartData}
            title="L2 交易量对比"
            color="#06b6d4"
            height={200}
          />
        </div>
      )}
    </div>
  );
}
