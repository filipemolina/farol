package searchpage

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/filipemolina/farol/src/cmds"
	"github.com/filipemolina/farol/src/constants"
	"github.com/filipemolina/farol/src/store"
)

// focusedPage builds a page over the given store, sized like a small terminal
// and focused — the two broadcasts AppModel issues before any keystroke can
// reach it (the page ignores KeyPressMsg while unfocused, so a test that
// skips this is testing nothing).
func focusedPage(t *testing.T, s *store.Store) Model {
	t.Helper()
	return focusAndSize(t, New(s).(Model))
}

func focusAndSize(t *testing.T, m Model) Model {
	t.Helper()
	out, _ := m.Update(cmds.SetBodyLayoutMsg{Height: 30, TerminalWidth: 100})
	m = out.(Model)
	out, _ = m.Update(cmds.SetFocusMsg(constants.COMPONENT_SEARCH_PAGE))
	return out.(Model)
}

// typeQuery drives the page's Update with a printable keystroke per rune,
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

// seedTasks creates one list and n distinctly-titled tasks in it.
func seedTasks(t *testing.T, s interface {
	CreateList(name, description string) (string, error)
	CreateTask(listID, title string, parentID *string, notes string) (string, error)
}, listName string, titles []string) string {
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
	return lid
}

// resultRows returns the rendered frame's lines that are result rows: they
// carry the "<list> › <title>" separator, which nothing else on the page
// (title chip, legend, empty-state card) contains.
func resultRows(m Model) []string {
	var rows []string
	for _, line := range strings.Split(m.View().Content, "\n") {
		if strings.Contains(ansi.Strip(line), "›") {
			rows = append(rows, line)
		}
	}
	return rows
}

// TestSubmitArchivedResultRoutesToArchivePage proves Enter on a result whose
// list is archived hands off to the Archive page (CloseSearchPage's follow is
// a cmds.OpenArchivedTaskMsg) rather than the JumpToTask that would try to
// make an archived list the active list.
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

	m := focusedPage(t, s)
	m = typeQuery(t, m, "collaborative")
	if len(m.results) == 0 {
		t.Fatal("no results for query")
	}
	if !m.results[0].Archived {
		t.Fatal("precondition: result must be flagged Archived")
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter produced no command")
	}
	msg, ok := cmd().(cmds.CloseSearchPageMsg)
	if !ok {
		t.Fatalf("Enter command produced %T, want cmds.CloseSearchPageMsg", cmd())
	}
	if msg.Follow == nil {
		t.Fatal("CloseSearchPage has no follow-up")
	}
	if _, ok := msg.Follow().(cmds.OpenArchivedTaskMsg); !ok {
		t.Fatalf("follow = %T, want cmds.OpenArchivedTaskMsg (archived result must not jump the active list)", msg.Follow())
	}
}

// TestSubmitActiveResultStillJumpsToTask proves the non-archived path is
// unchanged: Enter on an active-list result hands off to JumpToTask.
func TestSubmitActiveResultStillJumpsToTask(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"Buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk")
	if len(m.results) == 0 {
		t.Fatal("no results")
	}
	if m.results[0].Archived {
		t.Fatal("precondition: active-list result must not be Archived")
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	msg, ok := cmd().(cmds.CloseSearchPageMsg)
	if !ok {
		t.Fatalf("Enter command = %T, want cmds.CloseSearchPageMsg", cmd())
	}
	if _, ok := msg.Follow().(cmds.JumpToTaskMsg); !ok {
		t.Fatalf("follow = %T, want cmds.JumpToTaskMsg for an active-list result", msg.Follow())
	}
}

// TestSubmitWithNoResultsIsInert proves Enter on an empty result set neither
// closes the page nor emits a follow-up — there is nothing to open.
func TestSubmitWithNoResultsIsInert(t *testing.T) {
	s := testStore(t)
	m := focusedPage(t, s)
	m = typeQuery(t, m, "zzzz")
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("Enter with no results produced %v, want nil", cmd())
	}
}

// TestEscClosesOntoActiveWithNoFollow pins the locked decision: esc lands on
// the Active page (AppModel's side of CloseSearchPageMsg), and carries no
// follow-up — search is a visit, not a destination.
func TestEscClosesOntoActiveWithNoFollow(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"Buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk")

	_, escCmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if escCmd == nil {
		t.Fatal("esc produced no command")
	}
	msg, ok := escCmd().(cmds.CloseSearchPageMsg)
	if !ok {
		t.Fatalf("esc command = %T, want cmds.CloseSearchPageMsg", escCmd())
	}
	if msg.Follow != nil {
		t.Fatalf("esc follow = %v, want nil (esc must not jump anywhere)", msg.Follow)
	}
}

// TestDigitTwoLeavesForArchivePage proves 2 works from inside Search — the
// digits are a second way off the page, and they win over the input because
// the page matches them ahead of the default printable branch.
func TestDigitTwoLeavesForArchivePage(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"Buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk") // input non-empty: the digit still leaves

	_, cmd := m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	if cmd == nil {
		t.Fatal("2 produced no command")
	}
	if _, ok := cmd().(cmds.OpenArchivePageMsg); !ok {
		t.Fatalf("2 command = %T, want cmds.OpenArchivePageMsg", cmd())
	}
}

// TestDigitThreeOnSearchPageIsInert proves 3 on the Search page itself is an
// idempotent no-op that still consumes the keystroke — the page stays up and
// the digit never reaches the query input (docs/DESIGN.md §5).
func TestDigitThreeOnSearchPageIsInert(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"Buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk")

	out, cmd := m.Update(tea.KeyPressMsg{Code: '3', Text: "3"})
	m = out.(Model)
	if cmd != nil {
		t.Fatalf("3 on the Search page produced %v, want nil", cmd())
	}
	if m.input.Value() != "milk" {
		t.Errorf("query = %q, want unchanged (digits are never query characters)", m.input.Value())
	}
}

// TestCursorMovesAndIsVisible adapts the modal-era regression: up/down move
// the cursor AND the movement must be visible — exactly one result row
// carries the bold selected treatment, and pressing down moves that treatment
// to the next row.
func TestCursorMovesAndIsVisible(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"buy milk", "walk dog", "write report"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "w") // matches walk dog + write report
	if len(m.results) < 2 {
		t.Fatalf("seed: %d results for 'w', want at least 2", len(m.results))
	}

	boldRows := func(rows []string) int {
		n := 0
		for _, line := range rows {
			// lipgloss merges bold into one SGR ("\x1b[1;38;2…"), so match
			// the leading "1;" rather than a standalone "\x1b[1m".
			if strings.Contains(line, "\x1b[1;") {
				n++
			}
		}
		return n
	}

	before := resultRows(m)
	if n := boldRows(before); n != 1 {
		t.Fatalf("cursor row bold count = %d, want exactly 1", n)
	}

	m = pressKey(t, m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("cursor = %d after down, want 1", m.cursor)
	}
	after := resultRows(m)
	if n := boldRows(after); n != 1 {
		t.Errorf("bold count after down = %d, want exactly 1", n)
	}
	if rowsEqual(before, after) {
		t.Error("pressing down rendered identical frames — the cursor moved but nothing on screen shows it")
	}

	m = pressKey(t, m, tea.KeyUp)
	if m.cursor != 0 {
		t.Errorf("cursor = %d after up, want 0", m.cursor)
	}
	if !rowsEqual(before, resultRows(m)) {
		t.Error("up did not restore the original frame")
	}

	// The cursor never leaves the list.
	for i := 0; i < 10; i++ {
		m = pressKey(t, m, tea.KeyDown)
	}
	if m.cursor != len(m.results)-1 {
		t.Errorf("cursor = %d after overscrolling down, want clamped to %d", m.cursor, len(m.results)-1)
	}
}

// TestKeystrokesIgnoredWhileUnfocused proves the page only answers to keys
// while AppModel has focused it — a stray keystroke routed to the component
// set must not edit the persisted query.
func TestKeystrokesIgnoredWhileUnfocused(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"Buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk")

	out, _ := m.Update(cmds.SetFocusMsg(constants.COMPONENT_TASK_TREE))
	m = out.(Model)
	out, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m = out.(Model)

	if m.input.Value() != "milk" {
		t.Errorf("query = %q after an unfocused keystroke, want unchanged %q", m.input.Value(), "milk")
	}
}

// TestQueryResultsCursorPersistAcrossFocusLoss pins the page-form decision
// the modal never made: leaving and re-entering shows the search as you left
// it (docs/DESIGN.md §5).
func TestQueryResultsCursorPersistAcrossFocusLoss(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"buy milk", "walk dog", "write report"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "w")
	m = pressKey(t, m, tea.KeyDown)
	wantResults, wantCursor := len(m.results), m.cursor

	out, _ := m.Update(cmds.SetFocusMsg(constants.COMPONENT_TASK_TREE))
	m = out.(Model)
	out, _ = m.Update(cmds.SetFocusMsg(constants.COMPONENT_SEARCH_PAGE))
	m = out.(Model)

	if m.input.Value() != "w" {
		t.Errorf("query = %q after refocus, want %q", m.input.Value(), "w")
	}
	if len(m.results) != wantResults || m.cursor != wantCursor {
		t.Errorf("results/cursor = %d/%d after refocus, want %d/%d", len(m.results), m.cursor, wantResults, wantCursor)
	}
}

// TestRefreshReRunsPersistedQuery pins the staleness rule: a poll tick while
// the page is open re-runs the non-empty query, so a task created elsewhere
// shows up behind an already-open search (docs/DESIGN.md §5).
func TestRefreshReRunsPersistedQuery(t *testing.T) {
	s := testStore(t)
	lid := seedTasks(t, s, "Errands", []string{"Buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk")
	before := len(m.results)
	if before == 0 {
		t.Fatal("seed: no results for 'milk'")
	}

	if _, err := s.CreateTask(lid, "Milk carton", nil, ""); err != nil {
		t.Fatalf("create task: %v", err)
	}
	out, _ := m.Update(cmds.RefreshSearchMsg{})
	m = out.(Model)

	if len(m.results) != before+1 {
		t.Errorf("results after refresh = %d, want %d (the new matching task)", len(m.results), before+1)
	}
}

// TestResizeReflowsFromBroadcast proves sizing comes from the layout
// broadcast, not a snapshot taken at construction: a wider broadcast renders
// a wider page (docs/DESIGN.md §5 — the modal's fixed-at-open sizing died
// with the modal).
func TestResizeReflowsFromBroadcast(t *testing.T) {
	s := testStore(t)
	seedTasks(t, s, "Errands", []string{"buy milk"})

	m := focusedPage(t, s)
	m = typeQuery(t, m, "milk")

	maxWidth := func(mm Model) int {
		w := 0
		for _, line := range strings.Split(mm.View().Content, "\n") {
			if lw := ansi.StringWidth(line); lw > w {
				w = lw
			}
		}
		return w
	}

	narrow := maxWidth(setSize(t, m, 60))
	wide := maxWidth(setSize(t, m, 120))
	if narrow >= wide {
		t.Errorf("page width did not follow the broadcast: 60-col terminal rendered %d wide, 120-col %d wide", narrow, wide)
	}
}

func setSize(t *testing.T, m Model, width int) Model {
	t.Helper()
	out, _ := m.Update(cmds.SetBodyLayoutMsg{Height: 30, TerminalWidth: width})
	return out.(Model)
}

// TestEmptyStatesRenderGuidance proves both empty conditions render their
// message: an untouched query guides, a fruitless query reports no matches
// (the recessed-card pattern, docs/DESIGN.md §12).
func TestEmptyStatesRenderGuidance(t *testing.T) {
	s := testStore(t)
	m := focusedPage(t, s)

	fresh := ansi.Strip(m.View().Content)
	if !strings.Contains(fresh, "Type to search across every list") {
		t.Errorf("empty page does not show guidance:\n%s", fresh)
	}

	m = typeQuery(t, m, "zzzznothing")
	barren := ansi.Strip(m.View().Content)
	if !strings.Contains(barren, "No matches") {
		t.Errorf("fruitless query does not report no matches:\n%s", barren)
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
