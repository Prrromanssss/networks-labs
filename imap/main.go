package main

import (
	"fmt"
	"log"
	"net/mail"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	// Подключаемся к серверу
	c, err := client.DialTLS("imap.example.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	// Аутентификация
	if err := c.Login("your-email@example.com", "your-password"); err != nil {
		log.Fatal(err)
	}

	// Выбор папки
	mailbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Message count: %d\n", mailbox.Messages)

	// Получаем сообщение по индексу
	seqSet := new(imap.SeqSet)
	seqSet.Add("1") // Получаем первое сообщение

	// Используем правильный тип для FetchItem
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBodyStructure}

	messages := make(chan *imap.Message, 1)

	// Асинхронно извлекаем сообщения
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	// Получаем сообщение
	msg := <-messages
	fmt.Printf("From: %s\n", msg.Envelope.From[0].Address())

	// Парсим тело сообщения
	// Для получения текста сообщения используем правильный BodySectionName
	// BODY[TEXT] или BODY.PEEK[TEXT]
	section := imap.BodySectionName{BodyPartName: imap.BodyPartName{Specifier: "BODY", Path: []int{0}}}
	r := msg.GetBody(&section)
	if r == nil {
		log.Fatal("Message body not found")
	}

	msgContent, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}

	// Выводим тему и тело сообщения
	fmt.Println("Subject:", msgContent.Header.Get("Subject"))
	fmt.Println("Body:")
	fmt.Println(msgContent.Body)
}
