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

func createUserTable() error {
	queryUser := `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				username VARCHAR(255) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL,
				telegram_id BIGINT NOT NULL UNIQUE,
				birthday DATE);`
	if _, err := DB.Exec(queryUser); err != nil {
		return errors.New(500, fmt.Sprintf("could not create users table: %v", err))
	}

	logging.Logger.Println("Successfull DB User Table Creation")
	return nil
}

func createSubsTable() error {
	querySubscription := `CREATE TABLE subscriptions (
						id SERIAL PRIMARY KEY,
						user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
						subscribed_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
						UNIQUE (user_id, subscribed_user_id));`

	if _, err := DB.Exec(querySubscription); err != nil {
		return errors.New(500, fmt.Sprintf("could not create subcriptions table: %v", err))
	}
	logging.Logger.Println("Successfull DB Subscription Table Creation")
	return nil
}

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
		return errors.New(500, fmt.Sprintf("could not open db: %v", err))
	}

	if err := DB.Ping(); err != nil {
		return errors.New(500, fmt.Sprintf("could not ping db: %v", err))
	}

	logging.Logger.Println("Connected to the database successfully")

	if err = checkTables(); err != nil {
		logging.Logger.Fatalf("Couldn't create tables: %v", err)
	}

	return nil
}

func checkTables() error {
	query := `select relname from pg_stat_user_tables`
	rows, err := DB.Query(query)
	if err != nil {
		return errors.New(400, fmt.Sprintf("query error: %v", err))
	}

	rows.Next()
	if err := rows.Scan(); err != nil {
		if err := createUserTable(); err != nil {
			return err
		}
		if err := createSubsTable(); err != nil {
			return err
		}
		return nil
	}

	rows.Close()
	return nil
}

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, password, telegram_id, birthday) VALUES ($1, $2, $3, $4) RETURNING id`
	err := DB.QueryRow(query, user.Username, user.Password, user.TelegramID, user.Birthday).Scan(&user.ID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("could not create user: %v", err))
	}
	return nil
}

func CreateSubscription(sub *models.Subscription) error {
	query := `INSERT INTO subscriptions (subscriber_id, subscribed_user_id) VALUES ($1, $2) RETURNING id`
	err := DB.QueryRow(query, sub.SubscriberID, sub.SubscribedUserID).Scan(&sub.ID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("could not create subscription: %v", err))
	}
	return nil
}

func DeleteSubscription(subscriberID, subscribedUserID int64) error {
	query := `DELETE FROM subscriptions WHERE subscriber_id = $1 AND subscribed_user_id = $2`
	result, err := DB.Exec(query, subscriberID, subscribedUserID)
	if err != nil {
		return errors.New(400, fmt.Sprintf("could not delete subscription: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New(500, fmt.Sprintf("could not get affected rows: %v", err))
	}

	if rowsAffected == 0 {
		return errors.New(404, "subscription not found")
	}

	return nil
}

func GetSubscriptions(userID int64) ([]models.Subscription, error) {
	query := `SELECT id, subscriber_id, subscribed_user_id FROM subscriptions WHERE subscriber_id = $1`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, errors.New(400, fmt.Sprintf("could not get subscriptions: %v", err))
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(&sub.ID, &sub.SubscriberID, &sub.SubscribedUserID); err != nil {
			return nil, errors.New(400, fmt.Sprintf("could not scan subscription: %v", err))
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

func IsSubscribed(subscriberID, subscribedUserID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE subscriber_id = $1 AND subscribed_user_id = $2)`
	var exists bool
	err := DB.QueryRow(query, subscriberID, subscribedUserID).Scan(&exists)
	if err != nil {
		return false, errors.New(400, fmt.Sprintf("could not check subscription: %v", err))
	}
	return exists, nil
}

func GetSubscribers(userID int64) ([]models.User, error) {
	query := `SELECT users.id, users.username, users.password, users.telegram_id, users.birthday 
			FROM subscriptions 
			JOIN users ON subscriptions.subscriber_id = users.id 
			WHERE subscriptions.subscribed_user_id = $1`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, errors.New(400, fmt.Sprintf("could not get subscribers: %v", err))
	}
	defer rows.Close()

	var subscribers []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.TelegramID, &user.Birthday); err != nil {
			return nil, errors.New(400, fmt.Sprintf("could not scan user: %v", err))
		}
		subscribers = append(subscribers, user)
	}
	return subscribers, nil
}

func GetUsersWithBirthday(date string) ([]models.User, error) {
	query := `SELECT id, username, password, telegram_id, birthday FROM users WHERE to_char(birthday, 'MM-DD') = $1`
	rows, err := DB.Query(query, date)
	if err != nil {
		return nil, errors.New(400, fmt.Sprintf("could not get users with birthday: %v", err))
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.TelegramID, &user.Birthday); err != nil {
			return nil, errors.New(400, fmt.Sprintf("could not scan user: %v", err))
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
			return nil, errors.New(404, "user not found")
		}
		return nil, errors.New(500, fmt.Sprintf("could not get user by username: %v", err))
	}

	return &user, nil
}
