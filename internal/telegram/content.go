package telegram

import (
	"fmt"
	"path/filepath"
	"strings"
)

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

func (c ContentType) String() string {
	return string(c)
}

func (c ContentType) IsValid() bool {
	switch c {
	case ContentTypeImage, ContentTypeDoc, ContentTypeVideo:
		return true
	default:
		return false
	}
}

func ValidateContentType(hint ContentType, detected ContentType) (ContentType, error) {
	if !hint.IsValid() {
		return detected, fmt.Errorf("invalid content type hint: %s", hint)
	}
	if hint != "" && hint != detected {
		return detected, fmt.Errorf("content type mismatch: got %s but detected %s, using detected", hint, detected)
	}
	return detected, nil
}
