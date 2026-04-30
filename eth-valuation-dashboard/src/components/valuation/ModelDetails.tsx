import { useDashboardStore } from '../../store/dashboard';
import styles from './ValuationPanel.module.css';

function formatNumber(value: number, decimals = 2): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(decimals)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(decimals)}K`;
  return value.toFixed(decimals);
}

function getSignalLabel(signal: string): string {
  switch (signal) {
    case 'overvalued':
      return '高估';
    case 'undervalued':
      return '低估';
    case 'eth_undervalued':
      return 'ETH 相对低估';
    case 'eth_overvalued':
      return 'ETH 相对高估';
    default:
      return '中性';
  }
}

function getSignalBadgeClass(signal: string): string {
  switch (signal) {
    case 'overvalued':
    case 'eth_overvalued':
      return styles.badgeOvervalued;
    case 'undervalued':
    case 'eth_undervalued':
      return styles.badgeUndervalued;
    default:
      return styles.badgeNeutral;
  }
}

export function ModelDetails() {
  const valuationScore = useDashboardStore((s) => s.valuationScore);

  if (!valuationScore) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>估值模型详情</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  const { breakdown } = valuationScore;

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>估值模型详情</h3>

      <div className={styles.modelGrid}>
        {/* MVRV Ratio */}
        <div className={styles.modelCard}>
          <p className={styles.modelName}>MVRV Ratio</p>
          <p className={styles.modelValue}>{breakdown.mvrvScore.ratio.toFixed(3)}</p>
          <p className={styles.modelSignal}>
            <span className={`${styles.badge} ${getSignalBadgeClass(breakdown.mvrvScore.signal)}`}>
              {getSignalLabel(breakdown.mvrvScore.signal)}
            </span>
          </p>
        </div>

        {/* P/F Ratio */}
        <div className={styles.modelCard}>
          <p className={styles.modelName}>P/F Ratio (市费率)</p>
          <p className={styles.modelValue}>{formatNumber(breakdown.priceToFeeScore.ratio)}</p>
          <p className={styles.modelSignal}>
            <span className={`${styles.badge} ${getSignalBadgeClass(breakdown.priceToFeeScore.signal)}`}>
              {getSignalLabel(breakdown.priceToFeeScore.signal)}
            </span>
          </p>
        </div>

        {/* DCF Valuation */}
        <div className={styles.modelCard}>
          <p className={styles.modelName}>DCF 估值范围</p>
          <p className={styles.modelValue}>
            ${formatNumber(breakdown.dcfValuation.fairValueLow)} - ${formatNumber(breakdown.dcfValuation.fairValueHigh)}
          </p>
          <p className={styles.modelSignal}>
            <span className={`${styles.badge} ${styles.badgeNeutral}`}>
              中值 ${formatNumber(breakdown.dcfValuation.fairValueMid)}
            </span>
          </p>
        </div>

        {/* Stock-to-Flow */}
        <div className={styles.modelCard}>
          <p className={styles.modelName}>Stock-to-Flow</p>
          <p className={styles.modelValue}>${formatNumber(breakdown.stockToFlowScore.modelPrice)}</p>
          <p className={styles.modelSignal}>
            <span className={`${styles.badge} ${breakdown.stockToFlowScore.deviation > 0 ? styles.badgeOvervalued : styles.badgeUndervalued}`}>
              偏差 {breakdown.stockToFlowScore.deviation > 0 ? '+' : ''}{breakdown.stockToFlowScore.deviation.toFixed(1)}%
            </span>
          </p>
        </div>

        {/* NVT Ratio */}
        <div className={styles.modelCard}>
          <p className={styles.modelName}>NVT Ratio</p>
          <p className={styles.modelValue}>{breakdown.nvtScore.ratio.toFixed(2)}</p>
          <p className={styles.modelSignal}>
            <span className={`${styles.badge} ${getSignalBadgeClass(breakdown.nvtScore.signal)}`}>
              {getSignalLabel(breakdown.nvtScore.signal)}
            </span>
          </p>
        </div>

        {/* ETH/BTC */}
        <div className={styles.modelCard}>
          <p className={styles.modelName}>ETH/BTC 相对估值</p>
          <p className={styles.modelValue}>{breakdown.ethBtcScore.ratio.toFixed(5)}</p>
          <p className={styles.modelSignal}>
            <span className={`${styles.badge} ${getSignalBadgeClass(breakdown.ethBtcScore.signal)}`}>
              {getSignalLabel(breakdown.ethBtcScore.signal)}
            </span>
          </p>
        </div>
      </div>
    </div>
  );
}
