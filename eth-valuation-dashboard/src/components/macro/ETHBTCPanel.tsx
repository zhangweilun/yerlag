import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import styles from './MacroPanel.module.css';

function getSignalLabel(signal: string): string {
  switch (signal) {
    case 'eth_undervalued':
      return 'ETH 低估';
    case 'eth_overvalued':
      return 'ETH 高估';
    default:
      return '中性';
  }
}

function getSignalStyle(signal: string): string {
  switch (signal) {
    case 'eth_undervalued':
      return styles.badgeSuccess;
    case 'eth_overvalued':
      return styles.badgeDanger;
    default:
      return styles.badgeWarning;
  }
}

export function ETHBTCPanel() {
  const ethbtcData = useDashboardStore((s) => s.ethbtcData);

  if (!ethbtcData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>ETH/BTC 相对估值</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>ETH/BTC 相对估值</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>ETH/BTC 价格</p>
          <p className={styles.statValue}>{ethbtcData.ethBtcPrice.toFixed(6)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>30d 相关系数</p>
          <p className={styles.statValue}>{ethbtcData.correlation30d.toFixed(3)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>90d 相关系数</p>
          <p className={styles.statValue}>{ethbtcData.correlation90d.toFixed(3)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>ETH Dominance</p>
          <p className={styles.statValue}>{ethbtcData.ethDominance.toFixed(2)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>历史百分位</p>
          <p className={styles.statValue}>{ethbtcData.ethBtcPercentile.toFixed(1)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>信号</p>
          <p className={styles.statValue}>
            <span className={`${styles.badge} ${getSignalStyle(ethbtcData.ethBtcSignal)}`}>
              {getSignalLabel(ethbtcData.ethBtcSignal)}
            </span>
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={ethbtcData.ethBtcHistory}
          title="ETH/BTC 价格趋势"
          color="#f59e0b"
          height={220}
        />
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={ethbtcData.ethDominanceHistory}
          title="ETH Dominance 趋势"
          color="#8b5cf6"
          height={220}
        />
      </div>
    </div>
  );
}
