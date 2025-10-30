//nolint:goconst
package rbdb

import (
	"errors"

	"gorm.io/gorm"
	"raidbot.app/go/pkg/errcode"
)

func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(errors.Unwrap(err), gorm.ErrRecordNotFound)
}

func GormToErrcode(err error) error {
	if IsRecordNotFoundError(err) {
		return errcode.ERR_DB_NOT_FOUND.Wrap(err)
	}

	if err != nil {
		return errcode.ERR_DB_INTERNAL.Wrap(err)
	}

	return nil
}

// GetSortField returns the database column name for sorting
func GetSortField(sortBy string) string {
	switch sortBy {
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	case "views_count":
		return "views_count"
	case "likes_count":
		// likes_count is no longer a database column
		// We'll need to handle sorting after fetching the data
		return "created_at" // fallback to created_at for now
	case "name":
		return "name"
	default:
		return "created_at"
	}
}
