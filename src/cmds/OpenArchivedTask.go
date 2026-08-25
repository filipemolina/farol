package cmds

import tea "charm.land/bubbletea/v2"

// OpenArchivedTaskMsg asks AppModel to open the Archived Lists page, select
// the archived list that contains TaskID, and highlight that task in the
// preview. The search picker sends this instead of JumpToTask when the
// result's list is archived: an archived list cannot become the active list
// (Active vs Archived is a two-page split, docs/DESIGN.md §5), so "jump to
// it" means "show it on the Archive page", not "make it the active list".
type OpenArchivedTaskMsg struct {
	TaskID string
	ListID string
}

// OpenArchivedTask returns a command that opens the Archive page revealed on
// the given archived list and task.
func OpenArchivedTask(taskID, listID string) tea.Cmd {
	return func() tea.Msg { return OpenArchivedTaskMsg{TaskID: taskID, ListID: listID} }
}
