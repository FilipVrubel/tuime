package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down":
		if m.Cursor < len(m.Items)-1 {
			m.Cursor++
		}
	case "enter":
		switch m.Items[m.Cursor].Title {
		case "Quit":
			return m, tea.Quit
		case "Time Tracker":
			return m.startTracker()
		case "Pomodoro":
			return m.startPomodoro()
		case "Activities":
			return m.startActivities()
		case "Statistics":
			return m.startStatistics()
		case "Configuration":
			m.startConfig()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) viewHome() string {
	var b strings.Builder
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 25
	}
	centerStyle := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)

	b.WriteString(centerStyle.Render(TitleStyle.Render("tuime")))
	b.WriteString("\n\n")

	for i, item := range m.Items {
		cursor := "  "
		style := NormalStyle
		if m.Cursor == i {
			cursor = "> "
			style = SelectedStyle
		}
		b.WriteString(centerStyle.Render(fmt.Sprintf("%s%s", cursor, style.Render(item.Title))))
		if m.Cursor == i {
			b.WriteString(centerStyle.Render(HelpStyle.Render(item.Desc)))
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("up/down: navigate • enter: select • q: quit")))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}
