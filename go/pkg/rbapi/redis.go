package rbapi

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"rslbot.com/go/pkg/errcode"
	"rslbot.com/go/pkg/rbdb"
)

type RedisConfig struct {
	Addr      string
	Password  string
	DB        int
	MaxMemory string
	Testing   bool // Add testing flag to skip configuration for tests
}

type RedisStore struct {
	client *redis.Client
	config RedisConfig
}

type RedisUserSession struct {
	UsageID    string    `json:"usage_id"`
	IsPaidTier bool      `json:"is_paid_tier"`
	LastSeen   time.Time `json:"last_seen"`
}

const (
	freeKeyPrefix     = "free:"
	defaultSessionTTL = 24 * time.Hour
	maxSessionAge     = 30 * 24 * time.Hour
)

func NewRedisStore(config RedisConfig) (*RedisStore, error) {
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, errcode.ERR_REDIS_CONNECTION_ERROR.Wrap(err)
	}

	store := &RedisStore{
		client: client,
		config: config,
	}

	// Only configure if not in testing mode
	if !config.Testing {
		if err := store.configure(); err != nil {
			client.Close()
			return nil, err
		}
	}

	go store.periodicCleanup()

	return store, nil
}

func (rs *RedisStore) configure() error {
	ctx := context.Background()
	configs := []struct {
		key, value string
	}{
		{"maxmemory", rs.config.MaxMemory},
		{"maxmemory-policy", "volatile-ttl"},
		{"appendonly", "yes"},
		{"appendfsync", "everysec"},
	}

	for _, cfg := range configs {
		if err := rs.client.ConfigSet(ctx, cfg.key, cfg.value).Err(); err != nil {
			return errcode.ERR_REDIS_CONFIG_ERROR.Wrap(err)
		}
	}
	return nil
}

func (rs *RedisStore) Close() error {
	return rs.client.Close()
}

func (rs *RedisStore) periodicCleanup() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		if err := rs.cleanup(context.Background()); err != nil {
			log.Printf("Redis cleanup error: %v", err)
		}
	}
}

func (rs *RedisStore) cleanup(ctx context.Context) error {
	var cursor uint64
	for {
		keys, newCursor, err := rs.client.Scan(ctx, cursor, freeKeyPrefix+"*", 100).Result()
		if err != nil {
			return errcode.ERR_REDIS_SCAN_ERROR.Wrap(err)
		}

		for _, key := range keys {
			ttl, err := rs.client.TTL(ctx, key).Result()
			if err != nil {
				continue
			}
			if ttl == -1 || ttl > maxSessionAge {
				rs.client.Expire(ctx, key, maxSessionAge)
			}
		}

		if newCursor == 0 {
			break
		}
		cursor = newCursor
	}
	return nil
}

func (rs *RedisStore) CreateRedisFreeSession(ctx context.Context) (string, error) {
	// Generate a new usage ID using the same function as license system
	usageID, err := rbdb.GenerateUsageID()
	if err != nil {
		return "", err
	}

	session := RedisUserSession{
		UsageID:    usageID,
		IsPaidTier: false,
		LastSeen:   time.Now().UTC(),
	}

	// Serialize session data
	sessionData, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	// Store in Redis with TTL
	key := freeKeyPrefix + usageID
	err = rs.client.Set(ctx, key, sessionData, defaultSessionTTL).Err()
	if err != nil {
		return "", err
	}

	return usageID, nil
}

func (rs *RedisStore) TrackPaidSession(ctx context.Context, usageID string) error {
	session := RedisUserSession{
		UsageID:    usageID,
		IsPaidTier: true,
		LastSeen:   time.Now().UTC(),
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// Store in Redis with TTL, using a different prefix for paid users
	key := "paid:" + usageID
	return rs.client.Set(ctx, key, sessionData, defaultSessionTTL).Err()
}

func (rs *RedisStore) ValidateFreeSession(ctx context.Context, usageID string) error {
	key := freeKeyPrefix + usageID

	// Check if session exists
	exists, err := rs.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return redis.Nil
	}

	// Update last seen time
	var session RedisUserSession
	sessionData, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	err = json.Unmarshal(sessionData, &session)
	if err != nil {
		return err
	}

	session.LastSeen = time.Now().UTC()
	sessionData, err = json.Marshal(session)
	if err != nil {
		return err
	}

	// Reset TTL and update last seen
	return rs.client.Set(ctx, key, sessionData, defaultSessionTTL).Err()
}

func (rs *RedisStore) GetActiveUsers(ctx context.Context) (map[string]int, error) {
	activeUsers := map[string]int{
		"free": 0,
		"paid": 0,
	}

	thirtyMinutesAgo := time.Now().UTC().Add(-30 * time.Minute)

	// Count free tier users
	freeUsers, err := rs.client.Keys(ctx, freeKeyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	// Filter free users by last seen
	for _, key := range freeUsers {
		sessionData, err := rs.client.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip if we can't read the session
		}

		var session RedisUserSession
		if err := json.Unmarshal(sessionData, &session); err != nil {
			continue // Skip if we can't parse the session
		}

		if session.LastSeen.After(thirtyMinutesAgo) {
			activeUsers["free"]++
		}
	}

	// Count paid tier users
	paidUsers, err := rs.client.Keys(ctx, "paid:*").Result()
	if err != nil {
		return nil, err
	}

	// Filter paid users by last seen
	for _, key := range paidUsers {
		sessionData, err := rs.client.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip if we can't read the session
		}

		var session RedisUserSession
		if err := json.Unmarshal(sessionData, &session); err != nil {
			continue // Skip if we can't parse the session
		}

		if session.LastSeen.After(thirtyMinutesAgo) {
			activeUsers["paid"]++
		}
	}

	return activeUsers, nil
}
