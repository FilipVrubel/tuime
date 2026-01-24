package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startTracker() (tea.Model, tea.Cmd) {
	m.CurrentView = TimeTrackerView
	m.Running = true
	m.StartTime = time.Now()
	m.Elapsed = 0
	return m, TickCmd()
}

func (m Model) updateTracker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case " ":
		if m.Running {
			m.Running = false
		} else {
			m.StartTime = time.Now().Add(-m.Elapsed)
			m.Running = true
			return m, TickCmd()
		}
	case "esc":
		m.Running = false
		m.CurrentView = HomeView
	}
	return m, nil
}

func (m Model) viewTracker() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Time Tracker"))
	b.WriteString("\n\n")

	b.WriteString(TimeStyle.Render(FormatDuration(m.Elapsed)))
	b.WriteString("\n\n")

	status := "[Running]"
	if !m.Running {
		status = "[Paused]"
	}
	b.WriteString(NormalStyle.Render(status))
	b.WriteString("\n")

	b.WriteString(HelpStyle.Render("\nspace: pause/resume • esc: stop & back"))

	return b.String()
}
