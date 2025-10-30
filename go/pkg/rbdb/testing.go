package rbdb

import (
	"context"
	fmt "fmt"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"moul.io/zapgorm2"
)

func TestingSqliteDB(t *testing.T, logger *zap.Logger) (*gorm.DB, *snowflake.Node) {
	t.Helper()

	zapGormLogger := zapgorm2.New(logger.Named("rb-gorm"))
	defaultGormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: zapGormLogger,
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &defaultGormConfig)
	if err != nil {
		t.Fatalf("init in-memory sqlite server: %v", err)
	}

	sfn, err := snowflake.NewNode(1)
	if err != nil {
		t.Fatalf("init snowflake generator: %v", err)
	}

	ctx := context.TODO()
	err = configureDB(ctx, db, sfn)
	if err != nil {
		t.Fatalf("init rbdb: %v", err)
	}

	// enable foreign keys constraints on sqlite
	err = db.Exec("PRAGMA foreign_keys = ON").Error
	if err != nil {
		t.Fatalf("enabled foreign keys: %v", err)
	}

	TestingCreateEntities(t, db)

	return db, sfn
}

func TestingCreateEntities(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Transaction(func(tx *gorm.DB) error {

		users := []*UserORM{
			{
				DiscourseId: 7,
			},
			{
				DiscourseId: 1337,
			},
		}
		return tx.Create(users).Error
	}); err != nil {
		t.Fatalf("create testing entities: %v", err)
	}
}

func TestingGetUser(t *testing.T, db *gorm.DB, discourseId int64) *User {
	var userOrm UserORM

	// Use struct-based query instead of string-based query
	err := db.Where(&UserORM{DiscourseId: discourseId}).First(&userOrm).Error
	if err != nil {
		t.Fatalf("fetch user by ID: %v", err)
	}

	user, err := userOrm.ToPB(context.Background())
	if err != nil {
		t.Fatalf("convert user to protobuf: %v", err)
	}

	return &user
}

// Helper function to create a test payment
func TestingCreateTestPayment(t *testing.T, db *gorm.DB, user *User, duration LicenseKey_Duration) *Payment {
	payment := Payment{
		Provider:        Payment_PROVIDER_MANUAL,
		ReferenceId:     fmt.Sprintf("test-payment-%d", time.Now().UnixNano()),
		AmountInCents:   900, // 9 euros in cents
		Currency:        "eur",
		LicenseDuration: duration,
		UserId:          user.Id,
	}
	createdPayment, err := DefaultCreatePayment(context.Background(), &payment, db)
	if err != nil {
		t.Fatalf("create testing payment: %v", err)
	}
	return createdPayment
}
