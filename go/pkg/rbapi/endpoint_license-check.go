package rbapi

import (
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
	"rslbot.com/go/pkg/rbdb"
)

const checkSecret = "BzkE4gdxc9z956v"

type CheckLicenseRequest struct {
	LicenseKey string `json:"license_key"`
	UsageID    string `json:"usage_id,omitempty"`
	Secret     string `json:"secret"`
	IP         string `json:"ip"`
	Version    string `json:"version"`
}

// checkLicense handles license checking
func checkLicense(db *gorm.DB, redisStore *RedisStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CheckLicenseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Secret != checkSecret {
			// Leave it blank so people don't have any clue what's happening
			return
		}

		var response rbdb.LicenseResponse
		response.Timestamp = time.Now().UTC().Format(time.RFC3339)

		if req.LicenseKey != "" {
			// Paid tier check
			license, err := rbdb.CheckLicense(db, req.LicenseKey, req.UsageID)
			if err != nil {
				response.Status = faultString
				response.FaultString = err.Error()
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				}
				return
			}

			// Update last seen in Redis (for analytics)
			_ = redisStore.TrackPaidSession(r.Context(), req.UsageID)

			response.Status = "ok"
			response.Uses = license.Uses
			response.LicenseType = rbdb.MapToClientLicenseType(license.Duration, license.Tier)

			// Fetch offsets for the version if provided
			if req.Version != "" {
				offsets, err := rbdb.GetOffsetByVersion(db, req.Version)
				if err == nil && offsets != nil {
					response.Offsets = offsets
				}
			}
		} else {
			response.Status = faultString
			response.FaultString = "No license key provided"
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
