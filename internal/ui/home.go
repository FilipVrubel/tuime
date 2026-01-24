package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
		}
	}
	return m, nil
}

func (m Model) viewHome() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("tuime"))
	b.WriteString("\n\n")

	for i, item := range m.Items {
		cursor := "  "
		style := NormalStyle
		if m.Cursor == i {
			cursor = "> "
			style = SelectedStyle
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item.Title)))
		if m.Cursor == i {
			b.WriteString(fmt.Sprintf("    %s\n", HelpStyle.Render(item.Desc)))
		}
	}

	b.WriteString(HelpStyle.Render("\nup/down: navigate • enter: select • q: quit"))

	return b.String()
}
