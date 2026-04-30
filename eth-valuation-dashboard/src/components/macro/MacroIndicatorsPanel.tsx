import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
import styles from './MacroPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000_000) return `$${(value / 1_000_000_000).toFixed(decimals)}B`;
  if (value >= 1_000_000) return `$${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `$${(value / 1_000).toFixed(decimals)}K`;
  return `$${value.toFixed(decimals)}`;
}

function getFearGreedLabel(index: number): string {
  if (index <= 20) return '极度恐惧';
  if (index <= 40) return '恐惧';
  if (index <= 60) return '中性';
  if (index <= 80) return '贪婪';
  return '极度贪婪';
}

function getFearGreedStyle(index: number): string {
  if (index <= 20) return styles.badgeDanger;
  if (index <= 40) return styles.badgeWarning;
  if (index <= 60) return styles.badgeWarning;
  if (index <= 80) return styles.badgeSuccess;
  return styles.badgeSuccess;
}

export function MacroIndicatorsPanel() {
  const macroIndicators = useDashboardStore((s) => s.macroIndicators);

  if (!macroIndicators) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>宏观经济指标</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>宏观经济指标</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>美元指数 (DXY)</p>
          <p className={styles.statValue}>{macroIndicators.dxyIndex.toFixed(2)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>10Y 国债收益率</p>
          <p className={styles.statValue}>{macroIndicators.treasury10y.toFixed(3)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>纳斯达克 30d 相关性</p>
          <p className={styles.statValue}>{macroIndicators.nasdaqCorrelation30d.toFixed(3)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>纳斯达克 90d 相关性</p>
          <p className={styles.statValue}>{macroIndicators.nasdaqCorrelation90d.toFixed(3)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>联邦基金利率</p>
          <p className={styles.statValue}>{macroIndicators.fedFundsRate.toFixed(2)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>恐惧贪婪指数</p>
          <p className={styles.statValue}>
            {macroIndicators.fearGreedIndex}
            <span
              className={`${styles.badge} ${getFearGreedStyle(macroIndicators.fearGreedIndex)}`}
            >
              {getFearGreedLabel(macroIndicators.fearGreedIndex)}
            </span>
          </p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>稳定币市值</p>
          <p className={styles.statValue}>
            {formatNumber(macroIndicators.stablecoinMarketCap)}
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={macroIndicators.dxyHistory}
          title="DXY 趋势"
          color="#ef4444"
          height={200}
        />
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={macroIndicators.treasury10yHistory}
          title="10Y 国债收益率趋势"
          color="#f59e0b"
          height={200}
        />
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={macroIndicators.fearGreedHistory}
          title="恐惧贪婪指数趋势"
          color="#8b5cf6"
          height={200}
        />
      </div>

      <div className={styles.chartSection}>
        <LineChartComponent
          data={macroIndicators.stablecoinHistory}
          title="稳定币市值趋势"
          color="#10b981"
          height={200}
        />
      </div>
    </div>
  );
}
