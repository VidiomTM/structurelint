package tui

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/autofix"
	"github.com/Jonathangadeaharder/structurelint/internal/linter"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewModel(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test-rule", Path: "file.go", Message: "violation"},
	}
	m := NewModel(violations, true)
	assert.Equal(t, 1, len(m.violations))
	assert.True(t, m.fixableOnly)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, modeList, m.viewMode)
	assert.NotNil(t, m.fixEngine)
}

func TestNewModel_EmptyViolations(t *testing.T) {
	m := NewModel(nil, false)
	assert.Empty(t, m.violations)
}

func TestModel_Init(t *testing.T) {
	m := NewModel(nil, false)
	cmd := m.Init()
	assert.Nil(t, cmd)
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := NewModel(nil, false)
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	result, cmd := m.Update(msg)
	assert.Nil(t, cmd)

	updated := result.(Model)
	assert.Equal(t, 100, updated.width)
	assert.Equal(t, 50, updated.height)
}

func TestModel_Update_UnknownMsg(t *testing.T) {
	m := NewModel(nil, false)
	result, cmd := m.Update("unknown msg")
	assert.Nil(t, cmd)
	assert.Equal(t, m, result)
}

func TestModel_View_Quitting(t *testing.T) {
	m := NewModel(nil, false)
	m.quitting = true
	assert.Empty(t, m.View())
}

func TestModel_View_ListMode(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "rule-1", Path: "file.go", Message: "msg1"},
	}
	m := NewModel(violations, false)
	view := m.View()
	assert.Contains(t, view, "Structurelint - Interactive Mode")
	assert.Contains(t, view, "Found 1 violation")
}

func TestModel_View_ListMode_FixableOnly(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "rule-1", Path: "file.go", Message: "msg"},
	}
	m := NewModel(violations, true)
	view := m.View()
	assert.Contains(t, view, "(fixable only)")
}

func TestModel_View_DetailMode(t *testing.T) {
	violations := []linter.Violation{
		{
			Rule:     "test-rule",
			Path:     "test.go",
			Message:  "violation message",
			Expected: "expected value",
			Actual:   "actual value",
		},
	}
	m := NewModel(violations, false)
	m.viewMode = modeDetail
	view := m.View()
	assert.Contains(t, view, "Violation Details")
	assert.Contains(t, view, "expected value")
}

func TestModel_View_DetailMode_NoViolation(t *testing.T) {
	m := NewModel(nil, false)
	m.viewMode = modeDetail
	m.cursor = 5
	view := m.View()
	assert.Equal(t, "No violation selected", view)
}

func TestModel_View_DetailMode_WithSuggestions(t *testing.T) {
	violations := []linter.Violation{
		{
			Rule:        "test-rule",
			Path:        "test.go",
			Message:     "msg",
			Suggestions: []string{"fix by doing X", "or do Y"},
		},
	}
	m := NewModel(violations, false)
	m.viewMode = modeDetail
	view := m.View()
	assert.Contains(t, view, "fix by doing X")
}

func TestModel_View_DetailMode_WithContext(t *testing.T) {
	violations := []linter.Violation{
		{
			Rule:    "test",
			Path:    "test.go",
			Message: "msg",
			Context: "some context",
		},
	}
	m := NewModel(violations, false)
	m.viewMode = modeDetail
	view := m.View()
	assert.Contains(t, view, "some context")
}

func TestModel_View_DetailMode_AutoFixAvailable(t *testing.T) {
	violations := []linter.Violation{
		{
			Rule:    "test",
			Path:    "test.go",
			Message: "msg",
			AutoFix: &rules.AutoFix{FilePath: "fix.go", Content: "fixed"},
		},
	}
	m := NewModel(violations, false)
	m.viewMode = modeDetail
	view := m.View()
	assert.Contains(t, view, "Auto-fix available")
}

func TestModel_View_FixPreviewMode(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeFixPreview
	view := m.View()
	assert.Contains(t, view, "Fix Preview")
}

func TestModel_View_FixPreviewMode_WithFix(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeFixPreview
	m.selectedFix = &autofix.Fix{
		Description: "Fix test violation",
		Confidence:  0.9,
		Safe:        true,
		Actions:     []autofix.Action{},
	}
	view := m.View()
	assert.Contains(t, view, "90%")
	assert.Contains(t, view, "Apply this fix?")
}

func TestModel_View_FixPreviewMode_UnsafeFix(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeFixPreview
	m.selectedFix = &autofix.Fix{
		Description: "Unsafe fix",
		Confidence:  0.5,
		Safe:        false,
		Actions:     []autofix.Action{},
	}
	view := m.View()
	assert.Contains(t, view, "UNSAFE")
}

func TestModel_View_GraphMode(t *testing.T) {
	m := NewModel(nil, false)
	m.viewMode = modeGraph
	view := m.View()
	assert.Contains(t, view, "coming soon")
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hel...", truncate("hello world", 6))
	assert.Equal(t, "hello world", truncate("hello world", 20))
	assert.Equal(t, "tes...", truncate("test string longer", 6))
}

func TestHandleKeyPress_Quit(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeList

	m2 := NewModel(violations, false)
	m2.viewMode = modeDetail
	result2, _ := m2.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	assert.True(t, result2.(Model).quitting)
}

func TestHandleKeyPress_Navigate(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "a", Path: "a.go", Message: "a"},
		{Rule: "b", Path: "b.go", Message: "b"},
	}
	m := NewModel(violations, false)
	assert.Equal(t, 0, m.cursor)

	m2 := NewModel(violations, false)
	m2.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	m3 := NewModel(violations, false)
	m3.cursor = 1
	m3.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
}

func TestHandleKeyPress_EnterDetail(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, modeDetail, result.(Model).viewMode)
}

func TestHandleKeyPress_BackFromDetail(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeDetail
	m.statusMessage = "some msg"
	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, modeList, result.(Model).viewMode)
	assert.Empty(t, result.(Model).statusMessage)
}

func TestHandleKeyPress_CursorBounds(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "a", Path: "a.go", Message: "a"},
	}
	m := NewModel(violations, false)
	m.cursor = 0
	m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, m.cursor)

	m2 := NewModel(violations, false)
	m2.cursor = 0
	m2.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 0, m2.cursor)
}

func TestRenderList_WithTruncation(t *testing.T) {
	violations := make([]linter.Violation, 30)
	for i := 0; i < 30; i++ {
		violations[i] = linter.Violation{
			Rule: "rule", Path: "file.go", Message: "msg",
		}
	}
	m := NewModel(violations, false)
	m.cursor = 28
	view := m.renderList()
	assert.NotEmpty(t, view)
}

func TestHandleKeyPress_Fix(t *testing.T) {
	violations := []linter.Violation{
		{
			Rule:    "test",
			Path:    "test.go",
			Message: "msg",
		},
	}
	m := NewModel(violations, false)

	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	updated := result.(Model)
	assert.Equal(t, modeList, updated.viewMode)
}

func TestHandleKeyPress_FixWithAutoFix(t *testing.T) {
	violations := []linter.Violation{
		{
			Rule:    "test",
			Path:    "test.go",
			Message: "msg",
			AutoFix: &rules.AutoFix{FilePath: "fix.go", Content: "fixed"},
		},
	}
	m := NewModel(violations, false)
	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	updated := result.(Model)
	assert.Equal(t, modeFixPreview, updated.viewMode)
}

func TestHandleKeyPress_Graph(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	updated := result.(Model)
	assert.Equal(t, modeGraph, updated.viewMode)
}

func TestHandleFixPreviewKeys_ApplyFix(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeFixPreview
	m.selectedFix = &autofix.Fix{
		Description: "test fix",
		Actions:     []autofix.Action{},
	}

	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	updated := result.(Model)
	assert.Equal(t, modeList, updated.viewMode)
	assert.Nil(t, updated.selectedFix)
}

func TestHandleFixPreviewKeys_CancelFix(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeFixPreview
	m.selectedFix = &autofix.Fix{Description: "test fix"}

	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	updated := result.(Model)
	assert.Equal(t, modeList, updated.viewMode)
	assert.Contains(t, updated.statusMessage, "cancelled")
}

func TestHandleFixPreviewKeys_Esc(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "test", Path: "test.go", Message: "msg"},
	}
	m := NewModel(violations, false)
	m.viewMode = modeFixPreview
	m.selectedFix = &autofix.Fix{Description: "test"}

	result, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEsc})
	updated := result.(Model)
	assert.Equal(t, modeList, updated.viewMode)
}
