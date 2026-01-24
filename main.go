package main

import (
	"fmt"

	"tuime/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
