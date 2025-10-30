package rbapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"raidbot.app/go/pkg/rbdb"
)

// TestUser represents a test user with authentication info and license state
type TestUser struct {
	DiscourseID int64
	Username    string
	Email       string
	Token       string
	Signature   string
	User        *rbdb.User
	Licenses    []*rbdb.LicenseKey
	Description string
}

// GenerateTestSSOToken generates a valid SSO token and signature for a test user
func GenerateTestSSOToken(discourseID int64, username, email string) (token string, signature string) {
	// Build the SSO payload
	params := url.Values{}
	params.Set("admin", "false")
	params.Set("email", email)
	params.Set("external_id", fmt.Sprintf("%d", discourseID))
	params.Set("groups", "trust_level_0,trust_level_1,test_users")
	params.Set("moderator", "false")
	params.Set("name", username)
	params.Set("nonce", fmt.Sprintf("test-nonce-%d", time.Now().UnixNano()))
	params.Set("return_sso_url", "http://localhost:8085/callback")
	params.Set("username", username)

	// Encode the payload
	payload := params.Encode()
	token = base64.StdEncoding.EncodeToString([]byte(payload))

	// Generate the signature
	mac := hmac.New(sha256.New, []byte(DiscourseSecret))
	mac.Write([]byte(token))
	signature = hex.EncodeToString(mac.Sum(nil))

	return token, signature
}

// createTestUserWithLicense is a helper to create a test user with a specific license configuration
func createTestUserWithLicense(t *testing.T, svc Service, discourseID int64, userType string, duration rbdb.LicenseKey_Duration, amountInCents int64, description string) *TestUser {
	t.Helper()

	db := TestingSvcDB(t, svc)
	username := fmt.Sprintf("user_%s_%d", userType, discourseID)
	email := fmt.Sprintf("%s@test.com", username)

	// Generate SSO token
	token, signature := GenerateTestSSOToken(discourseID, username, email)

	// Create user in database
	user := &rbdb.User{
		DiscourseId: discourseID,
		Username:    username,
		Email:       email,
	}
	createdUser, err := rbdb.DefaultCreateUser(context.Background(), user, db)
	require.NoError(t, err)

	// Create payment and license
	payment := &rbdb.Payment{
		Provider:        rbdb.Payment_PROVIDER_MANUAL,
		ReferenceId:     fmt.Sprintf("test-payment-%s-%d", userType, discourseID),
		AmountInCents:   amountInCents,
		Currency:        "eur",
		LicenseDuration: duration,
		UserId:          createdUser.Id,
	}
	createdPayment, err := rbdb.DefaultCreatePayment(context.Background(), payment, db)
	require.NoError(t, err)

	license, err := rbdb.GenerateLicense(db, createdUser.Id, createdPayment.Id, duration, true)
	require.NoError(t, err)

	return &TestUser{
		DiscourseID: discourseID,
		Username:    username,
		Email:       email,
		Token:       token,
		Signature:   signature,
		User:        createdUser,
		Licenses:    []*rbdb.LicenseKey{license},
		Description: description,
	}
}

// CreateTestUserWithActiveLicense creates a test user with an active license
func CreateTestUserWithActiveLicense(t *testing.T, svc Service, discourseID int64) *TestUser {
	return createTestUserWithLicense(t, svc, discourseID, "active", rbdb.LicenseKey_ONE_MONTH, 900, "User with active monthly license")
}

// CreateTestUserWithoutLicense creates a test user without any license
func CreateTestUserWithoutLicense(t *testing.T, svc Service, discourseID int64) *TestUser {
	t.Helper()

	db := TestingSvcDB(t, svc)
	username := fmt.Sprintf("user_no_license_%d", discourseID)
	email := fmt.Sprintf("%s@test.com", username)

	// Generate SSO token
	token, signature := GenerateTestSSOToken(discourseID, username, email)

	// Create user in database
	user := &rbdb.User{
		DiscourseId: discourseID,
		Username:    username,
		Email:       email,
	}
	createdUser, err := rbdb.DefaultCreateUser(context.Background(), user, db)
	require.NoError(t, err)

	return &TestUser{
		DiscourseID: discourseID,
		Username:    username,
		Email:       email,
		Token:       token,
		Signature:   signature,
		User:        createdUser,
		Licenses:    []*rbdb.LicenseKey{},
		Description: "User without any license",
	}
}

// CreateTestUserWithExpiredLicense creates a test user with an expired license
func CreateTestUserWithExpiredLicense(t *testing.T, svc Service, discourseID int64) *TestUser {
	t.Helper()

	db := TestingSvcDB(t, svc)
	username := fmt.Sprintf("user_expired_%d", discourseID)
	email := fmt.Sprintf("%s@test.com", username)

	// Generate SSO token
	token, signature := GenerateTestSSOToken(discourseID, username, email)

	// Create user in database
	user := &rbdb.User{
		DiscourseId: discourseID,
		Username:    username,
		Email:       email,
	}
	createdUser, err := rbdb.DefaultCreateUser(context.Background(), user, db)
	require.NoError(t, err)

	// Create payment and expired license
	payment := &rbdb.Payment{
		Provider:        rbdb.Payment_PROVIDER_MANUAL,
		ReferenceId:     fmt.Sprintf("test-payment-expired-%d", discourseID),
		AmountInCents:   900,
		Currency:        "eur",
		LicenseDuration: rbdb.LicenseKey_ONE_DAY,
		UserId:          createdUser.Id,
	}
	createdPayment, err := rbdb.DefaultCreatePayment(context.Background(), payment, db)
	require.NoError(t, err)

	license, err := rbdb.GenerateLicense(db, createdUser.Id, createdPayment.Id, rbdb.LicenseKey_ONE_DAY, false)
	require.NoError(t, err)

	// Set EffectiveFrom to 10 days ago to make it expired
	expiredDate := time.Now().AddDate(0, 0, -10)
	license.EffectiveFrom = timestamppb.New(expiredDate)
	_, err = rbdb.DefaultStrictUpdateLicenseKey(context.Background(), license, db)
	require.NoError(t, err)

	return &TestUser{
		DiscourseID: discourseID,
		Username:    username,
		Email:       email,
		Token:       token,
		Signature:   signature,
		User:        createdUser,
		Licenses:    []*rbdb.LicenseKey{license},
		Description: "User with expired daily license",
	}
}

// CreateTestUserWithRevokedLicense creates a test user with a revoked license
func CreateTestUserWithRevokedLicense(t *testing.T, svc Service, discourseID int64) *TestUser {
	t.Helper()

	db := TestingSvcDB(t, svc)
	username := fmt.Sprintf("user_revoked_%d", discourseID)
	email := fmt.Sprintf("%s@test.com", username)

	// Generate SSO token
	token, signature := GenerateTestSSOToken(discourseID, username, email)

	// Create user in database
	user := &rbdb.User{
		DiscourseId: discourseID,
		Username:    username,
		Email:       email,
	}
	createdUser, err := rbdb.DefaultCreateUser(context.Background(), user, db)
	require.NoError(t, err)

	// Create payment and license
	payment := &rbdb.Payment{
		Provider:        rbdb.Payment_PROVIDER_MANUAL,
		ReferenceId:     fmt.Sprintf("test-payment-revoked-%d", discourseID),
		AmountInCents:   900,
		Currency:        "eur",
		LicenseDuration: rbdb.LicenseKey_ONE_MONTH,
		UserId:          createdUser.Id,
	}
	createdPayment, err := rbdb.DefaultCreatePayment(context.Background(), payment, db)
	require.NoError(t, err)

	license, err := rbdb.GenerateLicense(db, createdUser.Id, createdPayment.Id, rbdb.LicenseKey_ONE_MONTH, true)
	require.NoError(t, err)

	// Revoke the license
	license.Revoked = true
	_, err = rbdb.DefaultStrictUpdateLicenseKey(context.Background(), license, db)
	require.NoError(t, err)

	return &TestUser{
		DiscourseID: discourseID,
		Username:    username,
		Email:       email,
		Token:       token,
		Signature:   signature,
		User:        createdUser,
		Licenses:    []*rbdb.LicenseKey{license},
		Description: "User with revoked monthly license",
	}
}

// CreateTestUserWithLifetimeLicense creates a test user with a lifetime license
func CreateTestUserWithLifetimeLicense(t *testing.T, svc Service, discourseID int64) *TestUser {
	return createTestUserWithLicense(t, svc, discourseID, "lifetime", rbdb.LicenseKey_LIFETIME, 9900, "User with lifetime license")
}

// SetContextForTestUser creates a context with the test user's authentication
func SetContextForTestUser(ctx context.Context, t *testing.T, testUser *TestUser) context.Context {
	t.Helper()

	// Verify the token works
	userInfo, err := VerifySSO(testUser.Token, testUser.Signature, DiscourseSecret)
	require.NoError(t, err)
	require.Equal(t, testUser.DiscourseID, userInfo.ExternalId)

	// Set it in context
	return context.WithValue(ctx, userInfoCtx, userInfo)
}

// EnsureDefaultTestUserHasLicense ensures the default test user (discourse ID 7) has an active license
// This is needed for backward compatibility with existing tests
func EnsureDefaultTestUserHasLicense(t *testing.T, svc Service) {
	t.Helper()

	db := TestingSvcDB(t, svc)

	// Check if user exists, if not create it
	var userOrm rbdb.UserORM
	err := db.Where(&rbdb.UserORM{DiscourseId: 7}).First(&userOrm).Error
	if err != nil {
		// Create the user
		user := &rbdb.User{
			DiscourseId: 7,
			Username:    "test",
			Email:       "raidbotpoe@gmail.com",
		}
		createdUser, err := rbdb.DefaultCreateUser(context.Background(), user, db)
		require.NoError(t, err)
		userOrm.Id = createdUser.Id
	}

	// Check if user already has an active license
	hasLicense, err := rbdb.UserHasActiveLicense(db, userOrm.Id)
	require.NoError(t, err)

	if !hasLicense {
		// Create payment and license for the default test user
		payment := &rbdb.Payment{
			Provider:        rbdb.Payment_PROVIDER_MANUAL,
			ReferenceId:     "test-payment-default",
			AmountInCents:   900,
			Currency:        "eur",
			LicenseDuration: rbdb.LicenseKey_LIFETIME,
			UserId:          userOrm.Id,
		}
		createdPayment, err := rbdb.DefaultCreatePayment(context.Background(), payment, db)
		require.NoError(t, err)

		_, err = rbdb.GenerateLicense(db, userOrm.Id, createdPayment.Id, rbdb.LicenseKey_LIFETIME, true)
		require.NoError(t, err)
	}
}

// CreateTestUsersSet creates a standard set of test users with different license states
func CreateTestUsersSet(t *testing.T, svc Service) map[string]*TestUser {
	t.Helper()

	users := make(map[string]*TestUser)

	// Start with discourse ID 100 to avoid conflicts with existing test data
	users["active"] = CreateTestUserWithActiveLicense(t, svc, 100)
	users["no_license"] = CreateTestUserWithoutLicense(t, svc, 101)
	users["expired"] = CreateTestUserWithExpiredLicense(t, svc, 102)
	users["revoked"] = CreateTestUserWithRevokedLicense(t, svc, 103)
	users["lifetime"] = CreateTestUserWithLifetimeLicense(t, svc, 104)

	return users
}
