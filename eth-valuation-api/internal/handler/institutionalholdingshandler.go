package handler

import (
	"net/http"

	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetInstitutionalHoldingsHandler returns the institutional holdings data handler.
func GetInstitutionalHoldingsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement institutional holdings logic in subsequent tasks
		httpx.OkJson(w, map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    map[string]interface{}{},
			"meta":    map[string]interface{}{},
		})
	}
}
