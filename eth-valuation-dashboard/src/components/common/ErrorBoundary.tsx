import { Component, type ErrorInfo, type ReactNode } from 'react';

interface ErrorBoundaryProps {
  moduleName: string;
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error(`[ErrorBoundary] ${this.props.moduleName}:`, error, errorInfo);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render(): ReactNode {
    if (this.state.hasError) {
      return (
        <div
          role="alert"
          style={{
            padding: '24px',
            borderRadius: '8px',
            backgroundColor: 'var(--color-bg-card)',
            border: '1px solid var(--color-danger)',
            color: 'var(--color-text-primary)',
          }}
        >
          <h3 style={{ margin: '0 0 8px', color: 'var(--color-danger)' }}>
            {this.props.moduleName} - Error
          </h3>
          <p style={{ margin: '0 0 12px', color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>
            {this.state.error?.message ?? 'An unexpected error occurred.'}
          </p>
          <button
            onClick={this.handleRetry}
            style={{
              padding: '6px 14px',
              borderRadius: '4px',
              border: '1px solid var(--color-danger)',
              backgroundColor: 'transparent',
              color: 'var(--color-danger)',
              cursor: 'pointer',
              fontSize: '0.8rem',
            }}
          >
            Retry
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
