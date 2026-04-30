import { useDashboardStore } from '../../store/dashboard';
import type { Alert } from '../../api/types';
import styles from './Alert.module.css';

const SEVERITY_ORDER: Record<Alert['severity'], number> = {
  high: 0,
  medium: 1,
  low: 2,
};

function formatTime(timestamp: number): string {
  const date = new Date(timestamp * 1000);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return '刚刚';
  if (diffMin < 60) return `${diffMin}分钟前`;
  const diffHours = Math.floor(diffMin / 60);
  if (diffHours < 24) return `${diffHours}小时前`;
  return `${Math.floor(diffHours / 24)}天前`;
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

export function AlertBanner() {
  const activeAlerts = useDashboardStore((s) => s.activeAlerts);

  if (!activeAlerts || activeAlerts.length === 0) {
    return null;
  }

  const sortedAlerts = [...activeAlerts].sort(
    (a, b) => SEVERITY_ORDER[a.severity] - SEVERITY_ORDER[b.severity]
  );

  return (
    <div className={styles.banner} role="alert" aria-label="活跃预警">
      <div className={styles.bannerTitle}>
        <span>⚠️ 活跃预警</span>
        <span className={styles.alertCount}>{sortedAlerts.length}</span>
      </div>
      <div className={styles.alertList}>
        {sortedAlerts.map((alert) => (
          <div key={alert.id} className={styles.alertItem}>
            <span className={`${styles.severityBadge} ${getSeverityClass(alert.severity)}`}>
              {alert.severity}
            </span>
            <p className={styles.alertMessage}>{alert.title}: {alert.message}</p>
            <span className={styles.alertTime}>{formatTime(alert.triggeredAt)}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
