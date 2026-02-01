package db

import (
	"database/sql"
	"errors"
	"time"

	"tuime/internal/model"

	"github.com/jmoiron/sqlx"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionFilter struct {
	ActivityID *int
	StartDate  *time.Time
	EndDate    *time.Time
	Type       *model.SessionType
}

func CreateSession(db *sqlx.DB, session *model.Session) (*model.Session, error) {
	result, err := db.Exec(
		"INSERT INTO sessions (activity_id, started_at, ended_at, duration, type) VALUES (?, ?, ?, ?, ?)",
		session.ActivityID,
		session.StartedAt,
		session.EndedAt,
		session.Duration,
		session.Type,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	session.ID = int(id)
	return session, nil
}

func ListSessions(db *sqlx.DB, filter SessionFilter) ([]model.Session, error) {
	query := "SELECT id, activity_id, started_at, ended_at, duration, type FROM sessions WHERE 1=1"
	args := []interface{}{}

	if filter.ActivityID != nil {
		query += " AND activity_id = ?"
		args = append(args, *filter.ActivityID)
	}

	if filter.StartDate != nil {
		query += " AND started_at >= ?"
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		query += " AND started_at < ?"
		args = append(args, *filter.EndDate)
	}

	if filter.Type != nil {
		query += " AND type = ?"
		args = append(args, *filter.Type)
	}

	query += " ORDER BY started_at DESC"

	var sessions []model.Session
	err := db.Select(&sessions, query, args...)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func GetSession(db *sqlx.DB, id int) (*model.Session, error) {
	var session model.Session

	err := db.Get(&session, "SELECT id, activity_id, started_at, ended_at, duration, type FROM sessions WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

func UpdateSession(db *sqlx.DB, id int, endedAt time.Time, duration int) error {
	result, err := db.Exec("UPDATE sessions SET ended_at = ?, duration = ? WHERE id = ?", endedAt, duration, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func DeleteSession(db *sqlx.DB, id int) error {
	result, err := db.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func GetRecentSessions(db *sqlx.DB, limit int) ([]model.Session, error) {
	var sessions []model.Session
	err := db.Select(&sessions, `
		SELECT id, activity_id, started_at, ended_at, duration, type 
		FROM sessions 
		ORDER BY started_at DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func GetSessionsByDateRange(db *sqlx.DB, startDate, endDate time.Time) ([]model.Session, error) {
	var sessions []model.Session
	err := db.Select(&sessions, `
		SELECT id, activity_id, started_at, ended_at, duration, type 
		FROM sessions 
		WHERE started_at >= ? AND started_at < ?
		ORDER BY started_at DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

type ActivityStats struct {
	ActivityID   *int   `db:"activity_id"`
	ActivityName string `db:"activity_name"`
	TotalTime    int    `db:"total_time"`
	SessionCount int    `db:"session_count"`
}

func GetActivityStats(db *sqlx.DB, startDate, endDate *time.Time) ([]ActivityStats, error) {
	query := `
		SELECT 
			s.activity_id,
			COALESCE(a.name, 'No Activity') as activity_name,
			SUM(s.duration) as total_time,
			COUNT(s.id) as session_count
		FROM sessions s
		LEFT JOIN activities a ON s.activity_id = a.id
	`
	args := []interface{}{}

	if startDate != nil && endDate != nil {
		startStr := startDate.Format("2006-01-02 15:04:05")
		endStr := endDate.Format("2006-01-02 15:04:05")
		query += " WHERE SUBSTR(s.started_at, 1, 19) >= ? AND SUBSTR(s.started_at, 1, 19) < ?"
		args = append(args, startStr, endStr)
	}

	query += " GROUP BY s.activity_id, activity_name ORDER BY total_time DESC"

	var stats []ActivityStats
	err := db.Select(&stats, query, args...)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type ActivityTypeStats struct {
	ActivityID    *int   `db:"activity_id"`
	ActivityName  string `db:"activity_name"`
	PomodoroTime  int    `db:"pomodoro_time"`
	PomodoroCount int    `db:"pomodoro_count"`
	TrackerTime   int    `db:"tracker_time"`
	TrackerCount  int    `db:"tracker_count"`
	TotalTime     int    `db:"total_time"`
	TotalCount    int    `db:"total_count"`
}

func GetActivityTypeStats(db *sqlx.DB, startDate, endDate *time.Time) ([]ActivityTypeStats, error) {
	query := `
		SELECT 
			s.activity_id,
			COALESCE(a.name, 'No Activity') as activity_name,
			SUM(CASE WHEN s.type = 'pomodoro' THEN s.duration ELSE 0 END) as pomodoro_time,
			SUM(CASE WHEN s.type = 'pomodoro' THEN 1 ELSE 0 END) as pomodoro_count,
			SUM(CASE WHEN s.type = 'tracker' THEN s.duration ELSE 0 END) as tracker_time,
			SUM(CASE WHEN s.type = 'tracker' THEN 1 ELSE 0 END) as tracker_count,
			SUM(s.duration) as total_time,
			COUNT(s.id) as total_count
		FROM sessions s
		LEFT JOIN activities a ON s.activity_id = a.id
	`
	args := []interface{}{}

	if startDate != nil && endDate != nil {
		startStr := startDate.Format("2006-01-02 15:04:05")
		endStr := endDate.Format("2006-01-02 15:04:05")
		query += " WHERE SUBSTR(s.started_at, 1, 19) >= ? AND SUBSTR(s.started_at, 1, 19) < ?"
		args = append(args, startStr, endStr)
	}

	query += " GROUP BY s.activity_id, activity_name ORDER BY total_time DESC"

	var stats []ActivityTypeStats
	err := db.Select(&stats, query, args...)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type TypeStats struct {
	Type         model.SessionType `db:"type"`
	TotalTime    int               `db:"total_time"`
	SessionCount int               `db:"session_count"`
}

func GetTypeStats(db *sqlx.DB, startDate, endDate *time.Time) ([]TypeStats, error) {
	query := `
		SELECT 
			type,
			SUM(duration) as total_time,
			COUNT(id) as session_count
		FROM sessions
	`
	args := []interface{}{}

	if startDate != nil && endDate != nil {
		startStr := startDate.Format("2006-01-02 15:04:05")
		endStr := endDate.Format("2006-01-02 15:04:05")
		query += " WHERE SUBSTR(started_at, 1, 19) >= ? AND SUBSTR(started_at, 1, 19) < ?"
		args = append(args, startStr, endStr)
	}

	query += " GROUP BY type ORDER BY type"

	var stats []TypeStats
	err := db.Select(&stats, query, args...)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type OverallStats struct {
	TotalSessions int `db:"total_sessions"`
	TotalTime     int `db:"total_time"`
}

func GetOverallStats(db *sqlx.DB, startDate, endDate *time.Time) (*OverallStats, error) {
	query := `
		SELECT 
			COUNT(id) as total_sessions,
			COALESCE(SUM(duration), 0) as total_time
		FROM sessions
	`
	args := []interface{}{}

	if startDate != nil && endDate != nil {
		startStr := startDate.Format("2006-01-02 15:04:05")
		endStr := endDate.Format("2006-01-02 15:04:05")
		query += " WHERE SUBSTR(started_at, 1, 19) >= ? AND SUBSTR(started_at, 1, 19) < ?"
		args = append(args, startStr, endStr)
	}

	var stats OverallStats
	err := db.Get(&stats, query, args...)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

type DailySessionStats struct {
	Date          string `db:"date"`
	PomodoroCount int    `db:"pomodoro_count"`
	PomodoroTime  int    `db:"pomodoro_time"`
	TrackerCount  int    `db:"tracker_count"`
	TrackerTime   int    `db:"tracker_time"`
}

func GetDailySessionStats(db *sqlx.DB, startDate, endDate *time.Time) ([]DailySessionStats, error) {
	query := `
		SELECT 
			SUBSTR(started_at, 1, 10) as date,
			SUM(CASE WHEN type = 'pomodoro' THEN 1 ELSE 0 END) as pomodoro_count,
			SUM(CASE WHEN type = 'pomodoro' THEN duration ELSE 0 END) as pomodoro_time,
			SUM(CASE WHEN type = 'tracker' THEN 1 ELSE 0 END) as tracker_count,
			SUM(CASE WHEN type = 'tracker' THEN duration ELSE 0 END) as tracker_time
		FROM sessions
	`
	args := []interface{}{}

	if startDate != nil && endDate != nil {
		startStr := startDate.Format("2006-01-02 15:04:05")
		endStr := endDate.Format("2006-01-02 15:04:05")
		query += " WHERE SUBSTR(started_at, 1, 19) >= ? AND SUBSTR(started_at, 1, 19) < ?"
		args = append(args, startStr, endStr)
	}

	query += " GROUP BY SUBSTR(started_at, 1, 10) ORDER BY date ASC"

	var stats []DailySessionStats
	err := db.Select(&stats, query, args...)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
