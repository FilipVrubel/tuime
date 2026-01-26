package ui

import (
	"fmt"
	"time"

	"tuime/internal/db"
	"tuime/internal/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jmoiron/sqlx"
)

type TickMsg time.Time

func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

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

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
