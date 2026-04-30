import { useDashboardStore } from '../../store/dashboard';
import styles from './MarketPanel.module.css';

function formatUsd(value: number): string {
  return `$${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatSpread(value: number): string {
  return `${value >= 0 ? '+' : ''}${value.toFixed(4)}%`;
}

export function ExchangeSpread() {
  const marketData = useDashboardStore((s) => s.marketData);

  if (!marketData) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>交易所价差</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const { exchangePrices } = marketData;

  if (!exchangePrices || exchangePrices.length === 0) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>交易所价差</h3>
        <p className={styles.loading}>暂无数据</p>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>交易所价差</h3>

      <table className={styles.spreadTable}>
        <thead>
          <tr>
            <th>交易所</th>
            <th>价格</th>
            <th>价差</th>
          </tr>
        </thead>
        <tbody>
          {exchangePrices.map((ep) => (
            <tr key={ep.exchange}>
              <td>{ep.exchange}</td>
              <td>{formatUsd(ep.price)}</td>
              <td className={ep.spread >= 0 ? styles.positive : styles.negative}>
                {formatSpread(ep.spread)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
