package rbapi

import (
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
	"rslbot.com/go/pkg/rbdb"
)

const (
	activateSecret = "LXwkAuZa9vfMsxW"
	faultString    = "fault"
)

type ActivateLicenseRequest struct {
	LicenseKey string `json:"license_key"`
	Secret     string `json:"secret"`
	IP         string `json:"ip"`
	Version    string `json:"version"`
}

// activateLicense handles license activation
func activateLicense(db *gorm.DB, redisStore *RedisStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ActivateLicenseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Secret != activateSecret {
			// Leave it blank so people don't have any clue what's happening
			return
		}

		var response rbdb.LicenseResponse
		response.Timestamp = time.Now().UTC().Format(time.RFC3339)

		if req.LicenseKey != "" {
			// Paid tier activation
			license, err := rbdb.ActivateLicense(db, req.LicenseKey)
			if err != nil {
				response.Status = faultString
				response.FaultString = err.Error()
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				}
				return
			}

			// Track paid session in Redis
			_ = redisStore.TrackPaidSession(r.Context(), license.ActiveUsageId)

			response.Status = "ok"
			response.UsageID = license.ActiveUsageId
			response.Uses = license.Uses
			response.LicenseType = rbdb.MapToClientLicenseType(license.Duration, license.Tier)

			// Fetch offsets for the version if provided
			if req.Version != "" {
				offsets, err := rbdb.GetOffsetByVersion(db, req.Version)
				if err == nil && offsets != nil {
					response.Offsets = offsets
				}
			}
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
