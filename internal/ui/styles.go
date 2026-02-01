package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7FE5D0")).
			Padding(1, 2).
			Align(lipgloss.Center)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#7FE5D0")).
			Bold(true).
			Padding(0, 1)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A8A8A")).
			MarginTop(2).
			Align(lipgloss.Center)

	TimeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7FE5D0")).
			Padding(1, 3)

	PomodoroWorkStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF9999")).
				Padding(1, 3).
				Align(lipgloss.Center)

	PomodoroBreakStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#99CCFF")).
				Padding(1, 3).
				Align(lipgloss.Center)

	PomodoroWorkTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF9999")).
				MarginBottom(2).
				Padding(1, 2).
				Align(lipgloss.Center)

	PomodoroBreakTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#99CCFF")).
				MarginBottom(1).
				Padding(1, 2).
				Align(lipgloss.Center)
)
