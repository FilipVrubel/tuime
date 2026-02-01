package ui

import (
	"fmt"
	"strings"
	"time"

	"tuime/internal/db"

	"github.com/NimbleMarkets/ntcharts/barchart"
	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
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
		m.activityPage = 0
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
			m.statsViewMode = (m.statsViewMode + 1) % 4
			m.activityPage = 0
		case "up", "k":
			if (m.statsViewMode == 0 || m.statsViewMode == 2) && m.activityPage > 0 {
				m.activityPage--
			}
		case "down", "j":
			if m.statsViewMode == 0 || m.statsViewMode == 2 {
				maxPage := (len(m.activityStats) - 1) / 6
				if m.activityPage < maxPage {
					m.activityPage++
				}
			}
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
		daysFromMonday := int(now.Weekday()) - 1
		if daysFromMonday < 0 {
			daysFromMonday = 6
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
			filterDisplay.WriteString(" | ")
		}
	}
	b.WriteString(centerStyle.Render(filterDisplay.String()))
	b.WriteString("\n")

	views := []string{"Overview", "Time Trends", "Activity Breakdown", "Stacked Bar Chart"}
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
		return m.viewCombinedTimeSeries(&b, centerStyle, width, height)
	case 2:
		return m.viewActivityBreakdown(&b, centerStyle, width, height)
	case 3:
		return m.viewStackedBarChart(&b, centerStyle, width, height)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) viewStatsOverview(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7FE5D0"))
	b.WriteString(centerStyle.Render(cyanStyle.Render("Total Sessions: ") + NormalStyle.Render(fmt.Sprintf("%d", m.overallStats.TotalSessions))))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(cyanStyle.Render("Total Time: ") + NormalStyle.Render(formatSeconds(m.overallStats.TotalTime))))
	b.WriteString("\n\n")

	if len(m.typeStats) > 0 {
		b.WriteString(centerStyle.Render(cyanStyle.Render("By Type:")))
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
		itemsPerPage := 6
		totalPages := (len(m.activityStats) + itemsPerPage - 1) / itemsPerPage
		if m.activityPage >= totalPages {
			m.activityPage = totalPages - 1
		}
		if m.activityPage < 0 {
			m.activityPage = 0
		}

		start := m.activityPage * itemsPerPage
		end := start + itemsPerPage
		if end > len(m.activityStats) {
			end = len(m.activityStats)
		}
		pageStats := m.activityStats[start:end]

		if totalPages > 1 {
			b.WriteString(centerStyle.Render(cyanStyle.Render(fmt.Sprintf("By Activity (Page %d/%d):", m.activityPage+1, totalPages))))
		} else {
			b.WriteString(centerStyle.Render(cyanStyle.Render("By Activity:")))
		}
		b.WriteString("\n")

		for _, stat := range pageStats {
			b.WriteString(centerStyle.Render(fmt.Sprintf("  %s: %s (%d sessions)",
				stat.ActivityName,
				formatSeconds(stat.TotalTime),
				stat.SessionCount)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	paginationHelp := ""
	if len(m.activityStats) > 0 {
		itemsPerPage := 6
		totalPages := (len(m.activityStats) + itemsPerPage - 1) / itemsPerPage
		if totalPages > 1 {
			paginationHelp = " • ↑/↓: page"
		}
	}
	b.WriteString(centerStyle.Render(HelpStyle.Render(fmt.Sprintf("←/→: change period • tab: switch view%s • esc: back", paginationHelp))))

	return b.String()
}

func (m Model) viewCombinedTimeSeries(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	if len(m.dailyStats) == 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("No data available for this period")))
		b.WriteString("\n\n")
		b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
	}

	maxY := 0.0
	for _, stat := range m.dailyStats {
		pomodoroHours := float64(stat.PomodoroTime) / 3600.0
		trackerHours := float64(stat.TrackerTime) / 3600.0
		if pomodoroHours > maxY {
			maxY = pomodoroHours
		}
		if trackerHours > maxY {
			maxY = trackerHours
		}
	}
	if maxY == 0 {
		maxY = 0.1
	}

	chartWidth := width - 10
	if chartWidth < 40 {
		chartWidth = 40
	}

	chartHeight := height - 15
	if chartHeight < 10 {
		chartHeight = 10
	}
	if chartHeight > 12 {
		chartHeight = 12
	}

	chart := tslc.New(chartWidth, chartHeight,
		tslc.WithYRange(0, maxY*1.1))

	for _, stat := range m.dailyStats {
		date, _ := time.Parse("2006-01-02", stat.Date)
		pomodoroHours := float64(stat.PomodoroTime) / 3600.0
		chart.Push(tslc.TimePoint{Time: date, Value: pomodoroHours})
	}

	chart.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("9")))

	for _, stat := range m.dailyStats {
		date, _ := time.Parse("2006-01-02", stat.Date)
		trackerHours := float64(stat.TrackerTime) / 3600.0
		chart.PushDataSet("tracker", tslc.TimePoint{Time: date, Value: trackerHours})
	}

	chart.SetDataSetStyle("tracker", lipgloss.NewStyle().Foreground(lipgloss.Color("11")))

	if len(m.dailyStats) > 0 {
		totalCombined := 0.0
		for _, stat := range m.dailyStats {
			totalCombined += float64(stat.PomodoroTime+stat.TrackerTime) / 3600.0
		}
		avgHours := totalCombined / float64(len(m.dailyStats))
		for _, stat := range m.dailyStats {
			date, _ := time.Parse("2006-01-02", stat.Date)
			chart.PushDataSet("average", tslc.TimePoint{Time: date, Value: avgHours})
		}
		chart.SetDataSetStyle("average", lipgloss.NewStyle().Foreground(lipgloss.Color("10")))
	}

	chart.DrawBrailleAll()
	chartView := chart.View()

	b.WriteString(centerStyle.Render(NormalStyle.Render("Focus Time Over Time")))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("X-axis: Date | Y-axis: Hours")))
	b.WriteString("\n\n")

	chartLines := strings.Split(chartView, "\n")
	for _, line := range chartLines {
		b.WriteString(centerStyle.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	totalPomodoros := 0
	totalTrackers := 0
	totalPomodoroTime := 0
	totalTrackerTime := 0
	for _, stat := range m.dailyStats {
		totalPomodoros += stat.PomodoroCount
		totalTrackers += stat.TrackerCount
		totalPomodoroTime += stat.PomodoroTime
		totalTrackerTime += stat.TrackerTime
	}

	avgTotalTime := 0.0
	if len(m.dailyStats) > 0 {
		totalCombined := 0.0
		for _, stat := range m.dailyStats {
			totalCombined += float64(stat.PomodoroTime+stat.TrackerTime) / 60.0
		}
		avgTotalTime = totalCombined / float64(len(m.dailyStats))
	}

	legendPomodoroStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	legendTrackerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	legendAvgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	b.WriteString(centerStyle.Render(legendPomodoroStyle.Render("Pomodoro: ") +
		fmt.Sprintf("%d sessions, %s total", totalPomodoros, formatSeconds(totalPomodoroTime))))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(legendTrackerStyle.Render("Tracker: ") +
		fmt.Sprintf("%d sessions, %s total", totalTrackers, formatSeconds(totalTrackerTime))))
	b.WriteString("\n")
	if avgTotalTime > 0 {
		avgTotalStr := formatSeconds(int(avgTotalTime * 60))
		b.WriteString(centerStyle.Render(legendAvgStyle.Render("Daily Avg: ") + avgTotalStr))
	}
	b.WriteString("\n")

	b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) viewActivityBreakdown(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	if len(m.activityStats) == 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("No activity data available for this period")))
		b.WriteString("\n\n")
		b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
	}

	itemsPerPage := 6
	totalPages := (len(m.activityStats) + itemsPerPage - 1) / itemsPerPage
	if m.activityPage >= totalPages {
		m.activityPage = totalPages - 1
	}
	if m.activityPage < 0 {
		m.activityPage = 0
	}

	start := m.activityPage * itemsPerPage
	end := start + itemsPerPage
	if end > len(m.activityStats) {
		end = len(m.activityStats)
	}
	pageStats := m.activityStats[start:end]

	b.WriteString(centerStyle.Render(NormalStyle.Render(fmt.Sprintf("Activity Time Distribution (Page %d/%d)", m.activityPage+1, totalPages))))
	b.WriteString("\n\n")

	var totalTime int
	for _, stat := range m.activityStats {
		totalTime += stat.TotalTime
	}

	var listBuilder strings.Builder
	for i, stat := range pageStats {
		globalIndex := start + i
		minutes := float64(stat.TotalTime) / 60.0
		percentage := 0.0
		if totalTime > 0 {
			percentage = float64(stat.TotalTime) / float64(totalTime) * 100.0
		}
		line := fmt.Sprintf("%d. %s - %.1f min (%.1f%%)", globalIndex+1, stat.ActivityName, minutes, percentage)
		listBuilder.WriteString(line)
		listBuilder.WriteString("\n")
	}
	for i := len(pageStats); i < itemsPerPage; i++ {
		listBuilder.WriteString("\n")
	}
	b.WriteString(centerStyle.Render(NormalStyle.Render(listBuilder.String())))
	b.WriteString("\n")

	listAndUIHeight := 13
	itemsPerPageForWidth := 6
	chartWidth := width - 10
	if chartWidth < 80 {
		chartWidth = 80
	}

	chartHeight := height - listAndUIHeight
	if chartHeight < 12 {
		chartHeight = 12
	}
	if chartHeight > 12 {
		chartHeight = 12
	}

	yellowBarColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Background(lipgloss.Color("228"))

	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	bc := barchart.New(chartWidth, chartHeight,
		barchart.WithStyles(axisStyle, labelStyle))

	for i, stat := range pageStats {
		globalIndex := start + i
		minutes := float64(stat.TotalTime) / 60.0

		bc.Push(barchart.BarData{
			Label: fmt.Sprintf("%d", globalIndex+1),
			Values: []barchart.BarValue{
				{Value: minutes, Style: yellowBarColor},
			},
		})
	}

	for i := len(pageStats); i < itemsPerPageForWidth; i++ {
		bc.Push(barchart.BarData{
			Label: "",
			Values: []barchart.BarValue{
				{Value: 0, Style: yellowBarColor},
			},
		})
	}

	bc.Draw()
	chartView := bc.View()
	chartBox := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	b.WriteString(chartBox.Render(chartView))
	b.WriteString("\n\n")

	paginationHelp := ""
	if totalPages > 1 {
		paginationHelp = " • ↑/↓: page"
	}
	b.WriteString(centerStyle.Render(HelpStyle.Render(fmt.Sprintf("←/→: change period • tab: switch view%s • esc: back", paginationHelp))))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m Model) viewStackedBarChart(b *strings.Builder, centerStyle lipgloss.Style, width, height int) string {
	if len(m.dailyStats) == 0 {
		b.WriteString(centerStyle.Render(NormalStyle.Render("No data available for this period")))
		b.WriteString("\n\n")
		b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
	}

	chartWidth := width - 10
	if chartWidth < 100 {
		chartWidth = 100
	}

	chartHeight := height - 18
	if chartHeight < 12 {
		chartHeight = 12
	}
	if chartHeight > 12 {
		chartHeight = 12
	}

	pomodoroBarColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Background(lipgloss.Color("9"))
	trackerBarColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Background(lipgloss.Color("11"))

	pomodoroLegendColor := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	trackerLegendColor := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	bc := barchart.New(chartWidth, chartHeight,
		barchart.WithStyles(axisStyle, labelStyle))

	maxBars := 15
	step := 1
	if len(m.dailyStats) > maxBars {
		step = (len(m.dailyStats) + maxBars - 1) / maxBars
	}

	actualBarCount := 0
	for i := 0; i < len(m.dailyStats); i += step {
		actualBarCount++
	}

	labelInterval := 1
	if actualBarCount > 15 {
		labelInterval = 4
	} else if actualBarCount > 10 {
		labelInterval = 3
	} else if actualBarCount > 7 {
		labelInterval = 2
	}

	maxValue := 0.0

	barIndex := 0
	for i := 0; i < len(m.dailyStats); i += step {
		stat := m.dailyStats[i]
		date, _ := time.Parse("2006-01-02", stat.Date)

		pomodoroMinutes := float64(stat.PomodoroTime) / 60.0
		trackerMinutes := float64(stat.TrackerTime) / 60.0
		totalMinutes := pomodoroMinutes + trackerMinutes

		if totalMinutes > maxValue {
			maxValue = totalMinutes
		}

		var label string
		if barIndex%labelInterval == 0 {
			if m.statsFilterIndex == 3 {
				label = date.Format("Jan/06")
			} else if step > 1 {
				label = date.Format("Jan2")
			} else {
				label = date.Format("1/2")
			}
		} else {
			label = "" // Empty label to reduce crowding
		}

		bc.Push(barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{Value: pomodoroMinutes, Style: pomodoroBarColor},
				{Value: trackerMinutes, Style: trackerBarColor},
			},
		})
		barIndex++
	}

	for i := actualBarCount; i < maxBars; i++ {
		bc.Push(barchart.BarData{
			Label: "",
			Values: []barchart.BarValue{
				{Value: 0, Style: pomodoroBarColor},
			},
		})
	}

	bc.Draw()
	chartView := bc.View()

	maxValueStr := fmt.Sprintf("%.0f min", maxValue)
	if maxValue >= 60 {
		maxValueStr = fmt.Sprintf("%.1fh", maxValue/60.0)
	}

	b.WriteString(centerStyle.Render(NormalStyle.Render(fmt.Sprintf("Daily Focus Time (Stacked) - Peak: %s", maxValueStr))))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("Y-axis: Minutes | Red = Pomodoro, Yellow = Tracker")))
	b.WriteString("\n\n")

	chartBox := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	b.WriteString(chartBox.Render(chartView))
	b.WriteString("\n\n")

	totalPomodoros := 0
	totalTrackers := 0
	totalPomodoroTime := 0
	totalTrackerTime := 0
	for _, stat := range m.dailyStats {
		totalPomodoros += stat.PomodoroCount
		totalTrackers += stat.TrackerCount
		totalPomodoroTime += stat.PomodoroTime
		totalTrackerTime += stat.TrackerTime
	}

	b.WriteString(centerStyle.Render(pomodoroLegendColor.Render("Pomodoro: ") +
		fmt.Sprintf("%d sessions, %s total", totalPomodoros, formatSeconds(totalPomodoroTime))))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(trackerLegendColor.Render("Tracker: ") +
		fmt.Sprintf("%d sessions, %s total", totalTrackers, formatSeconds(totalTrackerTime))))
	b.WriteString("\n")
	b.WriteString(centerStyle.Render(HelpStyle.Render("←/→: change period • tab: switch view • esc: back")))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, b.String())
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
