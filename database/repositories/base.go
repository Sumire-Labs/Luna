package repositories

import (
	"database/sql"
)

// BaseRepository provides common database functionality
type BaseRepository struct {
	db *sql.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sql.DB) *BaseRepository {
	return &BaseRepository{
		db: db,
	}
}

// DB returns the database connection
func (r *BaseRepository) DB() *sql.DB {
	return r.db
}

// Transaction executes a function within a database transaction
func (r *BaseRepository) Transaction(fn func(*sql.Tx) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}