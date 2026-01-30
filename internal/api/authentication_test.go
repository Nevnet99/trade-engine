package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestHandleCreateUser(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := store.NewStorage(tx)
	s := NewServer(storage)
	ctx := context.Background()

	t.Run("Happy Path_Returns_201_Created", func(t *testing.T) {
		payload := map[string]string{
			"username": "bilbo_baggins",
			"password": "my_precious_ring",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		s.HandleCreateUser(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if response["user_id"] == "" {
			t.Error("Expected user_id in response, got empty string")
		}
	})

	t.Run("Duplicate User_Returns_409_Conflict", func(t *testing.T) {
		existingUser := &store.User{
			Username:     "gollum",
			PasswordHash: "fish",
		}
		_, err := storage.CreateUser(ctx, existingUser)
		if err != nil {
			t.Fatalf("Failed to seed user: %v", err)
		}

		payload := map[string]string{
			"username": "gollum",
			"password": "new_password",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		s.HandleCreateUser(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status 409 Conflict, got %d", w.Code)
		}
	})

	t.Run("Bad JSON_Returns_400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("{invalid-json")))
		w := httptest.NewRecorder()

		s.HandleCreateUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
		}
	})
}
