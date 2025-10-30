package rbapi

import (
	"context"
	"errors"

	"raidbot.app/go/pkg/errcode"
	"raidbot.app/go/pkg/rbdb"
)

// UserSyncDiscordRole syncs the user's Discord role based on their license status
func (svc *service) UserSyncDiscordRole(ctx context.Context, in *UserSyncDiscordRole_Input) (*UserSyncDiscordRole_Output, error) {
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

	// Check rate limit (30 attempts per hour)
	if err := svc.checkRateLimit(ctx, user.Id, rateLimitActionDiscordSync, rateLimitDiscordSync); err != nil {
		return nil, err
	}

	// Check if user has a lifetime license
	var licensesOrm []*rbdb.LicenseKeyORM
	err = svc.db.Where(&rbdb.LicenseKeyORM{
		UserId: user.Id,
	}).Find(&licensesOrm).Error
	if err != nil {
		return nil, rbdb.GormToErrcode(err)
	}

	hasLifetime := false
	for _, license := range licensesOrm {
		if license.Duration == int32(rbdb.LicenseKey_LIFETIME) && !license.Revoked {
			hasLifetime = true
			break
		}
	}

	if !hasLifetime {
		return &UserSyncDiscordRole_Output{
			Success:            false,
			Message:            "No lifetime license found for your account",
			HasLifetimeLicense: false,
			DiscordLinked:      false,
			RoleAssigned:       false,
		}, nil
	}

	// Get Discord ID from Discourse
	discordID, discourseErr := GetDiscordIDFromDiscourse(ctx, discourseUser.ExternalId)
	if discourseErr != nil {
		// If we can't fetch from Discourse, still return a structured response
		// The error likely means Discord isn't linked or there's an API issue
		//nolint:nilerr // Intentionally returning nil error to provide user-friendly response
		return &UserSyncDiscordRole_Output{
			Success:            false,
			Message:            "Unable to check Discord connection. Please ensure your Discord account is linked in your forum account settings under 'Associated Accounts'.",
			HasLifetimeLicense: true,
			DiscordLinked:      false,
			RoleAssigned:       false,
		}, nil
	}

	if discordID == "" {
		return &UserSyncDiscordRole_Output{
			Success:            false,
			Message:            "Discord account not linked to forum account. Please link your Discord account in your forum profile settings first.",
			HasLifetimeLicense: true,
			DiscordLinked:      false,
			RoleAssigned:       false,
		}, nil
	}

	// Assign Discord role
	err = AssignDiscordRole(ctx, discordID)
	if err != nil {
		// Check specific error types
		if errors.Is(err, errcode.ERR_DISCORD_USER_NOT_IN_GUILD) {
			return &UserSyncDiscordRole_Output{
				Success:            false,
				Message:            "You need to join our Discord server first before the role can be assigned.",
				HasLifetimeLicense: true,
				DiscordLinked:      true,
				RoleAssigned:       false,
			}, nil
		}
		if errors.Is(err, errcode.ERR_DISCORD_CONFIG_MISSING) {
			return &UserSyncDiscordRole_Output{
				Success:            false,
				Message:            "Discord integration is not configured. Please contact an administrator.",
				HasLifetimeLicense: true,
				DiscordLinked:      true,
				RoleAssigned:       false,
			}, nil
		}
		return nil, err
	}

	return &UserSyncDiscordRole_Output{
		Success:            true,
		Message:            "Discord lifetimer role successfully assigned!",
		HasLifetimeLicense: true,
		DiscordLinked:      true,
		RoleAssigned:       true,
	}, nil
}
