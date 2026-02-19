package telegram

import (
	"fmt"
	"os"
)

// validateFileForUpload checks if the file at the given path exists and is within the allowed size limit for Telegram uploads (50 MB). It returns an error if the file does not exist or exceeds the size limit.
func validateFileForUpload(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}
	if info.Size() > maxTelegramFileSize {
		return fmt.Errorf("file too large (%.2f MB, max 50 MB)", float64(info.Size())/(1024*1024))
	}
	return nil
}