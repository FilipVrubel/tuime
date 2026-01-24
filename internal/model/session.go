package model

import "time"

type SessionType string

const (
	PomodoroSession SessionType = "pomodoro"
	TrackerSession SessionType = "tracker"
)

type Session struct {
	ID         int
	ActivityID *int
	StartedAt  time.Time
	EndedAt    time.Time
	Duration   int // seconds
	Type       SessionType
}