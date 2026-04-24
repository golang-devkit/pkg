# Telegram Bot API Client

A simple, thread-safe, and feature-rich Telegram Bot API client for Go. It supports exponential backoff retries, rate-limiting, batch sending, media groups, and various keyboards out of the box.

## Installation

```bash
go get github.com/golang-devkit/pkg/telegram
```

## Initialization & Setup

To start using the library, initialize a new `Client` with your bot token. You can also pass various configuration options:

```go
package main

import (
	"log"
	"time"

	"github.com/golang-devkit/pkg/telegram"
)

func main() {
	client, err := telegram.New(
		"YOUR_BOT_TOKEN_HERE",
		telegram.WithTimeout(15 * time.Second),
		telegram.WithParseMode(telegram.HTML),
		telegram.WithBatchConcurrency(5),
		// Optional: configure retries
		telegram.WithRetry(telegram.RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   300 * time.Millisecond,
			MaxDelay:    3 * time.Second,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	
	// Ready to use client
}
```

## Basic Usage

### Sending Text Messages

You can send a message using either a numeric `Chat ID` (using `SendChatMessage`) or a `@username`/channel handle (using `SendMessage`).

```go
// Using Chat ID
_, err := client.SendChatMessage(ctx, 123456789, telegram.Content{
	Type: telegram.ContentText,
	Text: "Hello, <b>world!</b>",
})

// Using Handle
_, err := client.SendMessage(ctx, "@my_channel_handle", telegram.Content{
	Type: telegram.ContentText,
	Text: "Hello to the channel!",
})
```

### Keyboards & Reply Markups

The library provides a declarative way to create standard inline or reply keyboards.

```go
// Creating an Inline Keyboard
inlineKeyboard := telegram.InlineKeyboard(
	telegram.InlineRow(
		telegram.InlineButton("Click Me", "callback_data_1"),
		telegram.URLButton("Visit Site", "https://example.com"),
	),
)

// Sending a message with the keyboard
_, err = client.SendChatMessage(ctx, 123456789, telegram.Content{
	Type:        telegram.ContentText,
	Text:        "Please select an option:",
	ReplyMarkup: inlineKeyboard,
})
```

## Sending Media

This client supports sending any form of media natively. You can pass a file via Path, URL, File ID, or memory bytes.

### Single Photo/Document

```go
// 1. Send via local file path
_, err = client.SendChatMessage(ctx, 123456789, telegram.Content{
	Type:    telegram.ContentPhoto,
	File:    telegram.FileFromPath("./profile.jpg"),
	Caption: "Look at this photo!",
})

// 2. Send via memory buffer (bytes)
imageData := []byte{...}
_, err = client.SendChatMessage(ctx, 123456789, telegram.Content{
	Type:    telegram.ContentDocument,
	File:    telegram.FileFromBytes("report.pdf", imageData),
	Caption: "Here is your report.",
})

// 3. Send using an existing Telegram File ID
_, err = client.SendChatMessage(ctx, 123456789, telegram.Content{
	Type: telegram.ContentVideo,
	File: telegram.FileFromID("AgACAgUAAxkBA..."),
})
```

### Media Groups (Albums)

To send multiple files in a single album (up to 10):

```go
messages, err := client.SendChatMediaGroup(ctx, 123456789, telegram.Content{
	Type: telegram.ContentMediaGroup,
	Media: []telegram.MediaItem{
		{
			Type:    telegram.MediaPhoto,
			File:    telegram.FileFromPath("./photo1.jpg"),
			Caption: "First Photo",
		},
		{
			Type:    telegram.MediaPhoto,
			File:    telegram.FileFromPath("./photo2.jpg"),
		},
	},
})
```

## Advanced Features

### Batch Sending

When you need to broadcast a message to many users concurrently, use `SendBatch`. It automatically preserves concurrency limits and result ordering.

```go
batchItems := []telegram.BatchItem{
	{ChatID: 111111, Content: telegram.Content{Text: "Message 1"}},
	{ChatID: 222222, Content: telegram.Content{Text: "Message 2"}},
}

results, err := client.SendBatch(ctx, batchItems)
for _, res := range results {
	if res.Error != nil {
		log.Printf("Failed to send to user index %d: %v", res.Index, res.Error)
	} else {
		log.Printf("Successfully sent message %d", res.Message.MessageID)
	}
}
```

## Contributing

See standard guidelines for submitting PRs and checking changes. Ensure all tests in `client_test.go` pass before pushing.
