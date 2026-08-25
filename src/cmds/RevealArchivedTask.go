package cmds

import tea "charm.land/bubbletea/v2"

// RevealArchivedTaskMsg asks the Archive page to select the archived list
// ListID (scrolling it into view if needed) and, once that list's preview has
// loaded, to scroll to and highlight the task TaskID so its position within a
// long list is obvious. It is the Archive page's half of an archived search
// result's Enter: AppModel opens the page (OpenArchivePageMsg) and then hands
// it the reveal target.
type RevealArchivedTaskMsg struct {
	TaskID string
	ListID string
}

// RevealArchivedTask returns a command that tells the Archive page to
// reveal the given archived list and task.
func RevealArchivedTask(taskID, listID string) tea.Cmd {
	return func() tea.Msg { return RevealArchivedTaskMsg{TaskID: taskID, ListID: listID} }
}
