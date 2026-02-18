package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Shreehari-Acharya/vayuu/templates"
)

func InitializeTemplates(workDir string) error {
	return fs.WalkDir(templates.EmbeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(workDir, path), 0700)
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		return copyEmbeddedFile(templates.EmbeddedFS, path, filepath.Join(workDir, path))
	})
}

func copyEmbeddedFile(fsys fs.FS, src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}

	content, err := fs.ReadFile(fsys, src)
	if err != nil {
		return fmt.Errorf("read embedded %s: %w", src, err)
	}

	if err := os.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}

	fmt.Printf("  initialized: %s\n", filepath.Base(dst))
	return nil
}

func LoadTemplate(workDir, templateName string) string {
	content, err := os.ReadFile(filepath.Join(workDir, templateName))
	if err != nil {
		return ""
	}
	return string(content)
}
