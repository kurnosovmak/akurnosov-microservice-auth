package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return nil
}

func CreateUser(ctx context.Context, email, password string) (models.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	id := uuid.New().String()
	token := uuid.New().String()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback()

	// Проверка на существующий email
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists); err != nil {
		return models.User{}, err
	}

	if exists {
		return models.User{}, errors.New("email already exists")
	}

	// Создание пользователя
	_, err = tx.ExecContext(ctx, `INSERT INTO users (id, email, password, verification_token) VALUES ($1, $2, $3, $4)`,
		id, email, string(hashed), token)
	if err != nil {
		return models.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.User{}, err
	}

	return models.User{ID: id, Email: email, VerificationToken: token}, nil
}

func VerifyUser(ctx context.Context, token string) error {
	res, err := db.ExecContext(ctx, `UPDATE users SET is_verified = true, verification_token = '' WHERE verification_token = $1`, token)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("invalid token")
	}

	return nil
}

func AuthenticateUser(ctx context.Context, email, password string) (models.User, error) {
	var user models.User
	err := db.QueryRowContext(ctx, `SELECT id, email, password, is_verified, verification_token FROM users WHERE email = $1`, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.IsVerified, &user.VerificationToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return user, errors.New("wrong password")
	}

	if !user.IsVerified {
		// Генерируем новый токен верификации
		newToken, err := RegenerateVerificationToken(ctx, email)
		if err != nil {
			return user, err
		}
		user.VerificationToken = newToken
		return user, errors.New("email not verified")
	}

	// Очистка конфиденциальных данных перед возвратом
	user.Password = ""
	user.VerificationToken = ""

	return user, nil
}

func RegenerateVerificationToken(ctx context.Context, email string) (string, error) {
    token := uuid.New().String()
    
    result, err := db.ExecContext(ctx, 
        `UPDATE users SET verification_token = $1 WHERE email = $2 AND is_verified = false`,
        token, email)
    if err != nil {
        return "", err
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return "", err
    }
    
    if rows == 0 {
        return "", errors.New("user not found or already verified")
    }
    
    return token, nil
}
