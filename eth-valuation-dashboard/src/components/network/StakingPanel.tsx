import { useDashboardStore } from '../../store/dashboard';
import { PieChartComponent } from '../charts/PieChartComponent';
import { LineChartComponent } from '../charts/LineChartComponent';
import styles from './NetworkPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(decimals)}B`;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

export function StakingPanel() {
  const stakingData = useDashboardStore((s) => s.stakingData);

  if (!stakingData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>质押数据</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const liquidStakingPieData = stakingData.liquidStakingShares.map((item) => ({
    name: item.protocol,
    value: item.share,
  }));

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>质押数据</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>总质押量</p>
          <p className={styles.statValue}>{formatNumber(stakingData.totalStakedEth)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>质押占比</p>
          <p className={styles.statValue}>{stakingData.stakingPercentage.toFixed(2)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>活跃验证者</p>
          <p className={styles.statValue}>{formatNumber(stakingData.activeValidators, 0)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>质押收益率</p>
          <p className={styles.statValue}>
            <span className={styles.positive}>{stakingData.stakingYield.toFixed(2)}%</span>
          </p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>进入队列</p>
          <p className={styles.statValue}>
            {formatNumber(stakingData.entryQueueLength, 0)}
            <span className={styles.badge + ' ' + styles.badgeWarning}>
              {stakingData.entryWaitTime}
            </span>
          </p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>退出队列</p>
          <p className={styles.statValue}>
            {formatNumber(stakingData.exitQueueLength, 0)}
            <span className={styles.badge + ' ' + styles.badgeWarning}>
              {stakingData.exitWaitTime}
            </span>
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <PieChartComponent
          data={liquidStakingPieData}
          title="流动性质押份额"
          height={260}
        />
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={stakingData.validatorHistory}
          title="验证者数量趋势"
          color="#8b5cf6"
          height={220}
        />
      </div>
    </div>
  );
}
