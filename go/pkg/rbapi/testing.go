package rbapi

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"

	"raidbot.app/go/pkg/rbdb"
)

const (
	TestToken     = "YWRtaW49ZmFsc2UmZW1haWw9ZXhpbGVkYm90cG9lJTQwZ21haWwuY29tJmV4dGVybmFsX2lkPTcmZ3JvdXBzPXRydXN0X2xldmVsXzAlMkN0cnVzdF9sZXZlbF8xJTJDdGVzdF91c2VycyZtb2RlcmF0b3I9ZmFsc2UmbmFtZT1UZXN0K1VzZXImbm9uY2U9dGVzdC1ub25jZS0xMjMmcmV0dXJuX3Nzb191cmw9aHR0cCUzQSUyRiUyRmxvY2FsaG9zdCUzQTgwODUlMkZjYWxsYmFjayZ1c2VybmFtZT10ZXN0"
	TestSignature = "90756c4de4f8382d77556873744c4d3d19e63e9a660d8214a1e4df894db62ab1"
)

func TestingService(t *testing.T, opts ServiceOpts) (Service, func()) {
	t.Helper()

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	// Use TestingSqliteDB directly for test database
	db, sfn := rbdb.TestingSqliteDB(t, opts.Logger)

	// Initialize test Redis
	redisStore, redisCleanup := TestingRedisStore(t)

	svc := &service{
		logger:    opts.Logger,
		db:        db,
		sfn:       sfn,
		redis:     redisStore,
		startedAt: time.Now(),
	}

	// Give the default test user (discourse ID 7) a lifetime license
	// This user is used by TestingSetContextToken
	var userOrm rbdb.UserORM
	err := db.Where(&rbdb.UserORM{DiscourseId: 7}).First(&userOrm).Error
	if err == nil {
		// User exists, give them a license
		payment := &rbdb.Payment{
			Provider:        rbdb.Payment_PROVIDER_MANUAL,
			ReferenceId:     "test-payment-default",
			AmountInCents:   900,
			Currency:        "eur",
			LicenseDuration: rbdb.LicenseKey_LIFETIME,
			UserId:          userOrm.Id,
		}
		createdPayment, err := rbdb.DefaultCreatePayment(context.Background(), payment, db)
		require.NoError(t, err)

		_, err = rbdb.GenerateLicense(db, userOrm.Id, createdPayment.Id, rbdb.LicenseKey_LIFETIME, true)
		require.NoError(t, err)
	}

	cleanup := func() {
		sqlDB, err := svc.db.DB()
		if err == nil {
			sqlDB.Close()
		}
		redisCleanup()
	}

	return svc, cleanup
}

// TestingServer creates a new server instance for testing
func TestingServer(t *testing.T, ctx context.Context, opts ServerOpts) (*Server, Service, func()) {
	t.Helper()

	// Setup database
	svc, svcCleanup := TestingService(t, ServiceOpts{Logger: opts.Logger})
	db := TestingSvcDB(t, svc)
	redis := TestingSvcRedis(t, svc)

	if opts.Bind == "" {
		opts.Bind = "127.0.0.1:0"
	}

	// Create new server
	server, err := NewServer(ctx, svc, db, redis, opts)
	assert.NoError(t, err)

	cleanup := func() {
		server.Close()
		svcCleanup()
	}

	// Start server in background
	go func() {
		if err := server.Run(); err != nil {
			opts.Logger.Warn("server shutdown", zap.Error(err))
		}
	}()

	return server, svc, cleanup
}

func TestingSvcDB(t *testing.T, svc Service) *gorm.DB {
	t.Helper()

	typed := svc.(*service)
	return typed.db
}

func TestingSvcRedis(t *testing.T, svc Service) *RedisStore {
	t.Helper()

	typed := svc.(*service)
	return typed.redis
}

func TestingClient(t *testing.T, address string) (ServiceClient, func()) {
	t.Helper()

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()
	assert.NoError(t, err)
	c := NewServiceClient(conn)

	cleanup := func() {
		conn.Close()
	}
	return c, cleanup
}

func TestingSetContextToken(ctx context.Context, t *testing.T) context.Context {
	t.Helper()

	// Create a mock userInfo based on the test token
	userInfo, err := VerifySSO(TestToken, TestSignature, DiscourseSecret)
	require.NoError(t, err)

	// Set it directly in context
	return context.WithValue(ctx, userInfoCtx, userInfo)
}

// Helper to test if a context has valid auth
func TestingHasValidAuth(ctx context.Context, t *testing.T) {
	t.Helper()
	md, ok := metadata.FromIncomingContext(ctx)
	require.True(t, ok, "context should have metadata")

	auth, ok := md["authorization"]
	require.True(t, ok, "should have authorization")
	require.NotEmpty(t, auth, "should have auth token")
}

func TestingRedisStore(t *testing.T) (*RedisStore, func()) {
	t.Helper()

	// Start miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err, "failed to start miniredis")

	// Create RedisStore with miniredis connection and testing flag
	store, err := NewRedisStore(RedisConfig{
		Addr:    mr.Addr(),
		Testing: true, // Set testing flag to skip configuration
	})
	require.NoError(t, err, "failed to create redis store")

	cleanup := func() {
		store.Close()
		mr.Close()
	}

	return store, cleanup
}

// Helper function to test SSO verification
func TestingVerifySSO(t *testing.T) {
	t.Helper()
	discourseUser, err := VerifySSO(TestToken, TestSignature, DiscourseSecret)
	require.NoError(t, err, "should verify SSO token")
	require.Equal(t, 7, discourseUser.ExternalId, "should have correct discourse id")
}
