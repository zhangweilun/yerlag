package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DataFetcherConfig holds configuration for the DataFetcher.
type DataFetcherConfig struct {
	MaxRetries     int           `json:"maxRetries"`     // Maximum retry attempts, default 3
	RetryDelay     time.Duration `json:"retryDelay"`     // Base retry delay, default 1s
	RequestTimeout time.Duration `json:"requestTimeout"` // Per-request timeout
}

// FetchResult holds the result of a data fetch operation.
type FetchResult struct {
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`          // "live" | "cache"
	Timestamp int64       `json:"timestamp"`       // Unix timestamp
	Error     string      `json:"error,omitempty"` // Error message if any
}

// FetchRequest represents a single fetch request in a batch operation.
type FetchRequest struct {
	Key     string
	TTL     time.Duration
	Fetcher func() (interface{}, error)
}

// DataFetcher defines the interface for fetching data with caching support.
type DataFetcher interface {
	// Fetch retrieves data, preferring cache. Falls back to the fetcher function on cache miss.
	Fetch(ctx context.Context, key string, ttl time.Duration, fetcher func() (interface{}, error)) (*FetchResult, error)

	// ForceFetch skips cache and always calls the fetcher, then updates cache.
	ForceFetch(ctx context.Context, key string, ttl time.Duration, fetcher func() (interface{}, error)) (*FetchResult, error)

	// InvalidateCache removes the specified key from the cache.
	InvalidateCache(ctx context.Context, key string) error

	// FetchBatch executes multiple fetch requests sequentially.
	FetchBatch(ctx context.Context, requests []FetchRequest) ([]*FetchResult, error)

	// IsSourceUnavailable returns true if the given key's data source has been
	// marked unavailable due to consecutive failures.
	IsSourceUnavailable(key string) bool

	// GetConsecutiveFailures returns the current consecutive failure count for a key.
	GetConsecutiveFailures(key string) int
}

// cacheEntry is the JSON-serializable wrapper stored in Redis.
type cacheEntry struct {
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

// redisDataFetcher is the concrete implementation of DataFetcher backed by Redis.
type redisDataFetcher struct {
	rdb    redis.Cmdable
	config DataFetcherConfig

	mu                  sync.RWMutex
	consecutiveFailures map[string]int  // key -> consecutive failure count
	unavailable         map[string]bool // key -> whether source is marked unavailable
}

const (
	defaultMaxRetries = 3
	defaultRetryDelay = 1 * time.Second
	// After this many consecutive failures, the source is marked unavailable.
	maxConsecutiveFailures = 3
)

// NewDataFetcher creates a new DataFetcher backed by the given Redis client.
func NewDataFetcher(rdb redis.Cmdable, cfg DataFetcherConfig) DataFetcher {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = defaultRetryDelay
	}
	return &redisDataFetcher{
		rdb:                 rdb,
		config:              cfg,
		consecutiveFailures: make(map[string]int),
		unavailable:         make(map[string]bool),
	}
}

// Fetch checks Redis cache first. On cache hit, returns cached data.
// On cache miss, calls the fetcher function with retry, stores the result, and returns it.
// If the fetcher fails, it attempts to return stale cached data if available.
func (f *redisDataFetcher) Fetch(ctx context.Context, key string, ttl time.Duration, fetcher func() (interface{}, error)) (*FetchResult, error) {
	// Try cache first
	cached, err := f.getFromCache(ctx, key)
	if err == nil && cached != nil {
		return &FetchResult{
			Data:      cached.Data,
			Source:    "cache",
			Timestamp: cached.Timestamp,
		}, nil
	}

	// Cache miss or error — fetch live data
	return f.fetchAndCache(ctx, key, ttl, fetcher)
}

// ForceFetch always calls the fetcher function, skipping cache lookup.
func (f *redisDataFetcher) ForceFetch(ctx context.Context, key string, ttl time.Duration, fetcher func() (interface{}, error)) (*FetchResult, error) {
	return f.fetchAndCache(ctx, key, ttl, fetcher)
}

// InvalidateCache deletes the specified key from Redis.
func (f *redisDataFetcher) InvalidateCache(ctx context.Context, key string) error {
	return f.rdb.Del(ctx, key).Err()
}

// FetchBatch executes multiple fetch requests sequentially and returns all results.
func (f *redisDataFetcher) FetchBatch(ctx context.Context, requests []FetchRequest) ([]*FetchResult, error) {
	results := make([]*FetchResult, len(requests))
	for i, req := range requests {
		result, err := f.Fetch(ctx, req.Key, req.TTL, req.Fetcher)
		if err != nil {
			results[i] = &FetchResult{
				Source:    "live",
				Timestamp: time.Now().Unix(),
				Error:     err.Error(),
			}
		} else {
			results[i] = result
		}
	}
	return results, nil
}

// IsSourceUnavailable returns true if the data source for the given key
// has been marked unavailable due to consecutive failures.
func (f *redisDataFetcher) IsSourceUnavailable(key string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.unavailable[key]
}

// GetConsecutiveFailures returns the current consecutive failure count for a key.
func (f *redisDataFetcher) GetConsecutiveFailures(key string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.consecutiveFailures[key]
}

// fetchAndCache calls the fetcher with retry logic, caches the result on success,
// and falls back to stale cache on failure.
func (f *redisDataFetcher) fetchAndCache(ctx context.Context, key string, ttl time.Duration, fetcher func() (interface{}, error)) (*FetchResult, error) {
	data, err := f.fetchWithRetry(ctx, fetcher)
	if err != nil {
		// Record failure
		f.recordFailure(key)

		// Try to return stale cached data as fallback
		cached, cacheErr := f.getFromCache(ctx, key)
		if cacheErr == nil && cached != nil {
			return &FetchResult{
				Data:      cached.Data,
				Source:    "cache",
				Timestamp: cached.Timestamp,
				Error:     fmt.Sprintf("fetch failed: %v; returning stale cache", err),
			}, nil
		}

		return nil, fmt.Errorf("fetch failed for key %q and no cached data available: %w", key, err)
	}

	// Success — reset failure tracking
	f.recordSuccess(key)

	// Store in cache
	if storeErr := f.storeInCache(ctx, key, data, ttl); storeErr != nil {
		// Cache store failure is non-fatal; log but continue
		_ = storeErr
	}

	return &FetchResult{
		Data:      data,
		Source:    "live",
		Timestamp: time.Now().Unix(),
	}, nil
}

// fetchWithRetry calls the fetcher function with exponential backoff retry.
// Retries up to MaxRetries times with delays of RetryDelay * 2^attempt (1s, 2s, 4s).
func (f *redisDataFetcher) fetchWithRetry(ctx context.Context, fetcher func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	for attempt := 0; attempt < f.config.MaxRetries; attempt++ {
		data, err := fetcher()
		if err == nil {
			return data, nil
		}
		lastErr = err

		// Don't sleep after the last attempt
		if attempt < f.config.MaxRetries-1 {
			delay := f.config.RetryDelay * (1 << uint(attempt)) // 1s, 2s, 4s
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return nil, fmt.Errorf("all %d retries exhausted: %w", f.config.MaxRetries, lastErr)
}

// getFromCache retrieves and deserializes a cache entry from Redis.
func (f *redisDataFetcher) getFromCache(ctx context.Context, key string) (*cacheEntry, error) {
	val, err := f.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Decode the raw JSON data back to interface{}
	var data interface{}
	if err := json.Unmarshal(entry.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}
	entry.Data, _ = json.Marshal(data)

	return &cacheEntry{
		Data:      entry.Data,
		Timestamp: entry.Timestamp,
	}, nil
}

// storeInCache serializes and stores data in Redis with the given TTL.
func (f *redisDataFetcher) storeInCache(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for cache: %w", err)
	}

	entry := cacheEntry{
		Data:      dataBytes,
		Timestamp: time.Now().Unix(),
	}

	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	return f.rdb.Set(ctx, key, string(entryBytes), ttl).Err()
}

// recordFailure increments the consecutive failure count for a key.
// If the count reaches maxConsecutiveFailures, the source is marked unavailable.
func (f *redisDataFetcher) recordFailure(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.consecutiveFailures[key]++
	if f.consecutiveFailures[key] >= maxConsecutiveFailures {
		f.unavailable[key] = true
	}
}

// recordSuccess resets the consecutive failure count and clears the unavailable flag.
func (f *redisDataFetcher) recordSuccess(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.consecutiveFailures[key] = 0
	f.unavailable[key] = false
}
