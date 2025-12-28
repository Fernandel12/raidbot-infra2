package rbapi

import (
	"context"

	"gorm.io/gorm"
	"rslbot.com/go/pkg/errcode"
	"rslbot.com/go/pkg/rbdb"
)

func (svc *service) UserGetSession(ctx context.Context, in *UserGetSession_Input) (*UserGetSession_Output, error) {
	discourseUser, err := discourseUserFromContext(ctx)
	if err != nil {
		return nil, errcode.ERR_GET_USER_FROM_CTX.Wrap(err)
	}

	// Try loading from database
	user, err := svc.loadOrCreateUser(ctx, discourseUser)
	if err != nil {
		return nil, errcode.ERR_LOAD_OR_CREATE_USER.Wrap(err)
	}

	return &UserGetSession_Output{
		User: user,
	}, nil
}

func (svc *service) loadOrCreateUser(ctx context.Context, userInfo *rbdb.DiscourseUser) (*rbdb.User, error) {
	// First try to find by discourse id
	var userOrm rbdb.UserORM
	err := svc.db.Where(&rbdb.UserORM{DiscourseId: userInfo.ExternalId}).First(&userOrm).Error
	// If the user exists, return it
	if err == nil {
		// Check if email or username need updating
		needsUpdate := false
		if userOrm.Email != userInfo.Email {
			userOrm.Email = userInfo.Email
			needsUpdate = true
		}
		if userOrm.Username != userInfo.Username {
			userOrm.Username = userInfo.Username
			needsUpdate = true
		}

		if needsUpdate {
			err = svc.db.Save(&userOrm).Error
			if err != nil {
				return nil, rbdb.GormToErrcode(err)
			}
		}

		// Convert to PB type
		pbUser, err := userOrm.ToPB(ctx)
		if err != nil {
			return nil, errcode.ERR_USER_PROTOBUF_CONVERSION.Wrap(err)
		}
		return &pbUser, nil
	}

	// Only proceed to creation if the error is "record not found"
	if !rbdb.IsRecordNotFoundError(err) {
		return nil, rbdb.GormToErrcode(err)
	}
	// Create new user in a transaction
	newUser := &rbdb.User{
		DiscourseId: userInfo.ExternalId,
		Email:       userInfo.Email,
		Username:    userInfo.Username,
	}

	var createdUser *rbdb.User
	err = svc.db.Transaction(func(tx *gorm.DB) error {
		// Try to create the user
		var err error
		createdUser, err = rbdb.DefaultCreateUser(ctx, newUser, tx)
		if err != nil {
			// Check if the error is due to duplicate key - another request might have created it
			return rbdb.GormToErrcode(err)
		}

		// Create associated activity
		activity := &rbdb.Activity{
			Kind: rbdb.Activity_KIND_USER_REGISTER,
			User: createdUser,
		}

		_, err = rbdb.DefaultCreateActivity(ctx, activity, tx)
		return rbdb.GormToErrcode(err)
	})
	if err != nil {
		return nil, err
	}

	// Load the freshly created user to return (to ensure we have all fields correctly populated)
	err = svc.db.Where(&rbdb.UserORM{DiscourseId: userInfo.ExternalId}).First(&userOrm).Error
	if err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	pbUser, err := userOrm.ToPB(ctx)
	if err != nil {
		return nil, errcode.ERR_USER_PROTOBUF_CONVERSION.Wrap(err)
	}

	return &pbUser, nil
}
