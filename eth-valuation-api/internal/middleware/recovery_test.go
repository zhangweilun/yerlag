package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eth-valuation-api/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	handler := RecoveryMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"ok":true`)
}

func TestRecoveryMiddleware_CatchesPanic(t *testing.T) {
	handler := RecoveryMiddleware(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var resp types.APIResponse[any]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.Code)
	assert.Equal(t, "internal server error", resp.Message)
	assert.Nil(t, resp.Data)
}

func TestRecoveryMiddleware_CatchesNilPanic(t *testing.T) {
	handler := RecoveryMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var s *string
		// Dereference nil pointer to trigger a runtime panic.
		_ = *s
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp types.APIResponse[any]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.Code)
	assert.Equal(t, "internal server error", resp.Message)
}

func TestRecoveryMiddleware_ResponseFormat(t *testing.T) {
	handler := RecoveryMiddleware(func(w http.ResponseWriter, r *http.Request) {
		panic("test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	// Verify the response matches the APIResponse JSON structure exactly.
	var raw map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "code")
	assert.Contains(t, raw, "message")
	assert.Contains(t, raw, "data")
	assert.Contains(t, raw, "meta")

	assert.Equal(t, float64(500), raw["code"])
	assert.Equal(t, "internal server error", raw["message"])
	assert.Nil(t, raw["data"])
}
