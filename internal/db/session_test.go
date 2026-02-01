package db

import (
	"testing"
	"time"

	"tuime/internal/model"
)

func TestCreateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity, err := CreateActivity(db, "Test Activity")
	if err != nil {
		t.Fatalf("Failed to create activity: %v", err)
	}

	now := time.Now()
	tests := []struct {
		name    string
		session *model.Session
		wantErr bool
	}{
		{
			name: "pomodoro session with activity",
			session: &model.Session{
				ActivityID: &activity.ID,
				StartedAt:  now,
				EndedAt:    now.Add(25 * time.Minute),
				Duration:   1500,
				Type:       model.PomodoroSession,
			},
			wantErr: false,
		},
		{
			name: "tracker session without activity",
			session: &model.Session{
				ActivityID: nil,
				StartedAt:  now,
				EndedAt:    now.Add(2 * time.Hour),
				Duration:   7200,
				Type:       model.TrackerSession,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := CreateSession(db, tt.session)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.wantErr {
				if created.ID <= 0 {
					t.Error("Expected positive ID")
				}
				if created.Type != tt.session.Type {
					t.Errorf("Expected type %s, got %s", tt.session.Type, created.Type)
				}
			}
		})
	}
}

func TestListSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity1, _ := CreateActivity(db, "Activity 1")
	activity2, _ := CreateActivity(db, "Activity 2")

	now := time.Now()
	pomodoroType := model.PomodoroSession
	trackerType := model.TrackerSession

	sessions := []*model.Session{
		{
			ActivityID: &activity1.ID,
			StartedAt:  now.AddDate(0, 0, -2),
			EndedAt:    now.AddDate(0, 0, -2).Add(25 * time.Minute),
			Duration:   1500,
			Type:       pomodoroType,
		},
		{
			ActivityID: &activity2.ID,
			StartedAt:  now.AddDate(0, 0, -1),
			EndedAt:    now.AddDate(0, 0, -1).Add(1 * time.Hour),
			Duration:   3600,
			Type:       trackerType,
		},
		{
			ActivityID: &activity1.ID,
			StartedAt:  now,
			EndedAt:    now.Add(25 * time.Minute),
			Duration:   1500,
			Type:       pomodoroType,
		},
	}

	for _, s := range sessions {
		_, err := CreateSession(db, s)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	tests := []struct {
		name          string
		filter        SessionFilter
		expectedCount int
	}{
		{
			name:          "no filter",
			filter:        SessionFilter{},
			expectedCount: 3,
		},
		{
			name: "filter by activity 1",
			filter: SessionFilter{
				ActivityID: &activity1.ID,
			},
			expectedCount: 2,
		},
		{
			name: "filter by pomodoro type",
			filter: SessionFilter{
				Type: &pomodoroType,
			},
			expectedCount: 2,
		},
		{
			name: "filter by tracker type",
			filter: SessionFilter{
				Type: &trackerType,
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ListSessions(db, tt.filter)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d sessions, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	session := &model.Session{
		ActivityID: nil,
		StartedAt:  now,
		EndedAt:    now.Add(25 * time.Minute),
		Duration:   1500,
		Type:       model.PomodoroSession,
	}

	created, err := CreateSession(db, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	retrieved, err := GetSession(db, created.ID)
	if err != nil {
		t.Errorf("Failed to get session: %v", err)
		return
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}

	if retrieved.Duration != created.Duration {
		t.Errorf("Expected duration %d, got %d", created.Duration, retrieved.Duration)
	}

	_, err = GetSession(db, 999)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestUpdateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	session := &model.Session{
		ActivityID: nil,
		StartedAt:  now,
		EndedAt:    now.Add(25 * time.Minute),
		Duration:   1500,
		Type:       model.PomodoroSession,
	}

	created, err := CreateSession(db, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	newEndTime := now.Add(30 * time.Minute)
	newDuration := 1800

	err = UpdateSession(db, created.ID, newEndTime, newDuration)
	if err != nil {
		t.Errorf("Failed to update session: %v", err)
	}

	updated, err := GetSession(db, created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if updated.Duration != newDuration {
		t.Errorf("Expected duration %d, got %d", newDuration, updated.Duration)
	}

	err = UpdateSession(db, 999, newEndTime, newDuration)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestDeleteSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	session := &model.Session{
		ActivityID: nil,
		StartedAt:  now,
		EndedAt:    now.Add(25 * time.Minute),
		Duration:   1500,
		Type:       model.PomodoroSession,
	}

	created, err := CreateSession(db, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	err = DeleteSession(db, created.ID)
	if err != nil {
		t.Errorf("Failed to delete session: %v", err)
	}

	_, err = GetSession(db, created.ID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound after deletion, got %v", err)
	}

	err = DeleteSession(db, 999)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound for non-existing session, got %v", err)
	}
}

func TestGetRecentSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	for i := 0; i < 5; i++ {
		session := &model.Session{
			ActivityID: nil,
			StartedAt:  now.Add(time.Duration(i) * time.Hour),
			EndedAt:    now.Add(time.Duration(i)*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		}
		_, err := CreateSession(db, session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	recent, err := GetRecentSessions(db, 3)
	if err != nil {
		t.Errorf("Failed to get recent sessions: %v", err)
		return
	}

	if len(recent) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(recent))
	}

	for i := 0; i < len(recent)-1; i++ {
		if recent[i].StartedAt.Before(recent[i+1].StartedAt) {
			t.Error("Sessions not ordered by most recent first")
		}
	}
}

func TestGetSessionsByDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -5),
		now.AddDate(0, 0, -3),
		now.AddDate(0, 0, -1),
		now,
	}

	for _, date := range dates {
		session := &model.Session{
			ActivityID: nil,
			StartedAt:  date,
			EndedAt:    date.Add(25 * time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		}
		_, err := CreateSession(db, session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	startDate := now.AddDate(0, 0, -4)
	endDate := now.AddDate(0, 0, 1)

	sessions, err := GetSessionsByDateRange(db, startDate, endDate)
	if err != nil {
		t.Errorf("Failed to get sessions by date range: %v", err)
		return
	}

	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions in range, got %d", len(sessions))
	}

	for _, s := range sessions {
		if s.StartedAt.Before(startDate) || s.StartedAt.After(endDate) {
			t.Errorf("Session %d outside date range", s.ID)
		}
	}
}

func TestGetActivityStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity1, _ := CreateActivity(db, "Activity 1")
	activity2, _ := CreateActivity(db, "Activity 2")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	sessions := []*model.Session{
		{
			ActivityID: &activity1.ID,
			StartedAt:  today.Add(9 * time.Hour),
			EndedAt:    today.Add(9*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity1.ID,
			StartedAt:  today.Add(10 * time.Hour),
			EndedAt:    today.Add(10*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity2.ID,
			StartedAt:  today.Add(14 * time.Hour),
			EndedAt:    today.Add(14*time.Hour + 1*time.Hour),
			Duration:   3600,
			Type:       model.TrackerSession,
		},
	}

	for _, s := range sessions {
		_, err := CreateSession(db, s)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	startDate := today
	endDate := today.AddDate(0, 0, 1)

	stats, err := GetActivityStats(db, &startDate, &endDate)
	if err != nil {
		t.Fatalf("Failed to get activity stats: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected 2 activities, got %d", len(stats))
	}

	var activity1Stats *ActivityStats
	for i := range stats {
		if stats[i].ActivityID != nil && *stats[i].ActivityID == activity1.ID {
			activity1Stats = &stats[i]
			break
		}
	}

	if activity1Stats == nil {
		t.Fatal("Activity 1 stats not found")
	}

	if activity1Stats.TotalTime != 3000 {
		t.Errorf("Expected total time 3000, got %d", activity1Stats.TotalTime)
	}

	if activity1Stats.SessionCount != 2 {
		t.Errorf("Expected 2 sessions, got %d", activity1Stats.SessionCount)
	}
}

func TestGetTypeStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity, _ := CreateActivity(db, "Test Activity")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	sessions := []*model.Session{
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(9 * time.Hour),
			EndedAt:    today.Add(9*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(10 * time.Hour),
			EndedAt:    today.Add(10*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(14 * time.Hour),
			EndedAt:    today.Add(14*time.Hour + 1*time.Hour),
			Duration:   3600,
			Type:       model.TrackerSession,
		},
	}

	for _, s := range sessions {
		_, err := CreateSession(db, s)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	startDate := today
	endDate := today.AddDate(0, 0, 1)

	stats, err := GetTypeStats(db, &startDate, &endDate)
	if err != nil {
		t.Fatalf("Failed to get type stats: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected 2 session types, got %d", len(stats))
	}

	var pomodoroStats, trackerStats *TypeStats
	for i := range stats {
		if stats[i].Type == model.PomodoroSession {
			pomodoroStats = &stats[i]
		} else if stats[i].Type == model.TrackerSession {
			trackerStats = &stats[i]
		}
	}

	if pomodoroStats == nil {
		t.Fatal("Pomodoro stats not found")
	}

	if trackerStats == nil {
		t.Fatal("Tracker stats not found")
	}

	if pomodoroStats.TotalTime != 3000 {
		t.Errorf("Expected pomodoro total time 3000, got %d", pomodoroStats.TotalTime)
	}

	if trackerStats.TotalTime != 3600 {
		t.Errorf("Expected tracker total time 3600, got %d", trackerStats.TotalTime)
	}
}

func TestGetOverallStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity, _ := CreateActivity(db, "Test Activity")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	sessions := []*model.Session{
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(9 * time.Hour),
			EndedAt:    today.Add(9*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(10 * time.Hour),
			EndedAt:    today.Add(10*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(14 * time.Hour),
			EndedAt:    today.Add(14*time.Hour + 1*time.Hour),
			Duration:   3600,
			Type:       model.TrackerSession,
		},
	}

	for _, s := range sessions {
		_, err := CreateSession(db, s)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	startDate := today
	endDate := today.AddDate(0, 0, 1)

	stats, err := GetOverallStats(db, &startDate, &endDate)
	if err != nil {
		t.Fatalf("Failed to get overall stats: %v", err)
	}

	expectedTotalTime := 6600
	if stats.TotalTime != expectedTotalTime {
		t.Errorf("Expected total time %d, got %d", expectedTotalTime, stats.TotalTime)
	}

	expectedTotalSessions := 3
	if stats.TotalSessions != expectedTotalSessions {
		t.Errorf("Expected %d sessions, got %d", expectedTotalSessions, stats.TotalSessions)
	}
}

func TestGetDailySessionStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activity, _ := CreateActivity(db, "Test Activity")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	sessions := []*model.Session{
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(9 * time.Hour),
			EndedAt:    today.Add(9*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(10 * time.Hour),
			EndedAt:    today.Add(10*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.Add(14 * time.Hour),
			EndedAt:    today.Add(14*time.Hour + 1*time.Hour),
			Duration:   3600,
			Type:       model.TrackerSession,
		},
		{
			ActivityID: &activity.ID,
			StartedAt:  today.AddDate(0, 0, -1).Add(10 * time.Hour),
			EndedAt:    today.AddDate(0, 0, -1).Add(10*time.Hour + 25*time.Minute),
			Duration:   1500,
			Type:       model.PomodoroSession,
		},
	}

	for _, s := range sessions {
		_, err := CreateSession(db, s)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	startDate := today.AddDate(0, 0, -1)
	endDate := today.AddDate(0, 0, 1)

	stats, err := GetDailySessionStats(db, &startDate, &endDate)
	if err != nil {
		t.Fatalf("Failed to get daily stats: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected 2 days of stats, got %d", len(stats))
	}

	var todayStats *DailySessionStats
	for i := range stats {
		if stats[i].Date == today.Format("2006-01-02") {
			todayStats = &stats[i]
			break
		}
	}

	if todayStats == nil {
		t.Fatal("Today's stats not found")
	}

	if todayStats.PomodoroTime != 3000 {
		t.Errorf("Expected pomodoro time 3000, got %d", todayStats.PomodoroTime)
	}

	if todayStats.TrackerTime != 3600 {
		t.Errorf("Expected tracker time 3600, got %d", todayStats.TrackerTime)
	}

	if todayStats.PomodoroCount != 2 {
		t.Errorf("Expected 2 pomodoro sessions, got %d", todayStats.PomodoroCount)
	}

	if todayStats.TrackerCount != 1 {
		t.Errorf("Expected 1 tracker session, got %d", todayStats.TrackerCount)
	}
}
