package bot

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"BirthdayGreetings/internal/auth"
	"BirthdayGreetings/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot          *tgbotapi.BotAPI
	authService  *auth.AuthService
	userService  *service.UserService
	pendingCmd   map[int64]string
	sessionStore map[int64]bool
	mu           sync.Mutex
}

func NewBotService(authService *auth.AuthService, userService *service.UserService) (*BotService, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	return &BotService{
		bot:          bot,
		authService:  authService,
		userService:  userService,
		pendingCmd:   make(map[int64]string),
		sessionStore: make(map[int64]bool),
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
	if cmd, exists := s.pendingCmd[message.Chat.ID]; exists {
		s.handleCommandResponse(message, cmd)
		return
	}

	args := strings.Split(message.Text, " ")
	switch args[0] {
	case "/start":
		s.handleStartCommand(message)
	case "/login":
		s.handleLoginCommand(message)
	case "/register":
		s.handleRegisterCommand(message)
	default:
		if !s.isLoggedIn(message.Chat.ID) {
			msg := tgbotapi.NewMessage(message.Chat.ID, "You need to log in first. Use /login username password.")
			s.bot.Send(msg)
			return
		}
		switch args[0] {
		case "/setbirthday":
			s.handleSetBirthdayCommand(message)
		case "/userslist":
			s.handleUsersListCommand(message)
		case "/subscribe":
			s.handleSubscribeCommand(message)
		case "/unsubscribe":
			s.handleUnsubscribeCommand(message)
		case "/logout":
			s.handeLogoutCommand(message)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command.")
			s.bot.Send(msg)
		}
	}
}

func (s *BotService) isLoggedIn(chatID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionStore[chatID]
}

func (s *BotService) setLoggedIn(chatID int64, status bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionStore[chatID] = status
}

func (s *BotService) handleCommandResponse(message *tgbotapi.Message, cmd string) {
	args := strings.Split(message.Text, " ")

	switch cmd {
	case "/login":
		if len(args) != 2 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid login format. Use username password"))
			s.pendingCmd[message.Chat.ID] = "/login"
			return
		}
		s.handleLoginCommandArgs(message, args)
	case "/register":
		if len(args) != 2 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid register format. Use username password"))
			return
		}
		s.handleRegisterCommandArgs(message, args)
	case "/setbirthday":
		if len(args) != 1 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid format. Use YYYY-MM-DD"))
			s.pendingCmd[message.Chat.ID] = "/setbirthday"
			return
		}
		s.handleSetBirthdayCommandArgs(message, args[0])
	case "/subscribe":
		if len(args) != 1 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid format. Use username"))
			s.pendingCmd[message.Chat.ID] = "/subscribe"
			return
		}
		s.handleSubscribeCommandArgs(message, args[0])
	case "/unsubscribe":
		if len(args) != 1 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid format. Use username"))
			s.pendingCmd[message.Chat.ID] = "/unsubscribe"
			return
		}
		s.handleUnsubscribeCommandArgs(message, args[0])
	}
	delete(s.pendingCmd, message.Chat.ID)
}

func (s *BotService) handleStartCommand(message *tgbotapi.Message) {
	if s.isLoggedIn(message.Chat.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You can now subcribe to users birthday date, unsubscribe, set your birthday date and get all users list")
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome! You can register or login by sending your username and password in the following format:\n/login username password")
	s.bot.Send(msg)
}

func (s *BotService) handleLoginCommand(message *tgbotapi.Message) {
	if s.isLoggedIn(message.Chat.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You are already logged in.")
		s.bot.Send(msg)
		return
	}

	s.pendingCmd[message.Chat.ID] = "/login"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your username and password separated by a space.")
	s.bot.Send(msg)
}

func (s *BotService) handleLoginCommandArgs(message *tgbotapi.Message, args []string) {
	username := args[0]
	password := args[1]
	telegramID := message.From.ID

	name, err := s.authService.AuthenticateUser(username, password, telegramID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Login failed: "+err.Error())
		s.bot.Send(msg)
		return
	}
	s.setLoggedIn(message.Chat.ID, true)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Login successful! Welcome, "+name)
	s.bot.Send(msg)
}

func (s *BotService) handleRegisterCommand(message *tgbotapi.Message) {
	if s.isLoggedIn(message.Chat.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You are already logged in.")
		s.bot.Send(msg)
		return
	}

	user, err := s.userService.GetUserByTgID(message.From.ID)
	if err == nil && user != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You are already registered.")
		s.bot.Send(msg)
		return
	}

	s.pendingCmd[message.Chat.ID] = "/register"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your username and password separated by a space to register.")
	s.bot.Send(msg)
}

func (s *BotService) handleRegisterCommandArgs(message *tgbotapi.Message, args []string) {
	username := args[0]
	password := args[1]
	telegramID := message.From.ID

	err := s.authService.RegisterUser(username, password, telegramID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Registration failed: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Registration successful! You can now login with /login and then update your birthday date with /setbirthday")
	s.bot.Send(msg)
}

func (s *BotService) handleSetBirthdayCommand(message *tgbotapi.Message) {
	s.pendingCmd[message.Chat.ID] = "/setbirthday"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your birthday in the format YYYY-MM-DD.")
	s.bot.Send(msg)
}

func (s *BotService) handleSetBirthdayCommandArgs(message *tgbotapi.Message, birthday string) {
	telegramID := message.From.ID

	err := s.userService.SetUserBirthday(telegramID, birthday)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not set birthday: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Birthday set successfully.")
	s.bot.Send(msg)
}

func (s *BotService) handeLogoutCommand(message *tgbotapi.Message) {
	s.setLoggedIn(message.Chat.ID, false)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Logout")
	s.bot.Send(msg)
}

func (s *BotService) handleUsersListCommand(message *tgbotapi.Message) {
	users, err := s.userService.GetAllUsers()
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not retrieve users: "+err.Error())
		s.bot.Send(msg)
		return
	}

	var userList string
	for _, user := range users {
		if user.TelegramID == message.From.ID {
			continue
		}
		userList += fmt.Sprintf("Username: %s, Birthday: %s\n", user.Username, fmt.Sprint(user.Birthday.Day())+"."+user.Birthday.Month().String()+"."+fmt.Sprint(user.Birthday.Year()))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, userList)
	s.bot.Send(msg)
}

func (s *BotService) handleSubscribeCommand(message *tgbotapi.Message) {
	s.pendingCmd[message.Chat.ID] = "/subscribe"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter the username you want to subscribe to.")
	s.bot.Send(msg)
}

func (s *BotService) handleSubscribeCommandArgs(message *tgbotapi.Message, username string) {
	currentUser, err := s.userService.GetUserByTgID(message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not find current user: "+err.Error())
		s.bot.Send(msg)
		return
	}

	subscribedUser, err := s.userService.GetUserByName(username)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not find user: "+err.Error())
		s.bot.Send(msg)
		return
	}

	if currentUser.TelegramID == subscribedUser.TelegramID {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You can't subscribe yourself birthday date")
		s.bot.Send(msg)
		return
	}

	err = s.userService.SubscribeUser(currentUser.ID, subscribedUser.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not subscribe: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Subscribed to "+subscribedUser.Username+" successfully.")
	s.bot.Send(msg)
}

func (s *BotService) handleUnsubscribeCommand(message *tgbotapi.Message) {
	s.pendingCmd[message.Chat.ID] = "/unsubscribe"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter the username you want to unsubscribe from.")
	s.bot.Send(msg)
}

func (s *BotService) handleUnsubscribeCommandArgs(message *tgbotapi.Message, username string) {
	currentUser, err := s.userService.GetUserByTgID(message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not find current user: "+err.Error())
		s.bot.Send(msg)
		return
	}

	subscribedUser, err := s.userService.GetUserByName(username)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not find user: "+err.Error())
		s.bot.Send(msg)
		return
	}

	err = s.userService.UnubscribeUser(currentUser.ID, subscribedUser.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Could not unsubscribe: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Unsubscribed to "+subscribedUser.Username+" successfully.")
	s.bot.Send(msg)
}
