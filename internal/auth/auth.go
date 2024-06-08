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

func (s *AuthService) RegisterUser(username, password string, telegramID int64) (*models.User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, errors.New(500, fmt.Sprintf("could not hash password: %v", err))
	}

	user := &models.User{
		Username:   username,
		Password:   hashedPassword,
		TelegramID: telegramID,
	}

	err = db.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) AuthenticateUser(username, password string) (*models.User, error) {
	user, err := db.GetUserByName(username)
	if err != nil {
		return nil, errors.New(400, "invalid username or password")
	}

	if !CheckPasswordHash(password, user.Password) {
		return nil, errors.New(400, "invalid username or password")
	}

	return user, nil
}
