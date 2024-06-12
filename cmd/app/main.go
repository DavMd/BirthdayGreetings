package main

import (
	"BirthdayGreetings/internal/auth"
	"BirthdayGreetings/internal/bot"
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/logging"
	"BirthdayGreetings/internal/notification"
	"BirthdayGreetings/internal/service"
	"BirthdayGreetings/internal/subscription"
	"BirthdayGreetings/internal/telegram"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		logging.Logger.Fatalf("Error loading .env file: %v", err)
	}

	logging.Init("logs/app.log")

	err = db.Connect()
	if err != nil {
		logging.Logger.Fatalf("could not connect to the database: %v", err)
	}

	appID, err := strconv.Atoi(os.Getenv("TELEGRAM_APP_ID"))
	if err != nil {
		logging.Logger.Fatalf("Ошибка парсинга TELEGRAM_APP_ID: %v", err)
	}
	appHash := os.Getenv("TELEGRAM_APP_HASH")
	phoneNumber := os.Getenv("TELEGRAM_PHONE_NUMBER")
	password := os.Getenv("TELEGRAM_PASSWORD")

	subscriptionService := subscription.NewSubscriptionService()
	userService := service.NewUserService()
	authService := auth.NewAuthService(userService)
	telegramClient, err := telegram.NewClient(appID, appHash, phoneNumber, password)
	if err != nil {
		logging.Logger.Fatalf("ошибка в создании telegram_client: %v", err)
	}

	botService, err := bot.NewBotService(authService, userService, subscriptionService, telegramClient)
	if err != nil {
		logging.Logger.Fatalf("ошибка в создании bot service: %v", err)
	}
	go botService.Start()

	notificationService := notification.NewNotificationService(userService, botService, telegramClient)
	notificationService.StartCronJobs()

	select {}

}
