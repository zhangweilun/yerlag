package handler

import (
	"net/http"

	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetTVLDataHandler returns the TVL data handler.
func GetTVLDataHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement TVL data logic in subsequent tasks
		httpx.OkJson(w, map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    map[string]interface{}{},
			"meta":    map[string]interface{}{},
		})
	}
}
