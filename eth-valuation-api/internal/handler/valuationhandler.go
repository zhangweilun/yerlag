package handler

import (
	"net/http"

	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetValuationHandler returns the valuation score handler.
func GetValuationHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement valuation logic in subsequent tasks
		httpx.OkJson(w, map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    map[string]interface{}{},
			"meta":    map[string]interface{}{},
		})
	}
}
