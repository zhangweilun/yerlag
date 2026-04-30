import type { CSSProperties } from 'react';

interface ExportButtonProps {
  /** Click handler to trigger the export */
  onClick: () => void;
  /** Label describing the export format (e.g., "PNG", "SVG", "CSV") */
  format: string;
  /** Optional title for the button tooltip */
  title?: string;
  /** Optional custom styles */
  style?: CSSProperties;
}

/**
 * A small button component for triggering data/chart exports.
 * Displays a download icon and the format label.
 */
export function ExportButton({ onClick, format, title, style }: ExportButtonProps) {
  return (
    <button
      onClick={onClick}
      aria-label={title ?? `Export as ${format}`}
      title={title ?? `Export as ${format}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        background: 'none',
        border: '1px solid var(--color-border, #ddd)',
        borderRadius: '4px',
        padding: '4px 8px',
        cursor: 'pointer',
        color: 'var(--color-text-secondary, #666)',
        fontSize: '0.75rem',
        lineHeight: 1,
        ...style,
      }}
    >
      <span aria-hidden="true">⬇</span>
      <span>{format}</span>
    </button>
  );
}
