package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRedis creates a miniredis server and a go-redis client connected to it.
func newTestRedis(t *testing.T) (*miniredis.Miniredis, redis.Cmdable) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return mr, rdb
}

// newTestFetcher creates a DataFetcher with short retry delays for fast tests.
func newTestFetcher(t *testing.T, rdb redis.Cmdable) DataFetcher {
	t.Helper()
	return NewDataFetcher(rdb, DataFetcherConfig{
		MaxRetries:     3,
		RetryDelay:     1 * time.Millisecond, // fast retries for tests
		RequestTimeout: 5 * time.Second,
	})
}

func TestFetch_CacheMiss_CallsFetcher(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	expected := map[string]interface{}{"price": float64(3500)}
	result, err := df.Fetch(ctx, "test:price", 5*time.Minute, func() (interface{}, error) {
		return expected, nil
	})

	require.NoError(t, err)
	assert.Equal(t, "live", result.Source)
	assert.NotZero(t, result.Timestamp)

	// Verify data round-trips through JSON correctly
	dataBytes, _ := json.Marshal(result.Data)
	var got map[string]interface{}
	json.Unmarshal(dataBytes, &got)
	assert.Equal(t, float64(3500), got["price"])
}

func TestFetch_CacheHit_ReturnsCachedData(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	callCount := 0
	fetcher := func() (interface{}, error) {
		callCount++
		return map[string]interface{}{"value": float64(42)}, nil
	}

	// First call — cache miss
	_, err := df.Fetch(ctx, "test:val", 5*time.Minute, fetcher)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call — should hit cache
	result, err := df.Fetch(ctx, "test:val", 5*time.Minute, fetcher)
	require.NoError(t, err)
	assert.Equal(t, "cache", result.Source)
	assert.Equal(t, 1, callCount) // fetcher not called again
}

func TestFetch_CacheExpired_CallsFetcherAgain(t *testing.T) {
	mr, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	callCount := 0
	fetcher := func() (interface{}, error) {
		callCount++
		return map[string]interface{}{"v": float64(callCount)}, nil
	}

	// First call
	_, err := df.Fetch(ctx, "test:expire", 1*time.Second, fetcher)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Fast-forward time in miniredis to expire the key
	mr.FastForward(2 * time.Second)

	// Second call — cache expired, should call fetcher again
	result, err := df.Fetch(ctx, "test:expire", 1*time.Second, fetcher)
	require.NoError(t, err)
	assert.Equal(t, "live", result.Source)
	assert.Equal(t, 2, callCount)
}

func TestForceFetch_SkipsCache(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	callCount := 0
	fetcher := func() (interface{}, error) {
		callCount++
		return map[string]interface{}{"n": float64(callCount)}, nil
	}

	// Populate cache
	_, err := df.Fetch(ctx, "test:force", 5*time.Minute, fetcher)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// ForceFetch should skip cache and call fetcher
	result, err := df.ForceFetch(ctx, "test:force", 5*time.Minute, fetcher)
	require.NoError(t, err)
	assert.Equal(t, "live", result.Source)
	assert.Equal(t, 2, callCount)
}

func TestInvalidateCache_RemovesKey(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	// Populate cache
	_, err := df.Fetch(ctx, "test:inv", 5*time.Minute, func() (interface{}, error) {
		return "data", nil
	})
	require.NoError(t, err)

	// Invalidate
	err = df.InvalidateCache(ctx, "test:inv")
	require.NoError(t, err)

	// Next fetch should be a cache miss
	callCount := 0
	_, err = df.Fetch(ctx, "test:inv", 5*time.Minute, func() (interface{}, error) {
		callCount++
		return "new-data", nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestFetchBatch_ReturnsAllResults(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	requests := []FetchRequest{
		{Key: "batch:1", TTL: 5 * time.Minute, Fetcher: func() (interface{}, error) { return "a", nil }},
		{Key: "batch:2", TTL: 5 * time.Minute, Fetcher: func() (interface{}, error) { return "b", nil }},
		{Key: "batch:3", TTL: 5 * time.Minute, Fetcher: func() (interface{}, error) { return nil, errors.New("fail") }},
	}

	results, err := df.FetchBatch(ctx, requests)
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Equal(t, "live", results[0].Source)
	assert.Equal(t, "live", results[1].Source)
	// Third request fails with no cache, so it should have an error
	assert.NotEmpty(t, results[2].Error)
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	var attempts int32
	fetcher := func() (interface{}, error) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return nil, errors.New("transient error")
		}
		return "success", nil
	}

	result, err := df.Fetch(ctx, "test:retry", 5*time.Minute, fetcher)
	require.NoError(t, err)
	assert.Equal(t, "live", result.Source)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetry_AllRetriesExhausted(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	fetcher := func() (interface{}, error) {
		return nil, errors.New("permanent error")
	}

	_, err := df.Fetch(ctx, "test:allretry", 5*time.Minute, fetcher)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "all 3 retries exhausted")
}

func TestFetch_FallbackToStaleCache_OnFetcherFailure(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	// Populate cache with initial data
	_, err := df.Fetch(ctx, "test:fallback", 5*time.Minute, func() (interface{}, error) {
		return map[string]interface{}{"old": true}, nil
	})
	require.NoError(t, err)

	// ForceFetch with a failing fetcher — should fall back to cached data
	result, err := df.ForceFetch(ctx, "test:fallback", 5*time.Minute, func() (interface{}, error) {
		return nil, errors.New("api down")
	})
	require.NoError(t, err)
	assert.Equal(t, "cache", result.Source)
	assert.Contains(t, result.Error, "fetch failed")
}

func TestConsecutiveFailures_MarksSourceUnavailable(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	failingFetcher := func() (interface{}, error) {
		return nil, errors.New("fail")
	}

	// Fail 3 times (each Fetch call triggers 3 retries internally, but counts as 1 failure)
	for i := 0; i < 3; i++ {
		_, _ = df.Fetch(ctx, "test:unavail", 5*time.Minute, failingFetcher)
	}

	assert.True(t, df.IsSourceUnavailable("test:unavail"))
	assert.Equal(t, 3, df.GetConsecutiveFailures("test:unavail"))
}

func TestConsecutiveFailures_ResetOnSuccess(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	failingFetcher := func() (interface{}, error) {
		return nil, errors.New("fail")
	}

	// Fail twice
	for i := 0; i < 2; i++ {
		_, _ = df.Fetch(ctx, "test:reset", 5*time.Minute, failingFetcher)
	}
	assert.Equal(t, 2, df.GetConsecutiveFailures("test:reset"))

	// Succeed
	_, err := df.Fetch(ctx, "test:reset", 5*time.Minute, func() (interface{}, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, 0, df.GetConsecutiveFailures("test:reset"))
	assert.False(t, df.IsSourceUnavailable("test:reset"))
}

func TestFetch_ContextCancellation_StopsRetry(t *testing.T) {
	_, rdb := newTestRedis(t)
	// Use longer retry delay so we can cancel during backoff
	df := NewDataFetcher(rdb, DataFetcherConfig{
		MaxRetries: 3,
		RetryDelay: 500 * time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())

	var attempts int32
	fetcher := func() (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, errors.New("fail")
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := df.Fetch(ctx, "test:cancel", 5*time.Minute, fetcher)
	require.Error(t, err)
	// Should have attempted at most 2 times (first attempt + cancelled during backoff)
	assert.LessOrEqual(t, atomic.LoadInt32(&attempts), int32(2))
}

func TestFetch_JSONRoundTrip_ComplexData(t *testing.T) {
	_, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	original := map[string]interface{}{
		"price":  float64(3500.50),
		"name":   "ethereum",
		"active": true,
		"tags":   []interface{}{"defi", "l1"},
		"nested": map[string]interface{}{"key": "value"},
	}

	// Store via Fetch
	_, err := df.Fetch(ctx, "test:complex", 5*time.Minute, func() (interface{}, error) {
		return original, nil
	})
	require.NoError(t, err)

	// Retrieve from cache
	result, err := df.Fetch(ctx, "test:complex", 5*time.Minute, func() (interface{}, error) {
		t.Fatal("fetcher should not be called on cache hit")
		return nil, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "cache", result.Source)

	// Verify data integrity through JSON round-trip
	dataBytes, _ := json.Marshal(result.Data)
	var got map[string]interface{}
	err = json.Unmarshal(dataBytes, &got)
	require.NoError(t, err)
	assert.Equal(t, float64(3500.50), got["price"])
	assert.Equal(t, "ethereum", got["name"])
	assert.Equal(t, true, got["active"])
}

func TestFetch_DifferentTTLs(t *testing.T) {
	mr, rdb := newTestRedis(t)
	df := newTestFetcher(t, rdb)
	ctx := context.Background()

	// Store with short TTL
	_, err := df.Fetch(ctx, "test:short", 1*time.Second, func() (interface{}, error) {
		return "short-lived", nil
	})
	require.NoError(t, err)

	// Store with long TTL
	_, err = df.Fetch(ctx, "test:long", 1*time.Hour, func() (interface{}, error) {
		return "long-lived", nil
	})
	require.NoError(t, err)

	// Fast-forward 2 seconds
	mr.FastForward(2 * time.Second)

	// Short TTL key should be expired
	callCount := 0
	result, err := df.Fetch(ctx, "test:short", 1*time.Second, func() (interface{}, error) {
		callCount++
		return "refreshed", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "live", result.Source)
	assert.Equal(t, 1, callCount)

	// Long TTL key should still be cached
	result, err = df.Fetch(ctx, "test:long", 1*time.Hour, func() (interface{}, error) {
		t.Fatal("should not be called")
		return nil, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "cache", result.Source)
}

func TestNewDataFetcher_DefaultConfig(t *testing.T) {
	_, rdb := newTestRedis(t)

	// Zero config should get defaults
	df := NewDataFetcher(rdb, DataFetcherConfig{})
	require.NotNil(t, df)

	impl := df.(*redisDataFetcher)
	assert.Equal(t, defaultMaxRetries, impl.config.MaxRetries)
	assert.Equal(t, defaultRetryDelay, impl.config.RetryDelay)
}
