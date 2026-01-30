package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Nevnet99/trade-engine/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytePassword := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

type RegisterRequest struct {
	Username string
	Password string
}

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	request := RegisterRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(request.Password)

	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
	}

	partialUser := store.User{
		Username:     request.Username,
		PasswordHash: hashedPassword,
	}

	user, err := s.store.CreateUser(r.Context(), &partialUser)
	if err != nil {
		if err == store.ErrDuplicateUser {
			http.Error(w, "Username already taken", http.StatusConflict) // 409
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"user_id": user.ID})

}
