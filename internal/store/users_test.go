package store

import (
	"context"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestCreateUser(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	s := NewStorage(tx)
	ctx := context.Background()

	t.Run("Happy Path_CreatesUserAndWallets", func(t *testing.T) {
		user := &User{
			Username:     "samwise",
			PasswordHash: "hashed_potatoes",
		}

		createdUser, err := s.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		if createdUser.ID == "" {
			t.Error("Expected User ID to be generated, got empty string")
		}

		var walletCount int
		err = tx.QueryRow(ctx, "SELECT count(*) FROM wallets WHERE user_id = $1", createdUser.ID).Scan(&walletCount)
		if err != nil {
			t.Fatalf("Failed to count wallets: %v", err)
		}

		if walletCount != 2 {
			t.Errorf("Expected 2 default wallets, got %d", walletCount)
		}
	})

	t.Run("Error_DuplicateUsername", func(t *testing.T) {

		user := &User{
			Username:     "sauron",
			PasswordHash: "different_hash",
		}

		_, err := s.CreateUser(ctx, user)

		if err != ErrDuplicateUser {
			t.Errorf("Expected ErrDuplicateUser, got %v", err)
		}
	})
}
