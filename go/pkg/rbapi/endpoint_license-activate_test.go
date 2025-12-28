package rbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rslbot.com/go/internal/testutil"
	"rslbot.com/go/pkg/rbdb"
)

func TestActivateLicense(t *testing.T) {
	logger := testutil.Logger(t)
	ctx := context.Background()
	server, svc, cleanup := TestingServer(t, ctx, ServerOpts{
		Logger: logger,
	})
	defer cleanup()
	db := TestingSvcDB(t, svc)

	httpClient := &http.Client{}
	urlActivate := fmt.Sprintf("http://%s/license/activate", server.ListenerAddr())

	t.Run("wrong secret", func(t *testing.T) {
		reqBody := ActivateLicenseRequest{
			Secret: "blabla",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", urlActivate, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Read the response body
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// For wrong secret, we expect empty response body
		assert.Empty(t, string(respBody))

		// Status should still be 200 to not give any indication of what went wrong
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("successful paid activation", func(t *testing.T) {
		ctx = TestingSetContextToken(ctx, t)
		session, err := svc.UserGetSession(ctx, nil)
		require.NoError(t, err)
		payment := rbdb.TestingCreateTestPayment(t, db, session.User, rbdb.LicenseKey_ONE_MONTH)
		license, err := rbdb.GenerateLicense(db, session.User.Id, payment.Id, rbdb.LicenseKey_ONE_MONTH, rbdb.LicenseKey_TIER_PREMIUM, true)
		require.NoError(t, err)

		reqBody := ActivateLicenseRequest{
			Secret:     activateSecret,
			LicenseKey: license.Key,
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", urlActivate, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respData rbdb.LicenseResponse
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, "ok", respData.Status)
		assert.Equal(t, int64(1), respData.Uses)
		assert.NotEmpty(t, respData.UsageID)
		assert.NotEmpty(t, respData.Timestamp)
	})

	t.Run("invalid license key", func(t *testing.T) {
		reqBody := ActivateLicenseRequest{
			Secret:     activateSecret,
			LicenseKey: "invalid-key",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		req, err := http.NewRequest("POST", urlActivate, bytes.NewReader(body))
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

	t.Run("expired license activation", func(t *testing.T) {
		ctx = TestingSetContextToken(ctx, t)
		session, err := svc.UserGetSession(ctx, nil)
		require.NoError(t, err)

		// Create an expired license
		license := &rbdb.LicenseKey{
			Key:      "expired-test-key",
			Duration: rbdb.LicenseKey_ONE_MONTH,
			UserId:   session.User.Id,
		}

		// Create it with old date
		oldTime := time.Now().UTC().AddDate(0, -2, 0) // 2 months ago
		licenseOrm, err := license.ToORM(context.Background())
		require.NoError(t, err)
		licenseOrm.CreatedAt = &oldTime
		licenseOrm.EffectiveFrom = &oldTime
		err = db.Create(&licenseOrm).Error
		require.NoError(t, err)

		// Create a payment and link it to the license
		payment := rbdb.TestingCreateTestPayment(t, db, session.User, rbdb.LicenseKey_ONE_MONTH)
		paymentORM, err := payment.ToORM(context.Background())
		require.NoError(t, err)
		paymentORM.LicenseKeyId = &licenseOrm.Id
		err = db.Save(&paymentORM).Error
		require.NoError(t, err)

		// Try to activate expired license
		reqBody := ActivateLicenseRequest{
			Secret:     activateSecret,
			LicenseKey: license.Key,
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", urlActivate, bytes.NewReader(body))
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
		assert.Contains(t, respData.FaultString, "expired")
		assert.NotEmpty(t, respData.Timestamp)
	})

	t.Run("malformed request", func(t *testing.T) {
		req, err := http.NewRequest("POST", urlActivate, bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
