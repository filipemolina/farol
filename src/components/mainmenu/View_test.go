package mainmenu

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/filipemolina/farol/src/apptypes"
	"github.com/filipemolina/farol/src/cmds"
)

func TestHeaderRendersWordmark(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})

	out := updated.(Model).View().Content
	if !strings.Contains(out, "farol") {
		t.Errorf("header output does not contain wordmark:\n%s", out)
	}
}

// TestHeaderRendersAllThreeTabs proves the header always shows the three
// page tabs, digit-prefixed (1 Active, 2 Archived, 3 Search) — the primary
// navigation, never dropped for width the way the mode/version are
// (docs/DESIGN.md §5).
func TestHeaderRendersAllThreeTabs(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})

	out := ansi.Strip(updated.(Model).View().Content)
	if !strings.Contains(out, "1 Active") {
		t.Errorf("header does not render the Active tab:\n%s", out)
	}
	if !strings.Contains(out, "2 Archived") {
		t.Errorf("header does not render the Archived tab:\n%s", out)
	}
	if !strings.Contains(out, "3 Search") {
		t.Errorf("header does not render the Search tab:\n%s", out)
	}
}

// TestHeaderHighlightsActiveTabByDefault and
// TestHeaderHighlightsArchivedTabWhilePageIsOpen pin which tab carries the
// accent "▌" bar — the selected-tab cue ../cais's mainmenu uses — before and
// after Open/CloseArchivePageMsg.
func TestHeaderHighlightsActiveTabByDefault(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})

	// The selected tab is preceded by the accent "▌" bar (activeCellStyle's
	// one-column left padding, unlike the unselected cell's two) — stripping
	// ANSI first collapses each tab's separately-styled digit/label render
	// calls into one contiguous run, so this substring check is reliable.
	out := ansi.Strip(updated.(Model).View().Content)
	if !strings.Contains(out, "▌ 1 Active") {
		t.Errorf("the Active tab is not highlighted by default:\n%s", out)
	}
	if strings.Contains(out, "▌ 2 Archived") {
		t.Errorf("the Archived tab is highlighted by default, want Active:\n%s", out)
	}
}

func TestHeaderHighlightsArchivedTabWhilePageIsOpen(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})
	updated, _ = updated.(Model).Update(cmds.SetPageMsg{Page: apptypes.PageArchived})

	out := ansi.Strip(updated.(Model).View().Content)
	if !strings.Contains(out, "▌ 2 Archived") {
		t.Errorf("the Archived tab is not highlighted while the Archive page is open:\n%s", out)
	}
	if strings.Contains(out, "▌ 1 Active") {
		t.Errorf("the Active tab is still highlighted while the Archive page is open:\n%s", out)
	}
}

// TestHeaderHighlightsSearchTabWhilePageIsOpen pins the third tab's accent
// bar: while the Search page is open neither of the other two tabs may claim
// the highlight (docs/DESIGN.md §5).
func TestHeaderHighlightsSearchTabWhilePageIsOpen(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})
	updated, _ = updated.(Model).Update(cmds.SetPageMsg{Page: apptypes.PageSearch})

	out := ansi.Strip(updated.(Model).View().Content)
	if !strings.Contains(out, "▌ 3 Search") {
		t.Errorf("the Search tab is not highlighted while the Search page is open:\n%s", out)
	}
	if strings.Contains(out, "▌ 1 Active") || strings.Contains(out, "▌ 2 Archived") {
		t.Errorf("another tab is still highlighted while the Search page is open:\n%s", out)
	}
}

// TestHeaderTabsStayMutuallyExclusive is the regression for the desynced
// open/close flags: walking Archived -> Search -> Active must leave exactly
// one tab highlighted at every step, always the page on screen. The old
// two-bool tracking lit 2 and 3 together after 2 -> 3, then kept 2 lit on a
// later return to Active.
func TestHeaderTabsStayMutuallyExclusive(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})

	steps := []struct {
		page    apptypes.Page
		wantTab string
		notTabs []string
	}{
		{apptypes.PageArchived, "▌ 2 Archived", []string{"▌ 1 Active", "▌ 3 Search"}},
		{apptypes.PageSearch, "▌ 3 Search", []string{"▌ 1 Active", "▌ 2 Archived"}},
		{apptypes.PageActive, "▌ 1 Active", []string{"▌ 2 Archived", "▌ 3 Search"}},
	}
	for _, s := range steps {
		updated, _ = updated.(Model).Update(cmds.SetPageMsg{Page: s.page})
		out := ansi.Strip(updated.(Model).View().Content)
		if !strings.Contains(out, s.wantTab) {
			t.Errorf("page %d: %q is not highlighted:\n%s", s.page, s.wantTab, out)
		}
		for _, banned := range s.notTabs {
			if strings.Contains(out, banned) {
				t.Errorf("page %d: stale highlight %q still showing:\n%s", s.page, banned, out)
			}
		}
	}
}

// TestHeaderKeepsModeIndicatorAcrossModes is the regression for the
// three-tab budget: pending/complete used to shed at everyday widths while
// "all" (shorter) survived, reading as the indicator vanishing whenever the
// filter left all.
func TestHeaderKeepsModeIndicatorAcrossModes(t *testing.T) {
	for _, mode := range []string{"all", "pending", "complete"} {
		m := New()
		updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 72})
		updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: mode})

		out := ansi.Strip(updated.(Model).View().Content)
		if !strings.Contains(out, "⇅ "+mode) {
			t.Errorf("view mode %q missing from the header at 72 columns:\n%s", mode, out)
		}
	}
}

// TestHeaderRightClusterNeverDanglesSeparator pins the separator rule: " . "
// joins two present pieces and nothing else — no line ends in a bare dot,
// and an unstamped ("unknown") build renders no version piece at all.
func TestHeaderRightClusterNeverDanglesSeparator(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 100})
	updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: "all"})

	out := ansi.Strip(updated.(Model).View().Content)
	if strings.Contains(out, "unknown") {
		t.Errorf("unstamped version leaked into the header:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasSuffix(strings.TrimRight(line, " "), ".") {
			t.Errorf("header line ends in a dangling separator:\n%s", line)
		}
	}
}

// TestHeaderDefaultsToAllView pins the header's view-mode indicator to
// "all" before any SetTaskTreeViewMsg arrives, matching the tree's own
// ViewAll default (docs/DESIGN.md §6) — the header must never claim a mode
// the tree has not actually selected.
func TestHeaderDefaultsToAllView(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})

	out := updated.(Model).View().Content
	if !strings.Contains(out, "all") {
		t.Errorf("header output does not contain default view mode \"all\":\n%s", out)
	}
}

// TestHeaderTracksTaskTreeView proves the header re-renders the tree's
// reported mode, the only wiring that keeps the two in step (mainmenu.Model
// has no other way to learn the tree's view).
func TestHeaderTracksTaskTreeView(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})
	updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: "pending"})

	out := updated.(Model).View().Content
	if !strings.Contains(out, "pending") {
		t.Errorf("header output does not contain updated view mode \"pending\":\n%s", out)
	}
}

// TestHeaderShedsViewModeBeforeWordmark pins the terminal-width guard's
// shedding priority: narrowing from 80 columns, the view-mode indicator is
// the first thing to disappear, and the wordmark — the header's one
// non-negotiable element — is still standing once it does. The mode is the
// tree's transient state; the wordmark is the header's identity.
func TestHeaderShedsViewModeBeforeWordmark(t *testing.T) {
	m := New()

	var droppedAt int
	for w := 80; w > 0; w-- {
		updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: w})
		updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: "pending"})
		out := updated.(Model).View().Content
		if !strings.Contains(out, "pending") {
			droppedAt = w
			break
		}
	}
	if droppedAt == 0 {
		t.Fatal("narrowing to 1 column never dropped the view mode")
	}

	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: droppedAt})
	updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: "pending"})
	out := updated.(Model).View().Content
	if !strings.Contains(out, "farol") {
		t.Errorf("wordmark missing at width %d, the width the view mode first dropped at:\n%s", droppedAt, out)
	}
}

// TestHeaderBlanksViewModeWhilePageIsOpen proves the header drops the task
// tree's view-mode indicator while the Archive page is open — that mode
// describes a surface the page has replaced, so it must not still be shown.
// Which page is on screen is now carried by the tabs (TestHeaderHighlights*
// above), not by a text label taking over this slot.
func TestHeaderBlanksViewModeWhilePageIsOpen(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})
	updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: "pending"})
	updated, _ = updated.(Model).Update(cmds.SetPageMsg{Page: apptypes.PageArchived})

	out := updated.(Model).View().Content
	if strings.Contains(out, "pending") {
		t.Errorf("header still shows the stale task-tree view mode while the Archive page is open:\n%s", out)
	}
}

// TestHeaderRestoresTreeViewAfterArchivePageCloses proves closing the Archive
// page hands the mode slot back to the task tree's own view indicator.
func TestHeaderRestoresTreeViewAfterArchivePageCloses(t *testing.T) {
	m := New()
	updated, _ := m.(Model).Update(cmds.SetBodyLayoutMsg{TerminalWidth: 80})
	updated, _ = updated.(Model).Update(cmds.SetTaskTreeViewMsg{View: "pending"})
	updated, _ = updated.(Model).Update(cmds.SetPageMsg{Page: apptypes.PageArchived})
	updated, _ = updated.(Model).Update(cmds.SetPageMsg{Page: apptypes.PageActive})

	out := updated.(Model).View().Content
	if !strings.Contains(out, "pending") {
		t.Errorf("header did not restore the task-tree view mode after the Archive page closed:\n%s", out)
	}
}
