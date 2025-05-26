package storage

import (
	"database/sql"
	"errors"
	"os"

	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	return err
}

func CreateUser(email, password string) (models.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	id := uuid.New().String()
	token := uuid.New().String()

	_, err = db.Exec(`INSERT INTO users (id, email, password, verification_token) VALUES ($1, $2, $3, $4)`,
		id, email, string(hashed), token)
	if err != nil {
		return models.User{}, err
	}

	return models.User{ID: id, Email: email, VerificationToken: token}, nil
}

func VerifyUser(token string) error {
	res, err := db.Exec(`UPDATE users SET is_verified = true, verification_token = '' WHERE verification_token = $1`, token)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return errors.New("invalid token")
	}
	return nil
}

func AuthenticateUser(email, password string) (models.User, error) {
	var user models.User
	err := db.QueryRow(`SELECT id, password, is_verified FROM users WHERE email = $1`, email).
		Scan(&user.ID, &user.Password, &user.IsVerified)
	if err != nil {
		return user, errors.New("user not found")
	}

	if !user.IsVerified {
		return user, errors.New("email not verified")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return user, errors.New("wrong password")
	}

	user.Email = email
	return user, nil
}
