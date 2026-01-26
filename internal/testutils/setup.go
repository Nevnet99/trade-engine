package testutils

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type TestTx struct {
	pgx.Tx
}

func (t *TestTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return t, nil
}

func (t *TestTx) Commit(ctx context.Context) error {
	return nil
}

func (t *TestTx) Rollback(ctx context.Context) error {
	return nil
}

func SetupTestDB(t *testing.T) *TestTx {
	t.Helper()

	_ = godotenv.Load("../../.env")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}

	t.Cleanup(func() {
		tx.Rollback(context.Background())
	})

	return &TestTx{Tx: tx}
}
