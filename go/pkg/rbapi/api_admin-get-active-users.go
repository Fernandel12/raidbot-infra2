package rbapi

import (
	"context"

	"raidbot.app/go/pkg/errcode"
)

func (svc *service) AdminGetActiveUsers(ctx context.Context, in *AdminGetActiveUsers_Input) (*AdminGetActiveUsers_Output, error) {
	if !isAdmin(ctx) {
		return nil, errcode.ERR_RESTRICTED_AREA
	}

	// Get active users from Redis
	activeUsers, err := svc.redis.GetActiveUsers(ctx)
	if err != nil {
		return nil, errcode.ERR_REDIS_QUERY_ERROR.Wrap(err)
	}

	return &AdminGetActiveUsers_Output{
		FreeTier:   int32(activeUsers["free"]),
		PaidTier:   int32(activeUsers["paid"]),
		TotalUsers: int32(activeUsers["free"] + activeUsers["paid"]),
	}, nil
}
