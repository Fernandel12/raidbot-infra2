package rbdb

import (
	"gorm.io/gorm"
)

// GetOffsetByVersion retrieves offset data for a specific version
func GetOffsetByVersion(db *gorm.DB, version string) ([]byte, error) {
	var offsetOrm OffsetORM
	err := db.Where(&OffsetORM{Version: version}).First(&offsetOrm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No offset for this version, return nil without error
		}
		return nil, GormToErrcode(err)
	}
	return offsetOrm.Data, nil
}

// UpdateOffset creates or updates offset data for a specific version
func UpdateOffset(db *gorm.DB, version string, data []byte) error {
	var offsetOrm OffsetORM
	err := db.Where(&OffsetORM{Version: version}).First(&offsetOrm).Error

	if err == gorm.ErrRecordNotFound {
		// Create new offset
		offsetOrm = OffsetORM{
			Version: version,
			Data:    data,
		}
		return db.Create(&offsetOrm).Error
	}

	if err != nil {
		return GormToErrcode(err)
	}

	// Update existing offset
	offsetOrm.Data = data
	return db.Save(&offsetOrm).Error
}
