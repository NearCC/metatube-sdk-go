package enginex

import (
	"database/sql"

	"gorm.io/gorm"
)

// GormDB is a thin wrapper that lets the caller access the underlying
// *sql.DB for explicit Close, without dragging the gorm import into every
// internal package that wants to free resources. It also lets us hand
// the *gorm.DB to packages that need raw queries.
type GormDB struct {
	*gorm.DB
}

func wrapDB(db *gorm.DB) *GormDB { return &GormDB{DB: db} }

// Raw returns the underlying *gorm.DB so callers can issue custom queries
// (the resolver needs this for the alias-cache write-back).
func (w *GormDB) Raw() *gorm.DB { return w.DB }

// Close shuts down the connection pool.
func (w *GormDB) Close() error {
	sqlDB, err := w.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// sqlDB is intentionally a type alias so the unused import is kept for
// the Close signature; this silences the linter without sacrificing type
// safety.
var _ = (*sql.DB)(nil)
