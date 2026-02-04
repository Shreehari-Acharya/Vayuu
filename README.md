<div align="center">


<img src="assets/vayuu.png" alt="Vayuu Logo" width="600"/>

# Vayuu - Local AI Agent with Telegram Interface

> **A powerful, privacy-focused AI agent that runs on your local machine and uses Telegram as its interface**

</div>

Vayuu is an intelligent AI agent that you deploy on your local machine. It uses Telegram as a convenient chat interface to interact with you, while executing system commands, reading/writing files, browsing the web, creating documents, and managing your local environment. Built to run best with **Ollama** for complete privacy and local control, but compatible with any OpenAI-compatible API provider.

## Key Features

- **Local Deployment** - Runs on your machine, you control everything
- **Telegram Interface** - Chat with your agent from anywhere via Telegram
- **Local-First AI** - Designed for Ollama (free, private, no API costs)
- **Secure by Default** - Encrypted config, system keyring integration
- **System Integration** - Execute commands, manage files, create documents
- **Extensible Skills** - Document generation, web scraping, and more
- **Customizable** - Edit agent personality and behavior without recompiling

## Installation & Setup

### Prerequisites

1. **Go 1.21+** (for installation)
2. **Telegram Bot Token** - Get from [@BotFather](https://t.me/botfather)
3. **LLM Provider** - Choose one:
   - **Ollama Cloud** (Recommended) - Free, local, private
   - OpenAI API - Cloud-based, paid
   - Any OpenAI-compatible API

### Step 1: Install Ollama Cloud (Recommended)

Ollama Cloud provides free, local AI with complete privacy and no API costs.

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama server (must be running for Vayuu to work)
ollama serve

# In a new terminal, pull a model (recommended for Vayuu)
ollama pull kimi-k2.5:cloud

# Or try other models
# ollama pull deepseek-r1:14b
# ollama pull qwen2.5:32b
# ollama pull llama3.3:70b
```

**Important:** The Ollama server must be running before starting Vayuu. Run `ollama serve` in a terminal and keep it running.

**Why Ollama Cloud?**
- Completely free and private
- No API costs or rate limits
- Data never leaves your machine
- Fast local inference
- Multiple model options

### Step 2: Get Telegram Bot Token

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` and follow instructions
3. Copy the bot token (format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

### Step 3: Install Vayuu

```bash
# Install latest version (recommended)
go install github.com/Shreehari-Acharya/vayuu/cmd/vayuu@latest

# Or install specific version
go install github.com/Shreehari-Acharya/vayuu/cmd/vayuu@v0.1.0

# If you get a "package not found" error, bypass the proxy cache:
GOPROXY=direct go install github.com/Shreehari-Acharya/vayuu/cmd/vayuu@latest

# Or build from source
git clone https://github.com/Shreehari-Acharya/vayuu.git
cd vayuu
go build -o vayuu ./cmd/vayuu/

# Run interactive setup
vayuu setup
# Or if built from source: ./vayuu setup
```

### Step 4: Configure Vayuu

The interactive setup wizard will ask for:

1. **Telegram Bot Token** - From @BotFather
2. **Allowed username** - Telegram username allowed to use the bot - Important for security 
2. **API Key** - Press Enter if using Ollama (sets to "ollama")
3. **API Base URL** - Default: `http://localhost:11434/v1` (Ollama)
4. **Model Name** - e.g., `kimi-k2.5:cloud` or `deepseek-r1:14b`
5. **Workspace Directory** - Default: `~/.vayuu/workspace`

**Configuration is automatically:**
- ✅ Encrypted with AES-256-GCM
- ✅ Password stored in system keyring (no password to remember!)
- ✅ Templates copied to workspace for customization

### Step 5: Run Vayuu

```bash
# Start the agent
vayuu
```

The bot will start and connect to Telegram. Send a message to your bot on Telegram to begin interacting!

### Alternative: Using OpenAI or Other Providers

If you prefer to use OpenAI instead of Ollama:

```bash
vayuu setup
```

Then enter:
- **API Key**: Your OpenAI API key (sk-...)
- **API Base URL**: `https://api.openai.com/v1`
- **Model**: `gpt-4` or `gpt-3.5-turbo`

Or use any OpenAI-compatible provider (Anthropic, Together.ai, etc.)


## Available Tools

Vayuu comes with built-in tools that the agent can use:

| Tool | Description | Usage |
|------|-------------|-------|
| **read_file** | Read contents of files | Agent reads configuration, logs, source code |
| **write_file** | Write/create files | Agent creates documents, saves data |
| **edit_file** | Edit files via string replacement | Agent modifies configuration, updates code |
| **execute_command** | Execute bash commands | Agent installs packages, runs scripts |
| **send_file** | Send files to user via Telegram | Agent shares generated documents, logs |

## Skills System

Vayuu has specialized skills for complex tasks. Skills are documented in `~/.vayuu/workspace/skills/` and require external tools.

### Available Skills

#### 1. Visual Document Engineering
**Creates professional PDF, PPTX, and DOCX documents**

**Required Tools:**
```bash
# Install Chromium/Chrome for PDF generation
sudo apt install chromium-browser  # Ubuntu/Debian
# or
brew install chromium              # macOS

# Install Marp CLI for PPTX generation
npm install -g @marp-team/marp-cli
# or use npx without installation:
# npx @marp-team/marp-cli

# For more installation options, see:
# https://github.com/marp-team/marp-cli#install

# Install Pandoc for document conversion
sudo apt install pandoc            # Ubuntu/Debian
# or
brew install pandoc                # macOS
# or download from: https://pandoc.org/installing.html
```

**What it does:**
- Generate PDF reports from HTML
- Create PowerPoint presentations (PPTX) from Markdown using Marp
- Convert between document formats using Pandoc (Markdown, HTML, DOCX, PDF, etc.)
- Professional styling and formatting
- Markdown-based presentation creation

**Documentation:** `skills/visual_doc_eng.md`

#### 2. URL to Markdown Scraper
**Converts web pages to clean markdown**

**Required Tools:**
```bash
# Install Chromium for web scraping
sudo apt install chromium-browser

# Python packages (optional, for enhanced features)
pip install beautifulsoup4 requests
```

**What it does:**
- Scrape public web pages
- Convert to clean markdown
- Extract main content
- Remove ads and clutter

**Documentation:** `skills/url_to_markdown.md`

### Adding Custom Skills

1. Create a new markdown file in `~/.vayuu/workspace/skills/`
2. Document the skill, required tools, and code snippets
3. The agent will automatically discover and use it

Example skill format:
```markdown
# My Custom Skill

## Purpose
Describe what this skill does

## Required Tools
- tool1
- tool2

## Code Snippets
\`\`\`bash
# Example commands
\`\`\`

## Usage Tips
- Tip 1
- Tip 2
```

## Customization

### Agent Personality (SOUL.md)
The agent also updates the contents when you ask it to "update its personality" or "change its behavior".

Edit `~/.vayuu/workspace/SOUL.md` to customize your agent's:
- Identity and name
- Personality traits
- Communication style
- Values and principles
- Behavioral guidelines

Changes take effect immediately - no recompilation needed!

### User Profile (USER.md)
The agent also updates the contents when you ask/tell it more
about yourself.

Edit `~/.vayuu/workspace/USER.md` to tell the agent about you:
- Your preferences
- Work style
- Technical expertise
- Project context
- Communication preferences

### Template Files

All template files are in `~/.vayuu/workspace/`:
- `SOUL.md` - Agent identity and personality
- `USER.md` - User profile and preferences
- `skills/readme.md` - Skills index
- `skills/*.md` - Individual skill documentation

**The agent loads these at runtime**, so you can edit them anytime without restarting.

## Configuration Details

### Configuration File Locations

- **Encrypted Config**: `~/.vayuu/.vayuu.enc` (credentials)
- **Workspace**: `~/.vayuu/workspace/` (templates, memory)
- **Templates**: `~/.vayuu/workspace/*.md` (editable)
- **Memory**: `~/.vayuu/workspace/memory/` (conversation history)

### Environment Variables (Alternative to Setup)

For development or CI/CD:

```bash
export TELEGRAM_TOKEN="123456:ABC-DEF..."
export API_KEY="ollama"                              # or your API key
export API_BASE_URL="http://localhost:11434/v1"     # Ollama
export MODEL="kimi-k2.5:cloud"
export AGENT_WORKDIR="$HOME/.vayuu/workspace"

./vayuu
```

### Updating Configuration

To reconfigure:
```bash
vayuu setup
```

This will:
- Update credentials securely
- Keep your customized templates
- Preserve conversation history

## Security Features

### Automatic Security

- **System Keyring**: Password stored in OS credential manager
  - macOS: Keychain
  - Linux: Secret Service (GNOME Keyring, KDE Wallet) or pass
  - Windows: Credential Manager
- **AES-256-GCM**: Military-grade encryption for credentials
- **File Permissions**: Config files are 0600 (owner-only)
- **Workspace Isolation**: All operations confined to workspace directory
- **Thread-Safe**: Concurrent operations protected with mutexes

### Managing Credentials

**View stored password (macOS):**
```bash
security find-generic-password -s vayuu -w
```

**View stored password (Linux):**
```bash
secret-tool lookup service vayuu key encryption_password
```

**Delete stored password:**
Setup will regenerate it automatically on next run.

## Production Deployment

### Systemd Service (Linux)

Create `/etc/systemd/system/vayuu.service`:

```ini
[Unit]
Description=Vayuu AI Agent Bot
After=network.target

[Service]
Type=simple
User=vayuu
WorkingDirectory=/opt/vayuu
ExecStart=/opt/vayuu/vayuu
Restart=always
RestartSec=10

# Environment (optional, if keyring unavailable)
# Environment="VAYUU_PASSWORD=your-password"

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable vayuu
sudo systemctl start vayuu
sudo systemctl status vayuu
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o vayuu ./cmd/vayuu/

FROM alpine:latest
RUN apk add --no-cache ca-certificates chromium nodejs npm pandoc
RUN npm install -g @marp-team/marp-cli
COPY --from=builder /app/vayuu /usr/local/bin/
COPY templates /app/templates/
WORKDIR /app

ENV AGENT_WORKDIR=/data
VOLUME /data

CMD ["vayuu"]
```

Run with:
```bash
docker run -d \
  -e TELEGRAM_TOKEN="your-token" \
  -e API_KEY="ollama" \
  -e API_BASE_URL="http://host.docker.internal:11434/v1" \
  -e MODEL="kimi-k2.5:cloud" \
  -v vayuu-data:/data \
  vayuu:latest
```

### Ollama with Docker

If running both Vayuu and Ollama in containers:

```yaml
version: '3.8'
services:
  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama-data:/root/.ollama
    ports:
      - "11434:11434"

  vayuu:
    build: .
    environment:
      - TELEGRAM_TOKEN=${TELEGRAM_TOKEN}
      - API_KEY=ollama
      - API_BASE_URL=http://ollama:11434/v1
      - MODEL=kimi-k2.5:cloud
    volumes:
      - vayuu-data:/data
    depends_on:
      - ollama

volumes:
  ollama-data:
  vayuu-data:
```

## Troubleshooting

### Common Issues

**"Config file not found"**
```bash
vayuu setup
```

**"Failed to connect to Ollama Cloud"**
```bash
# Make sure Ollama server is running
ollama serve

# In another terminal, check if models are available
ollama list
```

**"Model not found"**
```bash
# Pull the model
ollama pull kimi-k2.5:cloud
```

**"Skill failed: command not found"**
Install the required tool for that skill (see Skills section)

**"Permission denied"**
```bash
# Fix workspace permissions
chmod 700 ~/.vayuu/workspace
chmod 600 ~/.vayuu/.vayuu.enc
```

**Keyring not available (headless server)**
```bash
# Use environment variable instead
VAYUU_PASSWORD="your-password" ./vayuu
```

### Logs and Debugging

The bot logs to stdout. Capture logs:
```bash
./vayuu 2>&1 | tee vayuu.log
```

Or with systemd:
```bash
journalctl -u vayuu -f
```

## Project Structure

```
vayuu/
├── config/              # Configuration management
│   ├── config.go        # Config loading (env vars, encrypted file)
│   ├── setup.go         # Interactive setup wizard
│   ├── keyring.go       # System keyring integration
│   └── templates.go     # Template management
├── cmd/
│   └── vayuu/           # Application entry point
│       └── main.go      # Main binary (produces "vayuu" executable)
├── internal/            # Private packages (not importable externally)
│   ├── agent/           # LLM agent logic
│   │   ├── agent.go     # Main agent loop
│   │   ├── llm.go       # LLM client wrapper
│   │   └── tools.go     # Tool definitions
│   ├── tools/           # Tool implementations
│   │   ├── common.go    # Shared utilities
│   │   ├── readFile.go
│   │   ├── writeToFile.go
│   │   ├── editFile.go
│   │   ├── executeCommands.go
│   │   └── sendFile.go
│   ├── telegram/        # Telegram bot integration
│   │   ├── bot.go
│   │   ├── handlers.go
│   │   └── sender.go
│   └── prompts/         # System prompts
│       └── prompts.go
└── templates/           # Default templates (distributed with binary)
    ├── SOUL.md          # Agent personality template
    ├── USER.md          # User profile template
    └── skills/          # Skill documentation
        ├── readme.md
        ├── visual_doc_eng.md
        └── url_to_markdown.md
```

## Updates and Maintenance

### Updating Vayuu

```bash
# If installed via go install
go install github.com/Shreehari-Acharya/vayuu/cmd/vayuu@latest

# Or if built from source
git pull
go build -o vayuu ./cmd/vayuu/

# Your config and customizations are preserved!
vayuu
```

### Updating Ollama Cloud Models

```bash
# Update existing model
ollama pull kimi-k2.5:cloud

# Try a new model
ollama pull deepseek-r1:14b

# Update config to use new model
vayuu setup
```

### Backing Up Configuration

```bash
# Backup everything
tar czf vayuu-backup.tar.gz ~/.vayuu/

# Restore
tar xzf vayuu-backup.tar.gz -C ~/
```
Tips and Best Practices

1. **Keep Ollama Server Running** - Make sure `ollama serve` is running before starting Vayuu
2. **Start with Ollama Cloud** - Free, fast, and private
3. **Customize Templates** - Edit SOUL.md to match your needs
4. **Install Skill Tools** - Unlock advanced capabilities
5. **Use Descriptive Requests** - Better prompts = better results
6. **Review Generated Files** - Always verify agent outputs
7. **Keep Workspace Clean** - Periodically review `memory/` folder
8. **Update Regularly** - Pull new features and improvements

##
## Contributing

Contributions are welcome! Areas for contribution:
- New skills and capabilities
- Additional LLM provider integrations
- Bug fixes and improvements



## Acknowledgments

- [Ollama](https://ollama.com) - Local LLM inference
- [OpenAI](https://openai.com) - API compatibility standard
- [Telegram Bot API](https://core.telegram.org/bots/api) - Bot platform

---

