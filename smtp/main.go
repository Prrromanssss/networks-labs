package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

func main() {
	// Параметры подключения
	smtpServer := "smtp.example.com:587" // Сервер и порт SMTP
	username := "your_email@example.com"
	password := "your_password"
	from := "your_email@example.com"
	to := "recipient@example.com"
	subject := "Hello from Go!"
	body := "This is a test email sent from a Go SMTP client."

	// Создание подключения
	conn, err := net.Dial("tcp", smtpServer)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Обертка TLS для шифрования
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: strings.Split(smtpServer, ":")[0],
	})
	client, err := smtp.NewClient(tlsConn, strings.Split(smtpServer, ":")[0])
	if err != nil {
		fmt.Println("Error creating SMTP client:", err)
		return
	}
	defer client.Quit()

	// Приветствие EHLO
	err = client.Hello("localhost")
	if err != nil {
		fmt.Println("Error with EHLO:", err)
		return
	}

	// Аутентификация с использованием AUTH LOGIN
	auth := smtp.PlainAuth("", username, password, strings.Split(smtpServer, ":")[0])
	err = client.Auth(auth)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		return
	}

	// Установка адреса отправителя и получателя
	err = client.Mail(from)
	if err != nil {
		fmt.Println("Error setting sender:", err)
		return
	}
	err = client.Rcpt(to)
	if err != nil {
		fmt.Println("Error setting recipient:", err)
		return
	}

	// Отправка данных письма
	wc, err := client.Data()
	if err != nil {
		fmt.Println("Error creating data writer:", err)
		return
	}
	defer wc.Close()

	// Форматирование письма
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)
	_, err = wc.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}

	fmt.Println("Email sent successfully!")
}
