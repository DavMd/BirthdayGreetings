package bot

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"BirthdayGreetings/internal/auth"
	"BirthdayGreetings/internal/logging"
	"BirthdayGreetings/internal/service"
	"BirthdayGreetings/internal/subscription"
	"BirthdayGreetings/internal/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot            *tgbotapi.BotAPI
	authService    *auth.AuthService
	userService    *service.UserService
	subService     *subscription.SubscriptionService
	telegramClient *telegram.Client
	pendingCmd     map[int64]string
	sessionStore   map[int64]bool
	mu             sync.Mutex
	adminID        int64
}

func NewBotService(authService *auth.AuthService, userService *service.UserService, subService *subscription.SubscriptionService, telegramClient *telegram.Client) (*BotService, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	adminID, err := strconv.Atoi(os.Getenv("ADMIN_ID"))
	if err != nil {
		logging.Logger.Fatalf("Ошибка парсинга ADMIN_ID: %v", err)
	}

	return &BotService{
		bot:            bot,
		authService:    authService,
		userService:    userService,
		subService:     subService,
		telegramClient: telegramClient,
		pendingCmd:     make(map[int64]string),
		sessionStore:   make(map[int64]bool),
		adminID:        int64(adminID),
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
			msg := tgbotapi.NewMessage(message.Chat.ID, "Вы должны сначала ввойти в аккаунт.\nИспользуйте команду /login username password.")
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
		case "/getallsubscriptions":
			s.handleGetAllUserSubscriptions(message)
		case "/logout":
			s.handeLogoutCommand(message)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда.")
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
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неверный формат входа. Используйте username password"))
			s.pendingCmd[message.Chat.ID] = "/login"
			return
		}
		s.handleLoginCommandArgs(message, args)
	case "/register":
		if len(args) != 2 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неверный формат регистрации. Используйте username password"))
			return
		}
		s.handleRegisterCommandArgs(message, args)
	case "/setbirthday":
		if len(args) != 1 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неверный формат даты. Используйте YYYY-MM-DD"))
			s.pendingCmd[message.Chat.ID] = "/setbirthday"
			return
		}
		s.handleSetBirthdayCommandArgs(message, args[0])
	case "/subscribe":
		if len(args) != 1 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неверный формат. Используйте username"))
			s.pendingCmd[message.Chat.ID] = "/subscribe"
			return
		}
		s.handleSubscribeCommandArgs(message, args[0])
	case "/unsubscribe":
		if len(args) != 1 {
			s.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неверный формат. Используйте username"))
			s.pendingCmd[message.Chat.ID] = "/unsubscribe"
			return
		}
		s.handleUnsubscribeCommandArgs(message, args[0])
	}
	delete(s.pendingCmd, message.Chat.ID)
}

func (s *BotService) handleStartCommand(message *tgbotapi.Message) {
	if s.isLoggedIn(message.Chat.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы можете теперь подписываться и отписываться на день рождения других пользователей, получать список своих подписок и других пользователей.")
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Добро пожаловать в бота BirthdayGreetings. Вы можете зарегистрироваться либо войти в свой аккаунт. Для этого введите /login или /register username и password.")
	s.bot.Send(msg)
}

func (s *BotService) handleLoginCommand(message *tgbotapi.Message) {
	if s.isLoggedIn(message.Chat.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже в аккаунте.")
		s.bot.Send(msg)
		return
	}

	s.pendingCmd[message.Chat.ID] = "/login"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйства введите username и password.")
	s.bot.Send(msg)
}

func (s *BotService) handleLoginCommandArgs(message *tgbotapi.Message, args []string) {
	username := args[0]
	password := args[1]
	telegramID := message.From.ID

	name, err := s.authService.AuthenticateUser(username, password, telegramID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка входа: "+err.Error())
		s.bot.Send(msg)
		return
	}
	s.setLoggedIn(message.Chat.ID, true)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Вход успешный. Добро пожаловать, "+name)
	s.bot.Send(msg)
}

func (s *BotService) handleRegisterCommand(message *tgbotapi.Message) {
	if s.isLoggedIn(message.Chat.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже вошли в аккаунт.")
		s.bot.Send(msg)
		return
	}

	user, err := s.userService.GetUserByTgID(message.From.ID)
	if err == nil && user != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ваш аккаунт уже зарегистрирован.")
		s.bot.Send(msg)
		return
	}

	s.pendingCmd[message.Chat.ID] = "/register"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйства введите username и password.")
	s.bot.Send(msg)
}

func (s *BotService) handleRegisterCommandArgs(message *tgbotapi.Message, args []string) {
	username := args[0]
	password := args[1]
	telegramID := message.From.ID

	err := s.authService.RegisterUser(username, password, telegramID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка регистрации: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Регистрация успешна! Теперь вы можете войти, используя команду /login.\nСразу же после этого введите вашу дату рождения командой /setbirthday")
	s.bot.Send(msg)
}

func (s *BotService) handleSetBirthdayCommand(message *tgbotapi.Message) {
	s.pendingCmd[message.Chat.ID] = "/setbirthday"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите дату рождения в формате YYYY-MM-DD.")
	s.bot.Send(msg)
}

func (s *BotService) handleSetBirthdayCommandArgs(message *tgbotapi.Message, birthday string) {
	telegramID := message.From.ID

	err := s.userService.SetUserBirthday(telegramID, birthday)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка обновления даты рождения: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Дата рождения успешна изменена.")
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
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка в получении пользователей: "+err.Error())
		s.bot.Send(msg)
		return
	}

	var userList string
	for _, user := range users {
		if user.TelegramID == message.From.ID {
			continue
		}
		userList += fmt.Sprintf("Пользователь: %s, Дата рождения: %s\n", user.Username, fmt.Sprint(user.Birthday.Day())+"."+user.Birthday.Month().String()+"."+fmt.Sprint(user.Birthday.Year()))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, userList)
	s.bot.Send(msg)
}

func (s *BotService) handleSubscribeCommand(message *tgbotapi.Message) {
	s.pendingCmd[message.Chat.ID] = "/subscribe"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите имя пользователя на которого хотите подписаться.")
	s.bot.Send(msg)
}

func (s *BotService) handleSubscribeCommandArgs(message *tgbotapi.Message, username string) {
	currentUser, err := s.userService.GetUserByTgID(message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка в поиске пользователя: "+err.Error())
		s.bot.Send(msg)
		return
	}

	subscribedUser, err := s.userService.GetUserByName(username)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Не удалось найти пользователя: "+err.Error())
		s.bot.Send(msg)
		return
	}

	if currentUser.TelegramID == subscribedUser.TelegramID {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы не можете подписаться сами на себя")
		s.bot.Send(msg)
		return
	}

	err = s.subService.SubscribeUser(currentUser.ID, subscribedUser.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка при подписке: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Успешная подписка на пользователя "+subscribedUser.Username+".")
	s.bot.Send(msg)
}

func (s *BotService) handleUnsubscribeCommand(message *tgbotapi.Message) {
	s.pendingCmd[message.Chat.ID] = "/unsubscribe"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите имя пользователя от которого хотите отписаться.")
	s.bot.Send(msg)
}

func (s *BotService) handleUnsubscribeCommandArgs(message *tgbotapi.Message, username string) {
	currentUser, err := s.userService.GetUserByTgID(message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка в поиске пользователя: "+err.Error())
		s.bot.Send(msg)
		return
	}

	subscribedUser, err := s.userService.GetUserByName(username)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка в поиске пользователя: "+err.Error())
		s.bot.Send(msg)
		return
	}

	err = s.subService.UnsubscribeUser(currentUser.ID, subscribedUser.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Отписка не удалась: "+err.Error())
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Вы успешно отписались от пользователя "+subscribedUser.Username+".")
	s.bot.Send(msg)
}

func (s *BotService) handleGetAllUserSubscriptions(message *tgbotapi.Message) {

	user, err := s.userService.GetUserByTgID(message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Не удалось получить пользователя: "+err.Error())
		s.bot.Send(msg)
		return
	}
	subscriptions, err := s.subService.GetSubscriptions(user.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Не удалось получить подписки пользователя: "+err.Error())
		s.bot.Send(msg)
		return
	}
	returnMessage := "Пользователь подписан на:\n"
	for _, sub := range subscriptions {
		returnMessage += sub.Username + "\n"
	}
	logging.Logger.Println(len(subscriptions))
	logging.Logger.Println(returnMessage)
	msg := tgbotapi.NewMessage(message.Chat.ID, returnMessage)
	s.bot.Send(msg)

}

func (s *BotService) SendMessageToChannel(ctx context.Context, channelID int64, message string) error {
	msg := tgbotapi.NewMessage(channelID, message)
	_, err := s.bot.Send(msg)
	return err
}

func (s *BotService) GetBotID() int64 {
	return s.bot.Self.ID
}

func (s *BotService) GetAdminID() int64 {
	return s.adminID
}
