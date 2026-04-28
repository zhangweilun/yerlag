package middleware

import "net/http"

// CorsMiddleware returns an HTTP middleware that sets CORS headers on every
// response and handles preflight OPTIONS requests.
//
// For development it allows all origins. Restrict AllowedOrigins in
// production by replacing "*" with the actual frontend domain.
func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Preflight requests should return immediately with 204.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
