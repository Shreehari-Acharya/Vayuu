package config

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	configFileName = "vayuuConfig.json"
	vayuuASCII     = `██╗   ██╗ █████╗ ██╗   ██╗██╗   ██╗██╗   ██╗
██║   ██║██╔══██╗╚██╗ ██╔╝██║   ██║██║   ██║
██║   ██║███████║ ╚████╔╝ ██║   ██║██║   ██║
╚██╗ ██╔╝██╔══██║  ╚██╔╝  ██║   ██║██║   ██║
 ╚████╔╝ ██║  ██║   ██║   ╚██████╔╝╚██████╔╝
  ╚═══╝  ╚═╝  ╚═╝   ╚═╝    ╚═════╝  ╚═════╝ 
                                            `
)

var defaultWorkDirPath, _ = defaultWorkDir()

var (
	fields = []promptRequest{
		{
			Label:    "Telegram Bot Token",
			Help:     "Create a bot at https://t.me/BotFather and paste the token here.",
			Required: true,
			Secret:   true,
		},
		{
			Label:    "Allowed Telegram Username (without @)",
			Help:     "Only this username can talk to the bot.",
			Required: true,
		},
		{
			Label:    "API Key (OpenAI compatible)",
			Help:     "Your LLM provider API key. For Ollama, press enter for ollama",
			Default:  "ollama",
			Required: true,
			Secret:   true,
		},
		{
			Label:    "API Base URL",
			Help:     "Example: http://localhost:11434/v1 for Ollama.",
			Default:  "http://localhost:11434/v1",
			Required: true,
		},
		{
			Label:    "Model Name",
			Help:     "The model identifier your provider expects.",
			Default:  "kimi-k2.5:cloud",
			Required: true,
		},
		{
			Label:    "Agent Work Directory",
			Help:     "Vayuu will read/write templates and files here.",
			Default:  defaultWorkDirPath,
			Required: true,
		},
	}
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	stepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Bold(true)
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	inputBoxStyle = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)
