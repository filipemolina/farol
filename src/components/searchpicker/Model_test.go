package searchpicker

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/filipemolina/farol/src/cmds"
)

// typeQuery drives the picker's Update with a printable keystroke per rune,
// the way the real key loop delivers typed text (bubbles v2 inserts from
// KeyPressMsg.Text, so a synthetic keypress must carry it).
func typeQuery(t *testing.T, m Model, query string) Model {
	t.Helper()
	for _, r := range query {
		out, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = out.(Model)
	}
	return m
}

// pressKey sends one non-printable keypress (arrows, enter, esc).
func pressKey(t *testing.T, m Model, code rune) Model {
	t.Helper()
	out, _ := m.Update(tea.KeyPressMsg{Code: code})
	return out.(Model)
}

// TestSubmitArchivedResultRoutesToArchivePage proves Enter on a result whose
// list is archived hands off to the Archive page (CloseModal's follow is a
// cmds.OpenArchivedTaskMsg) rather than the JumpToTask that would try to make
// an archived list the active list — the "loads then reverts" bug.
func TestSubmitArchivedResultRoutesToArchivePage(t *testing.T) {
	s := testStore(t)
	lid, err := s.CreateList("Farol v0.4", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	if _, err := s.CreateTask(lid, "Tag a list as collaborative", nil, ""); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := s.ArchiveList(lid); err != nil {
		t.Fatalf("archive list: %v", err)
	}

	pm := New(s, 100, 30).(Model)
	pm = typeQuery(t, pm, "collaborative")
	if len(pm.results) == 0 {
		t.Fatal("no results for query")
	}
	if !pm.results[0].Archived {
		t.Fatal("precondition: result must be flagged Archived")
	}

	out, cmd := pm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_ = out
	if cmd == nil {
		t.Fatal("Enter produced no command")
	}
	msg, ok := cmd().(cmds.CloseModalMsg)
	if !ok {
		t.Fatalf("Enter command produced %T, want cmds.CloseModalMsg", cmd())
	}
	if msg.Follow == nil {
		t.Fatal("CloseModal has no follow-up")
	}
	if _, ok := msg.Follow().(cmds.OpenArchivedTaskMsg); !ok {
		t.Fatalf("CloseModal follow = %T, want cmds.OpenArchivedTaskMsg (archived result must not jump the active list)", msg.Follow())
	}
}

// TestSubmitActiveResultStillJumpsToTask proves the non-archived path is
// unchanged: Enter on an active-list result still hands off to JumpToTask.
func TestSubmitActiveResultStillJumpsToTask(t *testing.T) {
	s := testStore(t)
	lid, err := s.CreateList("Errands", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	if _, err := s.CreateTask(lid, "Buy milk", nil, ""); err != nil {
		t.Fatalf("create task: %v", err)
	}

	pm := New(s, 100, 30).(Model)
	pm = typeQuery(t, pm, "milk")
	if len(pm.results) == 0 {
		t.Fatal("no results")
	}
	if pm.results[0].Archived {
		t.Fatal("precondition: active-list result must not be Archived")
	}
	out, cmd := pm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_ = out
	msg, ok := cmd().(cmds.CloseModalMsg)
	if !ok {
		t.Fatalf("Enter command = %T, want cmds.CloseModalMsg", cmd())
	}
	if _, ok := msg.Follow().(cmds.JumpToTaskMsg); !ok {
		t.Fatalf("CloseModal follow = %T, want cmds.JumpToTaskMsg for an active-list result", msg.Follow())
	}
}

// seedTasks creates one list and n distinctly-titled tasks in it.
func seedTasks(t *testing.T, s interface {
	CreateList(name, description string) (string, error)
	CreateTask(listID, title string, parentID *string, notes string) (string, error)
}, listName string, titles []string) {
	t.Helper()
	lid, err := s.CreateList(listName, "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	for _, title := range titles {
		if _, err := s.CreateTask(lid, title, nil, ""); err != nil {
			t.Fatalf("create task %q: %v", title, err)
		}
	}
}

// The result area is a fixed-height reservation: exactly visible lines on
// every query state — empty, no matches, a few matches, more matches than
// fit. This is the regression test for the picker resizing itself on every
// keystroke.
func TestResultLinesFixedHeight(t *testing.T) {
	s := testStore(t)
	titles := []string{"buy milk", "walk dog", "write report", "clean garage"}
	for i := 0; i < 40; i++ {
		titles = append(titles, strings.ToLower(strings.Repeat("item", 1)+strings.Repeat("x", i)))
	}
	seedTasks(t, s, "Errands", titles)

	const termHeight = 30

	for _, tc := range []struct {
		name  string
		query string
	}{
		{"empty query", ""},
		{"no matches", "zzzz"},
		{"few matches", "milk"},
		{"more matches than fit", "itemx"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pm := New(s, 100, termHeight).(Model)
			pm = typeQuery(t, pm, tc.query)
			lines := pm.resultLines()
			if len(lines) != pm.visible {
				t.Fatalf("resultLines returned %d lines, want exactly visible=%d", len(lines), pm.visible)
			}
		})
	}
}

// The whole modal box — height AND width — must be identical across query
// states. Width is pinned by truncating every row to rowWidth; height by the
// fixed line count. Either axis moving per keystroke reads as a different
// modal.
func TestViewBoxStableAcrossQueries(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{
		"buy milk",
		"a very long task title that would stretch the modal far beyond its hints line if it were not truncated to the fixed row width",
		"walk dog",
	})

	boxOf := func(content string) (lines, width int) {
		for _, l := range strings.Split(content, "\n") {
			lines++
			if w := ansi.StringWidth(l); w > width {
				width = w
			}
		}
		return lines, width
	}

	var wantLines, wantWidth = -1, -1
	for _, query := range []string{"", "b", "bu", "buy m", "a very long", "zzzz"} {
		pm := New(s, 100, 30).(Model)
		pm = typeQuery(t, pm, query)
		gotLines, gotWidth := boxOf(pm.View().Content)
		if wantLines == -1 {
			wantLines, wantWidth = gotLines, gotWidth
			continue
		}
		if gotLines != wantLines || gotWidth != wantWidth {
			t.Errorf("query %q renders %dx%d, want the stable %dx%d", query, gotLines, gotWidth, wantLines, wantWidth)
		}
	}
}

// Up/down move the cursor AND the movement must be visible: exactly one
// result row carries the bold cursor treatment, and pressing down moves that
// treatment to the next row. The first version of this component moved the
// cursor internally while rendering selected and unselected rows identically,
// which read as dead keys.
func TestCursorMovesAndIsVisible(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"buy milk", "walk dog", "write report"})

	pm := New(s, 100, 30).(Model)
	pm = typeQuery(t, pm, "w") // matches walk dog + write report
	if len(pm.results) < 2 {
		t.Fatalf("seed: %d results for 'w', want at least 2", len(pm.results))
	}

	boldRows := func(m Model) int {
		n := 0
		for _, line := range m.resultLines() {
			// lipgloss merges bold into one SGR ("\x1b[1;38;2…"), so match
			// the leading "1;" rather than a standalone "\x1b[1m".
			if strings.Contains(line, "\x1b[1;") {
				n++
			}
		}
		return n
	}

	if n := boldRows(pm); n != 1 {
		t.Fatalf("cursor row bold count = %d, want exactly 1", n)
	}

	before := pm.resultLines()
	pm = pressKey(t, pm, tea.KeyDown)
	if pm.cursor != 1 {
		t.Errorf("cursor = %d after down, want 1", pm.cursor)
	}
	after := pm.resultLines()
	if n := boldRows(pm); n != 1 {
		t.Errorf("bold count after down = %d, want exactly 1", n)
	}
	if rowsEqual(before, after) {
		t.Error("pressing down rendered identical frames — the cursor moved but nothing on screen shows it")
	}

	pm = pressKey(t, pm, tea.KeyUp)
	if pm.cursor != 0 {
		t.Errorf("cursor = %d after up, want 0", pm.cursor)
	}
	if !rowsEqual(before, pm.resultLines()) {
		t.Error("up did not restore the original frame")
	}

	// The cursor never leaves the list.
	for i := 0; i < 10; i++ {
		pm = pressKey(t, pm, tea.KeyDown)
	}
	if pm.cursor != len(pm.results)-1 {
		t.Errorf("cursor = %d after overscrolling down, want clamped to %d", pm.cursor, len(pm.results)-1)
	}
}

func rowsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
