package main

import (
	"fmt"
	"log"
	"time"

	"github.com/knadh/go-pop3" // Импорт библиотеки
)

func main() {
	// Настройки подключения
	opt := pop3.Opt{
		Host:          "pop.mailserver.com", // Укажите POP3-сервер
		Port:          110,                  // Порт (995 для TLS)
		DialTimeout:   time.Second * 5,      // Таймаут соединения
		TLSEnabled:    false,                // Включите TLS, если нужно
		TLSSkipVerify: false,                // Пропуск проверки сертификата, если нужно
	}

	// Создаём клиента
	client := pop3.New(opt)

	// Устанавливаем соединение
	conn, err := client.NewConn()
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Quit()

	// Авторизация
	username := "your-email@example.com"
	password := "your-password"
	if err := conn.Auth(username, password); err != nil {
		log.Fatalf("Ошибка авторизации: %v", err)
	}
	fmt.Println("Авторизация успешна.")

	// Получаем статистику: количество сообщений и общий размер
	count, size, err := conn.Stat()
	if err != nil {
		log.Fatalf("Ошибка получения статистики: %v", err)
	}
	fmt.Printf("Количество сообщений: %d, общий размер: %d байт\n", count, size)

	// Получаем список всех сообщений
	messages, err := conn.List(0)
	if err != nil {
		log.Fatalf("Ошибка получения списка сообщений: %v", err)
	}

	for _, msg := range messages {
		fmt.Printf("Сообщение ID=%d, размер=%d байт\n", msg.ID, msg.Size)

		// Получаем содержимое сообщения
		rawMessage, err := conn.RetrRaw(msg.ID)
		if err != nil {
			log.Printf("Ошибка получения сообщения ID=%d: %v\n", msg.ID, err)
			continue
		}

		// Выводим содержимое сообщения
		fmt.Printf("Содержимое сообщения ID=%d:\n%s\n", msg.ID, rawMessage.String())
		fmt.Println("---- Конец сообщения ----")
	}
}
