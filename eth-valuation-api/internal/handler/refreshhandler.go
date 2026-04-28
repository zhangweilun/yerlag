package handler

import (
	"net/http"

	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ForceRefreshHandler returns the force refresh handler.
func ForceRefreshHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement force refresh logic in subsequent tasks
		httpx.OkJson(w, map[string]interface{}{
			"code":    0,
			"message": "ok",
		})
	}
}
