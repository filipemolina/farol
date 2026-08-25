package cmds

import tea "charm.land/bubbletea/v2"

// OpenSearchPageMsg asks AppModel to switch the current page to the
// cross-list Search page (the F key, docs/DESIGN.md §5). It is a page switch,
// not a modal open: the Search page component is constructed once and kept in
// AppModel's component set, so this message only flips the page enum — it does
// not build a new surface.
type OpenSearchPageMsg struct{}

// OpenSearchPage switches the app to the Search page.
func OpenSearchPage() tea.Cmd {
	return func() tea.Msg { return OpenSearchPageMsg{} }
}
