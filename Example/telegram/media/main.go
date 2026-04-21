package main

import (
	"context"
	"log"
	"os"
	"time"

	"telegram"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := int64(123456789)

	client, err := telegram.New(token, telegram.WithTimeout(15*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	content := telegram.Content{
		Type:      telegram.ContentPhoto,
		File:      telegram.FileFromURL("https://telegram.org/img/t_logo.png"),
		Caption:   "Telegram logo",
		ParseMode: telegram.MarkdownV2,
	}

	ctx := context.Background()
	if _, err := client.SendChatMessage(ctx, chatID, content); err != nil {
		log.Fatal(err)
	}
}
