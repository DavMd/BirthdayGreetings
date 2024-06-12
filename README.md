# BirthdayGreetings

BirthdayGreetings - это сервис для удобного поздравления сотрудников с днем рождения.
С помощью данного сервиса можно подписываться на уведомления о днях рождения коллег и получать их через Telegram бота.

## Функциональность

- Регистрация и авторизация пользователей.
- Подписка на уведомления о днях рождения других пользователей.
- Оповещение через Telegram бота о днях рождения.
- Автоматическое создание и управление каналами в Telegram для рассылки уведомлений.

## Установка и настройка

### Предварительные требования

- Go 1.16 или выше
- PostgreSQL
- Telegram Bot API Token

### Файл окружения

```sh
POSTGRES_HOST=db_host
POSTGRES_PORT=db_port
POSTGRES_USER=db_user
POSTGRES_PASSWORD=db_pass
POSTGRES_DB=db_name
TELEGRAM_BOT_TOKEN=TgBot_Token
TELEGRAM_API_ID=tg_api_id
TELEGRAM_API_HASH=tg-api-hash
TELEGRAM_PHONE_NUMBER=tg-phone-number
TELEGRAM_AUTH_PASSWORD=tg-2FA-password
```

## Структура проекта

cmd/app/main.go
Основной файл для запуска приложения.

internal/auth/auth.go
Модуль для регистрации и авторизации пользователей.

internal/bot/bot.go
Модуль для работы с Telegram ботом, включает обработку команд и взаимодействие с пользователями.

internal/db/db.go
Модуль для подключения к базе данных и выполнения запросов.

internal/notification/notification.go
Модуль для управления уведомлениями, использует библиотеку cron для планирования задач.

## Команды бота

- /start - Начало работы с ботом.
- /login <username> <password> - Вход в аккаунт.
- /register <username> <password> - Регистрация нового аккаунта.

Следующие команды доступны только зарегистрированным пользователям.

- /logout - Выйти из аккаунта
- /setbirthday <YYYY-MM-DD> - Установка даты рождения.
- /userslist - Получение списка всех пользователей.
- /subscribe <username> - Подписка на уведомления о днях рождения указанного пользователя.
- /unsubscribe <username> - Отписка от уведомлений о днях рождения указанного пользователя.
- /getallsubscriptions - Получить список всех пользователей, на которых подписан.