package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/filipemolina/farol/src/apptypes"
	"github.com/filipemolina/farol/src/cmds"
	"github.com/filipemolina/farol/src/constants"
)

// TestPickerKeyOpensSearchPage proves F is wired in AppModel.Update: it flips
// the page enum to Search and moves focus onto the page (docs/DESIGN.md §5).
func TestPickerKeyOpensSearchPage(t *testing.T) {
	m := seedOneList(t)
	m = refresh(t, m, tea.WindowSizeMsg{Width: 100, Height: 40})

	m = refresh(t, m, tea.KeyPressMsg{Text: "F"})

	if !m.searchPageVisible() {
		t.Fatal("F did not open the Search page")
	}
	if m.focusedZone != constants.COMPONENT_SEARCH_PAGE {
		t.Fatalf("focusedZone = %d, want COMPONENT_SEARCH_PAGE", m.focusedZone)
	}
}

// TestDigitThreeOpensSearchPage proves 3 (keys.Global.PageSearch) opens the
// Search page from Active — the tab pattern the header advertises.
func TestDigitThreeOpensSearchPage(t *testing.T) {
	m := seedOneList(t)
	m = refresh(t, m, tea.WindowSizeMsg{Width: 100, Height: 40})

	m = refresh(t, m, tea.KeyPressMsg{Text: "3"})

	if !m.searchPageVisible() {
		t.Fatal("3 did not open the Search page")
	}
	if m.focusedZone != constants.COMPONENT_SEARCH_PAGE {
		t.Fatalf("focusedZone = %d, want COMPONENT_SEARCH_PAGE", m.focusedZone)
	}
}

// TestEscOnSearchPageReturnsToActive pins the esc contract: the page closes
// onto Active with focus back on the task tree — search is a visit, not a
// destination (docs/DESIGN.md §5).
func TestEscOnSearchPageReturnsToActive(t *testing.T) {
	m := seedOneList(t)
	m = refresh(t, m, tea.WindowSizeMsg{Width: 100, Height: 40})
	m = refresh(t, m, tea.KeyPressMsg{Text: "F"})

	m = refresh(t, m, tea.KeyPressMsg{Text: "esc"})

	if m.searchPageVisible() {
		t.Fatal("esc did not close the Search page")
	}
	if m.page != PageActive {
		t.Fatalf("page = %d after esc, want PageActive", m.page)
	}
	if m.focusedZone != constants.COMPONENT_TASK_TREE {
		t.Fatalf("focusedZone = %d, want COMPONENT_TASK_TREE", m.focusedZone)
	}
}

// TestPickerKeyWorksFromArchivePage proves F is live on the Archived page too
// — its capture owns every keypress, so the page's own handleKey matches F
// exactly the way it matches 1 (docs/DESIGN.md §5).
func TestPickerKeyWorksFromArchivePage(t *testing.T) {
	m := seedOneList(t)
	m = refresh(t, m, tea.WindowSizeMsg{Width: 100, Height: 40})
	m = refresh(t, m, tea.KeyPressMsg{Text: "2"})
	if !m.archivePageVisible() {
		t.Fatal("precondition: 2 did not open the Archive page")
	}

	m = refresh(t, m, tea.KeyPressMsg{Text: "F"})

	if !m.searchPageVisible() {
		t.Fatal("F did not open the Search page from the Archive page")
	}
}

// TestEnterOnSearchResultJumpsAndCloses proves the full Enter path at
// AppModel level: the page closes onto Active and the JumpToTask follow-up
// switches the active list — close first, then hand off.
func TestEnterOnSearchResultJumpsAndCloses(t *testing.T) {
	m := newTestModel(t, t.TempDir())
	lists, _ := m.store.ListLists()
	for _, l := range lists {
		m.store.DeleteList(l.ID)
	}
	listA, _ := m.store.CreateList("Alpha", "")
	listB, _ := m.store.CreateList("Beta", "")
	if _, err := m.store.CreateTask(listB, "quantum flute", nil, ""); err != nil {
		t.Fatalf("create task: %v", err)
	}
	m.store.CreateTask(listA, "An active task", nil, "")
	m = refresh(t, m, tea.WindowSizeMsg{Width: 100, Height: 40})
	m = refresh(t, m, cmds.RefreshLists(m.store)())
	m = refresh(t, m, cmds.RefreshTasks(m.store, listA, apptypes.SortManual)())

	m = refresh(t, m, tea.KeyPressMsg{Text: "F"})
	// Type the query one rune at a time, the way the key loop delivers it
	// (bubbles v2 inserts from KeyPressMsg.Text).
	for _, r := range "quantum" {
		m = refresh(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	out, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = out.(AppModel)
	if cmd == nil {
		t.Fatal("Enter produced no command")
	}
	// The page capture returns the page's own command directly — the same
	// early return the Archive page capture uses — so the close arrives
	// unbatched.
	msg, ok := cmd().(cmds.CloseSearchPageMsg)
	if !ok {
		t.Fatalf("Enter command produced %T, want cmds.CloseSearchPageMsg", cmd())
	}
	if msg.Follow == nil {
		t.Fatal("close carried no JumpToTask follow-up")
	}
	follow := msg.Follow // the tea.Cmd handed off behind the close
	if _, ok := follow().(cmds.JumpToTaskMsg); !ok {
		t.Fatalf("follow = %T, want cmds.JumpToTaskMsg", follow())
	}
	// The runtime resolves the close first (flipping the page), then runs the
	// follow-up — mirror that order here.
	m = refresh(t, m, msg)
	m = refresh(t, m, follow())
	if m.searchPageVisible() || m.page != PageActive {
		t.Fatalf("page = %d after Enter, want PageActive", m.page)
	}
	if m.activeListID != listB {
		t.Errorf("activeListID = %q, want %q (Enter must jump to the result's list)", m.activeListID, listB)
	}
}
