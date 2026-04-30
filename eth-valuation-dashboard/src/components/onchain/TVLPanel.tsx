import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import { PieChartComponent } from '../charts/PieChartComponent';
import styles from './OnChainPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(decimals)}B`;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

export function TVLPanel() {
  const tvlData = useDashboardStore((s) => s.tvlData);

  if (!tvlData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>TVL (总锁仓量)</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const protocolPieData = tvlData.topProtocols.map((p) => ({
    name: p.name,
    value: p.tvlUsd,
  }));

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>TVL (总锁仓量)</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>总 TVL (USD)</p>
          <p className={styles.statValue}>${formatNumber(tvlData.totalTvlUsd)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>总 TVL (ETH)</p>
          <p className={styles.statValue}>{formatNumber(tvlData.totalTvlEth)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>TVL/市值比</p>
          <p className={styles.statValue}>{(tvlData.tvlToMarketCapRatio * 100).toFixed(2)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>ETH TVL 市场份额</p>
          <p className={styles.statValue}>{(tvlData.ethTvlDominance * 100).toFixed(1)}%</p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <PieChartComponent
          data={protocolPieData}
          title="协议 TVL 分布"
          height={250}
        />
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={tvlData.dominanceHistory}
          title="市场份额趋势"
          color="#10b981"
          height={200}
        />
      </div>
    </div>
  );
}
