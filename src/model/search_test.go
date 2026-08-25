package model

import (
	"testing"

	"github.com/filipemolina/farol/src/apptypes"
	"github.com/filipemolina/farol/src/cmds"
)

// selectedID reports the task tree's current selection, which the picker's
// jump must move onto.
func selectedID(t *testing.T, m AppModel) string {
	t.Helper()
	tree, ok := m.components.TaskPanel.(interface{ SelectedID() string })
	if !ok {
		t.Fatalf("TaskPanel is %T, want selected-ID accessor", m.components.TaskPanel)
	}
	return tree.SelectedID()
}

// The global picker jumped to a task in another list: the active list must
// switch to the result's list and the tree's selection must land on that exact
// task. The jump handler switches the list synchronously and issues a refresh
// plus a select; those two commands may arrive in either order, and the
// selection must end on the target either way.
func TestJumpToTaskSwitchesListAndSelects(t *testing.T) {
	m := newTestModel(t, t.TempDir())
	// GetInitialModel creates a default list when the store is empty; remove it
	// so this test's first list is the one it creates below.
	lists, _ := m.store.ListLists()
	if len(lists) > 0 {
		m.store.DeleteList(lists[0].ID)
	}

	listA, err := m.store.CreateList("Alpha", "")
	if err != nil {
		t.Fatalf("create list A: %v", err)
	}
	listB, err := m.store.CreateList("Beta", "")
	if err != nil {
		t.Fatalf("create list B: %v", err)
	}
	if _, err := m.store.CreateTask(listA, "First", nil, ""); err != nil {
		t.Fatalf("create task A: %v", err)
	}
	target, err := m.store.CreateTask(listB, "Needle", nil, "")
	if err != nil {
		t.Fatalf("create task B: %v", err)
	}

	// Seed the app into list A as active with its tasks loaded: the first
	// lists refresh adopts the first list and requests its tasks.
	m = refresh(t, m, cmds.RefreshLists(m.store)())
	m = refresh(t, m, cmds.RefreshTasks(m.store, listA, apptypes.SortManual)())
	if m.activeListID != listA {
		t.Fatalf("seed: activeListID = %q, want %q", m.activeListID, listA)
	}

	// The picker's enter delivers the jump to AppModel, which switches the
	// active list synchronously.
	out, _ := m.Update(cmds.JumpToTaskMsg{TaskID: target, ListID: listB})
	m = out.(AppModel)
	if m.activeListID != listB {
		t.Errorf("activeListID = %q, want %q (switched to the result's list)", m.activeListID, listB)
	}

	// The jump queued a refresh and a select; drive them in both orders. The
	// selection must be the matched task either way.
	t.Run("select-then-refresh", func(t *testing.T) {
		// Select against the old (list A) rows: the target is not present, so
		// the tree must remember the request and honour it once list B loads.
		m2 := refresh(t, m, cmds.SelectTask(target)())
		m2 = refresh(t, m2, cmds.RefreshTasks(m2.store, listB, apptypes.SortManual)())
		if got := selectedID(t, m2); got != target {
			t.Errorf("selection = %q, want %q", got, target)
		}
	})

	t.Run("refresh-then-select", func(t *testing.T) {
		m2 := refresh(t, m, cmds.RefreshTasks(m.store, listB, apptypes.SortManual)())
		m2 = refresh(t, m2, cmds.SelectTask(target)())
		if got := selectedID(t, m2); got != target {
			t.Errorf("selection = %q, want %q", got, target)
		}
	})
}

// Regression: a poll refresh for the OLD list can race a picker jump. The poll
// tick snapshots activeListID when it fires; if that tick was queued just
// before the jump switched the active list, its RefreshTasksMsg lands AFTER the
// jump carrying the stale list's rows. The task tree adopts whatever list a
// refresh names, so before the fix that stale refresh reset the selection the
// jump had just landed on — "lands on the right task, then jumps to another
// one about a second later." AppModel now drops a refresh whose list is not the
// active list, so the stale rows never reach the tree and the jump survives.
func TestJumpToOtherListSurvivesStalePollRefresh(t *testing.T) {
	m := newTestModel(t, t.TempDir())
	lists, _ := m.store.ListLists()
	for _, l := range lists {
		m.store.DeleteList(l.ID)
	}

	listA, err := m.store.CreateList("Alpha", "")
	if err != nil {
		t.Fatalf("create list A: %v", err)
	}
	listB, err := m.store.CreateList("Beta", "")
	if err != nil {
		t.Fatalf("create list B: %v", err)
	}
	var listBTasks []string
	for _, title := range []string{"First", "Second", "Third", "Last"} {
		id, err := m.store.CreateTask(listB, title, nil, "")
		if err != nil {
			t.Fatalf("create task in B: %v", err)
		}
		listBTasks = append(listBTasks, id)
	}
	if _, err := m.store.CreateTask(listA, "A task", nil, ""); err != nil {
		t.Fatalf("create task in A: %v", err)
	}

	// Seed list A as active with its tasks loaded.
	m = refresh(t, m, cmds.RefreshLists(m.store)())
	m = refresh(t, m, cmds.RefreshTasks(m.store, listA, apptypes.SortManual)())
	if m.activeListID != listA {
		t.Fatalf("seed: activeListID = %q, want %q", m.activeListID, listA)
	}

	// Jump to "Third" in list B (a middle task, so a reset to a surviving row
	// would be visibly the wrong selection, not a clamp back to the target).
	target := listBTasks[2]
	m = refresh(t, m, cmds.JumpToTaskMsg{TaskID: target, ListID: listB})
	if got := selectedID(t, m); got != target {
		t.Fatalf("after jump: selected = %q, want target %q", got, target)
	}

	// A poll tick queued before the jump delivers the OLD list's refresh.
	m = refresh(t, m, cmds.RefreshTasks(m.store, listA, apptypes.SortManual)())
	if got := selectedID(t, m); got != target {
		t.Fatalf("after stale old-list refresh: selected = %q, want target %q (reset by a stale poll refresh)", got, target)
	}
	if m.activeListID != listB {
		t.Fatalf("after stale old-list refresh: activeListID = %q, want %q (stale refresh must not switch the active list back)", m.activeListID, listB)
	}
}
