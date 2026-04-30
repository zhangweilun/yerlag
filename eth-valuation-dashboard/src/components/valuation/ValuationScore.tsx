import { useDashboardStore } from '../../store/dashboard';
import { RadarChartComponent } from '../charts/RadarChartComponent';
import styles from './ValuationPanel.module.css';

function getStatusLabel(status: 'undervalued' | 'fair' | 'overvalued'): string {
  switch (status) {
    case 'undervalued':
      return '低估';
    case 'fair':
      return '合理';
    case 'overvalued':
      return '高估';
  }
}

function getScoreClass(status: 'undervalued' | 'fair' | 'overvalued'): string {
  switch (status) {
    case 'undervalued':
      return styles.scoreUndervalued;
    case 'fair':
      return styles.scoreFair;
    case 'overvalued':
      return styles.scoreOvervalued;
  }
}

function getBadgeClass(status: 'undervalued' | 'fair' | 'overvalued'): string {
  switch (status) {
    case 'undervalued':
      return styles.badgeUndervalued;
    case 'fair':
      return styles.badgeFair;
    case 'overvalued':
      return styles.badgeOvervalued;
  }
}

export function ValuationScore() {
  const valuationScore = useDashboardStore((s) => s.valuationScore);

  if (!valuationScore) {
    return (
      <div className={styles.panel}>
        <h3 className={styles.panelTitle}>综合估值评分</h3>
        <div className={styles.loading}>加载中...</div>
      </div>
    );
  }

  return (
    <div className={styles.panel}>
      <h3 className={styles.panelTitle}>综合估值评分</h3>

      <div className={styles.scoreSection}>
        <div className={`${styles.scoreCircle} ${getScoreClass(valuationScore.status)}`}>
          {Math.round(valuationScore.overall)}
        </div>
        <div className={styles.scoreInfo}>
          <p className={styles.scoreLabel}>ETH 当前估值状态</p>
          <p className={styles.scoreStatus}>
            <span className={`${styles.badge} ${getBadgeClass(valuationScore.status)}`}>
              {getStatusLabel(valuationScore.status)}
            </span>
          </p>
        </div>
      </div>

      <div className={styles.chartSection}>
        <RadarChartComponent
          data={valuationScore.radarData}
          title="各维度评分"
          height={280}
        />
      </div>
    </div>
  );
}
