package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) startConfig() {
	m.CurrentView = ConfigView
	m.configCursor = 0
	m.errorMsg = ""

	m.configInputs = make([]textinput.Model, 4)

	labels := []string{"Work Duration (minutes)", "Short Break (minutes)", "Long Break (minutes)", "Cycles Before Long Break"}
	values := []int{m.Config.WorkDuration, m.Config.ShortBreakDuration, m.Config.LongBreakDuration, m.Config.CyclesBeforeLongBreak}

	for i := range m.configInputs {
		ti := textinput.New()
		ti.Placeholder = labels[i]
		ti.CharLimit = 3
		ti.Width = 20
		ti.SetValue(strconv.Itoa(values[i]))

		if i == 0 {
			ti.Focus()
		}

		m.configInputs[i] = ti
	}
}

func (m Model) updateConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "q", "esc":
		if err := m.saveConfigFromInputs(); err != nil {
			m.errorMsg = fmt.Sprintf("Error parsing config: %v", err)
			return m, nil
		}
		if err := m.Config.Save(); err != nil {
			m.errorMsg = fmt.Sprintf("Error saving config: %v", err)
			return m, nil
		}
		m.CurrentView = HomeView
		m.Cursor = 0
		m.errorMsg = ""
		return m, nil

	case "up", "k":
		if m.configCursor > 0 {
			m.configInputs[m.configCursor].Blur()
			m.configCursor--
			m.configInputs[m.configCursor].Focus()
		}

	case "down", "j", "tab":
		if m.configCursor < len(m.configInputs)-1 {
			m.configInputs[m.configCursor].Blur()
			m.configCursor++
			m.configInputs[m.configCursor].Focus()
		}

	case "enter":
		if err := m.saveConfigFromInputs(); err != nil {
			m.errorMsg = fmt.Sprintf("Error parsing config: %v", err)
			return m, nil
		}
		if err := m.Config.Save(); err != nil {
			m.errorMsg = fmt.Sprintf("Error saving config: %v", err)
			return m, nil
		}
		m.CurrentView = HomeView
		m.Cursor = 0
		m.errorMsg = ""
		return m, nil

	default:
		m.configInputs[m.configCursor], cmd = m.configInputs[m.configCursor].Update(msg)
	}

	return m, cmd
}

func (m *Model) saveConfigFromInputs() error {
	val, err := strconv.Atoi(m.configInputs[0].Value())
	if err != nil || val <= 0 {
		return fmt.Errorf("work duration must be a positive number")
	}
	m.Config.WorkDuration = val

	val, err = strconv.Atoi(m.configInputs[1].Value())
	if err != nil || val <= 0 {
		return fmt.Errorf("short break must be a positive number")
	}
	m.Config.ShortBreakDuration = val

	val, err = strconv.Atoi(m.configInputs[2].Value())
	if err != nil || val <= 0 {
		return fmt.Errorf("long break must be a positive number")
	}
	m.Config.LongBreakDuration = val

	val, err = strconv.Atoi(m.configInputs[3].Value())
	if err != nil || val <= 0 {
		return fmt.Errorf("cycles must be a positive number")
	}
	m.Config.CyclesBeforeLongBreak = val

	return nil
}

func (m Model) viewConfig() string {
	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 25
	}

	centerStyle := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)

	var b strings.Builder

	b.WriteString(centerStyle.Render(TitleStyle.Render("Configuration")))
	b.WriteString("\n\n")

	labels := []string{
		"Work Duration (minutes):",
		"Short Break (minutes):",
		"Long Break (minutes):",
		"Cycles Before Long Break:",
	}

	for i, label := range labels {
		cursor := "  "
		if m.configCursor == i {
			cursor = "→ "
		}

		line := fmt.Sprintf("%s%-30s %s", cursor, label, m.configInputs[i].View())
		b.WriteString(centerStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if m.errorMsg != "" {
		b.WriteString(centerStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(m.errorMsg)))
		b.WriteString("\n")
	}

	b.WriteString(centerStyle.Render(HelpStyle.Render("↑/↓/tab navigate • type to edit • enter save • q/esc back")))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}
