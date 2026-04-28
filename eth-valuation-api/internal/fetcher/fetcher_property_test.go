package fetcher

// Feature: eth-valuation-dashboard, Property 13: 缓存存取往返正确性
//
// For any data object stored in the cache with a given TTL, fetching the same key
// before TTL expiry SHALL return the original data with source="cache". When the
// underlying API fetch fails and cached data exists, the system SHALL return the
// cached data with source="cache". When consecutive failures reach 3, the system
// SHALL flag the data source as unavailable.
//
// **Validates: Requirements 14.2, 14.4, 14.5**

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// newPropertyTestRedis creates a miniredis server and a go-redis client for property tests.
func newPropertyTestRedis(t *rapid.T) (*miniredis.Miniredis, redis.Cmdable) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return mr, rdb
}

// newPropertyTestFetcher creates a DataFetcher with fast retry delays for property tests.
func newPropertyTestFetcher(t *rapid.T, rdb redis.Cmdable) DataFetcher {
	t.Helper()
	return NewDataFetcher(rdb, DataFetcherConfig{
		MaxRetries:     3,
		RetryDelay:     1 * time.Millisecond,
		RequestTimeout: 5 * time.Second,
	})
}

// genCacheKey generates random cache key strings.
func genCacheKey() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		prefix := rapid.SampledFrom([]string{"price", "gas", "tvl", "market", "staking", "macro", "etf"}).Draw(t, "prefix")
		suffix := rapid.StringMatching(`[a-z0-9]{1,10}`).Draw(t, "suffix")
		return fmt.Sprintf("test:%s:%s", prefix, suffix)
	})
}

// genDataValue generates random data values suitable for JSON round-tripping.
func genDataValue() *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		choice := rapid.IntRange(0, 3).Draw(t, "dataType")
		switch choice {
		case 0:
			// string value
			return rapid.StringMatching(`[a-zA-Z0-9 ]{1,20}`).Draw(t, "strVal")
		case 1:
			// numeric value (float64 for JSON compatibility)
			return rapid.Float64Range(-1e9, 1e9).Draw(t, "numVal")
		case 2:
			// map value
			m := make(map[string]any)
			n := rapid.IntRange(1, 5).Draw(t, "mapSize")
			for i := range n {
				key := fmt.Sprintf("field_%d", i)
				m[key] = rapid.Float64Range(-1e6, 1e6).Draw(t, fmt.Sprintf("mapVal_%d", i))
			}
			return m
		default:
			// slice of floats
			n := rapid.IntRange(1, 5).Draw(t, "sliceSize")
			s := make([]any, n)
			for i := range n {
				s[i] = rapid.Float64Range(-1e6, 1e6).Draw(t, fmt.Sprintf("sliceVal_%d", i))
			}
			return s
		}
	})
}

// genTTL generates random TTL durations between 1 second and 1 hour.
func genTTL() *rapid.Generator[time.Duration] {
	return rapid.Custom(func(t *rapid.T) time.Duration {
		seconds := rapid.IntRange(1, 3600).Draw(t, "ttlSeconds")
		return time.Duration(seconds) * time.Second
	})
}

// normalizeViaJSON round-trips a value through JSON to normalize types
// (e.g., int -> float64, which is what JSON unmarshalling produces).
func normalizeViaJSON(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	err = json.Unmarshal(b, &out)
	return out, err
}

// TestProperty13_CacheRoundTrip verifies that for any random data stored via Fetch,
// a subsequent Fetch with the same key returns the same data with source="cache".
func TestProperty13_CacheRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mr, rdb := newPropertyTestRedis(t)
		_ = mr
		df := newPropertyTestFetcher(t, rdb)
		ctx := context.Background()

		key := genCacheKey().Draw(t, "key")
		data := genDataValue().Draw(t, "data")
		ttl := genTTL().Draw(t, "ttl")

		// First fetch: populates cache with live data
		result1, err := df.Fetch(ctx, key, ttl, func() (any, error) {
			return data, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "live", result1.Source)

		// Second fetch: should return cached data
		fetcherCalled := false
		result2, err := df.Fetch(ctx, key, ttl, func() (any, error) {
			fetcherCalled = true
			return nil, errors.New("should not be called")
		})
		require.NoError(t, err)
		assert.False(t, fetcherCalled, "fetcher should not be called on cache hit")
		assert.Equal(t, "cache", result2.Source)

		// Verify data integrity through JSON round-trip normalization
		expected, err := normalizeViaJSON(data)
		require.NoError(t, err)

		actual, err := normalizeViaJSON(result2.Data)
		require.NoError(t, err)

		assert.Equal(t, expected, actual, "cached data should match original data after JSON round-trip")
	})
}

// TestProperty13_TTLExpiry verifies that after TTL expires, the fetcher is called
// again and returns source="live".
func TestProperty13_TTLExpiry(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mr, rdb := newPropertyTestRedis(t)
		df := newPropertyTestFetcher(t, rdb)
		ctx := context.Background()

		key := genCacheKey().Draw(t, "key")
		data1 := genDataValue().Draw(t, "data1")
		data2 := genDataValue().Draw(t, "data2")
		// Use a short TTL range for expiry testing
		ttlSeconds := rapid.IntRange(1, 10).Draw(t, "ttlSeconds")
		ttl := time.Duration(ttlSeconds) * time.Second

		// First fetch: populates cache
		result1, err := df.Fetch(ctx, key, ttl, func() (any, error) {
			return data1, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "live", result1.Source)

		// Fast-forward past TTL expiry
		mr.FastForward(time.Duration(ttlSeconds+1) * time.Second)

		// After expiry, fetcher should be called again
		result2, err := df.Fetch(ctx, key, ttl, func() (any, error) {
			return data2, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "live", result2.Source, "after TTL expiry, source should be 'live'")

		// Verify the new data is returned
		expected, err := normalizeViaJSON(data2)
		require.NoError(t, err)
		actual, err := normalizeViaJSON(result2.Data)
		require.NoError(t, err)
		assert.Equal(t, expected, actual, "after TTL expiry, new data should be returned")
	})
}

// TestProperty13_FailureFallback verifies that when the fetcher fails but cache
// has data, the system returns cached data with source="cache".
func TestProperty13_FailureFallback(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		mr, rdb := newPropertyTestRedis(t)
		_ = mr
		df := newPropertyTestFetcher(t, rdb)
		ctx := context.Background()

		key := genCacheKey().Draw(t, "key")
		data := genDataValue().Draw(t, "data")
		ttl := genTTL().Draw(t, "ttl")

		// First fetch: populates cache
		_, err := df.Fetch(ctx, key, ttl, func() (any, error) {
			return data, nil
		})
		require.NoError(t, err)

		// ForceFetch with a failing fetcher: should fall back to cached data
		result, err := df.ForceFetch(ctx, key, ttl, func() (any, error) {
			return nil, errors.New("api failure")
		})
		require.NoError(t, err)
		assert.Equal(t, "cache", result.Source, "on fetch failure with cached data, source should be 'cache'")
		assert.Contains(t, result.Error, "fetch failed", "error field should indicate fetch failure")

		// Verify the cached data matches original
		expected, err := normalizeViaJSON(data)
		require.NoError(t, err)
		actual, err := normalizeViaJSON(result.Data)
		require.NoError(t, err)
		assert.Equal(t, expected, actual, "fallback data should match originally cached data")
	})
}

// TestProperty13_ConsecutiveFailures verifies that after 3 consecutive fetch failures,
// the system flags the data source as unavailable.
func TestProperty13_ConsecutiveFailures(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		_, rdb := newPropertyTestRedis(t)
		df := newPropertyTestFetcher(t, rdb)
		ctx := context.Background()

		key := genCacheKey().Draw(t, "key")

		failingFetcher := func() (any, error) {
			return nil, errors.New("persistent failure")
		}

		// Before any failures, source should not be unavailable
		assert.False(t, df.IsSourceUnavailable(key))
		assert.Equal(t, 0, df.GetConsecutiveFailures(key))

		// Fail 3 times (each Fetch triggers internal retries but counts as 1 failure)
		for i := range 3 {
			_, _ = df.Fetch(ctx, key, 5*time.Minute, failingFetcher)
			assert.Equal(t, i+1, df.GetConsecutiveFailures(key),
				"consecutive failures should increment after each failed Fetch")
		}

		// After 3 consecutive failures, source should be marked unavailable
		assert.True(t, df.IsSourceUnavailable(key),
			"source should be marked unavailable after 3 consecutive failures")
		assert.Equal(t, 3, df.GetConsecutiveFailures(key))
	})
}

// TestProperty13_SuccessResetsFailures verifies that a successful fetch after
// failures resets the consecutive failure counter and clears the unavailable flag.
func TestProperty13_SuccessResetsFailures(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		_, rdb := newPropertyTestRedis(t)
		df := newPropertyTestFetcher(t, rdb)
		ctx := context.Background()

		key := genCacheKey().Draw(t, "key")
		data := genDataValue().Draw(t, "data")
		// Generate a random number of failures (1-2) before success
		numFailures := rapid.IntRange(1, 2).Draw(t, "numFailures")

		failingFetcher := func() (any, error) {
			return nil, errors.New("transient failure")
		}

		// Accumulate some failures
		for range numFailures {
			_, _ = df.Fetch(ctx, key, 5*time.Minute, failingFetcher)
		}
		assert.Equal(t, numFailures, df.GetConsecutiveFailures(key))

		// Successful fetch should reset the counter
		result, err := df.Fetch(ctx, key, 5*time.Minute, func() (any, error) {
			return data, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "live", result.Source)
		assert.Equal(t, 0, df.GetConsecutiveFailures(key),
			"consecutive failures should be reset after successful fetch")
		assert.False(t, df.IsSourceUnavailable(key),
			"source should not be unavailable after successful fetch")
	})
}
