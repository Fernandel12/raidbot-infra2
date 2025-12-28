package rbapi

import (
	"context"

	"rslbot.com/go/pkg/errcode"
)

// UserLogout implements the UserLogout RPC method
func (svc *service) UserLogout(ctx context.Context, in *UserLogout_Input) (*UserLogout_Output, error) {
	// Get user info from context
	discourseUser, err := discourseUserFromContext(ctx)
	if err != nil {
		return nil, errcode.ERR_GET_USER_FROM_CTX.Wrap(err)
	}

	// Call Discourse API to log out the user
	err = LogoutUserFromDiscourse(ctx, discourseUser.ExternalId)
	if err != nil {
		return nil, errcode.ERR_API_LOGOUT.Wrap(err)
	}

	return &UserLogout_Output{
		Success: true,
	}, nil
}
