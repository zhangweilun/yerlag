import { useDashboardStore } from '../../store/dashboard';
import { BarChartComponent } from '../charts/BarChartComponent';
import styles from './InstitutionalPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(decimals)}B`;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

function formatUsd(value: number): string {
  return `$${formatNumber(value)}`;
}

export function ETFPanel() {
  const etfData = useDashboardStore((s) => s.etfData);

  if (!etfData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>ETF 持仓与资金流向</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>ETF 持仓与资金流向</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>总持仓量 (ETH)</p>
          <p className={styles.statValue}>{formatNumber(etfData.totalHoldingsEth)} ETH</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>总持仓量 (USD)</p>
          <p className={styles.statValue}>{formatUsd(etfData.totalHoldingsUsd)}</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>占总供应量</p>
          <p className={styles.statValue}>{etfData.holdingsPercentOfSupply.toFixed(2)}%</p>
        </div>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>累计净流入</p>
          <p className={styles.statValue}>
            <span className={etfData.cumulativeNetFlow >= 0 ? styles.positive : styles.negative}>
              {formatUsd(etfData.cumulativeNetFlow)}
            </span>
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <BarChartComponent
          data={etfData.netFlowHistory}
          title="净流入/流出历史"
          color="#3b82f6"
          height={220}
        />
      </div>

      {etfData.etfs.length > 0 && (
        <div className={styles.chartSection}>
          <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>发行商排名</h4>
          <table className={styles.dataTable}>
            <thead>
              <tr>
                <th>发行商</th>
                <th>代码</th>
                <th>持仓 (ETH)</th>
                <th>日净流入</th>
                <th>市场份额</th>
              </tr>
            </thead>
            <tbody>
              {etfData.etfs.map((etf) => (
                <tr key={etf.ticker}>
                  <td>{etf.issuer}</td>
                  <td>{etf.ticker}</td>
                  <td>{formatNumber(etf.holdingsEth)}</td>
                  <td className={etf.dailyNetFlowUsd >= 0 ? styles.positive : styles.negative}>
                    {formatUsd(etf.dailyNetFlowUsd)}
                  </td>
                  <td>{etf.marketShare.toFixed(1)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
