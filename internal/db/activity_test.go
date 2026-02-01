package db

import (
	"testing"
	"time"

	"tuime/internal/model"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestCreateActivity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		actName   string
		wantErr   error
		errString string
	}{
		{
			name:    "valid activity",
			actName: "Study Go",
			wantErr: nil,
		},
		{
			name:    "another valid activity",
			actName: "Read Books",
			wantErr: nil,
		},
		{
			name:      "duplicate activity",
			actName:   "Study Go",
			wantErr:   ErrActivityExists,
			errString: "activity already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity, err := CreateActivity(db, tt.actName)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if activity == nil {
				t.Error("Expected activity, got nil")
				return
			}

			if activity.Name != tt.actName {
				t.Errorf("Expected name %s, got %s", tt.actName, activity.Name)
			}

			if activity.ID <= 0 {
				t.Errorf("Expected positive ID, got %d", activity.ID)
			}

			if activity.CreatedAt.IsZero() {
				t.Error("Expected non-zero created time")
			}
		})
	}
}

func TestListActivities(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activities, err := ListActivities(db)
	if err != nil {
		t.Fatalf("Failed to list empty activities: %v", err)
	}
	if len(activities) != 0 {
		t.Errorf("Expected 0 activities, got %d", len(activities))
	}

	names := []string{"Coding", "Reading", "Exercise"}
	for _, name := range names {
		_, err := CreateActivity(db, name)
		if err != nil {
			t.Fatalf("Failed to create activity %s: %v", name, err)
		}
	}

	activities, err = ListActivities(db)
	if err != nil {
		t.Fatalf("Failed to list activities: %v", err)
	}

	if len(activities) != len(names) {
		t.Errorf("Expected %d activities, got %d", len(names), len(activities))
	}

	expectedOrder := []string{"Coding", "Exercise", "Reading"}
	for i, activity := range activities {
		if activity.Name != expectedOrder[i] {
			t.Errorf("Expected activity %d to be %s, got %s", i, expectedOrder[i], activity.Name)
		}
	}
}

func TestGetActivity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	created, err := CreateActivity(db, "Test Activity")
	if err != nil {
		t.Fatalf("Failed to create activity: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{
			name:    "existing activity",
			id:      created.ID,
			wantErr: nil,
		},
		{
			name:    "non-existing activity",
			id:      999,
			wantErr: ErrActivityNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity, err := GetActivity(db, tt.id)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if activity.ID != tt.id {
				t.Errorf("Expected ID %d, got %d", tt.id, activity.ID)
			}
		})
	}
}

func TestDeleteActivity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	created, err := CreateActivity(db, "To Delete")
	if err != nil {
		t.Fatalf("Failed to create activity: %v", err)
	}

	err = DeleteActivity(db, created.ID)
	if err != nil {
		t.Errorf("Failed to delete activity: %v", err)
	}

	_, err = GetActivity(db, created.ID)
	if err != ErrActivityNotFound {
		t.Errorf("Expected ErrActivityNotFound after deletion, got %v", err)
	}

	err = DeleteActivity(db, 999)
	if err != ErrActivityNotFound {
		t.Errorf("Expected ErrActivityNotFound for non-existing activity, got %v", err)
	}
}

func TestDeleteActivityWithSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity, err := CreateActivity(db, "Activity with Sessions")
	if err != nil {
		t.Fatalf("Failed to create activity: %v", err)
	}

	now := time.Now()
	session := &model.Session{
		ActivityID: &activity.ID,
		StartedAt:  now,
		EndedAt:    now.Add(25 * time.Minute),
		Duration:   1500,
		Type:       model.PomodoroSession,
	}

	createdSession, err := CreateSession(db, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	err = DeleteActivity(db, activity.ID)
	if err != nil {
		t.Errorf("Failed to delete activity: %v", err)
	}

	retrievedSession, err := GetSession(db, createdSession.ID)
	if err != nil {
		t.Fatalf("Failed to get session after activity deletion: %v", err)
	}

	if retrievedSession.ActivityID != nil {
		t.Errorf("Expected session activity_id to be NULL after activity deletion, got %d", *retrievedSession.ActivityID)
	}
}
