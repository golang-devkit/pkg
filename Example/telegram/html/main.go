package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-devkit/pkg/telegram"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	client, err := telegram.New(
		token,
		telegram.WithParseMode(telegram.HTML),
		telegram.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	content := telegram.Content{
		Type: telegram.ContentHTML,
		Text: "<b>Thông báo quan trọng</b>\nHello World!",
		ReplyMarkup: telegram.InlineKeyboard(
			telegram.InlineRow(
				telegram.InlineButton("Acknowledge", "ack"),
				telegram.URLButton("Open Docs", "https://core.telegram.org/bots/api"),
			),
		),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Send(ctx, "@channelusername", content); err != nil {
		log.Fatal(err)
	}
}
