import { useDashboardStore } from '../../store/dashboard';
import { PieChartComponent } from '../charts/PieChartComponent';
import styles from './NetworkPanel.module.css';

export function PerformancePanel() {
  const networkPerformance = useDashboardStore((s) => s.networkPerformance);

  if (!networkPerformance) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>网络性能</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const clientDiversityPieData = networkPerformance.clientDiversity.map((item) => ({
    name: item.client,
    value: item.share,
  }));

  const missedSlotsStatus =
    networkPerformance.missedSlotsRate > 5
      ? styles.badgeDanger
      : networkPerformance.missedSlotsRate > 2
        ? styles.badgeWarning
        : styles.badgeSuccess;

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>网络性能</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>平均出块时间</p>
          <p className={styles.statValue}>{networkPerformance.avgBlockTime.toFixed(2)}s</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>当前 TPS</p>
          <p className={styles.statValue}>{networkPerformance.currentTps.toFixed(1)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>TPS 利用率</p>
          <p className={styles.statValue}>{networkPerformance.tpsRatio.toFixed(1)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>24h Missed Slots</p>
          <p className={styles.statValue}>
            {networkPerformance.missedSlots24h}
            <span className={`${styles.badge} ${missedSlotsStatus}`}>
              {networkPerformance.missedSlotsRate.toFixed(2)}%
            </span>
          </p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>验证率</p>
          <p className={styles.statValue}>
            <span className={styles.positive}>
              {networkPerformance.attestationRate.toFixed(2)}%
            </span>
          </p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>区块利用率</p>
          <p className={styles.statValue}>{networkPerformance.blockUtilization.toFixed(1)}%</p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <PieChartComponent
          data={clientDiversityPieData}
          title="客户端多样性"
          height={260}
        />
      </div>
    </div>
  );
}
