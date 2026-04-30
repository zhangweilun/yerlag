package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"eth-valuation-api/internal/logic/alert"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetActiveAlertsHandler returns the active alerts handler.
func GetActiveAlertsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alertSvc := alert.NewAlertService(ctx.DB)
		alerts, err := alertSvc.GetActiveAlerts()
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		resp := types.SuccessResponse(alerts, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
			NextRefresh: time.Now().Add(30 * time.Second).Unix(),
		})
		httpx.OkJson(w, resp)
	}
}

// GetAlertHistoryHandler returns the alert history handler.
func GetAlertHistoryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		daysStr := r.URL.Query().Get("days")
		days := 30 // default to 30 days
		if daysStr != "" {
			if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
				days = d
			}
		}

		alertSvc := alert.NewAlertService(ctx.DB)
		alerts, err := alertSvc.GetAlertHistory(days)
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		resp := types.SuccessResponse(alerts, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
			NextRefresh: time.Now().Add(30 * time.Second).Unix(),
		})
		httpx.OkJson(w, resp)
	}
}

// createAlertRuleReq is the request body for creating an alert rule.
type createAlertRuleReq struct {
	MetricKey string          `json:"metricKey"`
	Condition json.RawMessage `json:"condition"`
	Threshold float64         `json:"threshold"`
	Severity  string          `json:"severity"`
	Enabled   bool            `json:"enabled"`
	Message   string          `json:"message"`
}

// CreateAlertRuleHandler returns the create alert rule handler.
func CreateAlertRuleHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createAlertRuleReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid request body: "+err.Error()))
			return
		}

		var condition alert.AlertCondition
		if err := json.Unmarshal(req.Condition, &condition); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid condition: "+err.Error()))
			return
		}

		rule := alert.AlertRule{
			MetricKey: req.MetricKey,
			Condition: condition,
			Threshold: req.Threshold,
			Severity:  req.Severity,
			Enabled:   req.Enabled,
			Message:   req.Message,
		}

		alertSvc := alert.NewAlertService(ctx.DB)
		if err := alertSvc.AddRule(rule); err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		resp := types.SuccessResponse(rule, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
		})
		httpx.OkJson(w, resp)
	}
}

// alertRulePathReq holds the path parameter for alert rule endpoints.
type alertRulePathReq struct {
	ID string `path:"id"`
}

// UpdateAlertRuleHandler returns the update alert rule handler.
func UpdateAlertRuleHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq alertRulePathReq
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid rule ID"))
			return
		}

		var req createAlertRuleReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid request body: "+err.Error()))
			return
		}

		var condition alert.AlertCondition
		if err := json.Unmarshal(req.Condition, &condition); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid condition: "+err.Error()))
			return
		}

		rule := alert.AlertRule{
			ID:        pathReq.ID,
			MetricKey: req.MetricKey,
			Condition: condition,
			Threshold: req.Threshold,
			Severity:  req.Severity,
			Enabled:   req.Enabled,
			Message:   req.Message,
		}

		alertSvc := alert.NewAlertService(ctx.DB)
		if err := alertSvc.UpdateRule(pathReq.ID, rule); err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		resp := types.SuccessResponse(rule, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
		})
		httpx.OkJson(w, resp)
	}
}

// DeleteAlertRuleHandler returns the delete alert rule handler.
func DeleteAlertRuleHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq alertRulePathReq
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid rule ID"))
			return
		}

		alertSvc := alert.NewAlertService(ctx.DB)
		if err := alertSvc.RemoveRule(pathReq.ID); err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, err.Error()))
			return
		}

		httpx.OkJson(w, types.SuccessResponse[any](nil, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
		}))
	}
}
