package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type menuItem struct {
	title string
	desc  string
}

type model struct {
	cursor int
	items  []menuItem
}

func initialModel() model {
	return model{
		items: []menuItem{
			{title: "Pomodoro", desc: "Start a focused work session"},
			{title: "Time Tracker", desc: "Track time manually"},
			{title: "Activities", desc: "Manage activity types"},
			{title: "Statistics", desc: "View your progress"},
			{title: "Quit", desc: "Exit the application"},
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if m.items[m.cursor].title == "Quit" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Welcome to Tuime"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if m.cursor == i {
			cursor = "> "
			style = selectedStyle
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item.title)))
		if m.cursor == i {
			b.WriteString(fmt.Sprintf("    %s\n", helpStyle.Render(item.desc)))
		}
	}

	b.WriteString(helpStyle.Render("\nup/down: navigate • enter: select • q: quit"))

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
