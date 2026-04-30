import { useDashboardStore } from '../../store/dashboard';
import { AreaChartComponent } from '../charts/AreaChartComponent';
import { PieChartComponent } from '../charts/PieChartComponent';
import styles from './OnChainPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

export function SupplyPanel() {
  const supplyData = useDashboardStore((s) => s.supplyData);

  if (!supplyData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>供应量</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const distributionData = [
    { name: '质押', value: supplyData.stakedAmount },
    { name: 'DeFi 锁仓', value: supplyData.defiLocked },
    { name: '交易所余额', value: supplyData.exchangeBalance },
    { name: '其他', value: supplyData.otherAmount },
  ];

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>
        供应量
        {supplyData.isDeflationary ? (
          <span className={`${styles.badge} ${styles.badgeSuccess}`}>通缩</span>
        ) : (
          <span className={`${styles.badge} ${styles.badgeWarning}`}>通胀</span>
        )}
      </h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>总供应量</p>
          <p className={styles.statValue}>{formatNumber(supplyData.totalSupply)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>净发行量</p>
          <p className={styles.statValue}>{formatNumber(supplyData.netIssuance)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>年化通胀率</p>
          <p className={styles.statValue}>{(supplyData.annualInflationRate * 100).toFixed(3)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>交易所余额</p>
          <p className={styles.statValue}>{formatNumber(supplyData.exchangeBalance)} ETH</p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <PieChartComponent
          data={distributionData}
          title="供应量分类分布"
          height={250}
        />
      </div>

      <div className={styles.chartSection}>
        <AreaChartComponent
          data={supplyData.exchangeBalanceHistory}
          title="交易所余额趋势"
          color="#f59e0b"
          height={200}
        />
      </div>
    </div>
  );
}
