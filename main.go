package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"tuime/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

