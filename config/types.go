package config

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
)

// Config holds all application configuration
type Config struct {
	TelegramToken   string
	ApiKey          string
	ApiBaseURL      string
	Model           string
	AgentWorkDir    string
	AllowedUsername string
	OllamaBaseURL   string
	OllamaModel     string
}

type promptRequest struct {
	Label    string
	Help     string
	Default  string
	Required bool
	Secret   bool
}

type setupResult struct {
	TelegramToken   string
	AllowedUsername string
	ApiKey          string
	ApiBaseURL      string
	Model           string
	AgentWorkDir    string
}

type setupModel struct {
	fields   []promptRequest
	inputs   []textinput.Model
	values   []string
	index    int
	errMsg   string
	done     bool
	progress progress.Model
}
