package telegram
	
import (
	"github.com/go-telegram/bot"
	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/agent"
	"github.com/Shreehari-Acharya/vayuu/internal/tools"
)

type Bot struct {
	bot           *bot.Bot
	agent         *agent.Agent
	cfg           *config.Config
	toolEnv       *tools.ToolEnv
	currentChatID int64
}
