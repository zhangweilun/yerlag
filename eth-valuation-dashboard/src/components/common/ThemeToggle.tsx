import { useDashboardStore } from '../../store/dashboard';

export function ThemeToggle() {
  const theme = useDashboardStore((state) => state.theme);
  const toggleTheme = useDashboardStore((state) => state.toggleTheme);

  return (
    <button
      onClick={toggleTheme}
      aria-label={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}
      title={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}
      style={{
        background: 'none',
        border: '1px solid var(--color-border)',
        borderRadius: '8px',
        padding: '8px 12px',
        cursor: 'pointer',
        color: 'var(--color-text-primary)',
        fontSize: '1.2rem',
        lineHeight: 1,
      }}
    >
      {theme === 'light' ? '🌙' : '☀️'}
    </button>
  );
}
