package model

import "time"

type SessionType string

const (
	PomodoroSession SessionType = "pomodoro"
	TrackerSession  SessionType = "tracker"
)

type Session struct {
	ID         int         `db:"id"`
	ActivityID *int        `db:"activity_id"`
	StartedAt  time.Time   `db:"started_at"`
	EndedAt    time.Time   `db:"ended_at"`
	Duration   int         `db:"duration"` // seconds
	Type       SessionType `db:"type"`
}
