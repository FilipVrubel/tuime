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
		query += " AND started_at <= ?"
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
		WHERE started_at >= ? AND started_at <= ?
		ORDER BY started_at DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}
