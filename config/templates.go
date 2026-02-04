package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type TemplateFile struct {
	Name    string
	Content string
}

// InitializeTemplates copies template files from source to workspace if they don't exist
// This is called during setup and also on first run if templates are missing
func InitializeTemplates(workDir string) error {
	// Get the source templates directory (relative to the binary location)
	sourceDir := getSourceTemplatesDir()

	if sourceDir == "" {
		// If we can't find source templates, create basic ones
		return createDefaultTemplates(workDir)
	}

	// Copy templates from source
	return copyTemplatesFromSource(sourceDir, workDir)
}

// getSourceTemplatesDir finds the templates directory relative to the binary
func getSourceTemplatesDir() string {
	// Try multiple possible locations
	possiblePaths := []string{
		"./templates",
		"../templates",
		"../../templates",
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return ""
}

// copyTemplatesFromSource copies templates from source directory to workspace
func copyTemplatesFromSource(sourceDir, workDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		// If source can't be read, fall back to defaults
		return createDefaultTemplates(workDir)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Handle subdirectories like 'skills'
			subDir := filepath.Join(sourceDir, entry.Name())
			targetDir := filepath.Join(workDir, entry.Name())

			if err := os.MkdirAll(targetDir, 0700); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			// Copy files from subdirectory
			subEntries, _ := os.ReadDir(subDir)
			for _, subEntry := range subEntries {
				if !subEntry.IsDir() {
					if err := copyFile(
						filepath.Join(subDir, subEntry.Name()),
						filepath.Join(targetDir, subEntry.Name()),
					); err != nil {
						return err
					}
				}
			}
		} else {
			// Copy root level files
			if err := copyFile(
				filepath.Join(sourceDir, entry.Name()),
				filepath.Join(workDir, entry.Name()),
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a file only if it doesn't already exist
func copyFile(src, dst string) error {
	// Don't overwrite existing files (respect user customizations)
	if _, err := os.Stat(dst); err == nil {
		return nil // File exists, skip
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}

	// Read source file
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", src, err)
	}

	// Write to destination
	if err := os.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dst, err)
	}

	fmt.Printf("✓ Initialized: %s\n", filepath.Base(dst))
	return nil
}

// createDefaultTemplates creates minimal default templates
func createDefaultTemplates(workDir string) error {
	defaults := map[string]string{
		"SOUL.md": `# SOUL.md — Your Agent Identity

You are Vayuu, an intelligent assistant. Define your personality and values here.`,
		"USER.md": `# USER.md — About Your User

Describe the user you're assisting here.`,
		"skills/readme.md": `# Skills

Document special skills and capabilities here.`,
	}

	for path, content := range defaults {
		fullPath := filepath.Join(workDir, path)

		// Don't overwrite existing files
		if _, err := os.Stat(fullPath); err == nil {
			continue
		}

		// Create directories
		if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
			return err
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}

		fmt.Printf("✓ Created default: %s\n", filepath.Base(path))
	}

	return nil
}

// LoadTemplate loads a template from workspace
// Returns empty string if not found (agent will handle gracefully)
func LoadTemplate(workDir, templateName string) string {
	templatePath := filepath.Join(workDir, templateName)
	
	if content, err := os.ReadFile(templatePath); err == nil {
		return string(content)
	}

	return ""
}

// HasCustomTemplate checks if user has customized a template
func HasCustomTemplate(workDir, templateName string) bool {
	templatePath := filepath.Join(workDir, templateName)
	_, err := os.Stat(templatePath)
	return err == nil
}

// ResetTemplate resets a template by deleting it (will be recreated from source on next init)
func ResetTemplate(workDir, templateName string) error {
	templatePath := filepath.Join(workDir, templateName)

	// Backup current version
	backupPath := templatePath + ".backup"
	if content, err := os.ReadFile(templatePath); err == nil {
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return fmt.Errorf("failed to backup template: %w", err)
		}
		fmt.Printf("✓ Backup created: %s\n", filepath.Base(templatePath)+".backup")
	}

	// Delete the file so it will be recreated from source
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("✓ Template will be reset on next run: %s\n", templateName)
	return nil
}
