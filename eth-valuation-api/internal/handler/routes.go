package handler

import (
	"net/http"

	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers registers all API route handlers with the go-zero server.
func RegisterHandlers(server *rest.Server, ctx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			// Overview
			{
				Method:  http.MethodGet,
				Path:    "/overview",
				Handler: GetOverviewHandler(ctx),
			},
			// On-Chain
			{
				Method:  http.MethodGet,
				Path:    "/onchain/burn",
				Handler: GetBurnDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/onchain/gas",
				Handler: GetGasDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/onchain/activity",
				Handler: GetActivityDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/onchain/tvl",
				Handler: GetTVLDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/onchain/supply",
				Handler: GetSupplyDataHandler(ctx),
			},
			// Market
			{
				Method:  http.MethodGet,
				Path:    "/market",
				Handler: GetMarketDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/market/price-history",
				Handler: GetPriceHistoryHandler(ctx),
			},
			// Valuation
			{
				Method:  http.MethodGet,
				Path:    "/valuation",
				Handler: GetValuationHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/valuation/dcf",
				Handler: GetDCFValuationHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/valuation/distribution/:metric",
				Handler: GetDistributionHandler(ctx),
			},
			// Institutional
			{
				Method:  http.MethodGet,
				Path:    "/institutional/etf",
				Handler: GetETFDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/institutional/grayscale",
				Handler: GetGrayscaleDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/institutional/holdings",
				Handler: GetInstitutionalHoldingsHandler(ctx),
			},
			// Network
			{
				Method:  http.MethodGet,
				Path:    "/network/staking",
				Handler: GetStakingDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/network/performance",
				Handler: GetNetworkPerformanceHandler(ctx),
			},
			// Macro
			{
				Method:  http.MethodGet,
				Path:    "/macro/ethbtc",
				Handler: GetETHBTCDataHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/macro/indicators",
				Handler: GetMacroIndicatorsHandler(ctx),
			},
			// Alerts
			{
				Method:  http.MethodGet,
				Path:    "/alerts",
				Handler: GetActiveAlertsHandler(ctx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/alerts/history",
				Handler: GetAlertHistoryHandler(ctx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/alerts/rules",
				Handler: CreateAlertRuleHandler(ctx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/alerts/rules/:id",
				Handler: UpdateAlertRuleHandler(ctx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/alerts/rules/:id",
				Handler: DeleteAlertRuleHandler(ctx),
			},
			// Export & Share
			{
				Method:  http.MethodPost,
				Path:    "/export/chart",
				Handler: ExportChartHandler(ctx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/export/csv",
				Handler: ExportCSVHandler(ctx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/share",
				Handler: GenerateShareLinkHandler(ctx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/refresh",
				Handler: ForceRefreshHandler(ctx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}
