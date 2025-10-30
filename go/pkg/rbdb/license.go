package rbdb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"raidbot.app/go/pkg/errcode"
)

// LicenseResponse matches the PHP expected format
type LicenseResponse struct {
	Status      string `json:"status,omitempty"`
	UsageID     string `json:"usage_id,omitempty"`
	Uses        int64  `json:"uses,omitempty"`
	FaultString string `json:"fault_string,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// IsLicenseExpired checks if a license is expired based on its duration and effective date
func IsLicenseExpired(license *LicenseKey) bool {
	// If EffectiveFrom is not set, license hasn't been activated yet, so not expired
	if license.EffectiveFrom == nil {
		return false
	}

	// Calculate expiration based on duration
	var expiration time.Time
	switch license.Duration {
	case LicenseKey_LIFETIME:
		return false
	case LicenseKey_ONE_DAY:
		expiration = license.EffectiveFrom.AsTime().AddDate(0, 0, 1)
	case LicenseKey_ONE_WEEK:
		expiration = license.EffectiveFrom.AsTime().AddDate(0, 0, 7)
	case LicenseKey_ONE_MONTH:
		expiration = license.EffectiveFrom.AsTime().AddDate(0, 1, 0)
	case LicenseKey_THREE_MONTHS:
		expiration = license.EffectiveFrom.AsTime().AddDate(0, 3, 0)
	case LicenseKey_ONE_YEAR:
		expiration = license.EffectiveFrom.AsTime().AddDate(1, 0, 0)
	default:
		return true
	}

	// Check if current time is after expiration + grace period
	// Add 2-hour grace period to avoid service interruption during renewal processing
	gracePeriod := 2 * time.Hour
	now := time.Now().UTC()
	return now.After(expiration.Add(gracePeriod))
}

// ValidateLicenseStatus checks if a license is valid
func ValidateLicenseStatus(db *gorm.DB, key string) (*LicenseKey, error) {
	var licensesOrm []*LicenseKeyORM
	if err := db.Where(&LicenseKeyORM{Key: key}).Find(&licensesOrm).Error; err != nil {
		return nil, GormToErrcode(err)
	}
	if len(licensesOrm) == 0 {
		return nil, errcode.ERR_LICENSE_NOT_FOUND.Wrap(fmt.Errorf("%s", key))
	}
	// Check for duplicate licenses in database
	if len(licensesOrm) > 1 {
		return nil, errcode.ERR_LICENSE_COLLISION.Wrap(fmt.Errorf("found %d licenses with key %s", len(licensesOrm), key))
	}

	licenseOrm := licensesOrm[0]
	license, err := licenseOrm.ToPB(context.Background())
	if err != nil {
		return nil, errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(fmt.Errorf("%s: %w", key, err))
	}

	if license.Revoked {
		return nil, errcode.ERR_LICENSE_REVOKED.Wrap(fmt.Errorf("%s", key))
	}

	if license.Duration == LicenseKey_LIFETIME {
		return &license, nil
	}

	// If EffectiveFrom is not set, license is valid and don't check expiry
	if license.EffectiveFrom == nil {
		return &license, nil
	}

	// Check for expiration
	expired := IsLicenseExpired(&license)
	if expired {
		return nil, errcode.ERR_LICENSE_EXPIRED.Wrap(fmt.Errorf("%s", key))
	}

	return &license, nil
}

func ActivateLicense(db *gorm.DB, key string) (*LicenseKey, error) {
	var license *LicenseKey
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		// Validate license first
		license, err = ValidateLicenseStatus(tx, key)
		if err != nil {
			return err
		}

		var licenseOrm *LicenseKeyORM
		if err := tx.Where(&LicenseKeyORM{Key: key}).First(&licenseOrm).Error; err != nil {
			return GormToErrcode(err)
		}

		// Generate new usage ID
		usageID, err := GenerateUsageID()
		if err != nil {
			return err
		}

		// Set EffectiveFrom if this is the first activation OR if EffectiveFrom is nil
		if licenseOrm.Uses == 0 || licenseOrm.EffectiveFrom == nil {
			now := time.Now().UTC()
			licenseOrm.EffectiveFrom = &now
		}

		licenseOrm.ActiveUsageId = usageID
		licenseOrm.Uses++

		// Update license with new usage ID, uses, and possibly EffectiveFrom
		err = tx.Save(licenseOrm).Error
		if err != nil {
			return GormToErrcode(err)
		}

		*license, err = licenseOrm.ToPB(context.Background())
		if err != nil {
			return errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(fmt.Errorf("%s: %w", key, err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return license, err
}

// Check validates a license and increments its use count
func CheckLicense(db *gorm.DB, key string, usageID string) (int64, error) {
	var license *LicenseKey
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		// Validate license status
		if license, err = ValidateLicenseStatus(tx, key); err != nil {
			return err
		}

		var licenseOrm *LicenseKeyORM
		if err := tx.Where(&LicenseKeyORM{Key: key}).First(&licenseOrm).Error; err != nil {
			return GormToErrcode(err)
		}

		// Verify usage ID matches
		if licenseOrm.ActiveUsageId != usageID {
			return errcode.ERR_LICENSE_INVALID_USAGE_ID.Wrap(fmt.Errorf("key: %s, expected: %s, got: %s", key, licenseOrm.ActiveUsageId, usageID))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return license.Uses, err
}

// Create generates a new license key for a user
func GenerateLicense(db *gorm.DB, userId int64, paymentId int64, duration LicenseKey_Duration, setEffectiveFromNow bool) (*LicenseKey, error) {
	// Generate random bytes for the key
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, errcode.ERR_LICENSE_RANDOM_GENERATION.Wrap(err)
	}

	// Hash the bytes to create the key
	hash := sha256.Sum256(bytes)
	key := hex.EncodeToString(hash[:16])

	// Check for collisions and retry if necessary
	for attempts := 0; attempts < 10; attempts++ {
		var exists bool
		if err := db.Model(&LicenseKeyORM{}).
			Select("1").
			Where(&LicenseKeyORM{Key: key}).
			Limit(1).
			Find(&exists).
			Error; err != nil {
			return nil, GormToErrcode(err)
		}

		if !exists {
			break
		}

		if attempts == 9 {
			return nil, errcode.ERR_LICENSE_COLLISION.Wrap(fmt.Errorf("max attempts: %d", attempts))
		}

		// Generate a new key for the next attempt
		if _, err := rand.Read(bytes); err != nil {
			return nil, errcode.ERR_LICENSE_RANDOM_GENERATION.Wrap(err)
		}
		hash = sha256.Sum256(bytes)
		key = hex.EncodeToString(hash[:16])
	}

	// Create the license in a transaction
	var createdLicense *LicenseKey
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		paymentORM := &PaymentORM{Id: paymentId}
		err = tx.Where(&paymentORM).First(paymentORM).Error
		if err != nil {
			return GormToErrcode(err)
		}

		// First create the license key
		license := &LicenseKey{
			Key:         key,
			Duration:    duration,
			Revoked:     false,
			UserId:      userId,
			Uses:        0,
			SandboxMode: paymentORM.SandboxMode,
		}

		if setEffectiveFromNow {
			license.EffectiveFrom = timestamppb.New(time.Now())
		}

		createdLicense, err = DefaultCreateLicenseKey(context.Background(), license, tx)
		if err != nil {
			return err
		}

		paymentORM.LicenseKeyId = &createdLicense.Id

		if err := tx.Save(&paymentORM).Error; err != nil {
			return GormToErrcode(err)
		}

		// Create license generation activity
		licenseActivityORM := &ActivityORM{
			Kind:         int32(Activity_KIND_LICENSE_GENERATION),
			UserId:       &userId,
			LicenseKeyId: &createdLicense.Id,
			PaymentId:    &paymentId,
		}

		err = tx.Create(&licenseActivityORM).Error
		if err != nil {
			return GormToErrcode((err))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdLicense, nil
}

// RenewLicense renews an expired license key
func RenewLicense(db *gorm.DB, licenseKeyId int64, userId int64, paymentId int64, forceRenew bool) (*LicenseKey, error) {
	var licenseKeyORM LicenseKeyORM
	err := db.Where(&LicenseKeyORM{Id: licenseKeyId}).First(&licenseKeyORM).Error
	if err != nil {
		return nil, GormToErrcode(err)
	}
	licenseKey, err := licenseKeyORM.ToPB(context.Background())
	if err != nil {
		return nil, errcode.ERR_LICENSE_PROTOBUF_CONVERSION.Wrap(err)
	}

	// Verify the license is not revoked
	if licenseKey.Revoked {
		return nil, errcode.ERR_LICENSE_REVOKED.Wrap(fmt.Errorf("license with key %s is revoked", licenseKey.Key))
	}

	// For LIFETIME licenses, no need to renew
	if licenseKey.Duration == LicenseKey_LIFETIME {
		return nil, errcode.ERR_LICENSE_INVALID_OPERATION.Wrap(fmt.Errorf("cannot renew LIFETIME license"))
	}

	// Check if the license is expired, unless forced renewal for subscriptions
	if !forceRenew && !IsLicenseExpired(&licenseKey) {
		return nil, errcode.ERR_LICENSE_NOT_YET_EXPIRED.Wrap(fmt.Errorf("license with key %s is still active", licenseKey.Key))
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		paymentORM := PaymentORM{Id: paymentId}
		err := tx.Where(&paymentORM).First(&paymentORM).Error
		if err != nil {
			return err
		}
		paymentORM.LicenseKeyId = &licenseKey.Id
		if err := tx.Save(&paymentORM).Error; err != nil {
			return err
		}

		// Update effective_from date to now
		licenseKey.EffectiveFrom = timestamppb.Now()
		_, err = DefaultStrictUpdateLicenseKey(context.Background(), &licenseKey, tx)
		if err != nil {
			return err
		}

		// Create license renewal activity
		licenseActivityORM := &ActivityORM{
			Kind:         int32(Activity_KIND_LICENSE_RENEWAL),
			UserId:       &userId,
			LicenseKeyId: &licenseKey.Id,
			PaymentId:    &paymentId,
		}

		return tx.Create(&licenseActivityORM).Error
	})
	if err != nil {
		return nil, GormToErrcode((err))
	}

	return &licenseKey, nil
}

func GenerateUsageID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", errcode.ERR_LICENSE_INVALID_USAGE_ID.Wrap(err)
	}
	return hex.EncodeToString(bytes), nil
}
