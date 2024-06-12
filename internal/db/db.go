package db

import (
	"database/sql"
	"fmt"
	"os"

	"BirthdayGreetings/internal/errors"
	"BirthdayGreetings/internal/logging"
	"BirthdayGreetings/internal/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// func createUserTable() error {
// 	queryUser := `CREATE TABLE users (
// 				id SERIAL PRIMARY KEY,
// 				username VARCHAR(255) NOT NULL UNIQUE,
// 				password VARCHAR(255) NOT NULL,
// 				telegram_id BIGINT NOT NULL UNIQUE,
// 				birthday DATE);`
// 	if _, err := DB.Exec(queryUser); err != nil {
// 		return errors.New(500, fmt.Sprintf("could not create users table: %v", err))
// 	}

// 	logging.Logger.Println("Successfull DB User Table Creation")
// 	return nil
// }

// func createSubsTable() error {
// 	querySubscription := `CREATE TABLE subscriptions (
// 						id SERIAL PRIMARY KEY,
// 						user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
// 						subscribed_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
// 						UNIQUE (user_id, subscribed_user_id));`

// 	if _, err := DB.Exec(querySubscription); err != nil {
// 		return errors.New(500, fmt.Sprintf("could not create subcriptions table: %v", err))
// 	}
// 	logging.Logger.Println("Successfull DB Subscription Table Creation")
// 	return nil
// }

func Connect() error {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return errors.New(500, fmt.Sprintf("ошибка подключения к базе данных: %v", err))
	}

	if err := DB.Ping(); err != nil {
		return errors.New(500, fmt.Sprintf("ошибка в пинге базы данных: %v", err))
	}

	logging.Logger.Println("Успешное подключение к базе данных.")

	return nil
}

// func checkTables() error {
// 	query := `select relname from pg_stat_user_tables`
// 	rows, err := DB.Query(query)
// 	if err != nil {
// 		return errors.New(400, fmt.Sprintf("query error: %v", err))
// 	}

// 	rows.Next()
// 	if err := rows.Scan(); err != nil {
// 		if err := createUserTable(); err != nil {
// 			return err
// 		}
// 		if err := createSubsTable(); err != nil {
// 			return err
// 		}
// 		return nil
// 	}

// 	rows.Close()
// 	return nil
// }

func CreateUser(user *models.User) error {
	query := `SELECT username, telegram_id FROM users WHERE username = $1 or telegram_id = $2)`
	var tempUser models.User
	err := DB.QueryRow(query, user.Username, user.TelegramID).Scan(&tempUser.Username, &tempUser.TelegramID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("Ошибка: %v", err))
	}

	if tempUser.Username == user.Username {
		return errors.New(409, "Пользователь с таким именем уже существует")
	}

	if tempUser.TelegramID == user.TelegramID {
		return errors.New(409, "На этот телеграмм аккаунт уже зарегистрирован пользователь")
	}

	query = `INSERT INTO users (username, password, telegram_id, birthday) VALUES ($1, $2, $3, $4) RETURNING id`
	err = DB.QueryRow(query, user.Username, user.Password, user.TelegramID, user.Birthday).Scan(&user.ID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("ошибка в создании пользователя: %v", err))
	}
	return nil
}

func CreateSubscription(sub *models.Subscription) error {
	query := `INSERT INTO subscriptions (user_id, subscribed_user_id) VALUES ($1, $2) RETURNING id`
	err := DB.QueryRow(query, sub.SubscriberID, sub.SubscribedUserID).Scan(&sub.ID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("не удалось подписаться: %v", err))
	}
	return nil
}

func DeleteSubscription(subscriberID, subscribedUserID int64) error {
	query := `DELETE FROM subscriptions WHERE user_id = $1 AND subscribed_user_id = $2`
	result, err := DB.Exec(query, subscriberID, subscribedUserID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("не удалось отписаться: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New(404, fmt.Sprintf("нет строк для изменения: %v", err))
	}

	if rowsAffected == 0 {
		return errors.New(404, "подписка не найдена")
	}

	return nil
}

func IsSubscribed(subscriberID, subscribedUserID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE subscriber_id = $1 AND subscribed_user_id = $2)`
	var exists bool
	err := DB.QueryRow(query, subscriberID, subscribedUserID).Scan(&exists)
	if err != nil {
		return false, errors.New(400, fmt.Sprintf("ошибка в проверке подписки: %v", err))
	}
	return exists, nil
}

func GetSubscribers(userID int64) ([]models.UserBirthLayout, error) {
	query := `SELECT users.username, users.telegram_id, users.birthday 
			FROM subscriptions 
			JOIN users ON subscriptions.subscribed_user_id = users.id 
			WHERE subscriptions.user_id = $1`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, errors.New(400, fmt.Sprintf("не удалось получить пользователей: %v", err))
	}
	defer rows.Close()

	var subscribers []models.UserBirthLayout
	for rows.Next() {
		var user models.UserBirthLayout
		if err := rows.Scan(&user.Username, &user.TelegramID, &user.Birthday); err != nil {
			return nil, errors.New(400, fmt.Sprintf("ошибка в получении пользователя: %v", err))
		}
		subscribers = append(subscribers, user)
	}
	return subscribers, nil
}

func GetUsersWithBirthday(date string) ([]models.UserBirthLayout, error) {
	query := `SELECT username, telegram_id, birthday FROM users WHERE to_char(birthday, 'MM-DD') = $1`
	rows, err := DB.Query(query, date)
	if err != nil {
		return nil, errors.New(400, fmt.Sprintf("ошибка в получении пользователей: %v", err))
	}
	defer rows.Close()

	var users []models.UserBirthLayout
	for rows.Next() {
		var user models.UserBirthLayout
		if err := rows.Scan(&user.Username, &user.TelegramID, &user.Birthday); err != nil {
			return nil, errors.New(400, fmt.Sprintf("ошибка в получении пользователя: %v", err))
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserByName(username string) (*models.User, error) {
	query := `SELECT id, username, password, telegram_id, birthday FROM users WHERE username = $1`
	row := DB.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.TelegramID, &user.Birthday)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(404, "пользователь не найден")
		}
		return nil, errors.New(400, fmt.Sprintf("не удалось получить пользователя: %v", err))
	}

	return &user, nil
}

func GetUserByTgID(telegramID int64) (*models.User, error) {
	query := `SELECT id, username, password, telegram_id, birthday FROM users WHERE telegram_id = $1`
	row := DB.QueryRow(query, telegramID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.TelegramID, &user.Birthday)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(404, "пользователь не найден")
		}
		return nil, errors.New(400, fmt.Sprintf("не удалось получить пользователя: %v", err))
	}

	return &user, nil
}

func GetAllUsers() ([]*models.UserBirthLayout, error) {
	rows, err := DB.Query(`SELECT username, birthday, telegram_id FROM users`)
	if err != nil {
		return nil, fmt.Errorf("ошибка в получении пользователей: %w", err)
	}
	defer rows.Close()

	var users []*models.UserBirthLayout
	for rows.Next() {
		var user models.UserBirthLayout
		err := rows.Scan(&user.Username, &user.Birthday, &user.TelegramID)
		if err != nil {
			return nil, errors.New(500, fmt.Sprintf("ошибка в получении пользователя: %v", err))
		}
		users = append(users, &user)
	}

	return users, nil
}

func SetUserBirthday(telegramID int64, birthday string) error {
	query := `UPDATE users SET birthday = $1 WHERE telegram_id = $2`
	_, err := DB.Exec(query, birthday, telegramID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("ошибка в обновлении дня рождения: %v", err))
	}
	return nil
}

func UpdateUser(user *models.User) error {
	query := `UPDATE users SET username = $1, password = $2, telegram_id = $3, birthday = $4 WHERE id = $5`
	_, err := DB.Exec(query, user.Username, user.Password, user.TelegramID, user.Birthday, user.ID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("ошибка в обновлении пользователя: %v", err))
	}
	return nil
}
