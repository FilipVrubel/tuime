package ui

import (
	"time"

	"tuime/internal/config"
	"tuime/internal/db"
	"tuime/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jmoiron/sqlx"
)

type View int

const (
	HomeView View = iota
	PomodoroView
	TimeTrackerView
	ActivitiesView
	StatisticsView
	ConfigView
)

type Phase int

const (
	WorkPhase Phase = iota
	BreakPhase
)

type MenuItem struct {
	Title string
	Desc  string
}

type Model struct {
	Cursor      int
	Items       []MenuItem
	CurrentView View

	// Tracker state
	Running                  bool
	StartTime                time.Time
	Elapsed                  time.Duration
	trackerSelectingActivity bool
	trackerActivityCursor    int

	// Pomodoro state
	WorkDuration              time.Duration
	ShortBreakDuration        time.Duration
	LongBreakDuration         time.Duration
	CyclesBeforeLongBreak     int
	CyclesDone                int
	Phase                     Phase
	Remaining                 time.Duration
	pomodoroSelectingActivity bool
	pomodoroActivityCursor    int

	// Session tracking state
	sessionStartedAt           time.Time
	selectedActivityForSession *int

	// Database
	DB *sqlx.DB

	// Configuration
	Config         *config.Config
	configCursor   int
	configInputs   []textinput.Model

	// Activities state
	activities       []model.Activity
	selectedActivity int
	inputMode        bool
	inputValue       string
	errorMsg         string

	// Statistics state
	statsFilterIndex int
	statsViewMode    int
	activityStats    []db.ActivityStats
	typeStats        []db.TypeStats
	overallStats     *db.OverallStats
	dailyStats       []db.DailySessionStats
}

func NewModel(db *sqlx.DB, cfg *config.Config) Model {
	return Model{
		DB:     db,
		Config: cfg,
		Items: []MenuItem{
			{Title: "Pomodoro", Desc: "Start a focused work session"},
			{Title: "Time Tracker", Desc: "Track time manually"},
			{Title: "Activities", Desc: "Manage activity types"},
			{Title: "Statistics", Desc: "View your progress"},
			{Title: "Configuration", Desc: "Customize settings"},
			{Title: "Quit", Desc: "Exit the application"},
		},
		CurrentView: HomeView,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case TickMsg:
		if m.Running && m.CurrentView == TimeTrackerView {
			m.Elapsed = time.Since(m.StartTime)
			return m, TickCmd()
		}
		if m.Running && m.CurrentView == PomodoroView {
			return m.tickPomodoro()
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.CurrentView {
		case HomeView:
			return m.updateHome(msg)
		case TimeTrackerView:
			return m.updateTracker(msg)
		case PomodoroView:
			return m.updatePomodoro(msg)
		case ActivitiesView:
			return m.updateActivities(msg)
		case StatisticsView:
			return m.updateStatistics(msg)
		case ConfigView:
			return m.updateConfig(msg)
		}

	default:
		if m.CurrentView == ActivitiesView {
			return m.updateActivities(msg)
		}
		if m.CurrentView == TimeTrackerView {
			return m.updateTracker(msg)
		}
		if m.CurrentView == PomodoroView {
			return m.updatePomodoro(msg)
		}
		if m.CurrentView == StatisticsView {
			return m.updateStatistics(msg)
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.CurrentView {
	case TimeTrackerView:
		return m.viewTracker()
	case PomodoroView:
		return m.viewPomodoro()
	case ActivitiesView:
		return m.viewActivities()
	case StatisticsView:
		return m.viewStatistics()
	case ConfigView:
		return m.viewConfig()
	default:
		return m.viewHome()
	}
}
