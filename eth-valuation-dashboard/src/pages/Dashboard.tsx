import { useEffect } from 'react';
import { useDashboardStore } from '../store/dashboard';
import { ErrorBoundary } from '../components/common/ErrorBoundary';
import { ThemeToggle } from '../components/common/ThemeToggle';
import { BurnDataPanel, GasDataPanel, ActivityPanel, TVLPanel, SupplyPanel } from '../components/onchain';
import { MarketOverview, PriceChart, ExchangeSpread } from '../components/market';
import { ValuationScore, ModelDetails } from '../components/valuation';
import { ETFPanel, GrayscalePanel, HoldingsPanel } from '../components/institutional';
import { StakingPanel, PerformancePanel } from '../components/network';
import { ETHBTCPanel, MacroIndicatorsPanel } from '../components/macro';
import { AlertBanner } from '../components/alert';
import { ShareDialog } from '../components/share';
import styles from './Dashboard.module.css';

export function Dashboard() {
  const startPolling = useDashboardStore((state) => state.startPolling);
  const stopPolling = useDashboardStore((state) => state.stopPolling);
  const refreshAll = useDashboardStore((state) => state.refreshAll);
  const isLoading = useDashboardStore((state) => state.isLoading);

  const marketData = useDashboardStore((state) => state.marketData);
  const valuationScore = useDashboardStore((state) => state.valuationScore);

  useEffect(() => {
    startPolling();
    return () => {
      stopPolling();
    };
  }, [startPolling, stopPolling]);

  const price = marketData?.currentPrice;
  const change24h = marketData?.priceChange24h;
  const rank = marketData?.marketCapRank;
  const score = valuationScore?.overall;
  const scoreStatus = valuationScore?.status;

  return (
    <div className={styles.dashboard}>
      <header className={styles.header}>
        <h1 className={styles.title}>ETH Valuation Dashboard</h1>
        <div className={styles.headerActions}>
          <button
            className={styles.refreshButton}
            onClick={() => void refreshAll()}
            disabled={isLoading['all'] === true}
            aria-label="Refresh all data"
          >
            {isLoading['all'] ? 'Refreshing...' : '↻ Refresh'}
          </button>
          <ShareDialog />
          <ThemeToggle />
        </div>
      </header>

      {/* Alert Banner */}
      <AlertBanner />

      {/* Top Summary Bar */}
      <div className={styles.summaryBar}>
        <div className={styles.summaryCard}>
          <p className={styles.summaryLabel}>ETH Price</p>
          <p className={styles.summaryValue}>
            {price != null ? `${price.toLocaleString()}` : '—'}
          </p>
        </div>
        <div className={styles.summaryCard}>
          <p className={styles.summaryLabel}>24h Change</p>
          <p
            className={`${styles.summaryValue} ${
              change24h != null
                ? change24h >= 0
                  ? styles.positive
                  : styles.negative
                : ''
            }`}
          >
            {change24h != null ? `${change24h >= 0 ? '+' : ''}${change24h.toFixed(2)}%` : '—'}
          </p>
        </div>
        <div className={styles.summaryCard}>
          <p className={styles.summaryLabel}>Market Cap Rank</p>
          <p className={styles.summaryValue}>
            {rank != null ? `#${rank}` : '—'}
          </p>
        </div>
        <div className={styles.summaryCard}>
          <p className={styles.summaryLabel}>Valuation Score</p>
          <p className={styles.summaryValue}>
            {score != null ? `${score.toFixed(0)}/100` : '—'}
            {scoreStatus && (
              <span style={{ fontSize: '0.75rem', marginLeft: '8px', color: 'var(--color-text-muted)' }}>
                ({scoreStatus})
              </span>
            )}
          </p>
        </div>
      </div>

      {/* Five Dimension Sections */}
      <div className={styles.sections}>
        <ErrorBoundary moduleName="On-Chain Data">
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>On-Chain Data</h2>
            <div className={styles.panelGrid}>
              <BurnDataPanel />
              <GasDataPanel />
              <ActivityPanel />
              <TVLPanel />
              <SupplyPanel />
            </div>
          </section>
        </ErrorBoundary>

        <ErrorBoundary moduleName="Market Data">
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Market Data</h2>
            <div className={styles.panelGrid}>
              <MarketOverview />
              <PriceChart />
              <ExchangeSpread />
            </div>
          </section>
        </ErrorBoundary>

        <ErrorBoundary moduleName="Valuation">
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Valuation</h2>
            <div className={styles.panelGrid}>
              <ValuationScore />
              <ModelDetails />
            </div>
          </section>
        </ErrorBoundary>

        <ErrorBoundary moduleName="Institutional Data">
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Institutional Data</h2>
            <div className={styles.panelGrid}>
              <ETFPanel />
              <GrayscalePanel />
              <HoldingsPanel />
            </div>
          </section>
        </ErrorBoundary>

        <ErrorBoundary moduleName="Network Health">
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Network Health</h2>
            <div className={styles.panelGrid}>
              <StakingPanel />
              <PerformancePanel />
            </div>
          </section>
        </ErrorBoundary>

        <ErrorBoundary moduleName="Macro Economy">
          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>Macro Economy</h2>
            <div className={styles.panelGrid}>
              <ETHBTCPanel />
              <MacroIndicatorsPanel />
            </div>
          </section>
        </ErrorBoundary>
      </div>
    </div>
  );
}
