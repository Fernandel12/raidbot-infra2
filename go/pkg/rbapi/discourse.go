package rbapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"rslbot.com/go/pkg/errcode"
	"rslbot.com/go/pkg/rbdb"
)

// TODO: add to environment
const (
	DiscourseAPIKey      = "8cd0974ce97eede11aa99f9b900cf607dae41f51cf11b0dcee7e4cdade362e26"
	DiscourseAPIUsername = "Fernandel"
	DiscoursePath        = "https://community.rslbot.com"
)

// VerifySSO verifies that the SSO payload was signed by our Discourse instance
func VerifySSO(sso, sig, secret string) (*rbdb.DiscourseUser, error) {
	// Verify signature first
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sso))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return nil, errcode.ERR_AUTH_INVALID_SSO_SIGNATURE
	}

	// Decode payload
	decodedBytes, err := base64.StdEncoding.DecodeString(sso)
	if err != nil {
		return nil, errcode.ERR_AUTH_INVALID_SSO_PAYLOAD.Wrap(err)
	}

	// Parse the query string
	params, err := url.ParseQuery(string(decodedBytes))
	if err != nil {
		return nil, errcode.ERR_AUTH_INVALID_SSO_FORMAT.Wrap(err)
	}

	// Extract user information
	groups := []string{}
	if groupsStr := params.Get("groups"); groupsStr != "" {
		groups = strings.Split(groupsStr, ",")
	}

	discourseId, err := strconv.ParseInt(params.Get("external_id"), 10, 64)
	if err != nil {
		return nil, errcode.ERR_DISCOURSE_ID_FROM_STRING_CONVERSION.Wrap(err)
	}

	userInfo := &rbdb.DiscourseUser{
		ExternalId: discourseId,
		Username:   params.Get("username"),
		Email:      params.Get("email"),
		Groups:     groups,
		Admin:      params.Get("admin") == "true",
	}

	// Validate required fields
	if userInfo.ExternalId == 0 || userInfo.Username == "" || userInfo.Email == "" {
		return nil, errcode.ERR_AUTH_MISSING_SSO_USER_INFO
	}

	return userInfo, nil
}

// VerifyTokenAndGetUser verifies a Discourse API token
func VerifyTokenAndGetUser(ctx context.Context, tokenString string) (*rbdb.DiscourseUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/session/current.json", DiscoursePath), nil)
	if err != nil {
		return nil, errcode.ERR_AUTH_DISCOURSE_REQUEST_ERROR.Wrap(err)
	}

	req.Header.Set("User-Api-Key", tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errcode.ERR_AUTH_DISCOURSE_API_ERROR.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errcode.ERR_AUTH_DISCOURSE_RESPONSE_ERROR.Wrap(
			fmt.Errorf("status %d", resp.StatusCode))
	}

	var response struct {
		CurrentUser struct {
			ID       int      `json:"id"`
			Username string   `json:"username"`
			Email    string   `json:"email"`
			Groups   []string `json:"groups"`
			Admin    bool     `json:"admin"`
		} `json:"current_user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errcode.ERR_AUTH_DISCOURSE_RESPONSE_ERROR.Wrap(err)
	}

	return &rbdb.DiscourseUser{
		ExternalId: int64(response.CurrentUser.ID),
		Username:   response.CurrentUser.Username,
		Email:      response.CurrentUser.Email,
		Groups:     response.CurrentUser.Groups,
		Admin:      response.CurrentUser.Admin,
	}, nil
}

// LogoutUser logs a user out from Discourse by their externalID
func LogoutUserFromDiscourse(ctx context.Context, externalID int64) error {
	// Create the request URL
	logoutURL := fmt.Sprintf("%s/admin/users/%d/log_out", DiscoursePath, externalID)

	// Create a new HTTP request - Discourse doesn't need form fields for this endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", logoutURL, nil)
	if err != nil {
		return errcode.ERR_AUTH_DISCOURSE_REQUEST_ERROR.Wrap(err)
	}

	// Set required headers - use simple Content-Type as in the curl example
	req.Header.Set("Content-Type", "multipart/form-data")
	req.Header.Set("Api-Key", DiscourseAPIKey)
	req.Header.Set("Api-Username", DiscourseAPIUsername)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errcode.ERR_AUTH_DISCOURSE_API_ERROR.Wrap(err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errcode.ERR_AUTH_DISCOURSE_LOGOUT_ERROR.Wrap(
			fmt.Errorf("status %d with body length %d", resp.StatusCode, len(body)))
	}

	return nil
}
