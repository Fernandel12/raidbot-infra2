package rbapi

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"rslbot.com/go/pkg/errcode"
)

const (
	// Rate limits per user (very permissive)
	rateLimitPickitCreate  = 50  // 50 creates per hour
	rateLimitPickitUpdate  = 200 // 200 updates per hour
	rateLimitPickitLike    = 500 // 500 likes per hour
	rateLimitDiscordSync   = 30  // 30 Discord sync attempts per hour
	rateLimitWindowMinutes = 60  // 1 hour window
)

// checkRateLimit checks if a user has exceeded the rate limit for a specific action
func (svc *service) checkRateLimit(ctx context.Context, userId int64, action string, limit int) error {
	if svc.redis == nil || svc.redis.client == nil {
		// If Redis is not available, skip rate limiting
		return nil
	}

	key := fmt.Sprintf("ratelimit:%s:%d", action, userId)

	// Increment counter and get new value
	// Using INCR with TTL to implement sliding window rate limiting
	val, err := svc.redis.client.Incr(ctx, key).Result()
	if err != nil {
		// If there's an error, allow the request but log it
		svc.logger.Warn("Failed to increment rate limit counter",
			zap.Error(err),
			zap.String("key", key))
		return nil
	}

	// Set TTL on first increment
	if val == 1 {
		ttl := time.Duration(rateLimitWindowMinutes) * time.Minute
		if err := svc.redis.client.Expire(ctx, key, ttl).Err(); err != nil {
			svc.logger.Warn("Failed to set TTL on rate limit key",
				zap.Error(err),
				zap.String("key", key))
		}
	}

	// Check if limit exceeded
	if int(val) > limit {
		return errcode.ERR_RATE_LIMIT_EXCEEDED.Wrap(fmt.Errorf("rate limit exceeded for action %s: %d/%d", action, val, limit))
	}

	return nil
}

// Rate limit keys
const (
	rateLimitActionPickitCreate = "pickit:create"
	rateLimitActionPickitUpdate = "pickit:update"
	rateLimitActionPickitLike   = "pickit:like"
	rateLimitActionBuildCreate  = "build:create"
	rateLimitActionBuildUpdate  = "build:update"
	rateLimitActionBuildLike    = "build:like"
	rateLimitActionDiscordSync  = "discord:sync"
)
