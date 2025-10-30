//nolint:dupl
package rbapi

import (
	"context"

	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

// UserGetLicenses implements the UserGetLicenses RPC method
// It retrieves all licenses associated with the authenticated user
func (svc *service) UserGetLicenses(ctx context.Context, in *UserGetLicenses_Input) (*UserGetLicenses_Output, error) {
	// Get user info from context
	discourseUser, err := discourseUserFromContext(ctx)
	if err != nil {
		return nil, errcode.ERR_GET_USER_FROM_CTX.Wrap(err)
	}

	// Try loading from database
	user, err := svc.loadOrCreateUser(ctx, discourseUser)
	if err != nil {
		return nil, errcode.ERR_LOAD_OR_CREATE_USER.Wrap(err)
	}

	// Find all licenses belonging to this user
	var licensesOrm []*rbdb.LicenseKeyORM
	err = svc.db.Where(&rbdb.LicenseKeyORM{UserId: user.Id}).Find(&licensesOrm).Error
	if err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	// Convert licenses to protobuf types
	licenses := make([]*rbdb.LicenseKey, 0, len(licensesOrm))
	for _, licenseOrm := range licensesOrm {
		licensePb, err := licenseOrm.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(err)
		}
		licenses = append(licenses, &licensePb)
	}

	// Return the list of licenses
	return &UserGetLicenses_Output{
		Licenses: licenses,
	}, nil
}
