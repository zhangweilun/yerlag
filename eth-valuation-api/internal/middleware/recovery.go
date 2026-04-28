package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"eth-valuation-api/internal/types"
)

// RecoveryMiddleware catches panics in downstream handlers and returns a
// structured JSON error response using the standard APIResponse format.
func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)

				resp := types.ErrorResponse(500, "internal server error")
				_ = json.NewEncoder(w).Encode(resp)
			}
		}()

		next(w, r)
	}
}
