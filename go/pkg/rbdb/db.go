package rbdb

import (
	"context"
	"reflect"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"moul.io/zapgorm2"

	"rslbot.com/go/pkg/errcode"
)

var Models = []interface{}{
	&ActivityORM{},
	&LicenseKeyORM{},
	&UserORM{},
	&PaymentORM{},
	&SubscriptionORM{},
}

type DBConfig struct {
	Logger *zap.Logger
	URN    string
}

func InitDB(ctx context.Context, cfg DBConfig) (*gorm.DB, *snowflake.Node, error) {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	// Initialize Snowflake node for ID generation
	sfn, err := snowflake.NewNode(1)
	if err != nil {
		return nil, nil, errcode.ERR_DB_INIT.Wrap(err)
	}

	// Setup GORM logger with Zap
	gLogger := zapgorm2.New(cfg.Logger.Named("gorm"))
	gLogger.IgnoreRecordNotFoundError = true
	gLogger.SetAsDefault()
	gLogger.LogMode(logger.Info)

	// GORM configuration
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: gLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(mysql.Open(cfg.URN), gormConfig)
	if err != nil {
		return nil, nil, errcode.ERR_DB_CONNECT.Wrap(err)
	}

	// Configure callbacks and migrations
	if err := configureDB(ctx, db, sfn); err != nil {
		return nil, nil, errcode.ERR_CONFIGURE_DB.Wrap(err)
	}

	return db, sfn, nil
}

func configureDB(ctx context.Context, db *gorm.DB, sfn *snowflake.Node) error {
	// Register Snowflake ID generator callback
	if err := db.Callback().Create().Before("gorm:create").Register(
		"snowflake_ids:before_create",
		generateSnowflakeIDs(ctx, sfn),
	); err != nil {
		return errcode.ERR_DB_ADD_CALLBACK.Wrap(err)
	}

	// Run migrations
	if err := db.AutoMigrate(Models...); err != nil {
		return errcode.ERR_DB_AUTO_MIGRATE.Wrap(err)
	}

	return nil
}

func generateSnowflakeIDs(ctx context.Context, sfn *snowflake.Node) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		// Only handle if we have a schema and it's a create operation
		if db.Statement.Schema == nil {
			return
		}

		// Look for the ID field
		field := db.Statement.Schema.PrioritizedPrimaryField
		if field == nil {
			return
		}

		// Set snowflake ID for the current record(s)
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				if err := field.Set(ctx, db.Statement.ReflectValue.Index(i), sfn.Generate().Int64()); err != nil {
					db.Logger.Error(ctx, "failed to set snowflake ID", err)
				}
			}
		case reflect.Struct:
			if err := field.Set(ctx, db.Statement.ReflectValue, sfn.Generate().Int64()); err != nil {
				db.Logger.Error(ctx, "failed to set snowflake ID", err)
			}
		}
	}
}
