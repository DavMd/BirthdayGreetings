package models

import "time"

type User struct {
	ID         int64     `json:"id" db:"id"`
	Username   string    `json:"username" db:"username"`
	Password   string    `json:"password" db:"password"`
	TelegramID int64     `json:"telegram_id" db:"telegram_id"`
	Birthday   time.Time `json:"birthday" db:"birthday"`
}
