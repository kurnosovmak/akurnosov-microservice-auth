package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/mailer"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/models"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/storage"

	"github.com/golang-jwt/jwt/v5"
)

// Структуры запросов и ответов
type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	Token string      `json:"token,omitempty"`
	User  models.User `json:"user,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				writeError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Вспомогательные функции
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// Обработчики
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := storage.CreateUser(ctx, req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := mailer.SendVerificationEmail(user.Email, user.VerificationToken); err != nil {
		log.Printf("Failed to send verification email: %v", err)
	}

	writeJSON(w, http.StatusCreated, AuthResponse{User: user})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "Token is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := storage.VerifyUser(ctx, token); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := storage.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		if err.Error() == "email not verified" && user.VerificationToken != "" {
			// Отправляем новый код подтверждения
			if err := mailer.SendVerificationEmail(user.Email, user.VerificationToken); err != nil {
				log.Printf("Failed to send verification email: %v", err)
			}
			writeError(w, http.StatusUnauthorized, "Email not verified. New verification code has been sent")
			return
		}
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{Token: tokenString, User: user})
}
