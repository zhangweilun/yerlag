import { useState, useCallback } from 'react';
import { generateShareLink } from '../../api/client';
import { useDashboardStore } from '../../store/dashboard';
import type { ShareResponse } from '../../api/types';
import styles from './Share.module.css';

/**
 * Serializes the current dashboard state into a JSON string for sharing.
 */
function serializeDashboardState(): string {
  const state = useDashboardStore.getState();
  const snapshot = {
    theme: state.theme,
    lastUpdated: state.lastUpdated,
  };
  return JSON.stringify(snapshot);
}

/**
 * Restores dashboard state from a share link's state payload.
 */
export function restoreFromShareState(stateJson: string): void {
  try {
    const parsed = JSON.parse(stateJson) as { theme?: 'light' | 'dark' };
    const store = useDashboardStore.getState();
    if (parsed.theme && parsed.theme !== store.theme) {
      store.toggleTheme();
    }
  } catch {
    // Invalid state JSON, ignore
  }
}

export function ShareDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [shareData, setShareData] = useState<ShareResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleOpen = useCallback(async () => {
    setIsOpen(true);
    setIsLoading(true);
    setError(null);
    setShareData(null);
    setCopied(false);

    try {
      const dashboardState = serializeDashboardState();
      const response = await generateShareLink({ dashboardState });
      setShareData(response.data);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to generate share link';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleClose = useCallback(() => {
    setIsOpen(false);
    setShareData(null);
    setError(null);
    setCopied(false);
  }, []);

  const handleCopy = useCallback(async () => {
    if (!shareData?.url) return;
    try {
      await navigator.clipboard.writeText(shareData.url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for environments without clipboard API
      const input = document.querySelector<HTMLInputElement>(`.${styles.linkInput}`);
      if (input) {
        input.select();
        document.execCommand('copy');
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }
    }
  }, [shareData]);

  return (
    <>
      <button
        className={styles.shareButton}
        onClick={handleOpen}
        aria-label="Share dashboard"
      >
        Share
      </button>

      {isOpen && (
        <div className={styles.overlay} onClick={handleClose} role="dialog" aria-modal="true" aria-labelledby="share-dialog-title">
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <div className={styles.header}>
              <h2 id="share-dialog-title" className={styles.title}>Share Dashboard</h2>
              <button className={styles.closeButton} onClick={handleClose} aria-label="Close dialog">
                ×
              </button>
            </div>

            <div className={styles.content}>
              <p className={styles.description}>
                Generate a shareable link to the current dashboard state.
              </p>

              {isLoading && (
                <div className={styles.loading}>Generating share link...</div>
              )}

              {error && (
                <p className={styles.error}>{error}</p>
              )}

              {shareData && (
                <>
                  <div className={styles.linkContainer}>
                    <input
                      className={styles.linkInput}
                      value={shareData.url}
                      readOnly
                      aria-label="Share link URL"
                    />
                    <button
                      className={`${styles.copyButton} ${copied ? styles.copied : ''}`}
                      onClick={handleCopy}
                    >
                      {copied ? 'Copied!' : 'Copy'}
                    </button>
                  </div>
                  {shareData.expiresAt > 0 && (
                    <p className={styles.expiry}>
                      Link expires: {new Date(shareData.expiresAt * 1000).toLocaleDateString()}
                    </p>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
