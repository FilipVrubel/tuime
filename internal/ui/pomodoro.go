package ui

import (
	"fmt"
	"strings"
	"time"

	"tuime/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startPomodoro() (tea.Model, tea.Cmd) {
	m.CurrentView = PomodoroView
	m.pomodoroSelectingActivity = true
	m.pomodoroActivityCursor = 0
	m.selectedActivityForSession = nil
	return m, loadActivitiesCmd(m.DB)
}

func (m Model) startPomodoroWithActivity() (tea.Model, tea.Cmd) {
	m.pomodoroSelectingActivity = false
	m.WorkDuration = 5 * time.Minute
	m.ShortBreakDuration = 5 * time.Minute
	m.LongBreakDuration = 15 * time.Minute
	m.CyclesBeforeLongBreak = 4
	m.CyclesDone = 0
	m.Phase = WorkPhase
	m.Remaining = m.WorkDuration
	m.Running = false
	m.sessionStartedAt = time.Now()
	return m, nil
}

func (m *Model) saveWorkSession() tea.Cmd {
	if m.Phase != WorkPhase {
		return nil
	}

	activeWorkTime := m.WorkDuration - m.Remaining
	if activeWorkTime <= 0 {
		return nil
	}

	endedAt := m.sessionStartedAt.Add(activeWorkTime)
	duration := int(activeWorkTime.Seconds())
	session := &model.Session{
		ActivityID: m.selectedActivityForSession,
		StartedAt:  m.sessionStartedAt,
		EndedAt:    endedAt,
		Duration:   duration,
		Type:       model.PomodoroSession,
	}
	return saveSessionCmd(m.DB, session)
}

func (m Model) updatePomodoro(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case activitiesLoadedMsg:
		m.activities = msg.activities
		if m.pomodoroActivityCursor >= len(m.activities)+1 {
			m.pomodoroActivityCursor = 0
		}
		return m, nil

	case sessionSavedMsg:
		return m, nil

	case sessionSaveErrorMsg:
		return m, nil

	case tea.KeyMsg:
		if m.pomodoroSelectingActivity {
			switch msg.String() {
			case "up", "k":
				if m.pomodoroActivityCursor > 0 {
					m.pomodoroActivityCursor--
				}
			case "down", "j":
				maxCursor := len(m.activities)
				if m.pomodoroActivityCursor < maxCursor {
					m.pomodoroActivityCursor++
				}
			case "enter":
				if m.pomodoroActivityCursor == 0 {
					m.selectedActivityForSession = nil
				} else {
					activityID := m.activities[m.pomodoroActivityCursor-1].ID
					m.selectedActivityForSession = &activityID
				}
				return m.startPomodoroWithActivity()
			case "esc":
				m.CurrentView = HomeView
				m.pomodoroSelectingActivity = false
			}
			return m, nil
		}

		switch msg.String() {
		case " ":
			if m.Running {
				m.Running = false
			} else {
				m.Running = true
				return m, TickCmd()
			}
		case "right":
			cmd := m.transitionPhase()
			m.Running = false
			return m, cmd
		case "esc":
			cmd := m.saveWorkSession()
			m.Running = false
			m.CurrentView = HomeView
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) transitionPhase() tea.Cmd {
	var cmd tea.Cmd
	if m.Phase == WorkPhase {
		activeWorkTime := m.WorkDuration - m.Remaining
		if activeWorkTime > 0 {
			endedAt := m.sessionStartedAt.Add(activeWorkTime)
			duration := int(activeWorkTime.Seconds())
			session := &model.Session{
				ActivityID: m.selectedActivityForSession,
				StartedAt:  m.sessionStartedAt,
				EndedAt:    endedAt,
				Duration:   duration,
				Type:       model.PomodoroSession,
			}
			cmd = saveSessionCmd(m.DB, session)
		}

		m.CyclesDone++
		m.Phase = BreakPhase
		if m.CyclesDone%m.CyclesBeforeLongBreak == 0 {
			m.Remaining = m.LongBreakDuration
		} else {
			m.Remaining = m.ShortBreakDuration
		}
	} else {
		m.Phase = WorkPhase
		m.Remaining = m.WorkDuration
		m.sessionStartedAt = time.Now()
	}
	return cmd
}

func (m Model) tickPomodoro() (tea.Model, tea.Cmd) {
	m.Remaining -= time.Second
	if m.Remaining <= 0 {
		m.Running = false
		fmt.Print("\a")
		cmd := m.transitionPhase()
		return m, cmd
	}
	return m, TickCmd()
}

func (m Model) viewPomodoro() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Pomodoro Timer"))
	b.WriteString("\n\n")

	if m.pomodoroSelectingActivity {
		b.WriteString(NormalStyle.Render("Select an activity (optional):"))
		b.WriteString("\n\n")

		cursor := "  "
		style := NormalStyle
		if m.pomodoroActivityCursor == 0 {
			cursor = "> "
			style = SelectedStyle
		}
		b.WriteString(cursor)
		b.WriteString(style.Render("None"))
		b.WriteString("\n")

		for i, activity := range m.activities {
			cursor = "  "
			style = NormalStyle
			if m.pomodoroActivityCursor == i+1 {
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
		b.WriteString(TimeStyle.Render(FormatDuration(m.Remaining)))
		b.WriteString("\n\n")

		phaseStr := "Work"
		if m.Phase == BreakPhase {
			phaseStr = "Break"
		}
		b.WriteString(NormalStyle.Render("Phase: " + phaseStr))
		b.WriteString("\n")

		status := "[Running]"
		if !m.Running {
			status = "[Paused]"
		}
		b.WriteString(NormalStyle.Render(status))
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf("Cycles completed: %d\n", m.CyclesDone))

		b.WriteString(HelpStyle.Render("\nspace: pause/resume • esc: stop & back • right: skip phase"))
	}

	return b.String()
}
