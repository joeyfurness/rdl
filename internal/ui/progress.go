package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the progress TUI.
var (
	styleFilename = lipgloss.NewStyle().Bold(true)
	styleDone     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleFail     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleStatus   = lipgloss.NewStyle().Faint(true)
)

// FileStatus tracks the download state of a single file.
type FileStatus struct {
	Filename       string
	TotalBytes     int64
	CompletedBytes int64
	Speed          int64
	Done           bool
	Error          string
	StartedAt      time.Time
}

// ProgressModel is a bubbletea model that displays per-file download progress.
type ProgressModel struct {
	Files      map[string]*FileStatus
	Order      []string
	TotalFiles int
	DoneCount  int
	FailCount  int

	bars  map[string]progress.Model
	width int
}

// Message types for the progress TUI.

// TickMsg triggers a periodic UI refresh.
type TickMsg time.Time

// FileStartMsg signals that a file download has started.
type FileStartMsg struct {
	Filename   string
	TotalBytes int64
}

// FileProgressMsg updates bytes downloaded and speed for a file.
type FileProgressMsg struct {
	Filename       string
	CompletedBytes int64
	Speed          int64
}

// FileDoneMsg signals that a file download completed successfully.
type FileDoneMsg struct {
	Filename   string
	TotalBytes int64
}

// FileErrorMsg signals that a file download failed.
type FileErrorMsg struct {
	Filename string
	Message  string
}

// AllDoneMsg signals that all downloads are complete and the TUI should exit.
type AllDoneMsg struct{}

// NewProgressModel creates a new ProgressModel.
func NewProgressModel(totalFiles int) ProgressModel {
	return ProgressModel{
		Files:      make(map[string]*FileStatus),
		Order:      nil,
		TotalFiles: totalFiles,
		bars:       make(map[string]progress.Model),
		width:      80,
	}
}

// Init implements tea.Model.
func (m ProgressModel) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update implements tea.Model.
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		for name, bar := range m.bars {
			bar.Width = m.barWidth()
			m.bars[name] = bar
		}

	case TickMsg:
		return m, tickCmd()

	case FileStartMsg:
		fs := &FileStatus{
			Filename:   msg.Filename,
			TotalBytes: msg.TotalBytes,
			StartedAt:  time.Now(),
		}
		m.Files[msg.Filename] = fs
		m.Order = append(m.Order, msg.Filename)
		bar := progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage())
		bar.Width = m.barWidth()
		m.bars[msg.Filename] = bar

	case FileProgressMsg:
		if fs, ok := m.Files[msg.Filename]; ok {
			fs.CompletedBytes = msg.CompletedBytes
			fs.Speed = msg.Speed
		}

	case FileDoneMsg:
		if fs, ok := m.Files[msg.Filename]; ok {
			fs.Done = true
			fs.CompletedBytes = msg.TotalBytes
			fs.TotalBytes = msg.TotalBytes
			m.DoneCount++
		}

	case FileErrorMsg:
		if fs, ok := m.Files[msg.Filename]; ok {
			fs.Error = msg.Message
			m.FailCount++
		}

	case AllDoneMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m ProgressModel) barWidth() int {
	w := m.width - 40
	if w < 10 {
		w = 10
	}
	if w > 60 {
		w = 60
	}
	return w
}

// View implements tea.Model.
func (m ProgressModel) View() string {
	var b strings.Builder

	for _, name := range m.Order {
		fs := m.Files[name]
		if fs == nil {
			continue
		}
		displayName := styleFilename.Render(truncateMiddle(fs.Filename, 30))

		switch {
		case fs.Error != "":
			fmt.Fprintf(&b, "%s %s\n", displayName, styleFail.Render("FAIL: "+fs.Error))
		case fs.Done:
			fmt.Fprintf(&b, "%s %s %s\n", displayName,
				styleDone.Render("done"),
				styleStatus.Render(formatBytesShort(fs.TotalBytes)))
		default:
			pct := 0.0
			if fs.TotalBytes > 0 {
				pct = float64(fs.CompletedBytes) / float64(fs.TotalBytes)
			}
			bar, ok := m.bars[name]
			barStr := ""
			if ok {
				barStr = bar.ViewAs(pct)
			}
			speed := styleStatus.Render(formatSpeed(fs.Speed))
			eta := styleStatus.Render(formatETA(fs.CompletedBytes, fs.TotalBytes, fs.Speed))
			fmt.Fprintf(&b, "%s %s %s %s\n", displayName, barStr, speed, eta)
		}
	}

	// Summary line.
	active := len(m.Order) - m.DoneCount - m.FailCount
	if active < 0 {
		active = 0
	}
	summary := styleStatus.Render(fmt.Sprintf("%d files | %d active | %d done | %d failed",
		m.TotalFiles, active, m.DoneCount, m.FailCount))
	fmt.Fprintf(&b, "\n%s\n", summary)

	return b.String()
}

// truncateMiddle truncates s in the middle with an ellipsis, preserving the file extension.
func truncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 5 {
		return s[:maxLen]
	}
	ext := filepath.Ext(s)
	// Ensure we have room for at least a few chars + ellipsis + extension.
	if len(ext)+4 > maxLen {
		// Extension too long; just truncate plainly.
		half := (maxLen - 1) / 2
		return s[:half] + "\u2026" + s[len(s)-half:]
	}
	available := maxLen - len(ext) - 1 // 1 for ellipsis
	front := available / 2
	back := available - front
	base := strings.TrimSuffix(s, ext)
	if back > len(base) {
		back = len(base)
	}
	return base[:front] + "\u2026" + base[len(base)-back:] + ext
}

// formatSpeed formats bytes per second as a human-readable speed string.
func formatSpeed(bytesPerSec int64) string {
	if bytesPerSec <= 0 {
		return "-- MB/s"
	}
	switch {
	case bytesPerSec >= 1<<30:
		return fmt.Sprintf("%.1f GB/s", float64(bytesPerSec)/float64(1<<30))
	case bytesPerSec >= 1<<20:
		return fmt.Sprintf("%.1f MB/s", float64(bytesPerSec)/float64(1<<20))
	case bytesPerSec >= 1<<10:
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B/s", bytesPerSec)
	}
}

// formatETA returns a human-readable estimated time remaining.
func formatETA(completed, total, speed int64) string {
	if speed <= 0 || total <= 0 || completed >= total {
		return ""
	}
	remaining := total - completed
	secs := remaining / speed
	if secs < 0 {
		return ""
	}
	if secs < 60 {
		return fmt.Sprintf("ETA %ds", secs)
	}
	m := secs / 60
	s := secs % 60
	if s == 0 {
		return fmt.Sprintf("ETA %dm", m)
	}
	return fmt.Sprintf("ETA %dm%ds", m, s)
}

// formatBytesShort formats byte counts in a compact human-readable form.
func formatBytesShort(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
