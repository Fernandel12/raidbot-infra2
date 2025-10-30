package rbapi

import (
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
	"raidbot.app/go/pkg/rbdb"
)

const checkSecret = "BzkE4gdxc9z956v"

type CheckLicenseRequest struct {
	LicenseKey string `json:"license_key"`
	UsageID    string `json:"usage_id,omitempty"`
	Secret     string `json:"secret"`
	IP         string `json:"ip"`
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
			uses, err := rbdb.CheckLicense(db, req.LicenseKey, req.UsageID)
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
			response.Uses = uses
		} else {
			// Free tier check
			if req.UsageID == "" {
				response.Status = faultString
				response.FaultString = "Missing usage id"
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				}
				return
			}

			// Validate free session
			err := redisStore.ValidateFreeSession(r.Context(), req.UsageID)
			if err != nil {
				response.Status = faultString
				response.FaultString = "Invalid or expired free tier session"
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				}
				return
			}

			response.Status = "ok"
			response.Uses = 1 // Always 1 for free tier
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
