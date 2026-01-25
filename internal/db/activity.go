package db

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"tuime/internal/model"

	"github.com/jmoiron/sqlx"
)

var (
    ErrActivityExists   = errors.New("activity already exists")
    ErrActivityNotFound = errors.New("activity not found")
)

func CreateActivity(db *sqlx.DB, name string) (*model.Activity, error) {
    now := time.Now()
    
    result, err := db.Exec(
        "INSERT INTO activities (name, created_at) VALUES (?, ?)",
        name,
        now,
    )
    if err != nil {
        if strings.Contains(err.Error(), "UNIQUE constraint failed") {
            return nil, ErrActivityExists
        }
        return nil, err
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }

    return &model.Activity{
        ID:        int(id),
        Name:      name,
        CreatedAt: now,
    }, nil
}

func ListActivities(db *sqlx.DB) ([]model.Activity, error) {
    var activities []model.Activity
    
    err := db.Select(&activities, "SELECT id, name, created_at FROM activities ORDER BY name")
    if err != nil {
        return nil, err
    }
    
    return activities, nil
}

func GetActivity(db *sqlx.DB, id int) (*model.Activity, error) {
    var activity model.Activity
    
    err := db.Get(&activity, "SELECT id, name, created_at FROM activities WHERE id = ?", id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrActivityNotFound
        }
        return nil, err
    }
    
    return &activity, nil
}

func DeleteActivity(db *sqlx.DB, id int) error {
    result, err := db.Exec("DELETE FROM activities WHERE id = ?", id)
    if err != nil {
        return err
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rows == 0 {
        return ErrActivityNotFound
    }
    
    return nil
}