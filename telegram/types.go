package telegram

// ResponseParameters contains special Telegram error metadata.
// ResponseParameters chứa metadata lỗi đặc biệt từ Telegram.
type ResponseParameters struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int   `json:"retry_after,omitempty"`
}

type apiResponse[T any] struct {
	Ok          bool                `json:"ok"`
	Result      T                   `json:"result"`
	Description string              `json:"description,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
}

// User is a minimal Telegram user object returned by send methods.
// User là object user Telegram tối thiểu được trả về bởi send methods.
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// Chat is a minimal Telegram chat object.
// Chat là object chat Telegram tối thiểu.
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// MessageEntity describes a formatting entity inside text or captions.
// MessageEntity mô tả entity định dạng trong text hoặc caption.
type MessageEntity struct {
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	URL           string `json:"url,omitempty"`
	Language      string `json:"language,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
	User          *User  `json:"user,omitempty"`
}

// PhotoSize is a Telegram photo size variant.
// PhotoSize là một biến thể kích thước ảnh trong Telegram.
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size,omitempty"`
}

// Audio is the Telegram audio object.
// Audio là object audio của Telegram.
type Audio struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Duration     int        `json:"duration"`
	Performer    string     `json:"performer,omitempty"`
	Title        string     `json:"title,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     int        `json:"file_size,omitempty"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
}

// Document is the Telegram document object.
// Document là object document của Telegram.
type Document struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     int        `json:"file_size,omitempty"`
}

// Video is the Telegram video object.
// Video là object video của Telegram.
type Video struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Duration     int        `json:"duration"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     int        `json:"file_size,omitempty"`
}

// Voice is the Telegram voice object.
// Voice là object voice của Telegram.
type Voice struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Duration     int    `json:"duration"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
}

// Sticker is the Telegram sticker object.
// Sticker là object sticker của Telegram.
type Sticker struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Type         string     `json:"type,omitempty"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	IsAnimated   bool       `json:"is_animated,omitempty"`
	IsVideo      bool       `json:"is_video,omitempty"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	Emoji        string     `json:"emoji,omitempty"`
	SetName      string     `json:"set_name,omitempty"`
}

// Animation is the Telegram animation object.
// Animation là object animation của Telegram.
type Animation struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Duration     int        `json:"duration"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     int        `json:"file_size,omitempty"`
}

// Message is a minimal, send-focused Telegram message object.
// Message là object message Telegram tối thiểu, tập trung cho use-case gửi tin.
type Message struct {
	MessageID       int                   `json:"message_id"`
	MessageThreadID int                   `json:"message_thread_id,omitempty"`
	Date            int64                 `json:"date"`
	Chat            Chat                  `json:"chat"`
	From            *User                 `json:"from,omitempty"`
	SenderChat      *Chat                 `json:"sender_chat,omitempty"`
	Text            string                `json:"text,omitempty"`
	Caption         string                `json:"caption,omitempty"`
	Entities        []MessageEntity       `json:"entities,omitempty"`
	CaptionEntities []MessageEntity       `json:"caption_entities,omitempty"`
	Photo           []PhotoSize           `json:"photo,omitempty"`
	Audio           *Audio                `json:"audio,omitempty"`
	Document        *Document             `json:"document,omitempty"`
	Video           *Video                `json:"video,omitempty"`
	Voice           *Voice                `json:"voice,omitempty"`
	Sticker         *Sticker              `json:"sticker,omitempty"`
	Animation       *Animation            `json:"animation,omitempty"`
	MediaGroupID    string                `json:"media_group_id,omitempty"`
	ReplyMarkup     *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// BatchItem is one outbound item for SendBatch.
// BatchItem là một item outbound cho SendBatch.
type BatchItem struct {
	To      string
	ChatID  int64
	Content Content
}

// BatchResult contains the per-item outcome of SendBatch.
// BatchResult chứa kết quả theo từng item của SendBatch.
type BatchResult struct {
	Index    int
	To       string
	ChatID   int64
	Message  *Message
	Messages []Message
	Error    error
}
