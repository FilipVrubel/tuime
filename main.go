package main

import (
	"fmt"
	"strings"
	"time"

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

type view int

const (
	homeView = iota
	pomodoroView
	timeTrackerView
	activitiesView
	statisticsView
)
type model struct {
	cursor int
	items  []menuItem
	currentView view

	running bool
	startTime time.Time
	elapsed time.Duration
}

type tickMsg time.Time

func initialModel() model {
	return model{
		items: []menuItem{
			{title: "Pomodoro", desc: "Start a focused work session"},
			{title: "Time Tracker", desc: "Track time manually"},
			{title: "Activities", desc: "Manage activity types"},
			{title: "Statistics", desc: "View your progress"},
			{title: "Quit", desc: "Exit the application"},
		},
		currentView: homeView,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "q":
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
        switch m.items[m.cursor].title {
        case "Quit":
            return m, tea.Quit
        case "Time Tracker":
            m.currentView = timeTrackerView
            m.running = true
            m.startTime = time.Now()
            m.elapsed = 0
            return m, tickCmd() 
        }
    }
    return m, nil
}

func (m model) updateTracker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case " ":
        if m.running {
            m.running = false
        } else {
            m.startTime = time.Now().Add(-m.elapsed)
            m.running = true
            return m, tickCmd()
        }
    case "esc":
        m.running = false
        m.currentView = homeView
    }
    return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    
    case tickMsg:
        if m.running {
            m.elapsed = time.Since(m.startTime)
            return m, tickCmd()
        }
        return m, nil

    case tea.KeyMsg:
        if msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
        switch m.currentView {
        case homeView:
            return m.updateHome(msg)
        case timeTrackerView:
            return m.updateTracker(msg)
        }
    }
    return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case timeTrackerView:
		return m.viewTracker()
	default:
		return m.viewHome()
	}
}

func (m model) viewHome() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("tuime"))
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

func (m model) viewTracker() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Time Tracker"))
	b.WriteString("\n\n")

	timeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))
	b.WriteString(timeStyle.Render(formatDuration(m.elapsed)))
	b.WriteString("\n\n")

	status := "[Running]"
	if !m.running {
		status = "[Paused]"
	}
	b.WriteString(normalStyle.Render(status))
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("\nspace: pause/resume • esc: stop & back"))

	return b.String()
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
