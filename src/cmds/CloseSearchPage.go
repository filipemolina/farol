package cmds

import tea "charm.land/bubbletea/v2"

// CloseSearchPageMsg asks AppModel to leave the Search page and return to the
// Active page (esc, or a digit that lands on Active). Only AppModel changes
// the page enum and focus, so the component never mutates the layout itself.
// Follow, if set, is appended to the batch of commands run once the page is
// gone (mirroring CloseArchivePageMsg) — the Search page uses it to hand off a
// JumpToTask / OpenArchivedTask after closing.
type CloseSearchPageMsg struct {
	Follow tea.Cmd
}

// CloseSearchPage leaves the Search page, running follow once it is gone (nil
// for a plain dismiss).
func CloseSearchPage(follow tea.Cmd) tea.Cmd {
	return func() tea.Msg { return CloseSearchPageMsg{Follow: follow} }
}
