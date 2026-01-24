package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	TimeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))
)
