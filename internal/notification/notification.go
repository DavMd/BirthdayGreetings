package notification

import (
	"BirthdayGreetings/internal/bot"
	"BirthdayGreetings/internal/logging"
	"BirthdayGreetings/internal/service"
	"BirthdayGreetings/internal/telegram"
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type NotificationService struct {
	botService      *bot.BotService
	telegramService *telegram.Client
	userService     *service.UserService
	cronScheduler   *cron.Cron
}

func NewNotificationService(userService *service.UserService, botService *bot.BotService, telegramService *telegram.Client) *NotificationService {
	return &NotificationService{
		userService:     userService,
		botService:      botService,
		telegramService: telegramService,
		cronScheduler:   cron.New(cron.WithSeconds()),
	}
}

func (s *NotificationService) StartCronJobs() {
	_, err := s.cronScheduler.AddFunc("0 0 9 * * *", func() {
		s.handleDailyBirthdayNotifications()
	})
	if err != nil {
		logging.Logger.Fatalf("Ошибка в установке уведомлений: %v", err)
	}

	s.cronScheduler.Start()
}

func (s *NotificationService) handleDailyBirthdayNotifications() {
	ctx := context.Background()

	today := time.Now().Add(time.Hour * 24)
	todayDate := fmt.Sprint(today.Day()) + " " + today.Month().String()
	users, err := s.userService.GetUsersWithBirthday(today.Format("2000-01-02"))
	if err != nil {
		logging.Logger.Printf(err.Error())
		return
	}

	channel, err := s.telegramService.CreateChannel(ctx, "Поздравление с днем рождения", "Канал для уведомления о днем рождении пользователей")
	if err != nil {
		logging.Logger.Printf("Ошибка в создании канала: %v", err)
		return
	}

	allUsers, err := s.userService.GetAllUsers()
	if err != nil {
		logging.Logger.Printf(err.Error())
		return
	}

	userTgIDs := make([]int64, 0, len(allUsers))
	for _, user := range allUsers {
		if user.TelegramID != s.botService.GetAdminID() || fmt.Sprint(user.Birthday.Day())+" "+user.Birthday.Month().String() == todayDate {
			userTgIDs = append(userTgIDs, user.TelegramID)
		}
	}

	err = s.telegramService.AddUsersToChannel(ctx, channel, userTgIDs)
	if err != nil {
		logging.Logger.Printf("Ошибка в добавлении пользователей в канал: %v", err)
		return
	}

	err = s.telegramService.AddBotToChannel(ctx, channel, s.botService.GetBotID())
	if err != nil {
		logging.Logger.Printf("Ошибка в добавлении бота в канал: %v", err)
		return
	}

	if len(users) > 0 {
		birthdayUsernames := ""
		for _, user := range users {
			birthdayUsernames += user.Username + ", "
		}
		birthdayUsernames = birthdayUsernames[:len(birthdayUsernames)-2]

		message := "Сегодня день рождение у " + birthdayUsernames + " 🎉"
		err = s.botService.SendMessageToChannel(ctx, channel.ID, message)
		if err != nil {
			logging.Logger.Printf("Ошибка в отправлении сообщения в канал: %v", err)
			return
		}
	}

	logging.Logger.Println("Уведомления успешно отправлено.")
}

func (s *NotificationService) StopCronJobs() {
	s.cronScheduler.Stop()
}
