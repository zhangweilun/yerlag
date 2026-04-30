package handler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// exportCSVReq is the request body for CSV export.
type exportCSVReq struct {
	// Filename is the desired filename for the exported CSV (without extension).
	Filename string `json:"filename"`
	// Headers defines the column headers for the CSV.
	Headers []string `json:"headers"`
	// Records is a JSON array of objects to export as CSV rows.
	Records []map[string]interface{} `json:"records"`
}

// ExportCSVHandler returns the CSV export handler.
// It accepts a POST request with data records (JSON array), converts to CSV format,
// and returns as a downloadable file.
func ExportCSVHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req exportCSVReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid request body: "+err.Error()))
			return
		}

		if len(req.Records) == 0 {
			httpx.OkJson(w, types.ErrorResponse(400, "records cannot be empty"))
			return
		}

		// Determine headers: use provided headers or extract from first record
		headers := req.Headers
		if len(headers) == 0 {
			headerSet := make(map[string]struct{})
			for _, record := range req.Records {
				for k := range record {
					headerSet[k] = struct{}{}
				}
			}
			headers = make([]string, 0, len(headerSet))
			for k := range headerSet {
				headers = append(headers, k)
			}
			sort.Strings(headers)
		}

		// Build CSV content
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)

		// Write header row
		if err := writer.Write(headers); err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, "failed to write CSV headers: "+err.Error()))
			return
		}

		// Write data rows
		for _, record := range req.Records {
			row := make([]string, len(headers))
			for i, h := range headers {
				if val, ok := record[h]; ok {
					row[i] = fmt.Sprintf("%v", val)
				}
			}
			if err := writer.Write(row); err != nil {
				httpx.OkJson(w, types.ErrorResponse(500, "failed to write CSV row: "+err.Error()))
				return
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, "CSV write error: "+err.Error()))
			return
		}

		// Determine filename
		filename := req.Filename
		if filename == "" {
			filename = fmt.Sprintf("export_%d", time.Now().Unix())
		}

		// Set response headers for file download
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

// exportChartReq is the request body for chart export.
type exportChartReq struct {
	// ChartType is the type of chart to export (e.g., "line", "bar", "pie").
	ChartType string `json:"chartType"`
	// Format is the desired output format ("png" or "svg").
	Format string `json:"format"`
	// Width is the chart width in pixels.
	Width int `json:"width"`
	// Height is the chart height in pixels.
	Height int `json:"height"`
	// Data is the chart data configuration.
	Data interface{} `json:"data"`
	// Title is the chart title.
	Title string `json:"title"`
}

// exportChartResp is the response for chart export.
type exportChartResp struct {
	// Message indicates the export status.
	Message string `json:"message"`
	// Format is the requested output format.
	Format string `json:"format"`
	// Width is the requested width.
	Width int `json:"width"`
	// Height is the requested height.
	Height int `json:"height"`
	// Note provides additional information about the export capability.
	Note string `json:"note"`
}

// ExportChartHandler returns the chart export handler.
// It accepts a POST request with chart configuration. In production, this would
// use a headless browser to render the chart. Currently returns a placeholder response.
func ExportChartHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req exportChartReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid request body: "+err.Error()))
			return
		}

		if req.ChartType == "" {
			httpx.OkJson(w, types.ErrorResponse(400, "chartType is required"))
			return
		}

		// Default dimensions
		if req.Width <= 0 {
			req.Width = 800
		}
		if req.Height <= 0 {
			req.Height = 600
		}
		if req.Format == "" {
			req.Format = "png"
		}

		// Validate format
		if req.Format != "png" && req.Format != "svg" {
			httpx.OkJson(w, types.ErrorResponse(400, "format must be 'png' or 'svg'"))
			return
		}

		// In production, this would use a headless browser (e.g., chromedp) to render
		// the chart server-side. For now, return a placeholder response indicating
		// the server acknowledges the export request.
		resp := types.SuccessResponse(exportChartResp{
			Message: "chart export request accepted",
			Format:  req.Format,
			Width:   req.Width,
			Height:  req.Height,
			Note:    "server-side chart rendering requires headless browser integration",
		}, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
		})
		httpx.OkJson(w, resp)
	}
}
