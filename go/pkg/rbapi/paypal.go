package rbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/plutov/paypal/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

var (
	PayPalClientID     string
	PayPalClientSecret string
	PayPalWebhookID    string
	paypalSandboxMode  bool
	paypalSuccessURL   string
	paypalCancelURL    string
	paypalAPIBase      string
)

// Hardcoded PayPal configuration for different environments
const (
	PayPalClientIDSandboxTest     = "AYNx89eBn4ZHTZuWSFm4Vs3YMdj_NCHb0IVzrbAapUx3PdIjhITkfEA82YY50dt6YZVjLtlehB09qiEZ"
	PayPalClientSecretSandboxTest = "EO9BFlVy-pJacLZLGGkqx36VwulYIaflxZF1K69oBbM4obepbrOYmnVbqieSMx9H8VJbZx2XcGTFCImz"
	PayPalWebhookIDSandboxTest    = "7ML47262H43557450"
	PayPalSuccessURLDev           = "http://localhost:5173/payment/success"
	PayPalSuccessURLProd          = "https://raidbot.app/payment/success"
	PayPalCancelURLDev            = "http://localhost:5173/purchase"
	PayPalCancelURLProd           = "https://raidbot.app/purchase"
)

// SetupPayPal initializes the PayPal configuration
func SetupPayPal() {
	if PayPalClientID == "" {
		PayPalClientID = PayPalClientIDSandboxTest
		PayPalClientSecret = PayPalClientSecretSandboxTest
		PayPalWebhookID = PayPalWebhookIDSandboxTest
		paypalSandboxMode = true
		paypalSuccessURL = PayPalSuccessURLDev
		paypalCancelURL = PayPalCancelURLDev
		paypalAPIBase = paypal.APIBaseSandBox
	} else {
		paypalSandboxMode = false
		paypalSuccessURL = PayPalSuccessURLProd
		paypalCancelURL = PayPalCancelURLProd
		paypalAPIBase = paypal.APIBaseLive
	}
}

// CreatePayPalClient returns a new PayPal client with OAuth token
func CreatePayPalClient(ctx context.Context) (*paypal.Client, error) {
	client, err := paypal.NewClient(PayPalClientID, PayPalClientSecret, paypalAPIBase)
	if err != nil {
		return nil, errcode.ERR_PAYMENT_CREATE_PAYPAL_OAUTH_TOKEN.Wrap(err)
	}

	// Get OAuth token
	_, err = client.GetAccessToken(ctx)
	if err != nil {
		return nil, errcode.ERR_PAYMENT_CREATE_PAYPAL_OAUTH_TOKEN.Wrap(err)
	}

	return client, nil
}

// paypalWebhookHandler handles incoming webhooks from PayPal
func paypalWebhookHandler(db *gorm.DB, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Info("Received PayPal webhook request", zap.String("path", r.URL.Path))

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Failed to read PayPal webhook body", zap.Error(err))
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Restore the body so it can be read again
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Verify the webhook signature
		paypalClient, err := CreatePayPalClient(r.Context())
		if err != nil {
			logger.Error("Failed to create PayPal client", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		verifyResponse, err := paypalClient.VerifyWebhookSignature(r.Context(), r, PayPalWebhookID)
		if err != nil || verifyResponse == nil || verifyResponse.VerificationStatus != "SUCCESS" {
			logger.Error("Invalid PayPal webhook signature", zap.Error(err))
			signatureErr := errcode.ERR_PAYMENT_PAYPAL_WEBHOOK_SIGNATURE_INVALID.Wrap(err)
			logger.Error("PayPal webhook signature verification failed", zap.Error(signatureErr))
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		// Parse the webhook event
		var event paypal.AnyEvent
		if err := json.Unmarshal(body, &event); err != nil {
			logger.Error("Failed to parse PayPal webhook event", zap.Error(err))
			parseErr := errcode.ERR_PAYMENT_PAYPAL_EVENT_PARSING_ERROR.Wrap(err)
			logger.Error("Failed to parse PayPal webhook", zap.Error(parseErr))
			http.Error(w, "Failed to parse event", http.StatusBadRequest)
			return
		}

		// Process the webhook event
		err = processPayPalWebhookEvent(ctx, event, db, logger)
		if err != nil {
			logger.Error("Error processing PayPal webhook", zap.Error(err), zap.String("event_type", event.EventType))
		}

		// Always return 200 OK to PayPal to acknowledge receipt
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error("Failed to write response", zap.Error(err))
		}
	}
}

// processPayPalWebhookEvent processes different PayPal webhook events
func processPayPalWebhookEvent(ctx context.Context, event paypal.AnyEvent, db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Processing PayPal webhook event", zap.String("event_type", event.EventType))

	switch event.EventType {
	case paypal.EventCheckoutOrderApproved:
		return handleCheckoutOrderApproved(ctx, event, logger)
	case paypal.EventPaymentCaptureCompleted:
		return handlePaymentCaptureCompleted(ctx, event, db, logger)
	case "PAYMENT.CAPTURE.PENDING": // if pending, still consider that the payment was successful (paypal holding funds on seller's end)
		return handlePaymentCaptureCompleted(ctx, event, db, logger)
	case paypal.EventPaymentCaptureDenied:
		logger.Info("Payment capture denied", zap.String("event_id", event.ID))
	case paypal.EventPaymentCaptureRefunded:
		logger.Info("Payment capture refunded", zap.String("event_id", event.ID))
	default:
		// Only log important events
		logger.Debug("Received unhandled PayPal event", zap.String("event_type", event.EventType))
	}

	return nil
}

// handleCheckoutOrderApproved processes an approved order event and captures the payment
// This function only captures the payment, license creation happens in handlePaymentCaptureCompleted
func handleCheckoutOrderApproved(ctx context.Context, event paypal.AnyEvent, logger *zap.Logger) error {
	var orderData map[string]interface{}
	if err := json.Unmarshal(event.Resource, &orderData); err != nil {
		return errcode.ERR_PAYMENT_PAYPAL_EVENT_PARSING_ERROR.Wrap(err)
	}

	// Get the order ID
	orderID, ok := orderData["id"].(string)
	if !ok || orderID == "" {
		return errcode.ERR_PAYMENT_PAYPAL_ORDER_ID_MISSING.Wrap(fmt.Errorf("from event ID: %s", event.ID))
	}

	logger.Info("Processing PayPal order approval", zap.String("order_id", orderID))

	// Create PayPal client
	paypalClient, err := CreatePayPalClient(ctx)
	if err != nil {
		return errcode.ERR_PAYMENT_CREATE_PAYPAL_OAUTH_TOKEN.Wrap(err)
	}

	// Capture the payment
	captureResult, err := paypalClient.CaptureOrder(ctx, orderID, paypal.CaptureOrderRequest{})
	if err != nil {
		logger.Error("Failed to capture PayPal payment", zap.Error(err), zap.String("order_id", orderID))
		return errcode.ERR_PAYMENT_PAYPAL_ORDER_CAPTURE_FAILED.Wrap(err)
	}

	logger.Info("Successfully captured payment", zap.String("order_id", orderID), zap.String("status", captureResult.Status))
	return nil
}

// handlePaymentCaptureCompleted processes a successful payment capture
// This function is responsible for creating/renewing licenses after payment is captured
func handlePaymentCaptureCompleted(ctx context.Context, event paypal.AnyEvent, db *gorm.DB, logger *zap.Logger) error {
	// Extract the capture data
	var captureData map[string]interface{}
	if err := json.Unmarshal(event.Resource, &captureData); err != nil {
		return errcode.ERR_PAYMENT_PAYPAL_EVENT_PARSING_ERROR.Wrap(err)
	}

	// Get the capture ID from the resource
	captureID, ok := captureData["id"].(string)
	if !ok || captureID == "" {
		return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("missing capture ID in event %s", event.ID))
	}

	// Get links to find the order URL
	links, ok := captureData["links"].([]interface{})
	if !ok {
		return errcode.ERR_PAYMENT_PAYPAL_ORDER_LINKS_MISSING.Wrap(fmt.Errorf("capture ID: %s", captureID))
	}

	// Find the order ID from the links
	var orderID string
	for _, link := range links {
		linkMap, ok := link.(map[string]interface{})
		if !ok {
			continue
		}

		rel, ok := linkMap["rel"].(string)
		if !ok || rel != "up" {
			continue
		}

		href, ok := linkMap["href"].(string)
		if !ok {
			continue
		}

		// Extract order ID from URL
		parts := strings.Split(href, "/")
		if len(parts) > 0 {
			orderID = parts[len(parts)-1]
			break
		}
	}

	if orderID == "" {
		return errcode.ERR_PAYMENT_PAYPAL_ORDER_ID_MISSING.Wrap(fmt.Errorf("capture ID: %s", captureID))
	}

	logger.Info("Processing PayPal payment capture", zap.String("capture_id", captureID), zap.String("order_id", orderID))

	// Check if this payment has already been processed
	var existingPayment rbdb.PaymentORM
	err := db.Where(&rbdb.PaymentORM{ReferenceId: captureID}).First(&existingPayment).Error
	if err == nil {
		logger.Info("Payment already processed", zap.String("capture_id", captureID), zap.Int64("payment_id", existingPayment.Id))
		return nil
	} else if !rbdb.IsRecordNotFoundError(err) {
		return rbdb.GormToErrcode(err)
	}

	// Fetch the order details from PayPal
	paypalClient, err := CreatePayPalClient(ctx)
	if err != nil {
		return errcode.ERR_PAYMENT_CREATE_PAYPAL_OAUTH_TOKEN.Wrap(err)
	}

	order, err := paypalClient.GetOrder(ctx, orderID)
	if err != nil {
		return errcode.ERR_PAYMENT_RETRIEVE_PAYPAL_ORDER.Wrap(err)
	}

	// Extract metadata from order custom_id
	if len(order.PurchaseUnits) == 0 {
		return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("missing purchase units for order %s", orderID))
	}

	customID := order.PurchaseUnits[0].CustomID
	var metadata map[string]string
	if err := json.Unmarshal([]byte(customID), &metadata); err != nil {
		return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(err)
	}

	// Extract user ID
	userIDStr, hasUserID := metadata["user_id"]
	if !hasUserID || userIDStr == "" {
		return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("missing user_id for order %s", orderID))
	}

	userId, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return errcode.ERR_USER_ID_FROM_STRING_CONVERSION.Wrap(err)
	}

	// Find the user
	var userORM rbdb.UserORM
	err = db.Where(&rbdb.UserORM{Id: userId}).First(&userORM).Error
	if err != nil {
		if rbdb.IsRecordNotFoundError(err) {
			return errcode.ERR_USER_NOT_FOUND.Wrap(fmt.Errorf("%d", userId))
		}
		return rbdb.GormToErrcode(err)
	}

	user, err := userORM.ToPB(ctx)
	if err != nil {
		return errcode.ERR_USER_PROTOBUF_CONVERSION.Wrap(err)
	}

	// Determine if this is a renewal
	isRenewalStr := metadata["is_renewal"]
	isRenewal := isRenewalStr == "true"

	// Get sandbox mode
	sandboxModeStr := metadata["sandbox_mode"]
	sandboxMode := sandboxModeStr == "true"

	// Calculate amount in cents (PayPal uses string for amount)
	amountInCents := int64(0)
	if order.PurchaseUnits[0].Amount != nil {
		amountFloat, err := strconv.ParseFloat(order.PurchaseUnits[0].Amount.Value, 64)
		if err == nil {
			amountInCents = int64(amountFloat * 100)
		}
	}

	// Extract currency code
	currency := ""
	if order.PurchaseUnits[0].Amount != nil {
		currency = strings.ToLower(order.PurchaseUnits[0].Amount.Currency)
	}

	// Create the payment record
	payment := &rbdb.Payment{
		Provider:      rbdb.Payment_PROVIDER_PAYPAL,
		ReferenceId:   captureID,
		AmountInCents: amountInCents,
		Currency:      currency,
		SandboxMode:   sandboxMode,
		UserId:        user.Id,
	}

	if order.Payer != nil {
		// Get email
		if order.Payer.EmailAddress != "" {
			payment.BillingEmail = order.Payer.EmailAddress
		}

		// Get name
		if order.Payer.Name != nil {
			fullName := fmt.Sprintf("%s %s",
				order.Payer.Name.GivenName,
				order.Payer.Name.Surname)
			payment.BillingName = strings.TrimSpace(fullName)
		}
	}

	// Process the payment based on whether it's a renewal or new license
	if isRenewal {
		// Handle license renewal
		licenseIDStr, hasLicenseID := metadata["license_id"]
		if !hasLicenseID || licenseIDStr == "" {
			return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("missing license_id for renewal, order %s", orderID))
		}

		licenseID, err := strconv.ParseInt(licenseIDStr, 10, 64)
		if err != nil {
			return errcode.ERR_LICENSE_KEY_ID_FROM_STRING_CONVERSION.Wrap(err)
		}

		payment.IsRenewal = true

		// Process renewal
		err = db.Transaction(func(tx *gorm.DB) error {
			// Find the license to renew
			var licenseKeyORM rbdb.LicenseKeyORM
			if err := tx.Where(&rbdb.LicenseKeyORM{Id: licenseID}).First(&licenseKeyORM).Error; err != nil {
				if rbdb.IsRecordNotFoundError(err) {
					return errcode.ERR_LICENSE_NOT_FOUND.Wrap(fmt.Errorf("license with ID %d not found", licenseID))
				}
				return rbdb.GormToErrcode(err)
			}

			// Ensure license belongs to the right user
			if licenseKeyORM.UserId != userId {
				return errcode.ERR_AUTH_NO_PERMISSION.Wrap(fmt.Errorf("license %d", licenseID))
			}

			// Renew the license key
			pbLicenseKey, err := licenseKeyORM.ToPB(ctx)
			if err != nil {
				return errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(err)
			}

			// Create payment first
			payment.LicenseDuration = pbLicenseKey.Duration
			createdPayment, err := rbdb.DefaultCreatePayment(ctx, payment, tx)
			if err != nil {
				return rbdb.GormToErrcode(err)
			}

			updatedLicense, err := rbdb.RenewLicense(tx, licenseID, userId, createdPayment.Id, false)
			if err != nil {
				return err
			}

			// Create license purchase activity
			licenseActivityORM := &rbdb.ActivityORM{
				Kind:         int32(rbdb.Activity_KIND_PAYMENT_RECEIVED),
				UserId:       &user.Id,
				LicenseKeyId: &updatedLicense.Id,
				PaymentId:    &createdPayment.Id,
			}

			err = tx.Create(&licenseActivityORM).Error
			if err != nil {
				return rbdb.GormToErrcode(err)
			}

			logger.Info("License renewed via PayPal", zap.String("license_key", updatedLicense.Key), zap.Int64("payment_id", createdPayment.Id))
			return nil
		})

		if err != nil {
			logger.Error("Failed to process license renewal via PayPal", zap.Error(err), zap.String("capture_id", captureID))
			return err
		}
	} else {
		// Handle new license creation
		durationStr, hasDuration := metadata["duration"]
		if !hasDuration || durationStr == "" {
			return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("missing duration for order %s", orderID))
		}

		// Parse license duration
		value, exists := rbdb.LicenseKey_Duration_value[durationStr]
		if !exists {
			return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("invalid duration: %s", durationStr))
		}

		licenseDuration := rbdb.LicenseKey_Duration(value)
		if licenseDuration == rbdb.LicenseKey_UNSPECIFIED {
			return errcode.ERR_PAYMENT_PAYPAL_METADATA_ERROR.Wrap(fmt.Errorf("duration UNSPECIFIED for order %s", orderID))
		}

		payment.LicenseDuration = licenseDuration
		payment.IsRenewal = false

		// Process new license creation
		err = db.Transaction(func(tx *gorm.DB) error {
			// Create payment first
			createdPayment, err := rbdb.DefaultCreatePayment(ctx, payment, tx)
			if err != nil {
				return rbdb.GormToErrcode(err)
			}

			// Generate new license
			licenseKey, err := rbdb.GenerateLicense(tx, user.Id, createdPayment.Id, licenseDuration, true)
			if err != nil {
				return errcode.ERR_GENERATE_LICENSE.Wrap(err)
			}

			// Create license purchase activity
			licenseActivityORM := &rbdb.ActivityORM{
				Kind:         int32(rbdb.Activity_KIND_PAYMENT_RECEIVED),
				UserId:       &user.Id,
				LicenseKeyId: &licenseKey.Id,
				PaymentId:    &createdPayment.Id,
			}

			err = tx.Create(&licenseActivityORM).Error
			if err != nil {
				return rbdb.GormToErrcode(err)
			}

			logger.Info("New license generated via PayPal", zap.String("license_key", licenseKey.Key), zap.Int64("payment_id", createdPayment.Id))
			return nil
		})

		if err != nil {
			logger.Error("Failed to process new license creation via PayPal", zap.Error(err), zap.String("capture_id", captureID))
			return err
		}
	}

	logger.Info("PayPal payment processed successfully", zap.String("capture_id", captureID), zap.String("order_id", orderID))
	return nil
}
