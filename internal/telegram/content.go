package telegram

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ContentType represents the type of content being sent to Telegram (e.g., image, document, video).
func DetectContentType(pathOrContent string) (ContentType, error) {
	ext := strings.ToLower(filepath.Ext(pathOrContent))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return ContentTypeImage, nil
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".flv":
		return ContentTypeVideo, nil
	case ".pdf", ".txt", ".md", ".zip", ".rar", ".7z", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return ContentTypeDoc, nil
	default:
		return ContentTypeDoc, nil
	}
}

// ContentType defines the type of content being sent to Telegram, such as image, document, or video.
func (c ContentType) String() string {
	return string(c)
}

// IsValid checks if the ContentType is one of the recognized types (image, document, video).
func (c ContentType) IsValid() bool {
	switch c {
	case ContentTypeImage, ContentTypeDoc, ContentTypeVideo:
		return true
	default:
		return false
	}
}

// EncodeContentEntries encodes a slice of MemoryEntry into the provided encoder, skipping entries with empty content.
func ValidateContentType(hint ContentType, detected ContentType) (ContentType, error) {
	if !hint.IsValid() {
		return detected, fmt.Errorf("invalid content type hint: %s", hint)
	}
	if hint != "" && hint != detected {
		return detected, fmt.Errorf("content type mismatch: got %s but detected %s, using detected", hint, detected)
	}
	return detected, nil
}
