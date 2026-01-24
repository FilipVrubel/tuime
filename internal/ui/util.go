package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg time.Time

func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
