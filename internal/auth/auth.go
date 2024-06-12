package auth

import (
	"BirthdayGreetings/internal/errors"
	"BirthdayGreetings/internal/models"
	"BirthdayGreetings/internal/service"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *service.UserService
}

func NewAuthService(userService *service.UserService) *AuthService {
	return &AuthService{
		userService: userService,
	}
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
		return errors.New(401, fmt.Sprintf("не удалось хэшировать пароль: %v", err))
	}

	user := &models.User{
		Username:   username,
		Password:   hashedPassword,
		TelegramID: telegramID,
	}

	return s.userService.CreateUser(user)
}

func (s *AuthService) AuthenticateUser(username, password string, telegramID int64) (string, error) {
	user, err := s.userService.GetUserByName(username)
	if err != nil {
		return "", err
	}

	if user.TelegramID != telegramID {
		return "", errors.New(401, "неверный телеграмм аккаунт")
	}

	if !CheckPasswordHash(password, user.Password) {
		return "", errors.New(401, "неверный логин и/или пароль")
	}

	return username, nil
}
