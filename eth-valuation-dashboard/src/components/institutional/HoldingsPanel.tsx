import { useDashboardStore } from '../../store/dashboard';
import { LineChartComponent } from '../charts/LineChartComponent';
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

export function HoldingsPanel() {
  const institutionalHoldings = useDashboardStore((s) => s.institutionalHoldings);

  if (!institutionalHoldings) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>机构持仓</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>机构持仓</h3>

      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <p className={styles.statLabel}>CME 期货未平仓量</p>
          <p className={styles.statValue}>{formatUsd(institutionalHoldings.cmeFuturesOI)}</p>
        </div>
      </div>

      {institutionalHoldings.institutions.length > 0 && (
        <div className={styles.chartSection}>
          <h4 style={{ margin: '0 0 8px', fontSize: '0.875rem' }}>大型机构持仓汇总</h4>
          <table className={styles.dataTable}>
            <thead>
              <tr>
                <th>机构名称</th>
                <th>持仓 (ETH)</th>
                <th>持仓 (USD)</th>
              </tr>
            </thead>
            <tbody>
              {institutionalHoldings.institutions.map((inst) => (
                <tr key={inst.name}>
                  <td>{inst.name}</td>
                  <td>{formatNumber(inst.holdingsEth)}</td>
                  <td>{formatUsd(inst.holdingsUsd)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className={styles.chartSection}>
        <LineChartComponent
          data={institutionalHoldings.cmeFuturesHistory}
          title="CME 期货 OI 趋势"
          color="#f59e0b"
          height={220}
        />
      </div>
    </div>
  );
}
