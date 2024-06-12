package logging

import (
	"log"
	"os"
)

var (
	Logger *log.Logger
)

func Init(logFile string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка в открытии лог файла: %v", err)
	}

	Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Инициализация логгера")
}
