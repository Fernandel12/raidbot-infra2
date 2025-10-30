package rbapi

import (
	"context"

	"gorm.io/gorm"
	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

func (svc *service) AdminRevokeLicense(ctx context.Context, in *AdminRevokeLicense_Input) (*AdminRevokeLicense_Output, error) {
	if !isAdmin(ctx) {
		return nil, errcode.ERR_RESTRICTED_AREA
	}

	if in == nil || in.Key == "" {
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

	// Create output object
	output := &AdminRevokeLicense_Output{}

	// Perform operations in a transaction
	err = svc.db.Transaction(func(tx *gorm.DB) error {
		// Load the license key
		var licenseKeyORM rbdb.LicenseKeyORM
		if err := tx.Where(&rbdb.LicenseKeyORM{Key: in.Key}).First(&licenseKeyORM).Error; err != nil {
			return rbdb.GormToErrcode(err)
		}

		// Check if already revoked
		if licenseKeyORM.Revoked {
			return errcode.ERR_LICENSE_ALREADY_REVOKED
		}

		// Revoke the license
		licenseKeyORM.Revoked = true

		// Update the license in the database
		if err := tx.Save(&licenseKeyORM).Error; err != nil {
			return rbdb.GormToErrcode(err)
		}

		// Get the updated license for the response
		updatedLicense, err := licenseKeyORM.ToPB(ctx)
		if err != nil {
			return errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(err)
		}

		licenseRevocationActivityORM := &rbdb.ActivityORM{
			Kind:         int32(rbdb.Activity_KIND_ADMIN_LICENSE_REVOCATION),
			UserId:       &adminUser.Id,
			LicenseKeyId: &updatedLicense.Id,
		}

		err = tx.Create(&licenseRevocationActivityORM).Error
		if err != nil {
			return rbdb.GormToErrcode(err)
		}

		output.LicenseKey = &updatedLicense

		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
