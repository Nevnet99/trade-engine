package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

var ErrDuplicateUser = fmt.Errorf("username already taken")

func (s *Storage) CreateUser(ctx context.Context, user *User) (*User, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	insertUserQuery := `
    INSERT INTO users (username, password_hash) 
    VALUES ($1, $2) 
    RETURNING id
  `

	err = tx.QueryRow(ctx, insertUserQuery, user.Username, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateUser
		}

		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	insertWalletsQuery := `
    INSERT INTO wallets (user_id, asset, balance, locked)
    VALUES 
        ($1, 'USD', 0, 0),
        ($1, 'BTC', 0, 0)
  `

	if _, err := tx.Exec(ctx, insertWalletsQuery, user.ID); err != nil {
		return nil, fmt.Errorf("failed to insert default wallets: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}
