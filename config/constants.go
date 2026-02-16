package config

const (
	configFileName = "vayuuConfig.json"
	pathOfTemplate = "../templates"
	vayuuASCII     = `██╗   ██╗ █████╗ ██╗   ██╗██╗   ██╗██╗   ██╗
██║   ██║██╔══██╗╚██╗ ██╔╝██║   ██║██║   ██║
██║   ██║███████║ ╚████╔╝ ██║   ██║██║   ██║
╚██╗ ██╔╝██╔══██║  ╚██╔╝  ██║   ██║██║   ██║
 ╚████╔╝ ██║  ██║   ██║   ╚██████╔╝╚██████╔╝
  ╚═══╝  ╚═╝  ╚═╝   ╚═╝    ╚═════╝  ╚═════╝ 
                                            `

)

var defaultWorkDirPath, _ = defaultWorkDir()

var fields = []promptRequest{
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