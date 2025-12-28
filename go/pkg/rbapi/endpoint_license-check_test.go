package rbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rslbot.com/go/internal/testutil"
	"rslbot.com/go/pkg/rbdb"
)

func TestCheckLicense(t *testing.T) {
	logger := testutil.Logger(t)
	ctx := context.Background()
	server, svc, cleanup := TestingServer(t, ctx, ServerOpts{
		Logger: logger,
	})
	defer cleanup()
	db := TestingSvcDB(t, svc)

	httpClient := &http.Client{}

	urlActivate := fmt.Sprintf("http://%s/license/activate", server.ListenerAddr())
	urlCheck := fmt.Sprintf("http://%s/license/check", server.ListenerAddr())
	t.Run("successful check with valid usage ID", func(t *testing.T) {
		ctx = TestingSetContextToken(ctx, t)
		session, err := svc.UserGetSession(ctx, nil)
		require.NoError(t, err)
		payment := rbdb.TestingCreateTestPayment(t, db, session.User, rbdb.LicenseKey_ONE_MONTH)
		license, err := rbdb.GenerateLicense(db, session.User.Id, payment.Id, rbdb.LicenseKey_ONE_MONTH, rbdb.LicenseKey_TIER_PREMIUM, true)
		require.NoError(t, err)

		// First activate the license
		reqBodyActivate := ActivateLicenseRequest{
			Secret:     activateSecret,
			LicenseKey: license.Key,
		}
		body, err := json.Marshal(reqBodyActivate)
		require.NoError(t, err)
		req, err := http.NewRequest("POST", urlActivate, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var respData rbdb.LicenseResponse
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		assert.Equal(t, license.Uses+1, respData.Uses)
		assert.NotEmpty(t, respData.UsageID)

		// then perform a check
		reqBodyCheck := CheckLicenseRequest{
			Secret:     checkSecret,
			LicenseKey: license.Key,
			UsageID:    respData.UsageID,
		}
		body, err = json.Marshal(reqBodyCheck)
		require.NoError(t, err)
		req, err = http.NewRequest("POST", urlCheck, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err = httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		assert.Equal(t, "ok", respData.Status)
		assert.NotEmpty(t, respData.Timestamp)
	})

	t.Run("check with invalid license key", func(t *testing.T) {
		reqBody := CheckLicenseRequest{
			Secret:     checkSecret,
			LicenseKey: "invalid-key",
			UsageID:    "1",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", urlCheck, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respData rbdb.LicenseResponse
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, faultString, respData.Status)
		assert.NotEmpty(t, respData.FaultString)
		assert.NotEmpty(t, respData.Timestamp)
	})

	t.Run("wrong secret", func(t *testing.T) {
		reqBody := CheckLicenseRequest{
			Secret:  "wrong-secret",
			UsageID: "some-id",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", urlCheck, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Empty(t, string(respBody))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("malformed request", func(t *testing.T) {
		req, err := http.NewRequest("POST", urlCheck, bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
