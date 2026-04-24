package telegram

// ReplyMarkup is implemented by Telegram keyboard payloads.
// ReplyMarkup được implement bởi các payload keyboard của Telegram.
type ReplyMarkup interface {
	isReplyMarkup()
}

// InlineKeyboardMarkup is Telegram inline keyboard markup.
// InlineKeyboardMarkup là markup cho inline keyboard của Telegram.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

func (*InlineKeyboardMarkup) isReplyMarkup() {}

// InlineKeyboardButton is one button inside an inline keyboard row.
// InlineKeyboardButton là một nút trong một hàng inline keyboard.
type InlineKeyboardButton struct {
	Text                         string `json:"text"`
	URL                          string `json:"url,omitempty"`
	CallbackData                 string `json:"callback_data,omitempty"`
	SwitchInlineQuery            string `json:"switch_inline_query,omitempty"`
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat,omitempty"`
}

// ReplyKeyboardMarkup is Telegram reply keyboard markup.
// ReplyKeyboardMarkup là markup cho reply keyboard của Telegram.
type ReplyKeyboardMarkup struct {
	Keyboard              [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard        bool               `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard       bool               `json:"one_time_keyboard,omitempty"`
	InputFieldPlaceholder string             `json:"input_field_placeholder,omitempty"`
	Selective             bool               `json:"selective,omitempty"`
	IsPersistent          bool               `json:"is_persistent,omitempty"`
}

func (*ReplyKeyboardMarkup) isReplyMarkup() {}

// KeyboardButton is one button in a custom reply keyboard.
// KeyboardButton là một nút trong custom reply keyboard.
type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact,omitempty"`
	RequestLocation bool   `json:"request_location,omitempty"`
}

// ReplyKeyboardRemove removes the current custom keyboard.
// ReplyKeyboardRemove dùng để gỡ custom keyboard hiện tại.
type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
	Selective      bool `json:"selective,omitempty"`
}

func (*ReplyKeyboardRemove) isReplyMarkup() {}

// ForceReply asks Telegram clients to show the reply UI.
// ForceReply yêu cầu Telegram client hiển thị UI reply.
type ForceReply struct {
	ForceReply            bool   `json:"force_reply"`
	InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
	Selective             bool   `json:"selective,omitempty"`
}

func (*ForceReply) isReplyMarkup() {}

// InlineKeyboard creates an inline keyboard markup.
// InlineKeyboard tạo inline keyboard markup.
func InlineKeyboard(rows ...[]InlineKeyboardButton) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: rows}
}

// InlineRow creates one inline keyboard row.
// InlineRow tạo một hàng inline keyboard.
func InlineRow(buttons ...InlineKeyboardButton) []InlineKeyboardButton {
	return buttons
}

// InlineButton creates a callback-data inline button.
// InlineButton tạo inline button dùng callback data.
func InlineButton(text, callbackData string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		CallbackData: callbackData,
	}
}

// URLButton creates a URL inline button.
// URLButton tạo inline button mở URL.
func URLButton(text, rawURL string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text: text,
		URL:  rawURL,
	}
}

// ReplyKeyboard creates a reply keyboard markup.
// ReplyKeyboard tạo reply keyboard markup.
func ReplyKeyboard(rows ...[]KeyboardButton) *ReplyKeyboardMarkup {
	return &ReplyKeyboardMarkup{Keyboard: rows}
}

// ReplyRow creates one reply keyboard row.
// ReplyRow tạo một hàng reply keyboard.
func ReplyRow(buttons ...KeyboardButton) []KeyboardButton {
	return buttons
}

// ReplyButton creates a plain text reply keyboard button.
// ReplyButton tạo nút reply keyboard text thường.
func ReplyButton(text string) KeyboardButton {
	return KeyboardButton{Text: text}
}

// RemoveKeyboard removes the custom reply keyboard.
// RemoveKeyboard gỡ custom reply keyboard.
func RemoveKeyboard(selective bool) *ReplyKeyboardRemove {
	return &ReplyKeyboardRemove{
		RemoveKeyboard: true,
		Selective:      selective,
	}
}

// ForceReplyMarkup creates a force-reply markup.
// ForceReplyMarkup tạo force-reply markup.
func ForceReplyMarkup(selective bool, placeholder string) *ForceReply {
	return &ForceReply{
		ForceReply:            true,
		Selective:             selective,
		InputFieldPlaceholder: placeholder,
	}
}
