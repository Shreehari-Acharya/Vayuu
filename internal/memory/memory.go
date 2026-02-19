package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openai/openai-go/v3"
)

// NewFileMemoryWriter creates a FileMemoryWriter rooted at the provided work directory.
func NewFileMemoryWriter(workDir string) *FileMemoryWriter {
	return &FileMemoryWriter{
		Dir:     filepath.Join(workDir, memoryDirName),
		MaxSize: defaultMemoryMaxSize,
		Clock:   time.Now,
	}
}

// Write appends user and assistant messages to the memory store.
func (w *FileMemoryWriter) Write(messages []openai.ChatCompletionMessageParamUnion) error {
	if w == nil {
		return nil
	}
	if w.Dir == "" {
		return fmt.Errorf("memory writer directory is empty")
	}
	if w.Clock == nil {
		w.Clock = time.Now
	}
	if w.MaxSize <= 0 {
		w.MaxSize = defaultMemoryMaxSize
	}

	if err := os.MkdirAll(w.Dir, 0755); err != nil {
		return err
	}

	now := w.Clock()
	filePath := filepath.Join(w.Dir, fmt.Sprintf("%s.jsonl", now.Format(dayFileLayout)))

	if err := w.rotateIfNeeded(filePath); err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	timestamp := now.Format(clockLayout)

	for _, msg := range messages {
		entry, ok := buildMemoryEntry(msg, timestamp)
		if !ok {
			continue
		}

		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}

	return nil
}

func (w *FileMemoryWriter) rotateIfNeeded(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.Size() <= w.MaxSize {
		return nil
	}

	archivePath := fmt.Sprintf("%s.%d", filePath, w.Clock().Unix())
	if err := os.Rename(filePath, archivePath); err != nil {
		return fmt.Errorf("failed to archive memory: %w", err)
	}

	return nil
}

func buildMemoryEntry(msg openai.ChatCompletionMessageParamUnion, timestamp string) (MemoryEntry, bool) {
	entry := MemoryEntry{Timestamp: timestamp}

	switch {
	case msg.OfUser != nil:
		entry.Role = "user"
		if msg.OfUser.Content.OfString.String() != "" {
			entry.Content = msg.OfUser.Content.OfString.Value
		}
	case msg.OfAssistant != nil:
		entry.Role = "assistant"
		if msg.OfAssistant.Content.OfString.String() != "" {
			entry.Content = CleanThinkingTags(msg.OfAssistant.Content.OfString.Value)
		}
	default:
		return MemoryEntry{}, false
	}

	if entry.Content == "" {
		return MemoryEntry{}, false
	}

	return entry, true
}
