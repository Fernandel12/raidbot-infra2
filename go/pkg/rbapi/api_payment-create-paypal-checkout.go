package rbapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/plutov/paypal/v4"
	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

// PaymentCreatePayPalCheckout implements the API endpoint for creating a PayPal checkout session
// It takes either a license_duration for new licenses or a renewal_key_id for renewals
func (svc *service) PaymentCreatePayPalCheckout(ctx context.Context, in *PaymentCreatePayPalCheckout_Input) (*PaymentCreatePayPalCheckout_Output, error) {
	// Get user info from context
	discourseUser, err := discourseUserFromContext(ctx)
	if err != nil {
		return nil, errcode.ERR_GET_USER_FROM_CTX.Wrap(err)
	}

	// Validate input parameters - need either duration or renewal key ID
	if in.RenewalKeyId == 0 && in.LicenseDuration == rbdb.LicenseKey_UNSPECIFIED {
		return nil, errcode.ERR_MISSING_INPUT.Wrap(fmt.Errorf("must provide either license duration or renewal key ID"))
	}

	// Flag to track if this is a renewal
	isRenewal := in.RenewalKeyId > 0
	var licenseToRenew *rbdb.LicenseKey
	var licenseDuration rbdb.LicenseKey_Duration

	// Handle renewal case - verify license
	if isRenewal {
		// Find the license by ID
		var licenseKeyORM rbdb.LicenseKeyORM
		if err := svc.db.
			Preload("User").
			Where(&rbdb.LicenseKeyORM{Id: in.RenewalKeyId}).
			First(&licenseKeyORM).
			Error; err != nil {
			if rbdb.IsRecordNotFoundError(err) {
				return nil, errcode.ERR_LICENSE_NOT_FOUND.Wrap(fmt.Errorf("license with ID %d not found", in.RenewalKeyId))
			}
			return nil, rbdb.GormToErrcode(err)
		}

		// Convert to protobuf
		pbLicenseKey, err := licenseKeyORM.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(err)
		}
		licenseToRenew = &pbLicenseKey

		// Validate the license is not revoked
		if licenseToRenew.Revoked {
			return nil, errcode.ERR_LICENSE_REVOKED.Wrap(fmt.Errorf("license with ID %d", in.RenewalKeyId))
		}

		// Make sure the license belongs to the current user
		if licenseToRenew.User.DiscourseId != discourseUser.ExternalId {
			return nil, errcode.ERR_AUTH_NO_PERMISSION.Wrap(fmt.Errorf("license %d", in.RenewalKeyId))
		}

		// Check if the license is expired
		expired := rbdb.IsLicenseExpired(licenseToRenew)
		if !expired {
			return nil, errcode.ERR_LICENSE_NOT_YET_EXPIRED.Wrap(fmt.Errorf("license: %d - %s", licenseToRenew.Id, licenseToRenew.Key))
		}

		// For renewals, use the same duration as the original license
		licenseDuration = licenseToRenew.Duration
	} else {
		// For new licenses, use the specified duration
		licenseDuration = in.LicenseDuration
	}

	// Get the price for the selected duration
	amountInCents := getPriceInCentsForDuration(licenseDuration)

	// Validate amount
	if amountInCents == PriceNotAvailable {
		return nil, errcode.ERR_PAYMENT_INVALID_DURATION_PRICING.Wrap(fmt.Errorf("duration: %s", licenseDuration.String()))
	}

	// Try loading from database
	user, err := svc.loadOrCreateUser(ctx, discourseUser)
	if err != nil {
		return nil, errcode.ERR_LOAD_OR_CREATE_USER.Wrap(err)
	}

	// Get human-friendly strings for the checkout
	name := GenerateLicenseDisplayName(licenseDuration, isRenewal, in.RenewalKeyId, false)

	// Create PayPal client
	ppClient, err := CreatePayPalClient(ctx)
	if err != nil {
		return nil, errcode.ERR_PAYMENT_CREATE_PAYPAL_OAUTH_TOKEN.Wrap(err)
	}

	// Create metadata map for PayPal
	metadata := map[string]string{
		"user_id":      fmt.Sprintf("%d", user.Id),
		"is_renewal":   strconv.FormatBool(isRenewal),
		"duration":     licenseDuration.String(),
		"sandbox_mode": strconv.FormatBool(paypalSandboxMode),
	}

	// If this is a renewal, include the license key ID
	if isRenewal && in.RenewalKeyId > 0 {
		metadata["license_id"] = fmt.Sprintf("%d", in.RenewalKeyId)
	}

	// Convert metadata to JSON string for custom_id
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(err)
	}

	// Format amount as string with 2 decimal places
	amountString := fmt.Sprintf("%.2f", float64(amountInCents)/100.0)

	// Create PayPal order
	order, err := ppClient.CreateOrder(
		ctx,
		paypal.OrderIntentCapture,
		[]paypal.PurchaseUnitRequest{
			{
				ReferenceID: fmt.Sprintf("ref_%d", user.Id),
				CustomID:    string(metadataJSON),
				Amount: &paypal.PurchaseUnitAmount{
					Value:    amountString,
					Currency: "EUR",
					Breakdown: &paypal.PurchaseUnitAmountBreakdown{
						ItemTotal: &paypal.Money{
							Value:    amountString,
							Currency: "EUR",
						},
					},
				},
				Items: []paypal.Item{
					{
						Name: name,
						UnitAmount: &paypal.Money{
							Value:    amountString,
							Currency: "EUR",
						},
						Quantity: "1",
						Category: paypal.ItemCategoryDigitalGood,
						ImageURL: "https://raidbot.app/images/eb2avatar.png",
					},
				},
			},
		},
		nil,
		&paypal.ApplicationContext{
			ReturnURL:          paypalSuccessURL,
			CancelURL:          paypalCancelURL,
			UserAction:         paypal.UserActionPayNow,
			ShippingPreference: paypal.ShippingPreferenceNoShipping,
			LandingPage:        "BILLING",
		},
	)

	if err != nil {
		return nil, errcode.ERR_PAYMENT_CREATE_PAYPAL_CHECKOUT_SESSION.Wrap(err)
	}

	// Find the approval URL
	var approvalURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	if approvalURL == "" {
		return nil, errcode.ERR_PAYMENT_PAYPAL_APPROVAL_URL_MISSING.Wrap(fmt.Errorf("order ID: %s", order.ID))
	}

	// Return the order info
	return &PaymentCreatePayPalCheckout_Output{
		OrderId:     order.ID,
		CheckoutUrl: approvalURL,
	}, nil
}
