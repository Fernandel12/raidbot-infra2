package rbapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"gorm.io/gorm"
	"rslbot.com/go/pkg/rbdb"
)

type UpdateOffsetsRequest struct {
	Version string          `json:"version"`
	Offsets json.RawMessage `json:"offsets"`
}

type UpdateOffsetsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// updateOffsets handles offset updates for a specific version
// Requires Discourse SSO admin authentication
func updateOffsets(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "authorization required",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Extract token from Bearer header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify SSO token
		if !strings.HasPrefix(tokenString, "SSO_") {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "invalid token format, SSO token required",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Extract SSO and signature from the token
		parts := strings.Split(strings.TrimPrefix(tokenString, "SSO_"), ".")
		if len(parts) != 2 {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "invalid SSO token format",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		// URL decode the token part if needed
		decodedToken := parts[0]
		if unescaped, err := url.QueryUnescape(decodedToken); err == nil {
			decodedToken = unescaped
		}

		// Verify SSO signature and get user info
		userInfo, err := VerifySSO(decodedToken, parts[1], DiscourseSecret)
		if err != nil {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "invalid SSO token",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Check if user is admin
		if !userInfo.Admin {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "admin access required",
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Parse request body
		var req UpdateOffsetsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "invalid request body",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Validate required fields
		if req.Version == "" {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "version is required",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if len(req.Offsets) == 0 {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: "offsets is required",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Update offsets in database
		err = rbdb.UpdateOffset(db, req.Version, req.Offsets)
		if err != nil {
			response := UpdateOffsetsResponse{
				Status:  "error",
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := UpdateOffsetsResponse{
			Status:  "ok",
			Message: "offsets updated successfully",
		}
		json.NewEncoder(w).Encode(response)
	}
}
