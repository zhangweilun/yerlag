import { useEffect, useState } from 'react';
import { getAlertHistory } from '../../api/client';
import type { Alert } from '../../api/types';
import styles from './Alert.module.css';

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp * 1000);
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function getSeverityClass(severity: Alert['severity']): string {
  switch (severity) {
    case 'high':
      return styles.severityHigh;
    case 'medium':
      return styles.severityMedium;
    case 'low':
      return styles.severityLow;
  }
}

export function AlertHistory() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetchHistory() {
      try {
        setLoading(true);
        const response = await getAlertHistory(30);
        if (!cancelled) {
          setAlerts(response.data);
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : '获取预警历史失败';
          setError(message);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void fetchHistory();
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>预警历史（近 30 天）</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>预警历史（近 30 天）</h3>
        <div className={styles.error}>{error}</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>预警历史（近 30 天）</h3>
      {alerts.length === 0 ? (
        <div className={styles.emptyState}>暂无预警记录</div>
      ) : (
        <div className={styles.historyList}>
          {alerts.map((alert) => (
            <div key={alert.id} className={styles.historyItem}>
              <span className={`${styles.severityBadge} ${getSeverityClass(alert.severity)}`}>
                {alert.severity}
              </span>
              <div className={styles.historyContent}>
                <p className={styles.historyTitle}>{alert.title}</p>
                <p className={styles.historyMessage}>{alert.message}</p>
                <div className={styles.historyMeta}>
                  <span className={styles.alertTime}>{formatTimestamp(alert.triggeredAt)}</span>
                  <span className={styles.historyMetric}>{alert.metricKey}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
