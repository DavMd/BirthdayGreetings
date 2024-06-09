package main

import (
	"BirthdayGreetings/internal/auth"
	"BirthdayGreetings/internal/bot"
	"BirthdayGreetings/internal/db"
	"BirthdayGreetings/internal/logging"
	"BirthdayGreetings/internal/service"

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

	authService := auth.NewAuthService()
	userService := service.NewUserService()

	botService, err := bot.NewBotService(authService, userService)
	if err != nil {
		logging.Logger.Fatalf("could not create bot service: %v", err)
	}
	go botService.Start()

	logging.Logger.Println("Birthday Greetings Service is running")

	select {}

	//logging.Logger.Println("Birthday Greetings Service is stop")

}
