package main

import (
	"fmt"
	"os"
	"path/filepath"

	"tuime/internal/config"
	"tuime/internal/db"
	"tuime/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dbDir := filepath.Join(home, ".local", "share", "tuime")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		panic(err)
	}
	dbPath := filepath.Join(dbDir, "tuime.db")

	database, err := db.Open(dbPath)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		panic(err)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v. Using defaults.\n", err)
		cfg = config.DefaultConfig()
	}

	p := tea.NewProgram(ui.NewModel(database, cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
