package config

import (
	"crypto/rand"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "vayuu"
	keyringKey     = "encryption_password"
)

// GetOrCreateKeystorePassword gets encryption password from keyring, or creates one if needed
func GetOrCreateKeystorePassword() (string, error) {
	// Try to get existing password from keyring
	password, err := keyring.Get(keyringService, keyringKey)
	if err == nil && password != "" {
		return password, nil
	}

	// If not found, generate a new one
	password, err = GenerateSecurePassword()
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Store in keyring
	if err := keyring.Set(keyringService, keyringKey, password); err != nil {
		fmt.Printf("Warning: Could not store password in system keyring: %v\n", err)
		fmt.Printf("Falling back to password prompt mode.\n")
		// Don't fail - user can use password mode instead
		return "", nil
	}

	return password, nil
}

// GenerateSecurePassword generates a random secure password
func GenerateSecurePassword() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	const length = 32

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randInt(0, len(charset))]
	}
	return string(b), nil
}

// randInt generates a random integer in range [min, max)
func randInt(min, max int) int {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return min + int(b[0])%(max-min)
}

// ClearKeystorePassword removes the password from keyring
func ClearKeystorePassword() error {
	return keyring.Delete(keyringService, keyringKey)
}

// TryGetKeystorePassword attempts to get password from keyring without error
func TryGetKeystorePassword() string {
	password, err := keyring.Get(keyringService, keyringKey)
	if err != nil {
		return ""
	}
	return password
}
