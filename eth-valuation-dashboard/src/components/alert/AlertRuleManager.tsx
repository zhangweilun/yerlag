import { useState } from 'react';
import { useDashboardStore } from '../../store/dashboard';
import type { AlertCondition } from '../../api/types';
import styles from './Alert.module.css';

const METRIC_OPTIONS = [
  { value: 'burn_daily', label: '每日销毁量' },
  { value: 'gas_avg_gwei', label: 'Gas 均价 (Gwei)' },
  { value: 'tvl_dominance', label: 'TVL 市场份额 (%)' },
  { value: 'etf_net_flow', label: 'ETF 净流入 (USD)' },
  { value: 'grayscale_premium', label: '灰度溢价/折价率 (%)' },
  { value: 'validator_exit_queue', label: '验证者退出队列' },
  { value: 'missed_slots_rate', label: 'Missed Slots 率 (%)' },
  { value: 'exchange_balance', label: '交易所余额 (ETH)' },
  { value: 'eth_price', label: 'ETH 价格 (USD)' },
  { value: 'nvt_ratio', label: 'NVT Ratio' },
];

const CONDITION_OPTIONS: { value: AlertCondition['type']; label: string }[] = [
  { value: 'gt', label: '大于' },
  { value: 'lt', label: '小于' },
  { value: 'gt_percent_change', label: '涨幅超过 (%)' },
  { value: 'lt_percent_change', label: '跌幅超过 (%)' },
];

const SEVERITY_OPTIONS: { value: 'high' | 'medium' | 'low'; label: string }[] = [
  { value: 'high', label: '高' },
  { value: 'medium', label: '中' },
  { value: 'low', label: '低' },
];

export function AlertRuleManager() {
  const setAlertRule = useDashboardStore((s) => s.setAlertRule);
  const isLoading = useDashboardStore((s) => s.isLoading['alertRule'] ?? false);
  const ruleError = useDashboardStore((s) => s.errors['alertRule'] ?? null);

  const [metricKey, setMetricKey] = useState(METRIC_OPTIONS[0].value);
  const [conditionType, setConditionType] = useState<AlertCondition['type']>('gt');
  const [threshold, setThreshold] = useState('');
  const [severity, setSeverity] = useState<'high' | 'medium' | 'low'>('medium');
  const [message, setMessage] = useState('');
  const [submitted, setSubmitted] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const thresholdNum = parseFloat(threshold);
    if (isNaN(thresholdNum)) return;

    const ruleMessage = message || `${metricKey} ${conditionType} ${thresholdNum}`;

    await setAlertRule({
      metricKey,
      condition: { type: conditionType },
      threshold: thresholdNum,
      severity,
      enabled: true,
      message: ruleMessage,
    });

    if (!ruleError) {
      setThreshold('');
      setMessage('');
      setSubmitted(true);
      setTimeout(() => setSubmitted(false), 3000);
    }
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>预警规则管理</h3>
      <form onSubmit={handleSubmit}>
        <div className={styles.ruleForm}>
          <div className={styles.formGroup}>
            <label className={styles.formLabel} htmlFor="alert-metric">指标</label>
            <select
              id="alert-metric"
              className={styles.formSelect}
              value={metricKey}
              onChange={(e) => setMetricKey(e.target.value)}
            >
              {METRIC_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div className={styles.formGroup}>
            <label className={styles.formLabel} htmlFor="alert-condition">条件</label>
            <select
              id="alert-condition"
              className={styles.formSelect}
              value={conditionType}
              onChange={(e) => setConditionType(e.target.value as AlertCondition['type'])}
            >
              {CONDITION_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div className={styles.formGroup}>
            <label className={styles.formLabel} htmlFor="alert-threshold">阈值</label>
            <input
              id="alert-threshold"
              className={styles.formInput}
              type="number"
              step="any"
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
              placeholder="输入阈值"
              required
            />
          </div>

          <div className={styles.formGroup}>
            <label className={styles.formLabel} htmlFor="alert-severity">严重程度</label>
            <select
              id="alert-severity"
              className={styles.formSelect}
              value={severity}
              onChange={(e) => setSeverity(e.target.value as 'high' | 'medium' | 'low')}
            >
              {SEVERITY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div className={styles.formGroup}>
            <label className={styles.formLabel} htmlFor="alert-message">消息（可选）</label>
            <input
              id="alert-message"
              className={styles.formInput}
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="自定义预警消息"
            />
          </div>

          <div className={styles.formGroup}>
            <label className={styles.formLabel}>&nbsp;</label>
            <button
              type="submit"
              className={styles.submitButton}
              disabled={isLoading || !threshold}
            >
              {isLoading ? '创建中...' : '创建规则'}
            </button>
          </div>
        </div>
      </form>

      {ruleError && <p className={styles.error}>{ruleError}</p>}
      {submitted && !ruleError && (
        <p style={{ color: 'var(--color-success)', fontSize: '0.75rem', margin: '0.5rem 0 0' }}>
          ✓ 规则创建成功
        </p>
      )}
    </div>
  );
}
