package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startPomodoro() (tea.Model, tea.Cmd) {
	m.CurrentView = PomodoroView
	m.WorkDuration = 5 * time.Second       // TODO: change to 25 * time.Minute
	m.ShortBreakDuration = 5 * time.Second // TODO: change to 5 * time.Minute
	m.LongBreakDuration = 15 * time.Minute
	m.CyclesBeforeLongBreak = 4
	m.CyclesDone = 0
	m.Phase = WorkPhase
	m.Remaining = m.WorkDuration
	m.Running = false
	return m, nil
}

func (m Model) updatePomodoro(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case " ":
		if m.Running {
			m.Running = false
		} else {
			m.Running = true
			return m, TickCmd()
		}
	case "right":
		m.transitionPhase()
		m.Running = false
	case "esc":
		m.Running = false
		m.CurrentView = HomeView
	}
	return m, nil
}

func (m *Model) transitionPhase() {
	if m.Phase == WorkPhase {
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
	}
}

func (m Model) tickPomodoro() (tea.Model, tea.Cmd) {
	m.Remaining -= time.Second
	if m.Remaining <= 0 {
		m.Running = false
		fmt.Print("\a")
		m.transitionPhase()
	}
	return m, TickCmd()
}

func (m Model) viewPomodoro() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Pomodoro Timer"))
	b.WriteString("\n\n")

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

	return b.String()
}
