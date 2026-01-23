package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type Storage struct {
	db DBTX
}

func NewStorage(db DBTX) *Storage {
	return &Storage{
		db: db,
	}
}

func NewStorageFromPool(pool *pgxpool.Pool) *Storage {
	return &Storage{db: pool}
}

// Errors

var ErrValidation = errors.New("validation error")
