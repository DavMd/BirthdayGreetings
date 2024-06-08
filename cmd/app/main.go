package main

import (
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/logging"
	"log"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	logging.Init("logs/app.log")

	err = db.Connect()
	if err != nil {
		logging.Logger.Fatalf("could not connect to the database: %v", err)
	}

	logging.Logger.Println("Birthday Greetings Service is running")

	logging.Logger.Println("Birthday Greetings Service is stop")
}

func Test() {
	// users := make([]models.User, 0, 1)

	// users = append(users, CreateUser("test user 3", "test user 3", time.Date(2000, time.August, 27, 0, 0, 0, 0, time.Local)))

	// if err := db.CreateUser(&users[0]); err != nil {
	// 	logging.Logger.Fatalf("could not create user: %v", err)
	// }
	// logging.Logger.Println(users[0].Username + " добавлен в базу данных")

	// subscription := models.Subscription{SubscriberID: 1, SubscribedUserID: 3}

	// if err := db.CreateSubscription(&subscription); err != nil {
	// 	logging.Logger.Fatalf("could not create subscription: %v", err)
	// }

	// logging.Logger.Println("Пользователь " +
	// 	fmt.Sprint(subscription.SubscriberID) +
	// 	" подписался на уведомление о др пользователя " +
	// 	fmt.Sprint(subscription.SubscribedUserID))

	// subscriptions, err := db.GetSubscriptions(1)

	// if err != nil {
	// 	logging.Logger.Fatalf("could not scan subscription: %v", err)
	// }
	// for _, subsubscription := range subscriptions {
	// 	logging.Logger.Println("Пользователь " +
	// 		fmt.Sprint(subsubscription.SubscriberID) +
	// 		" подписан на уведомление о др пользователя " +
	// 		fmt.Sprint(subsubscription.SubscribedUserID))
	// }

	// if err := db.DeleteSubscription(1, 3); err != nil {
	// 	logging.Logger.Fatalf("Fatal Error: %v", err)
	// }

	// logging.Logger.Println("Подписка пользователя 1 удалена с пользователя 3")

	// IsSubscribed, err := db.IsSubscribed(1, 2)

	// if err != nil {
	// 	logging.Logger.Fatalf("Fatal Error: %v", err)
	// }

	// logging.Logger.Println("Подписка пользователя 1 на пользователя 2 - " + fmt.Sprint(IsSubscribed))
}

// func CreateUser(username, email string, birthday time.Time) models.User {

// 	user := models.User{
// 		Username: username,
// 		Email:    email,
// 		Birthday: birthday,
// 	}

// 	return user
// }
