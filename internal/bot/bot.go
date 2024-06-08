package bot

import (
	"os"
	"strings"

	"BirthdayGreetings/internal/auth"
	"BirthdayGreetings/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot         *tgbotapi.BotAPI
	authService *auth.AuthService
}

func NewBotService(authService *auth.AuthService) (*BotService, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	return &BotService{
		bot:         bot,
		authService: authService,
	}, nil
}

func (s *BotService) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			s.handleMessage(update.Message)
		}
	}
}

func (s *BotService) handleMessage(message *tgbotapi.Message) {
	switch message.Text {
	case "/start":
		s.handleStartCommand(message)
	default:
		s.handleRegistration(message)
	}
}

func (s *BotService) handleStartCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome! Please register by sending your username and password in the following format:\nusername password")
	s.bot.Send(msg)
}

func (s *BotService) handleRegistration(message *tgbotapi.Message) {
	args := strings.Split(message.Text, " ")
	if len(args) != 2 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid format. Please send your username and password  in the following format:\nusername password")
		s.bot.Send(msg)
		return
	}

	username := args[0]
	password := args[1]
	telegramID := message.From.ID

	user, err := s.authService.RegisterUser(username, password, telegramID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Registration failed: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Registration successful! Welcome, "+user.Username)
	s.bot.Send(msg)
}

func (s *BotService) SendMessage(telegramID int64, message string) {
	msg := tgbotapi.NewMessage(telegramID, message)
	_, err := s.bot.Send(msg)
	if err != nil {
		logging.Logger.Printf("could not send message: %v", err)
	}
}
