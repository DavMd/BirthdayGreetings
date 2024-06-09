package auth

import (
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/errors"
	"BirthdayGreetings/internal/models"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) RegisterUser(username, password string, telegramID int64) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return errors.New(500, fmt.Sprintf("could not hash password: %v", err))
	}

	user := &models.User{
		Username:   username,
		Password:   hashedPassword,
		TelegramID: telegramID,
	}

	return db.CreateUser(user)
}

func (s *AuthService) AuthenticateUser(username, password string, telegramID int64) (string, error) {
	user, err := db.GetUserByName(username)
	if err != nil {
		return "", err
	}

	if user.TelegramID != telegramID {
		return "", errors.New(401, "incorrect telegram account")
	}

	if !CheckPasswordHash(password, user.Password) {
		return "", errors.New(401, "invalid username or password")
	}

	return username, nil
}
