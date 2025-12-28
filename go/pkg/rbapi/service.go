package rbapi

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"rslbot.com/go/pkg/rbdb"
)

type Service interface {
	ServiceServer
	Close() error
	DB() *gorm.DB
	Redis() *RedisStore
}

type ServiceOpts struct {
	Logger             *zap.Logger
	DBUrn              string
	CORSAllowedOrigins string
	RedisConfig        RedisConfig
}

type service struct {
	UnimplementedServiceServer
	startedAt time.Time
	db        *gorm.DB
	logger    *zap.Logger
	sfn       *snowflake.Node
	redis     *RedisStore
}

func NewService(ctx context.Context, opts ServiceOpts) (Service, error) {
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	// Initialize database
	db, sfn, err := rbdb.InitDB(ctx, rbdb.DBConfig{
		Logger: opts.Logger.Named("db"),
		URN:    opts.DBUrn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis
	redis, err := NewRedisStore(opts.RedisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	svc := &service{
		startedAt: time.Now(),
		logger:    opts.Logger,
		db:        db,
		sfn:       sfn,
		redis:     redis,
	}

	return svc, nil
}

func (s *service) DB() *gorm.DB {
	return s.db
}

func (s *service) Redis() *RedisStore {
	return s.redis
}

func (s *service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	if err := s.redis.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}
	return nil
}

// Ensure service implements Service interface
var _ Service = (*service)(nil)
