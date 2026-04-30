package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"eth-valuation-api/internal/model"
	"eth-valuation-api/internal/svc"
	"eth-valuation-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// generateShareLinkReq is the request body for generating a share link.
type generateShareLinkReq struct {
	// DashboardState is the JSON-serialized dashboard state to save.
	DashboardState json.RawMessage `json:"dashboardState"`
	// SnapshotData is optional additional snapshot data.
	SnapshotData json.RawMessage `json:"snapshotData,omitempty"`
	// ExpiresInHours is how many hours the link should remain valid (default 72).
	ExpiresInHours int `json:"expiresInHours,omitempty"`
}

// generateShareLinkResp is the response for share link generation.
type generateShareLinkResp struct {
	// ID is the unique share link identifier.
	ID string `json:"id"`
	// ExpiresAt is when the share link expires.
	ExpiresAt time.Time `json:"expiresAt"`
}

// GenerateShareLinkHandler returns the share link generation handler.
// It accepts a POST with dashboard state JSON, saves to DB using ShareLinkModel,
// and returns the share link ID.
func GenerateShareLinkHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req generateShareLinkReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, types.ErrorResponse(400, "invalid request body: "+err.Error()))
			return
		}

		if len(req.DashboardState) == 0 {
			httpx.OkJson(w, types.ErrorResponse(400, "dashboardState is required"))
			return
		}

		// Generate a unique ID
		id, err := generateID()
		if err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, "failed to generate share link ID: "+err.Error()))
			return
		}

		// Default expiration: 72 hours
		expiresInHours := req.ExpiresInHours
		if expiresInHours <= 0 {
			expiresInHours = 72
		}
		expiresAt := time.Now().Add(time.Duration(expiresInHours) * time.Hour)

		// Build the model
		shareLink := model.ShareLinkModel{
			ID:             id,
			DashboardState: string(req.DashboardState),
			SnapshotData:   string(req.SnapshotData),
			ExpiresAt:      expiresAt,
		}

		// Save to database
		if err := ctx.DB.Create(&shareLink).Error; err != nil {
			httpx.OkJson(w, types.ErrorResponse(500, "failed to save share link: "+err.Error()))
			return
		}

		resp := types.SuccessResponse(generateShareLinkResp{
			ID:        id,
			ExpiresAt: expiresAt,
		}, types.Meta{
			LastUpdated: time.Now().Unix(),
			Source:      "live",
		})
		httpx.OkJson(w, resp)
	}
}

// generateID creates a random 16-byte hex string for use as a share link ID.
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
