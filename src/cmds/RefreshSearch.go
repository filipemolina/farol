package cmds

import tea "charm.land/bubbletea/v2"

// RefreshSearchMsg asks the Search page to re-run its persisted query against
// the store. AppModel issues it on every poll tick while the Search page is
// open (and once on open) so a non-empty query's results never go stale behind
// the cursor (docs/DESIGN.md §5). The component owns the query and the store
// handle, so the message carries no payload.
type RefreshSearchMsg struct{}

// RefreshSearch re-runs the Search page's current query.
func RefreshSearch() tea.Cmd {
	return func() tea.Msg { return RefreshSearchMsg{} }
}
