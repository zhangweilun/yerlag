import { useDashboardStore } from '../../store/dashboard';
import styles from './MarketPanel.module.css';

function formatUsd(value: number): string {
  if (value >= 1_000_000_000) return `$${(value / 1_000_000_000).toFixed(2)}B`;
  if (value >= 1_000_000) return `$${(value / 1_000_000).toFixed(2)}M`;
  if (value >= 1_000) return `$${(value / 1_000).toFixed(2)}K`;
  return `$${value.toFixed(2)}`;
}

function formatPercent(value: number): string {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
}

export function MarketOverview() {
  const marketData = useDashboardStore((s) => s.marketData);

  if (!marketData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>市场总览</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const changeClass = marketData.priceChange24h >= 0 ? styles.positive : styles.negative;

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>市场总览</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>实时价格</p>
          <p className={styles.statValue}>{formatUsd(marketData.currentPrice)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>24h 涨跌幅</p>
          <p className={`${styles.statValue} ${changeClass}`}>
            {formatPercent(marketData.priceChange24h)}
          </p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>24h 交易量</p>
          <p className={styles.statValue}>{formatUsd(marketData.volume24h)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>流通市值</p>
          <p className={styles.statValue}>{formatUsd(marketData.circulatingMarketCap)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>ATH 回撤</p>
          <p className={`${styles.statValue} ${styles.negative}`}>
            {formatPercent(-marketData.athDrawdown)}
          </p>
        </div>
      </div>
    </div>
  );
}
