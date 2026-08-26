package model

import (
	"testing"

	"github.com/filipemolina/farol/src/apptypes"
	"github.com/filipemolina/farol/src/cmds"
	"github.com/filipemolina/farol/src/constants"
)

// OpenArchivedTaskMsg — The Search page's Enter on a result whose list is
// archived — must open the Archive page and, crucially, must NOT switch the
// active list to the archived one. Archive lists cannot be an active list, so
// the old JumpToTask path briefly loaded the archived list on the Tasks panel
// and then the next RefreshLists (which excludes archived lists) reverted it —
// the "loads then snaps back" symptom.
func TestOpenArchivedTaskMsgOpensArchivePageNotActiveList(t *testing.T) {
	m := newTestModel(t, t.TempDir())
	lists, _ := m.store.ListLists()
	for _, l := range lists {
		m.store.DeleteList(l.ID)
	}
	listA, _ := m.store.CreateList("Alpha", "")
	archived, _ := m.store.CreateList("Farol v0.4", "")
	if _, err := m.store.CreateTask(archived, "Tag a list as collaborative", nil, ""); err != nil {
		t.Fatalf("create archived task: %v", err)
	}
	m.store.CreateTask(listA, "An active task", nil, "")
	m = refresh(t, m, cmds.RefreshLists(m.store)())
	m = refresh(t, m, cmds.RefreshTasks(m.store, listA, apptypes.SortManual)())
	if m.activeListID != listA {
		t.Fatalf("seed: activeListID = %q", m.activeListID)
	}

	m = refresh(t, m, cmds.OpenArchivedTaskMsg{TaskID: "some-task", ListID: archived})

	if !m.archivePageVisible() {
		t.Fatal("OpenArchivedTaskMsg should open the Archive page")
	}
	if m.focusedZone != constants.COMPONENT_ARCHIVE_PAGE {
		t.Fatalf("focusedZone = %d, want Archive page (%d)", m.focusedZone, constants.COMPONENT_ARCHIVE_PAGE)
	}
	if m.activeListID != listA {
		t.Errorf("activeListID = %q, want %q (archived results must not switch the active list)", m.activeListID, listA)
	}
}
