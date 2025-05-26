package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/mailer"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/storage"

	"github.com/golang-jwt/jwt/v5"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	user, err := storage.CreateUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = mailer.SendVerificationEmail(user.Email, user.VerificationToken)
	w.WriteHeader(http.StatusCreated)
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if err := storage.VerifyUser(token); err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}
	w.Write([]byte("Email verified successfully"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	user, err := storage.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
}
