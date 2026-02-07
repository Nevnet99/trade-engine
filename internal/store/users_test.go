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

func TestGetUserByUsername(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	s := NewStorage(tx)
	ctx := context.Background()

	targetUser := &User{
		Username:     "gandalf",
		PasswordHash: "you_shall_not_pass",
	}

	if _, err := s.CreateUser(ctx, targetUser); err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	t.Run("Happy Path_UserFound", func(t *testing.T) {
		found, err := s.GetUserByUsername(ctx, "gandalf")
		if err != nil {
			t.Fatalf("Expected to find user, got error: %v", err)
		}

		if found.Username != "gandalf" {
			t.Errorf("Expected username 'gandalf', got %s", found.Username)
		}
		if found.PasswordHash != "you_shall_not_pass" {
			t.Errorf("Expected password hash to match, got %s", found.PasswordHash)
		}
		if found.ID == "" {
			t.Error("Expected User ID to be populated")
		}
	})

	t.Run("Error_UserNotFound", func(t *testing.T) {
		_, err := s.GetUserByUsername(ctx, "saruman_the_missing")

		if err != ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}
