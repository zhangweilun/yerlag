import type { CSSProperties } from 'react';
import { useDashboardStore } from '../../store/dashboard';

interface ModuleStatusProps {
  /** Key used in the store's lastUpdated and errors records */
  moduleKey: string;
  /** Optional label for the module */
  label?: string;
}

const containerStyle: CSSProperties = {
  display: 'flex',
  flexWrap: 'wrap',
  alignItems: 'center',
  gap: '8px',
  padding: '8px 0',
  fontSize: '0.75rem',
  color: 'var(--color-text-muted, #888)',
};

const timestampStyle: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: '4px',
};

const errorStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '6px',
  padding: '6px 10px',
  borderRadius: '4px',
  backgroundColor: 'rgba(239, 68, 68, 0.1)',
  border: '1px solid rgba(239, 68, 68, 0.3)',
  color: 'var(--color-danger, #ef4444)',
  fontSize: '0.75rem',
};

const warningStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '6px',
  padding: '6px 10px',
  borderRadius: '4px',
  backgroundColor: 'rgba(245, 158, 11, 0.1)',
  border: '1px solid rgba(245, 158, 11, 0.3)',
  color: 'var(--color-warning, #f59e0b)',
  fontSize: '0.75rem',
};

function formatTimestamp(ts: number): string {
  const date = new Date(ts);
  const now = Date.now();
  const diffMs = now - ts;
  const diffSec = Math.floor(diffMs / 1000);

  if (diffSec < 60) return 'just now';
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
  return date.toLocaleString();
}

export function ModuleStatus({ moduleKey, label }: ModuleStatusProps) {
  const lastUpdated = useDashboardStore((state) => state.lastUpdated[moduleKey]);
  const error = useDashboardStore((state) => state.errors[moduleKey]);

  const hasError = Boolean(error);
  const isStale = lastUpdated != null && Date.now() - lastUpdated > 10 * 60 * 1000; // >10 min

  return (
    <div style={containerStyle}>
      {/* Last updated timestamp */}
      {lastUpdated != null && (
        <span style={timestampStyle}>
          {label ? `${label}: ` : ''}Last updated: {formatTimestamp(lastUpdated)}
        </span>
      )}

      {/* Error message */}
      {hasError && (
        <div style={errorStyle} role="alert">
          <span>⚠</span>
          <span>{error}</span>
        </div>
      )}

      {/* Data source unavailable warning (error + stale data) */}
      {hasError && isStale && (
        <div style={warningStyle} role="status">
          <span>⚡</span>
          <span>Data source unavailable — showing cached data</span>
        </div>
      )}
    </div>
  );
}
