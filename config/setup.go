package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunSetup runs the interactive configuration setup
func RunSetup() error {

	result, err := runSetupForm()
	if err != nil {
		return err
	}

	agentWorkDir := result.AgentWorkDir

	agentWorkDir, err = normalizePath(agentWorkDir)
	if err != nil {
		return err
	}
	if err := ensureWorkDir(agentWorkDir, true); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	config := &Config{
		TelegramToken:   result.TelegramToken,
		ApiKey:          result.ApiKey,
		ApiBaseURL:      result.ApiBaseURL,
		Model:           result.Model,
		AgentWorkDir:    agentWorkDir,
		AllowedUsername: result.AllowedUsername,
	}

	if err := config.validate(); err != nil {
		return err
	}

	configPath := getConfigPath()
	if err := saveToFile(configPath, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to: %s\n", configPath)
	fmt.Println("Permissions set to user-only access (0600)")

	// Initialize default templates in workspace (only if they don't exist)
	fmt.Println("\nInitializing agent templates...")
	if err := InitializeTemplates(agentWorkDir); err != nil {
		fmt.Printf("Warning: failed to initialize templates: %v\n", err)
		// Don't fail setup if templates can't be created
	} else {
		fmt.Println("Templates ready (you can customize them in your workspace)")
	}

	fmt.Println("\nSetup complete! You can now run: vayuu")

	return nil
}


func runSetupForm() (setupResult, error) {

	model := newSetupModel(fields)
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return setupResult{}, err
	}

	resultModel, ok := finalModel.(setupModel)
	if !ok {
		return setupResult{}, fmt.Errorf("unexpected setup model")
	}
	if resultModel.errMsg != "" {
		return setupResult{}, errors.New(resultModel.errMsg)
	}
	if len(resultModel.values) < len(fields) {
		return setupResult{}, fmt.Errorf("setup cancelled")
	}

	return setupResult{
		TelegramToken:   resultModel.values[0],
		AllowedUsername: resultModel.values[1],
		ApiKey:          resultModel.values[2],
		ApiBaseURL:      resultModel.values[3],
		Model:           resultModel.values[4],
		AgentWorkDir:    resultModel.values[5],
	}, nil
}

func newSetupModel(fields []promptRequest) setupModel {
	inputs := make([]textinput.Model, len(fields))
	values := make([]string, len(fields))
	progressBar := progress.New(progress.WithDefaultGradient())

	for i, field := range fields {
		input := textinput.New()
		input.Prompt = "› "
		input.Placeholder = field.Default
		input.CharLimit = 256
		if field.Secret {
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '•'
		}
		if i == 0 {
			input.Focus()
		}
		inputs[i] = input
	}

	return setupModel{
		fields: fields,
		inputs: inputs,
		values: values,
		index:  0,
		progress: progressBar,
	}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.errMsg = "setup cancelled"
			m.done = true
			return m, tea.Quit
		case "enter":
			value := strings.TrimSpace(m.inputs[m.index].Value())
			if value == "" {
				value = m.fields[m.index].Default
			}
			if m.fields[m.index].Required && value == "" {
				m.errMsg = "this field is required"
				return m, nil
			}
			m.errMsg = ""
			m.values[m.index] = value
			m.progress.SetPercent(m.progressPercent())
			if m.index == len(m.inputs)-1 {
				m.done = true
				return m, tea.Quit
			}
			m.inputs[m.index].Blur()
			m.index++
			m.inputs[m.index].Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.index], cmd = m.inputs[m.index].Update(msg)
	return m, cmd
}

func (m setupModel) View() string {
	if m.done {
		return successStyle.Render("Saving configuration...") + "\n"
	}

	current := m.fields[m.index]
	progressLine := m.progress.ViewAs(m.progressPercent())
	lines := []string{
		headerStyle.Render(vayuuASCII),
		subtitleStyle.Render("Let’s configure Vayuu. Press Enter to accept defaults."),
		"",
		fmt.Sprintf("%s %s", stepStyle.Render(fmt.Sprintf("Step %d of %d", m.index+1, len(m.fields))), progressLine),
		"",
		labelStyle.Render(current.Label),
	}
	if current.Default != "" {
		lines = append(lines, hintStyle.Render(fmt.Sprintf("Default: %s", current.Default)))
	}
	if current.Help != "" {
		lines = append(lines, helpStyle.Render(current.Help))
	}
	lines = append(lines, "", inputBoxStyle.Render(m.inputs[m.index].View()))
	if m.errMsg != "" {
		lines = append(lines, errorStyle.Render(fmt.Sprintf("Error: %s", m.errMsg)))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m setupModel) progressPercent() float64 {
	if len(m.fields) == 0 {
		return 0
	}
	return float64(m.index) / float64(len(m.fields))
}

var (
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	stepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	inputBoxStyle = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)
