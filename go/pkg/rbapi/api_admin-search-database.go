package rbapi

import (
	"context"
	"strconv"

	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

func (svc *service) AdminSearchDatabase(ctx context.Context, in *AdminSearchDatabase_Input) (*AdminSearchDatabase_Output, error) {
	// Check admin permissions
	if !isAdmin(ctx) {
		return nil, errcode.ERR_RESTRICTED_AREA
	}

	if in == nil || in.SearchTerm == "" {
		return nil, errcode.ERR_MISSING_INPUT
	}

	searchTerm := in.SearchTerm
	output := &AdminSearchDatabase_Output{}

	// Try to parse as an integer for ID searches
	searchID, err := strconv.ParseInt(searchTerm, 10, 64)
	isIDSearch := err == nil

	// --------------------------------
	// Search Users
	// --------------------------------
	var usersOrm []*rbdb.UserORM

	// Use protobuf struct fields directly
	userQuery := svc.db.Where(&rbdb.UserORM{Email: searchTerm})

	if isIDSearch {
		userQuery = userQuery.Or(&rbdb.UserORM{Id: searchID})
		userQuery = userQuery.Or(&rbdb.UserORM{DiscourseId: searchID})
	}

	if err := userQuery.Find(&usersOrm).Error; err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	for _, userOrm := range usersOrm {
		userPb, err := userOrm.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_USER_PROTOBUF_CONVERSION.Wrap(err)
		}
		output.Users = append(output.Users, &userPb)
	}

	// --------------------------------
	// Search LicenseKeys
	// --------------------------------
	var licensesOrm []*rbdb.LicenseKeyORM

	licenseQuery := svc.db.Where(&rbdb.LicenseKeyORM{Key: searchTerm})
	licenseQuery = licenseQuery.Or(&rbdb.LicenseKeyORM{ActiveUsageId: searchTerm})

	if isIDSearch {
		licenseQuery = licenseQuery.Or(&rbdb.LicenseKeyORM{Id: searchID})
		licenseQuery = licenseQuery.Or(&rbdb.LicenseKeyORM{UserId: searchID})
	}

	if err := licenseQuery.Find(&licensesOrm).Error; err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	for _, licenseOrm := range licensesOrm {
		licensePb, err := licenseOrm.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(err)
		}
		output.LicenseKeys = append(output.LicenseKeys, &licensePb)
	}

	// --------------------------------
	// Search Payments
	// --------------------------------
	var paymentsOrm []*rbdb.PaymentORM

	paymentQuery := svc.db.Where(&rbdb.PaymentORM{ReferenceId: searchTerm})
	paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{BillingEmail: searchTerm})
	paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{BillingName: searchTerm})
	paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{Currency: searchTerm})

	if isIDSearch {
		paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{Id: searchID})
		paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{UserId: searchID})

		// Add foreign key searches for license key and subscription
		paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{LicenseKeyId: &searchID})
		paymentQuery = paymentQuery.Or(&rbdb.PaymentORM{SubscriptionId: &searchID})
	}

	if err := paymentQuery.Find(&paymentsOrm).Error; err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	for _, paymentOrm := range paymentsOrm {
		paymentPb, err := paymentOrm.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_PAYMENT_PROTOBUF_CONVERSION.Wrap(err)
		}

		// For LicenseKeyId, create a minimal LicenseKey object with just the ID
		if paymentOrm.LicenseKeyId != nil && paymentPb.LicenseKey == nil {
			paymentPb.LicenseKey = &rbdb.LicenseKey{
				Id: *paymentOrm.LicenseKeyId,
			}
		}

		// For SubscriptionId, create a minimal Subscription object with just the ID
		if paymentOrm.SubscriptionId != nil && paymentPb.Subscription == nil {
			paymentPb.Subscription = &rbdb.Subscription{
				Id: *paymentOrm.SubscriptionId,
			}
		}

		output.Payments = append(output.Payments, &paymentPb)
	}

	// --------------------------------
	// Search Subscriptions
	// --------------------------------
	var subscriptionsOrm []*rbdb.SubscriptionORM

	subscriptionQuery := svc.db.Where(&rbdb.SubscriptionORM{StripeSubscriptionId: searchTerm})
	subscriptionQuery = subscriptionQuery.Or(&rbdb.SubscriptionORM{StripeCustomerId: searchTerm})

	if isIDSearch {
		subscriptionQuery = subscriptionQuery.Or(&rbdb.SubscriptionORM{Id: searchID})
		subscriptionQuery = subscriptionQuery.Or(&rbdb.SubscriptionORM{UserId: searchID})
		subscriptionQuery = subscriptionQuery.Or(&rbdb.SubscriptionORM{LicenseKeyId: searchID})
	}

	if err := subscriptionQuery.Find(&subscriptionsOrm).Error; err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	for _, subscriptionOrm := range subscriptionsOrm {
		subscriptionPb, err := subscriptionOrm.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_SUBSCRIPTION_PROTOBUF_CONVERSION.Wrap(err)
		}
		output.Subscriptions = append(output.Subscriptions, &subscriptionPb)
	}

	return output, nil
}
