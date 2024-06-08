package models

import "time"

type User struct {
	ID       int64     `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	Email    string    `json:"email" db:"email"`
	Birthday time.Time `json:"birthday" db:"birthday"`
}
