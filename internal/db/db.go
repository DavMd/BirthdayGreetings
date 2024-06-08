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
	return nil
}

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, email, birthday) VALUES ($1, $2, $3) RETURNING id`
	err := DB.QueryRow(query, user.Username, user.Email, user.Birthday).Scan(&user.ID)
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
