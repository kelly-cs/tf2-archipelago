package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleFocus = lipgloss.NewStyle().Bold(true)
	styleWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

// labelWidth keeps the values in one column whichever tab is on screen, which
// is what the window's own label column is for.
const labelWidth = 24

// view draws the settings over the whole screen: the tabs, the rows of the one
// that is open, the help under the row that is focused, and the keys.
func (f *settingsForm) view(width, height int) string {
	var out strings.Builder
	out.WriteString(f.tabLine())
	out.WriteString("\n\n")

	rows := f.rows(width, f.bodyHeight(height))
	out.WriteString(strings.Join(rows, "\n"))
	out.WriteString("\n")

	out.WriteString(f.footer(width))
	return out.String()
}

// bodyHeight is what the rows have after the tab line, the help and the keys.
func (f *settingsForm) bodyHeight(height int) int {
	return max(height-7, 3)
}

func (f *settingsForm) tabLine() string {
	rendered := make([]string, 0, len(f.tabs))
	for i, tab := range f.tabs {
		if i == f.tab {
			rendered = append(rendered, styleTabOn.Render(tab.title))
			continue
		}
		rendered = append(rendered, styleTab.Render(tab.title))
	}
	return strings.Join(rendered, "  ")
}

// rows is the fields of the open tab, scrolled to keep the focused one on
// screen. The Missions tab has a row per mission, which is more than any
// screen has lines.
func (f *settingsForm) rows(width, height int) []string {
	fields := f.fields()
	f.offset = min(f.offset, max(f.focused, 0))
	if f.focused >= f.offset+height {
		f.offset = f.focused - height + 1
	}
	if f.focused < f.offset {
		f.offset = f.focused
	}

	end := min(f.offset+height, len(fields))
	rows := make([]string, 0, height)
	for i := f.offset; i < end; i++ {
		rows = append(rows, f.row(fields[i], i == f.focused, width))
	}
	for len(rows) < height {
		rows = append(rows, "")
	}
	return rows
}

func (f *settingsForm) row(row field, focused bool, width int) string {
	marker, label := "  ", row.Label()
	if focused {
		marker = "> "
		label = styleFocus.Render(label)
	}
	pad := max(labelWidth-lipgloss.Width(row.Label()), 1)
	return truncate(marker+label+strings.Repeat(" ", pad)+row.Value(), width)
}

// footer is the help for the focused row, then the keys, then whatever the
// last save complained about.
func (f *settingsForm) footer(width int) string {
	var out strings.Builder

	fields := f.fields()
	if f.focused < len(fields) {
		out.WriteString(styleMuted.Render(truncate(fields[f.focused].Help(), width)))
	}
	out.WriteString("\n")

	pairs := [][2]string{
		{"↑↓", "row"},
		{"←→", "change"},
		{"tab", "tab"},
		{"ctrl+s", "save"},
		{"esc", "cancel"},
	}
	parts := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		parts = append(parts, styleKey.Render(pair[0])+" "+pair[1])
	}
	out.WriteString(styleMuted.Render(strings.Join(parts, "   ")))

	if f.warn != "" {
		out.WriteString("\n")
		out.WriteString(styleWarn.Render(truncate(f.warn, width)))
	}
	return out.String()
}

// summary is the line the main screen shows while the settings are shut: what
// the run is set to, which is what the window's own title bar cannot say.
func (m *model) summary() string {
	return fmt.Sprintf("%d missions, %s, goal %s",
		m.settings.MvmMissionCount, m.settings.MvmDifficulty, m.settings.MvmGoal)
}
