package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/golang-jwt/jwt"
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
		return
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

type LoginRequest struct {
	Username string
	Password string
}

func createJWT(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "default-dev-secret-do-not-use-in-prod"
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *Server) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	request := LoginRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := s.store.GetUserByUsername(r.Context(), request.Username)

	if err != nil {
		if err == store.ErrUserNotFound {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Failed to get the user by username", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jwt, err := createJWT(user.ID)

	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    jwt,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"user_id": user.ID})

}
