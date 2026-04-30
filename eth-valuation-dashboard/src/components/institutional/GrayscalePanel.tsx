import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import styles from './InstitutionalPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(decimals)}B`;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

export function GrayscalePanel() {
  const grayscaleData = useDashboardStore((s) => s.grayscaleData);

  if (!grayscaleData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>灰度信托 (ETHE)</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const isPremium = grayscaleData.premiumDiscount >= 0;

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>灰度信托 (ETHE)</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>持仓量</p>
          <p className={styles.statValue}>{formatNumber(grayscaleData.holdingsEth)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>资产净值 (NAV)</p>
          <p className={styles.statValue}>${formatNumber(grayscaleData.nav)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>{isPremium ? '溢价率' : '折价率'}</p>
          <p className={styles.statValue}>
            <span className={isPremium ? styles.positive : styles.negative}>
              {isPremium ? '+' : ''}{grayscaleData.premiumDiscount.toFixed(2)}%
            </span>
            <span className={`${styles.badge} ${isPremium ? styles.badgeSuccess : styles.badgeDanger}`}>
              {isPremium ? '溢价' : '折价'}
            </span>
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={grayscaleData.premiumHistory}
          title="溢价/折价率趋势"
          color="#8b5cf6"
          height={220}
        />
      </div>
    </div>
  );
}
