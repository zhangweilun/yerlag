import type { CSSProperties } from 'react';

interface LoadingSkeletonProps {
  /** Number of skeleton lines to render */
  lines?: number;
  /** Whether to show a spinner instead of skeleton lines */
  variant?: 'skeleton' | 'spinner';
  /** Optional height for the container */
  height?: string;
}

const pulseKeyframes = `
@keyframes skeleton-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}
@keyframes skeleton-spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
`;

const containerStyle: CSSProperties = {
  padding: '16px',
  display: 'flex',
  flexDirection: 'column',
  gap: '12px',
};

const lineStyle: CSSProperties = {
  height: '14px',
  borderRadius: '4px',
  backgroundColor: 'var(--color-bg-card, #2a2a3e)',
  animation: 'skeleton-pulse 1.5s ease-in-out infinite',
};

const spinnerContainerStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  padding: '32px',
};

const spinnerStyle: CSSProperties = {
  width: '28px',
  height: '28px',
  border: '3px solid var(--color-bg-card, #2a2a3e)',
  borderTopColor: 'var(--color-primary, #6366f1)',
  borderRadius: '50%',
  animation: 'skeleton-spin 0.8s linear infinite',
};

export function LoadingSkeleton({ lines = 3, variant = 'skeleton', height }: LoadingSkeletonProps) {
  const wrapperStyle: CSSProperties = height ? { ...containerStyle, height } : containerStyle;

  return (
    <>
      <style>{pulseKeyframes}</style>
      <div style={wrapperStyle} role="status" aria-label="Loading">
        {variant === 'spinner' ? (
          <div style={spinnerContainerStyle}>
            <div style={spinnerStyle} />
          </div>
        ) : (
          Array.from({ length: lines }, (_, i) => (
            <div
              key={i}
              style={{
                ...lineStyle,
                width: i === lines - 1 ? '60%' : '100%',
                animationDelay: `${i * 0.15}s`,
              }}
            />
          ))
        )}
      </div>
    </>
  );
}
