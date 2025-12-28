package rbapi

import (
	"fmt"

	"rslbot.com/go/pkg/rbdb"
)

// Constants for payment providers
const (
	// Base URLs for redirects
	PaymentSuccessBaseURL = "http://localhost:8080/payment/success"
	PaymentCancelBaseURL  = "http://localhost:8080/payment/cancel"
	PriceNotAvailable     = -1
)

// getPriceInCentsForDuration determines the price based on license duration
// Centralized pricing for all payment providers
func getPriceInCentsForDuration(duration rbdb.LicenseKey_Duration) int64 {
	switch duration {
	case rbdb.LicenseKey_ONE_WEEK:
		return 950
	case rbdb.LicenseKey_ONE_MONTH:
		return 1900
	case rbdb.LicenseKey_SIX_MONTHS:
		return 7900
	case rbdb.LicenseKey_ONE_YEAR:
		return 12900
	case rbdb.LicenseKey_LIFETIME:
		return 19900
	default:
		return PriceNotAvailable
	}
}

// GenerateLicenseStrings creates human-friendly strings for license checkout
// Returns both a name and description for the Stripe product data
func GenerateLicenseDisplayName(duration rbdb.LicenseKey_Duration, isRenewal bool, renewalKeyId int64, isSubscription bool) string {
	// Format durations in a friendly way
	var durationText string

	switch duration {
	case rbdb.LicenseKey_LIFETIME:
		durationText = "Lifetime"
	case rbdb.LicenseKey_ONE_WEEK:
		durationText = "1-Week"
	case rbdb.LicenseKey_ONE_MONTH:
		durationText = "1-Month"
	case rbdb.LicenseKey_SIX_MONTHS:
		durationText = "6-Month"
	case rbdb.LicenseKey_ONE_YEAR:
		durationText = "1-Year"
	default:
		durationText = "Unknown Duration"
	}

	// Create primary name string (appears in the line item)
	var name string
	if isRenewal && renewalKeyId > 0 {
		name = fmt.Sprintf("EB2 - %s License Renewal", durationText)
	} else {
		name = fmt.Sprintf("EB2 - %s License", durationText)
	}

	return name
}
