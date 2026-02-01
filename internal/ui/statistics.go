package ui

import (
	"fmt"
	"strings"
	"time"

	"tuime/internal/db"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
)

type statsLoadedMsg struct {
	activityStats []db.ActivityStats
	typeStats     []db.TypeStats
	overall       *db.OverallStats
	dailyStats    []db.DailySessionStats
}

type statsErrorMsg struct {
	err error
}

func loadStatsCmd(database *sqlx.DB, startDate, endDate *time.Time) tea.Cmd {
	return func() tea.Msg {
		activityStats, err := db.GetActivityStats(database, startDate, endDate)
		if err != nil {
			return statsErrorMsg{err: err}
		}

		typeStats, err := db.GetTypeStats(database, startDate, endDate)
		if err != nil {
			return statsErrorMsg{err: err}
		}

		overall, err := db.GetOverallStats(database, startDate, endDate)
		if err != nil {
			return statsErrorMsg{err: err}
		}

		dailyStats, err := db.GetDailySessionStats(database, startDate, endDate)
		if err != nil {
			return statsErrorMsg{err: err}
		}

		return statsLoadedMsg{
			activityStats: activityStats,
			typeStats:     typeStats,
			overall:       overall,
			dailyStats:    dailyStats,
		}
	}
}

func (m Model) startStatistics() (tea.Model, tea.Cmd) {
	m.CurrentView = StatisticsView
	m.statsFilterIndex = 3
	m.statsViewMode = 0
	return m, loadStatsCmd(m.DB, nil, nil)
}

func (m Model) updateStatistics(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statsLoadedMsg:
		m.activityStats = msg.activityStats
		m.typeStats = msg.typeStats
		m.overallStats = msg.overall
		m.dailyStats = msg.dailyStats
		return m, nil

	case statsErrorMsg:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.statsFilterIndex > 0 {
				m.statsFilterIndex--
				return m.loadStatsForFilter()
			}
		case "right", "l":
			if m.statsFilterIndex < 3 {
				m.statsFilterIndex++
				return m.loadStatsForFilter()
			}
		case "tab":
			m.statsViewMode = (m.statsViewMode + 1) % 3
		case "esc":
			m.CurrentView = HomeView
		}
	}
	return m, nil
}

func (m Model) loadStatsForFilter() (tea.Model, tea.Cmd) {
	now := time.Now()
	var startDate, endDate *time.Time

	switch m.statsFilterIndex {
	case 0:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		startDate = &start
		endDate = &end
	case 1:
		// Calculate the start of this week (Monday)
		daysFromMonday := int(now.Weekday()) - 1
		if daysFromMonday < 0 {
			daysFromMonday = 6 // Sunday is 6 days after Monday
		}
		start := now.AddDate(0, 0, -daysFromMonday)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end := start.AddDate(0, 0, 7)
		startDate = &start
		endDate = &end
	case 2:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0)
		startDate = &start
		endDate = &end
	case 3:
		startDate = nil
		endDate = nil
	}

	return m, loadStatsCmd(m.DB, startDate, endDate)
}

func (m Model) viewStatistics() string {
	var b strings.Builder

	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 25
	}
	centerStyle := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)

	b.WriteString(centerStyle.Render(TitleStyle.Render("Statistics")))
	b.WriteString("\n\n")

	filters := []string{"Today", "This Week", "This Month", "All Time"}
	var filterDisplay strings.Builder
	for i, filter := range filters {
		if i == m.statsFilterIndex {
			filterDisplay.WriteString(SelectedStyle.Render(filter))
		} else {
			filterDisplay.WriteString(NormalStyle.Render(filter))
		}
		if i < len(filters)-1 {
			filterDisplay.WriteString("  ")
		}
	}
	b.WriteString(centerStyle.Render(filterDisplay.String()))
	b.WriteString("\n")

	views := []string{"Overview", "Pomodoro Chart", "Tracker Chart"}
	var viewDisplay strings.Builder
	for i, view := range views {
		if i == m.statsViewMode {
			viewDisplay.WriteString(SelectedStyle.Render(view))
		} else {
			viewDisplay.WriteString(NormalStyle.Render(view))
		}
		if i < len(views)-1 {
			viewDisplay.WriteString(" | ")
		}
	}
	b.WriteString(centerStyle.Render(viewDisplay.String()))
	b.WriteString("\n\n")

	if m.overallStats == nil {
		b.WriteString(centerStyle.Render(NormalStyle.Render("Loading...")))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
	}

	switch m.statsViewMode {
	case 0:
		return m.viewStatsOverview(&b, centerStyle, width, height)
	case 1:
		return m.viewPomodoroTimeSeries(&b, centerStyle, width, height)
	case 2:
		return m.viewTrackerTimeSeries(&b, centerStyle, width, height)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) viewStatsOverview(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	b.WriteString(centerStyle.Render(NormalStyle.Render(fmt.Sprintf("Total Sessions: %d", m.overallStats.TotalSessions))))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(NormalStyle.Render(fmt.Sprintf("Total Time: %s", formatSeconds(m.overallStats.TotalTime)))))
	b.WriteString("\n\n")

	if len(m.typeStats) > 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("By Type:")))
		b.WriteString("\n")
		for _, stat := range m.typeStats {
			b.WriteString(centerStyle.Render(fmt.Sprintf("  %s: %s (%d sessions)",
				stat.Type,
				formatSeconds(stat.TotalTime),
				stat.SessionCount)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(m.activityStats) > 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("By Activity:")))
		b.WriteString("\n\n")

		chart := m.createActivityChart()
		if chart != "" {
			chartLines := strings.Split(chart, "\n")
			for _, line := range chartLines {
				b.WriteString(centerStyle.Render(line))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		for _, stat := range m.activityStats {
			b.WriteString(centerStyle.Render(fmt.Sprintf("  %s: %s (%d sessions)",
				stat.ActivityName,
				formatSeconds(stat.TotalTime),
				stat.SessionCount)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))

	return b.String()
}

func (m Model) viewPomodoroTimeSeries(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	if len(m.dailyStats) == 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("No pomodoro data available for this period")))
		b.WriteString("\n\n")
		b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
	}

	minX := 0.0
	maxX := float64(len(m.dailyStats) - 1)
	if maxX < 1 {
		maxX = 1
		minX = -0.5
	}
	minY := 0.0
	maxY := 0.0

	for _, stat := range m.dailyStats {
		minutes := float64(stat.PomodoroTime) / 60.0
		if minutes > maxY {
			maxY = minutes
		}
	}

	if maxY == 0 {
		maxY = 1
	}

	xLabelFormatter := func(i int, v float64) string {
		idx := int(v)
		if idx >= 0 && idx < len(m.dailyStats) {
			date := m.dailyStats[idx].Date
			if len(date) >= 10 {
				if m.statsFilterIndex == 3 {
					if t, err := time.Parse("2006-01-02", date); err == nil {
						return t.Format("Jan/06")
					}
				}
				return date[5:10]
			}
		}
		return fmt.Sprintf("%.0f", v)
	}

	yLabelFormatter := func(i int, v float64) string {
		return fmt.Sprintf("%.0fm", v)
	}

	lc := linechart.New(60, 12, minX, maxX, minY, maxY*1.1,
		linechart.WithXLabelFormatter(xLabelFormatter),
		linechart.WithYLabelFormatter(yLabelFormatter))

	for i, stat := range m.dailyStats {
		minutes := float64(stat.PomodoroTime) / 60.0
		point := canvas.Float64Point{X: float64(i), Y: minutes}

		lc.DrawBrailleCircle(point, 0.3)

		if i > 0 {
			prevMinutes := float64(m.dailyStats[i-1].PomodoroTime) / 60.0
			prevPoint := canvas.Float64Point{X: float64(i - 1), Y: prevMinutes}
			lc.DrawBrailleLine(prevPoint, point)
		}
	}

	lc.DrawXYAxisAndLabel()
	chart := lc.View()

	b.WriteString(centerStyle.Render(NormalStyle.Render("Pomodoro Sessions Over Time")))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("X-axis: Date (MM-DD) | Y-axis: Minutes")))
	b.WriteString("\n\n")

	chartLines := strings.Split(chart, "\n")
	for _, line := range chartLines {
		b.WriteString(centerStyle.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	totalPomodoros := 0
	totalTime := 0
	for _, stat := range m.dailyStats {
		totalPomodoros += stat.PomodoroCount
		totalTime += stat.PomodoroTime
	}

	b.WriteString(centerStyle.Render(fmt.Sprintf("Total Pomodoros: %d | Total Time: %s",
		totalPomodoros, formatSeconds(totalTime))))
	b.WriteString("\n\n")

	b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) viewTrackerTimeSeries(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	if len(m.dailyStats) == 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("No tracker data available for this period")))
		b.WriteString("\n\n")
		b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
	}

	minX := 0.0
	maxX := float64(len(m.dailyStats) - 1)
	if maxX < 1 {
		maxX = 1
		minX = -0.5
	}
	minY := 0.0
	maxY := 0.0

	for _, stat := range m.dailyStats {
		minutes := float64(stat.TrackerTime) / 60.0
		if minutes > maxY {
			maxY = minutes
		}
	}

	if maxY == 0 {
		maxY = 1
	}

	xLabelFormatter := func(i int, v float64) string {
		idx := int(v)
		if idx >= 0 && idx < len(m.dailyStats) {
			date := m.dailyStats[idx].Date
			if len(date) >= 10 {
				if m.statsFilterIndex == 3 {
					if t, err := time.Parse("2006-01-02", date); err == nil {
						return t.Format("Jan/06")
					}
				}
				return date[5:10]
			}
		}
		return fmt.Sprintf("%.0f", v)
	}

	yLabelFormatter := func(i int, v float64) string {
		return fmt.Sprintf("%.0fm", v)
	}

	lc := linechart.New(60, 12, minX, maxX, minY, maxY*1.1,
		linechart.WithXLabelFormatter(xLabelFormatter),
		linechart.WithYLabelFormatter(yLabelFormatter))

	for i, stat := range m.dailyStats {
		minutes := float64(stat.TrackerTime) / 60.0
		point := canvas.Float64Point{X: float64(i), Y: minutes}

		lc.DrawBrailleCircle(point, 0.3)

		if i > 0 {
			prevMinutes := float64(m.dailyStats[i-1].TrackerTime) / 60.0
			prevPoint := canvas.Float64Point{X: float64(i - 1), Y: prevMinutes}
			lc.DrawBrailleLine(prevPoint, point)
		}
	}

	lc.DrawXYAxisAndLabel()
	chart := lc.View()

	b.WriteString(centerStyle.Render(NormalStyle.Render("Tracker Sessions Over Time")))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("X-axis: Date (MM-DD) | Y-axis: Minutes")))
	b.WriteString("\n\n")

	chartLines := strings.Split(chart, "\n")
	for _, line := range chartLines {
		b.WriteString(centerStyle.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	totalSessions := 0
	totalTime := 0
	for _, stat := range m.dailyStats {
		totalSessions += stat.TrackerCount
		totalTime += stat.TrackerTime
	}

	b.WriteString(centerStyle.Render(fmt.Sprintf("Total Sessions: %d | Total Time: %s",
		totalSessions, formatSeconds(totalTime))))
	b.WriteString("\n\n")

	b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) createActivityChart() string {
	if len(m.activityStats) == 0 {
		return ""
	}

	bc := barchart.New(50, 10)

	for _, stat := range m.activityStats {
		minutes := float64(stat.TotalTime) / 60.0
		label := stat.ActivityName
		if len(label) > 15 {
			label = label[:12] + "..."
		}
		bc.Push(barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{Value: minutes},
			},
		})
	}

	bc.Draw()
	return bc.View()
}

func formatSeconds(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
