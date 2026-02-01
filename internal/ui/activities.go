package ui

import (
	"fmt"
	"strings"

	"tuime/internal/db"
	"tuime/internal/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
)

type activitiesLoadedMsg struct {
	activities []model.Activity
}

type activityCreatedMsg struct {
	activity *model.Activity
}

type activityDeletedMsg struct {
	id int
}

type activityErrorMsg struct {
	err error
}

func loadActivitiesCmd(database *sqlx.DB) tea.Cmd {
	return func() tea.Msg {
		activities, err := db.ListActivities(database)
		if err != nil {
			return activityErrorMsg{err: err}
		}
		return activitiesLoadedMsg{activities: activities}
	}
}

func createActivityCmd(database *sqlx.DB, name string) tea.Cmd {
	return func() tea.Msg {
		activity, err := db.CreateActivity(database, name)
		if err != nil {
			return activityErrorMsg{err: err}
		}
		return activityCreatedMsg{activity: activity}
	}
}

func deleteActivityCmd(database *sqlx.DB, id int) tea.Cmd {
	return func() tea.Msg {
		err := db.DeleteActivity(database, id)
		if err != nil {
			return activityErrorMsg{err: err}
		}
		return activityDeletedMsg{id: id}
	}
}

func (m Model) startActivities() (Model, tea.Cmd) {
	m.CurrentView = ActivitiesView
	m.inputMode = false
	m.inputValue = ""
	m.errorMsg = ""
	m.selectedActivity = 0
	return m, loadActivitiesCmd(m.DB)
}

func (m Model) updateActivities(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errorMsg = ""

		if m.inputMode {
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.inputValue) != "" {
					m.inputMode = false
					name := strings.TrimSpace(m.inputValue)
					m.inputValue = ""
					return m, createActivityCmd(m.DB, name)
				}
			case "esc":
				m.inputMode = false
				m.inputValue = ""
			case "backspace":
				if len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:len(m.inputValue)-1]
				}
			default:
				if len(msg.String()) == 1 && len(m.inputValue) < 50 {
					m.inputValue += msg.String()
				}
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if m.selectedActivity > 0 {
					m.selectedActivity--
				}
			case "down", "j":
				if m.selectedActivity < len(m.activities)-1 {
					m.selectedActivity++
				}
			case "n":
				m.inputMode = true
				m.inputValue = ""
			case "d":
				if len(m.activities) > 0 && m.selectedActivity < len(m.activities) {
					id := m.activities[m.selectedActivity].ID
					return m, deleteActivityCmd(m.DB, id)
				}
			case "esc":
				m.CurrentView = HomeView
				return m, nil
			}
		}

	case activitiesLoadedMsg:
		m.activities = msg.activities
		if m.selectedActivity >= len(m.activities) {
			m.selectedActivity = len(m.activities) - 1
		}
		if m.selectedActivity < 0 {
			m.selectedActivity = 0
		}

	case activityCreatedMsg:
		return m, loadActivitiesCmd(m.DB)

	case activityDeletedMsg:
		return m, loadActivitiesCmd(m.DB)

	case activityErrorMsg:
		m.errorMsg = msg.err.Error()
	}

	return m, nil
}

func (m Model) viewActivities() string {
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

	b.WriteString(centerStyle.Render(TitleStyle.Render("Activities")))
	b.WriteString("\n\n")

	if m.inputMode {
		b.WriteString(centerStyle.Render("New activity: " + m.inputValue + "█"))
		b.WriteString("\n\n")
		b.WriteString(centerStyle.Render(HelpStyle.Render("enter: save • esc: cancel")))
	} else {
		if len(m.activities) == 0 {
			b.WriteString(centerStyle.Render(NormalStyle.Render("No activities yet. Press 'n' to create one.")))
			b.WriteString("\n\n")
		} else {
			for i, activity := range m.activities {
				cursor := "  "
				style := NormalStyle
				if i == m.selectedActivity {
					cursor = "> "
					style = SelectedStyle
				}
				b.WriteString(centerStyle.Render(cursor + style.Render(activity.Name)))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
		b.WriteString(centerStyle.Render(HelpStyle.Render("n: new • d: delete • ↑/↓: navigate • esc: back")))
	}

	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		b.WriteString(centerStyle.Render(errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))))
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}
