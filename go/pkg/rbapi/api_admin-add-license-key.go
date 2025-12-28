package rbapi

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"rslbot.com/go/pkg/errcode"
	"rslbot.com/go/pkg/rbdb"
)

func (svc *service) AdminAddLicenseKey(ctx context.Context, in *AdminAddLicenseKey_Input) (*AdminAddLicenseKey_Output, error) {
	if !isAdmin(ctx) {
		return nil, errcode.ERR_RESTRICTED_AREA
	}

	if in == nil || (in.UserId == 0 && in.UserEmail == "") || in.Duration == rbdb.LicenseKey_UNSPECIFIED {
		return nil, errcode.ERR_MISSING_INPUT
	}

	discourseUser, err := discourseUserFromContext(ctx)
	if err != nil {
		return nil, errcode.ERR_GET_USER_FROM_CTX.Wrap(err)
	}

	// Load the user from the database
	adminUser, err := svc.loadOrCreateUser(ctx, discourseUser)
	if err != nil {
		return nil, errcode.ERR_LOAD_OR_CREATE_USER.Wrap(err)
	}

	// Find the user by ID or email
	var userORM rbdb.UserORM
	var query *gorm.DB

	if in.UserId != 0 {
		// Use ID if provided
		query = svc.db.Where(&rbdb.UserORM{Id: in.UserId})
	} else {
		// Use email as fallback
		query = svc.db.Where(&rbdb.UserORM{Email: in.UserEmail})
	}

	err = query.First(&userORM).Error
	if err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	out := &AdminAddLicenseKey_Output{}
	err = svc.db.Transaction(func(tx *gorm.DB) error {
		payment, err := rbdb.DefaultCreatePayment(ctx, &rbdb.Payment{
			Provider:    rbdb.Payment_PROVIDER_MANUAL,
			ReferenceId: fmt.Sprintf("MANUAL-%d-%d", adminUser.Id, time.Now().UnixNano()),
			UserId:      adminUser.Id,
		}, svc.db)
		if err != nil {
			return rbdb.GormToErrcode(err)
		}

		out.LicenseKey, err = rbdb.GenerateLicense(svc.db, userORM.Id, payment.Id, in.Duration, in.Tier, false)
		if err != nil {
			return errcode.ERR_GENERATE_LICENSE.Wrap(err)
		}

		licenseCreationActivityORM := &rbdb.ActivityORM{
			Kind:         int32(rbdb.Activity_KIND_ADMIN_LICENSE_CREATION),
			UserId:       &adminUser.Id,
			LicenseKeyId: &out.LicenseKey.Id,
			PaymentId:    &payment.Id,
		}

		err = tx.Create(&licenseCreationActivityORM).Error
		if err != nil {
			return rbdb.GormToErrcode(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
