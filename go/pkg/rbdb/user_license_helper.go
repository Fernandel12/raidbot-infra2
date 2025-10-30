package rbdb

import (
	"context"

	"gorm.io/gorm"
)

// UserHasActiveLicense checks if a user has at least one active (non-expired, non-revoked) license
func UserHasActiveLicense(db *gorm.DB, userId int64) (bool, error) {
	// Find all licenses belonging to this user
	var licensesOrm []*LicenseKeyORM
	err := db.Where(&LicenseKeyORM{UserId: userId}).Find(&licensesOrm).Error
	if err != nil {
		return false, GormToErrcode(err)
	}

	// Check if at least one license is active
	for _, licenseOrm := range licensesOrm {
		licensePb, err := licenseOrm.ToPB(context.Background())
		if err != nil {
			continue // Skip this license if we can't convert it
		}

		// Skip revoked licenses
		if licensePb.Revoked {
			continue
		}

		// Check if license is not expired
		if !IsLicenseExpired(&licensePb) {
			return true, nil
		}
	}

	return false, nil
}
