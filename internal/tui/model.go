package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Jonathangadeaharder/structurelint/internal/autofix"
	"github.com/Jonathangadeaharder/structurelint/internal/linter"
)

// View modes
type viewMode int

const (
	modeList viewMode = iota
	modeDetail
	modeFixPreview
	modeGraph
)

// Model holds the TUI state
type Model struct {
	violations    []linter.Violation
	fixableOnly   bool
	cursor        int
	viewMode      viewMode
	width         int
	height        int
	fixEngine     *autofix.Engine
	selectedFix   *autofix.Fix
	statusMessage string
	quitting      bool
}

// NewModel creates a new TUI model
func NewModel(violations []linter.Violation, fixableOnly bool) Model {
	return Model{
		violations:  violations,
		fixableOnly: fixableOnly,
		cursor:      0,
		viewMode:    modeList,
		fixEngine:   autofix.NewEngine(true), // Start in preview mode
	}
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("211"))

	quitKeyBindings = key.NewBinding(key.WithKeys("q", "ctrl+c"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Background(lipgloss.Color("235"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case modeList:
		return m.handleListKeys(msg)
	case modeDetail:
		return m.handleDetailKeys(msg)
	case modeFixPreview:
		return m.handleFixPreviewKeys(msg)
	}
	return m, nil
}

// handleListKeys handles keys in list view
func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitKeyBindings):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.violations)-1 {
			m.cursor++
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "space"))):
		m.viewMode = modeDetail

	case key.Matches(msg, key.NewBinding(key.WithKeys("f"))):
		// Check if current violation is fixable
		if m.cursor < len(m.violations) && m.violations[m.cursor].AutoFix != nil {
			m.viewMode = modeFixPreview
			// Generate fix preview
			fixes, err := m.fixEngine.GenerateFixes([]linter.Violation{m.violations[m.cursor]}, nil)
			if err == nil && len(fixes) > 0 {
				m.selectedFix = fixes[0]
			}
		} else {
			m.statusMessage = "❌ This violation cannot be auto-fixed"
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
		m.viewMode = modeGraph
		m.statusMessage = "📊 Graph view (coming soon)"
	}

	return m, nil
}

// handleDetailKeys handles keys in detail view
func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitKeyBindings):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "backspace"))):
		m.viewMode = modeList
		m.statusMessage = ""
	}

	return m, nil
}

// handleFixPreviewKeys handles keys in fix preview view
func (m Model) handleFixPreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitKeyBindings):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "backspace"))):
		m.viewMode = modeList
		m.statusMessage = ""
		m.selectedFix = nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "y"))):
		// Apply the fix
		if m.selectedFix != nil {
			// Switch engine to non-dry-run mode
			m.fixEngine = autofix.NewEngine(false)
			applied, err := m.fixEngine.ApplyFixes([]*autofix.Fix{m.selectedFix})
			if err != nil {
				m.statusMessage = errorStyle.Render(fmt.Sprintf("❌ Failed to apply fix: %v", err))
			} else if applied > 0 {
				m.statusMessage = successStyle.Render("✓ Fix applied successfully!")
				// Remove fixed violation from list
				m.violations = append(m.violations[:m.cursor], m.violations[m.cursor+1:]...)
				if m.cursor >= len(m.violations) && m.cursor > 0 {
					m.cursor--
				}
			}
		}
		m.viewMode = modeList
		m.selectedFix = nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
		m.viewMode = modeList
		m.statusMessage = "Fix cancelled"
		m.selectedFix = nil
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.viewMode {
	case modeList:
		return m.renderList()
	case modeDetail:
		return m.renderDetail()
	case modeFixPreview:
		return m.renderFixPreview()
	case modeGraph:
		return m.renderGraph()
	}

	return ""
}

// renderList renders the violation list view
func (m Model) renderList() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Structurelint - Interactive Mode")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Summary
	summary := fmt.Sprintf("Found %d violation(s)", len(m.violations))
	if m.fixableOnly {
		summary += " (fixable only)"
	}
	b.WriteString(headerStyle.Render(summary))
	b.WriteString("\n\n")

	// Violations list
	visibleStart := m.cursor - 10
	if visibleStart < 0 {
		visibleStart = 0
	}
	visibleEnd := visibleStart + 20
	if visibleEnd > len(m.violations) {
		visibleEnd = len(m.violations)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		v := m.violations[i]

		// Format violation line
		prefix := "  "
		if i == m.cursor {
			prefix = "▶ "
		}

		// Add fix indicator
		fixIndicator := "  "
		if v.AutoFix != nil {
			fixIndicator = "🔧"
		}

		line := fmt.Sprintf("%s%s %-30s %s", prefix, fixIndicator, truncate(v.Rule, 28), truncate(v.Path, 50))

		// Style based on cursor position
		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Show truncation indicators
	if visibleStart > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more above ...", visibleStart)))
		b.WriteString("\n")
	}
	if visibleEnd < len(m.violations) {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more below ...", len(m.violations)-visibleEnd)))
		b.WriteString("\n")
	}

	// Status message
	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(m.statusMessage)
		b.WriteString("\n")
	}

	// Help text
	help := helpStyle.Render("↑/↓: Navigate | Enter: Details | f: Fix | g: Graph | q: Quit")
	b.WriteString("\n")
	b.WriteString(help)

	return b.String()
}

// renderDetail renders the detailed view of a violation
func (m Model) renderDetail() string {
	if m.cursor >= len(m.violations) {
		return "No violation selected"
	}

	v := m.violations[m.cursor]

	var b strings.Builder

	// Title
	title := titleStyle.Render("Violation Details")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Details
	details := fmt.Sprintf("%s\n\nRule:     %s\nFile:     %s\nMessage:  %s\n",
		headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		v.Rule,
		v.Path,
		v.Message,
	)

	if v.Expected != "" {
		details += fmt.Sprintf("Expected: %s\n", v.Expected)
	}
	if v.Actual != "" {
		details += fmt.Sprintf("Actual:   %s\n", v.Actual)
	}
	if v.Context != "" {
		details += fmt.Sprintf("Context:  %s\n", v.Context)
	}

	// Suggestions
	if len(v.Suggestions) > 0 {
		details += "\nSuggestions:\n"
		for i, s := range v.Suggestions {
			details += fmt.Sprintf("  %d. %s\n", i+1, s)
		}
	}

	// Auto-fix indicator
	if v.AutoFix != nil {
		details += "\n" + successStyle.Render("✓ Auto-fix available (press 'f' to preview)")
	} else {
		details += "\n" + infoStyle.Render("ℹ Manual fix required")
	}

	box := detailBoxStyle.Render(details)
	b.WriteString(box)

	// Help
	help := helpStyle.Render("\nEsc: Back | f: Fix | q: Quit")
	b.WriteString(help)

	return b.String()
}

// renderFixPreview renders the fix preview view
func (m Model) renderFixPreview() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Fix Preview")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.selectedFix == nil {
		b.WriteString(errorStyle.Render("No fix available"))
		return b.String()
	}

	// Fix details
	details := fmt.Sprintf("Description: %s\nConfidence:  %.0f%%\nSafe:        %v\n\nActions:\n",
		m.selectedFix.Description,
		m.selectedFix.Confidence*100,
		m.selectedFix.Safe,
	)

	for i, action := range m.selectedFix.Actions {
		details += fmt.Sprintf("  %d. %s\n", i+1, action.Describe())
	}

	box := detailBoxStyle.Render(details)
	b.WriteString(box)

	// Warning for unsafe fixes
	if !m.selectedFix.Safe {
		warning := errorStyle.Render("\n⚠ WARNING: This fix is marked as UNSAFE. Review carefully before applying.")
		b.WriteString(warning)
		b.WriteString("\n")
	}

	// Prompt
	prompt := "\n" + headerStyle.Render("Apply this fix?")
	b.WriteString(prompt)

	// Help
	help := helpStyle.Render("\ny/Enter: Apply | n/Esc: Cancel | q: Quit")
	b.WriteString(help)

	return b.String()
}

// renderGraph renders the dependency graph view
func (m Model) renderGraph() string {
	var b strings.Builder

	title := titleStyle.Render("Dependency Graph")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Placeholder for graph view
	b.WriteString(infoStyle.Render("📊 Dependency graph visualization coming soon..."))
	b.WriteString("\n\n")

	b.WriteString("This view will show:")
	b.WriteString("\n  • File dependencies")
	b.WriteString("\n  • Import relationships")
	b.WriteString("\n  • Circular dependency detection")
	b.WriteString("\n  • Layer violations")

	help := helpStyle.Render("\n\nEsc: Back | q: Quit")
	b.WriteString(help)

	return b.String()
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
