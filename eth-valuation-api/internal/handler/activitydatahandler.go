package handler

import (
	"net/http"

	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetActivityDataHandler returns the on-chain activity data handler.
func GetActivityDataHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement activity data logic in subsequent tasks
		httpx.OkJson(w, map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    map[string]interface{}{},
			"meta":    map[string]interface{}{},
		})
	}
}
