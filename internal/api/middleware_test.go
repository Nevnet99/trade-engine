package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(string)

		if !ok {
			t.Error("Middleware failed to inject UserID into context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userID))
	})

	s := &Server{}
	protectedRoute := s.AuthMiddleware(nextHandler)

	t.Run("Refuse Request_Without_Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		protectedRoute.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("Refuse Request_With_Invalid_Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		req.AddCookie(&http.Cookie{Name: "auth_token", Value: "garbage_token_string"})

		protectedRoute.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("Refuse Request_With_Wrong_Signature", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "hacker",
		})

		badString, _ := token.SignedString([]byte("wrong-secret-key"))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: badString})

		protectedRoute.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 (Signature Invalid), got %d", w.Code)
		}
	})

	t.Run("Allow Request_With_Valid_Token", func(t *testing.T) {

		os.Setenv("JWT_SECRET", "test-secret")
		defer os.Unsetenv("JWT_SECRET")

		claims := jwt.MapClaims{
			"sub": "user_123",
			"exp": time.Now().Add(time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		validTokenString, _ := token.SignedString([]byte("test-secret"))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		req.AddCookie(&http.Cookie{Name: "auth_token", Value: validTokenString})

		protectedRoute.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
		}
		if w.Body.String() != "user_123" {
			t.Errorf("Expected body to be 'user_123', got %s", w.Body.String())
		}
	})
}
