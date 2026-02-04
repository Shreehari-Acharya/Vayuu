package config

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const (
	configFileName = ".vayuu.enc"
)

// RunSetup runs the interactive configuration setup
func RunSetup() error {
	fmt.Println("=== Vayuu Configuration Setup ===")
	fmt.Println("This wizard will help you configure Vayuu securely.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Collect configuration values
	telegramToken, err := promptSecret("Telegram Bot Token", "")
	if err != nil {
		return err
	}

	allowedUsername, err := promptInput(reader, "Allowed Telegram Username (without @)", "")
	if err != nil {
		return err
	}

	apiKey, err := promptSecret("API Key - openAI compatible services. Press enter if using ollama", "ollama")
	if err != nil {
		return err
	}

	apiBaseURL, err := promptInput(reader, "API Base URL - openAI compatible", "http://localhost:11434/v1")
	if err != nil {
		return err
	}

	model, err := promptInput(reader, "Model Name", "kimi-k2.5:cloud")
	if err != nil {
		return err
	}

	defaultWorkDir := filepath.Join(os.Getenv("HOME"), ".vayuu", "workspace")
	agentWorkDir, err := promptInput(reader, "Agent Work Directory", defaultWorkDir)
	if err != nil {
		return err
	}

	// Expand tilde in path
	if strings.HasPrefix(agentWorkDir, "~/") {
		home, _ := os.UserHomeDir()
		agentWorkDir = filepath.Join(home, agentWorkDir[2:])
	}

	// Create work directory if it doesn't exist
	if err := os.MkdirAll(agentWorkDir, 0700); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Prepare config data
	configData := fmt.Sprintf("TELEGRAM_TOKEN=%s\nAPI_KEY=%s\nAPI_BASE_URL=%s\nMODEL=%s\nAGENT_WORKDIR=%s\nALLOWED_USERNAME=%s\n",
		telegramToken, apiKey, apiBaseURL, model, agentWorkDir, allowedUsername)

	// Get or create encryption password using system keyring
	fmt.Println("\nSecuring configuration...")
	password, err := GetOrCreateKeystorePassword()
	if err != nil {
		return err
	}

	// If keyring is not available, ask for password
	if password == "" {
		fmt.Print("Enter encryption password (or press Enter for auto-generated): ")
		passBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		password = strings.TrimSpace(string(passBytes))
		if password == "" {
			password, err = GenerateSecurePassword()
			if err != nil {
				return err
			}
			fmt.Printf("Generated password: %s\n", password)
		} else if len(password) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}
	}

	// Encrypt and save
	configPath := getConfigPath()
	if err := encryptAndSave(configData, password, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✓ Configuration saved securely to: %s\n", configPath)
	fmt.Println("✓ Password stored in system keyring - no need to remember it!")
	fmt.Println("✓ Configuration will be automatically unlocked when you run Vayuu")

	// Initialize default templates in workspace (only if they don't exist)
	fmt.Println("\nInitializing agent templates...")
	if err := InitializeTemplates(agentWorkDir); err != nil {
		fmt.Printf("Warning: failed to initialize templates: %v\n", err)
		// Don't fail setup if templates can't be created
	} else {
		fmt.Println("✓ Templates ready (you can customize them in your workspace)")
	}

	fmt.Println("\n✓ Setup complete! You can now run: ./vayuu")

	return nil
}

// LoadEncryptedConfig loads configuration from encrypted file
func LoadEncryptedConfig(password string) (*Config, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found. Run 'vayuu setup' first")
	}

	// Decrypt config
	configData, err := decryptAndLoad(configPath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt config (wrong password?): %w", err)
	}

	// Parse config data
	cfg := &Config{}
	lines := strings.SplitSeq(configData, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		switch key {
		case "TELEGRAM_TOKEN":
			cfg.TelegramToken = value
		case "API_KEY":
			cfg.ApiKey = value
		case "API_BASE_URL":
			cfg.ApiBaseURL = value
		case "MODEL":
			cfg.Model = value
		case "AGENT_WORKDIR":
			cfg.AgentWorkDir = value
		case "ALLOWED_USERNAME":
			cfg.AllowedUsername = value
		}
		
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Helper functions

func promptInput(reader *bufio.Reader, prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return input, nil
}

func promptSecret(prompt, defaultValue string) (string, error) {
	fmt.Printf("%s: ", prompt)
	secret, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	value := strings.TrimSpace(string(secret))
	if value == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return value, nil
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vayuu", configFileName)
}

func encryptAndSave(data, password, filepath string) error {
	// Ensure directory exists
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Derive key from password
	key := sha256.Sum256([]byte(password))

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Encode to base64 for storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Write to file with restricted permissions
	return os.WriteFile(filepath, []byte(encoded), 0600)
}

func decryptAndLoad(filepath, password string) (string, error) {
	// Read file
	encoded, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return "", err
	}

	// Derive key from password
	key := sha256.Sum256([]byte(password))

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
