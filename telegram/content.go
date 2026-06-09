package telegram

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ParseMode controls how Telegram parses formatting entities.
// ParseMode điều khiển cách Telegram parse định dạng.
type ParseMode string

const (
	// HTML enables HTML parsing.
	// HTML bật chế độ parse HTML.
	HTML ParseMode = "HTML"
	// Markdown enables legacy Markdown parsing.
	// Markdown bật chế độ parse Markdown cũ.
	Markdown ParseMode = "Markdown"
	// MarkdownV2 enables Telegram MarkdownV2 parsing.
	// MarkdownV2 bật chế độ parse Telegram MarkdownV2.
	MarkdownV2 ParseMode = "MarkdownV2"
)

func (m ParseMode) validate() error {
	switch m {
	case "", HTML, Markdown, MarkdownV2:
		return nil
	default:
		return fmt.Errorf("telegram: unsupported parse mode %q", m)
	}
}

// ContentType selects which Telegram send* method will be used.
// ContentType chọn Telegram send* method sẽ được gọi.
type ContentType string

const (
	// ContentText sends plain text using sendMessage.
	// ContentText gửi text thường bằng sendMessage.
	ContentText ContentType = "text"
	// ContentHTML sends text using HTML parse mode.
	// ContentHTML gửi text với HTML parse mode.
	ContentHTML ContentType = "html"
	// ContentMarkdown sends text using legacy Markdown parse mode.
	// ContentMarkdown gửi text với Markdown parse mode cũ.
	ContentMarkdown ContentType = "markdown"
	// ContentMarkdownV2 sends text using MarkdownV2 parse mode.
	// ContentMarkdownV2 gửi text với MarkdownV2 parse mode.
	ContentMarkdownV2 ContentType = "markdown_v2"
	// ContentPhoto sends a photo.
	// ContentPhoto gửi ảnh.
	ContentPhoto ContentType = "photo"
	// ContentDocument sends a document or attachment.
	// ContentDocument gửi file document hoặc attachment.
	ContentDocument ContentType = "document"
	// ContentVideo sends a video.
	// ContentVideo gửi video.
	ContentVideo ContentType = "video"
	// ContentAudio sends an audio track.
	// ContentAudio gửi audio track.
	ContentAudio ContentType = "audio"
	// ContentVoice sends a voice note.
	// ContentVoice gửi voice note.
	ContentVoice ContentType = "voice"
	// ContentSticker sends a sticker.
	// ContentSticker gửi sticker.
	ContentSticker ContentType = "sticker"
	// ContentAnimation sends a GIF/animation.
	// ContentAnimation gửi GIF/animation.
	ContentAnimation ContentType = "animation"
	// ContentMediaGroup sends an album of media.
	// ContentMediaGroup gửi album media.
	ContentMediaGroup ContentType = "media_group"
)

// InputFile describes a Telegram file source.
// InputFile mô tả nguồn file cho Telegram.
type InputFile struct {
	FileID      string
	URL         string
	Path        string
	Data        []byte
	Name        string
	ContentType string
}

// FileFromID uses a Telegram file_id without re-uploading the file.
// FileFromID dùng Telegram file_id mà không upload lại file.
func FileFromID(fileID string) *InputFile {
	return &InputFile{FileID: strings.TrimSpace(fileID)}
}

// FileFromURL lets Telegram fetch a file from a remote URL.
// FileFromURL để Telegram tự tải file từ URL.
func FileFromURL(rawURL string) *InputFile {
	return &InputFile{URL: strings.TrimSpace(rawURL)}
}

// FileFromPath uploads a local file path.
// FileFromPath upload file từ đường dẫn cục bộ.
func FileFromPath(filePath string) *InputFile {
	return &InputFile{Path: filePath}
}

// FileFromBytes uploads an in-memory file buffer.
// FileFromBytes upload file từ buffer trong bộ nhớ.
func FileFromBytes(name string, data []byte) *InputFile {
	return &InputFile{Name: name, Data: append([]byte(nil), data...)}
}

func (f *InputFile) validate() error {
	if f == nil {
		return errors.New("telegram: file is required")
	}

	count := 0
	if strings.TrimSpace(f.FileID) != "" {
		count++
	}
	if strings.TrimSpace(f.URL) != "" {
		count++
	}
	if strings.TrimSpace(f.Path) != "" {
		count++
	}
	if len(f.Data) > 0 {
		count++
	}
	if count != 1 {
		return errors.New("telegram: exactly one file source must be provided")
	}
	if f.URL != "" {
		if _, err := url.ParseRequestURI(f.URL); err != nil {
			return fmt.Errorf("telegram: invalid file URL: %w", err)
		}
	}
	if len(f.Data) > 0 && strings.TrimSpace(f.Name) == "" {
		return errors.New("telegram: file name is required for in-memory uploads")
	}
	return nil
}

func (f *InputFile) isUpload() bool {
	return f != nil && (strings.TrimSpace(f.Path) != "" || len(f.Data) > 0)
}

func (f *InputFile) fileName() string {
	if f == nil {
		return "file"
	}
	if strings.TrimSpace(f.Name) != "" {
		return f.Name
	}
	if strings.TrimSpace(f.Path) != "" {
		return filepath.Base(f.Path)
	}
	if strings.TrimSpace(f.URL) != "" {
		parsed, err := url.Parse(f.URL)
		if err == nil {
			base := path.Base(parsed.Path)
			if base != "" && base != "." && base != "/" {
				return base
			}
		}
	}
	return "file"
}

func (f *InputFile) valueOrAttach(attachName string) string {
	if f == nil {
		return ""
	}
	if f.isUpload() {
		return "attach://" + attachName
	}
	if f.FileID != "" {
		return f.FileID
	}
	return f.URL
}

func (f *InputFile) open() (io.ReadCloser, error) {
	if f == nil {
		return nil, errors.New("telegram: file is required")
	}
	if strings.TrimSpace(f.Path) != "" {
		file, err := os.Open(f.Path)
		if err != nil {
			return nil, fmt.Errorf("telegram: open file %q: %w", f.Path, err)
		}
		return file, nil
	}
	if len(f.Data) > 0 {
		return io.NopCloser(bytes.NewReader(f.Data)), nil
	}
	return nil, errors.New("telegram: file source is not uploadable")
}

// MediaType represents an item type inside a media group.
// MediaType biểu diễn kiểu item trong media group.
type MediaType string

const (
	// MediaPhoto represents InputMediaPhoto.
	// MediaPhoto tương ứng InputMediaPhoto.
	MediaPhoto MediaType = "photo"
	// MediaVideo represents InputMediaVideo.
	// MediaVideo tương ứng InputMediaVideo.
	MediaVideo MediaType = "video"
	// MediaDocument represents InputMediaDocument.
	// MediaDocument tương ứng InputMediaDocument.
	MediaDocument MediaType = "document"
	// MediaAudio represents InputMediaAudio.
	// MediaAudio tương ứng InputMediaAudio.
	MediaAudio MediaType = "audio"
)

// MediaItem is one item inside a Telegram media group.
// MediaItem là một phần tử trong Telegram media group.
type MediaItem struct {
	Type      MediaType
	File      *InputFile
	Caption   string
	ParseMode ParseMode
}

func (m MediaItem) validate(defaultParseMode ParseMode) error {
	switch m.Type {
	case MediaPhoto, MediaVideo, MediaDocument, MediaAudio:
	default:
		return fmt.Errorf("telegram: unsupported media group item type %q", m.Type)
	}
	if err := m.File.validate(); err != nil {
		return err
	}
	if mode := m.effectiveParseMode(defaultParseMode); mode != "" {
		if err := mode.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (m MediaItem) effectiveParseMode(defaultParseMode ParseMode) ParseMode {
	if m.ParseMode != "" {
		return m.ParseMode
	}
	return defaultParseMode
}

// Content represents a single Telegram outgoing payload.
// Content biểu diễn một payload gửi ra Telegram.
type Content struct {
	Type        ContentType
	Text        string
	Caption     string
	ParseMode   ParseMode
	File        *InputFile
	Media       []MediaItem
	ReplyMarkup ReplyMarkup

	DisableNotification      bool
	DisableNotificationSet   bool
	DisableWebPagePreview    bool
	DisableWebPagePreviewSet bool
	ProtectContent           bool
	ProtectContentSet        bool
}

func (c Content) validate(defaultParseMode ParseMode) error {
	if c.Type == "" {
		c.Type = ContentText
	}

	switch c.Type {
	case ContentText, ContentHTML, ContentMarkdown, ContentMarkdownV2:
		if strings.TrimSpace(c.Text) == "" {
			return errors.New("telegram: text content cannot be empty")
		}
	case ContentPhoto, ContentDocument, ContentVideo, ContentAudio, ContentVoice, ContentSticker, ContentAnimation:
		if err := c.File.validate(); err != nil {
			return err
		}
	case ContentMediaGroup:
		if len(c.Media) < 2 || len(c.Media) > 10 {
			return errors.New("telegram: media group must contain 2 to 10 items")
		}
		if c.ReplyMarkup != nil {
			return errors.New("telegram: reply markup is not supported for media groups")
		}
		audioOrDoc := ""
		for _, item := range c.Media {
			if err := item.validate(defaultParseMode); err != nil {
				return err
			}
			if item.Type == MediaAudio || item.Type == MediaDocument {
				if audioOrDoc == "" {
					audioOrDoc = string(item.Type)
				}
				if audioOrDoc != string(item.Type) {
					return errors.New("telegram: documents and audios cannot be mixed in one media group")
				}
			}
		}
	default:
		return fmt.Errorf("telegram: unsupported content type %q", c.Type)
	}

	if c.Type == ContentSticker && strings.TrimSpace(c.Caption) != "" {
		return errors.New("telegram: sticker content does not support captions")
	}

	if mode := c.effectiveParseMode(defaultParseMode); mode != "" {
		if err := mode.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c Content) effectiveParseMode(defaultParseMode ParseMode) ParseMode {
	if c.ParseMode != "" {
		return c.ParseMode
	}
	switch c.Type {
	case ContentHTML:
		return HTML
	case ContentMarkdown:
		return Markdown
	case ContentMarkdownV2:
		return MarkdownV2
	default:
		return defaultParseMode
	}
}
