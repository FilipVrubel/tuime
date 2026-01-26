package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jmoiron/sqlx"
	"tuime/internal/db"
	"tuime/internal/model"
)

type sessionSavedMsg struct {
	session *model.Session
}

type sessionSaveErrorMsg struct {
	err error
}

func saveSessionCmd(database *sqlx.DB, session *model.Session) tea.Cmd {
	return func() tea.Msg {
		savedSession, err := db.CreateSession(database, session)
		if err != nil {
			return sessionSaveErrorMsg{err: err}
		}
		return sessionSavedMsg{session: savedSession}
	}
}

func (m Model) startTracker() (tea.Model, tea.Cmd) {
	m.CurrentView = TimeTrackerView
	m.trackerSelectingActivity = true
	m.trackerActivityCursor = 0
	m.selectedActivityForSession = nil
	return m, loadActivitiesCmd(m.DB)
}

func (m Model) startTrackerWithActivity() (tea.Model, tea.Cmd) {
	m.trackerSelectingActivity = false
	m.Running = true
	m.StartTime = time.Now()
	m.sessionStartedAt = time.Now()
	m.Elapsed = 0
	return m, TickCmd()
}

func (m Model) updateTracker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case activitiesLoadedMsg:
		m.activities = msg.activities
		if m.trackerActivityCursor >= len(m.activities)+1 {
			m.trackerActivityCursor = 0
		}
		return m, nil

	case sessionSavedMsg:
		return m, nil

	case sessionSaveErrorMsg:
		m.errorMsg = msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		if m.trackerSelectingActivity {
			switch msg.String() {
			case "up", "k":
				if m.trackerActivityCursor > 0 {
					m.trackerActivityCursor--
				}
			case "down", "j":
				maxCursor := len(m.activities)
				if m.trackerActivityCursor < maxCursor {
					m.trackerActivityCursor++
				}
			case "enter":
				if m.trackerActivityCursor == 0 {
					m.selectedActivityForSession = nil
				} else {
					activityID := m.activities[m.trackerActivityCursor-1].ID
					m.selectedActivityForSession = &activityID
				}
				return m.startTrackerWithActivity()
			case "esc":
				m.CurrentView = HomeView
				m.trackerSelectingActivity = false
			}
			return m, nil
		}

		switch msg.String() {
		case " ":
			if m.Running {
				m.Running = false
			} else {
				m.StartTime = time.Now().Add(-m.Elapsed)
				m.Running = true
				return m, TickCmd()
			}
		case "enter":
			if m.Running || m.Elapsed > 0 {
				endTime := time.Now()
				durationSeconds := int(m.Elapsed.Seconds())
				
				session := &model.Session{
					ActivityID: m.selectedActivityForSession,
					StartedAt:  m.sessionStartedAt,
					EndedAt:    endTime,
					Duration:   durationSeconds,
					Type:       model.TrackerSession,
				}
				
				m.Running = false
				m.CurrentView = HomeView
				m.Elapsed = 0
				return m, saveSessionCmd(m.DB, session)
			}
			m.CurrentView = HomeView
		case "esc":
			m.Running = false
			m.CurrentView = HomeView
			m.Elapsed = 0
		}
	}
	return m, nil
}

func (m Model) viewTracker() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Time Tracker"))
	b.WriteString("\n\n")

	if m.trackerSelectingActivity {
		b.WriteString(NormalStyle.Render("Select an activity (optional):"))
		b.WriteString("\n\n")

		cursor := "  "
		style := NormalStyle
		if m.trackerActivityCursor == 0 {
			cursor = "> "
			style = SelectedStyle
		}
		b.WriteString(cursor)
		b.WriteString(style.Render("None"))
		b.WriteString("\n")

		for i, activity := range m.activities {
			cursor = "  "
			style = NormalStyle
			if m.trackerActivityCursor == i+1 {
				cursor = "> "
				style = SelectedStyle
			}
			b.WriteString(cursor)
			b.WriteString(style.Render(activity.Name))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("↑/↓: navigate • enter: select • esc: back"))
	} else {
		b.WriteString(TimeStyle.Render(FormatDuration(m.Elapsed)))
		b.WriteString("\n\n")

		status := "[Running]"
		if !m.Running {
			status = "[Paused]"
		}
		b.WriteString(NormalStyle.Render(status))
		b.WriteString("\n")

		b.WriteString(HelpStyle.Render("\nspace: pause/resume • enter: save & back • esc: cancel"))
	}

	return b.String()
}
