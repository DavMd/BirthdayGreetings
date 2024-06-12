package service

import (
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/models"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) CreateUser(user *models.User) error {
	return db.CreateUser(user)
}

func (s *UserService) GetAllUsers() ([]*models.UserBirthLayout, error) {
	users, err := db.GetAllUsers()
	return users, err
}

func (s *UserService) SetUserBirthday(telegramID int64, birthday string) error {
	return db.SetUserBirthday(telegramID, birthday)
}

func (s *UserService) UpdateUser(user *models.User) error {
	return db.UpdateUser(user)
}

func (s *UserService) GetUserByName(username string) (*models.User, error) {
	users, err := db.GetUserByName(username)
	return users, err
}

func (s *UserService) GetUserByTgID(telegramID int64) (*models.User, error) {
	users, err := db.GetUserByTgID(telegramID)
	return users, err
}

func (s *UserService) GetUsersWithBirthday(date string) ([]models.UserBirthLayout, error) {
	users, err := db.GetUsersWithBirthday(date)
	return users, err
}
