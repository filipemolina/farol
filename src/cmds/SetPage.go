package cmds

import (
	tea "charm.land/bubbletea/v2"

	"github.com/filipemolina/farol/src/apptypes"
)

// SetPageMsg tells the header which top-level page is on screen. AppModel
// owns the page enum; the header only displays it. One broadcast per
// transition replaces the open/close message pairs whose independent flags
// desynced the tab highlight whenever a transition skipped the other page's
// close message — 2 -> 3 lit both tabs, and the stale flag kept the tree's
// view-mode slot blanked after landing back on Active.
type SetPageMsg struct {
	Page apptypes.Page
}

// SetPage broadcasts the current top-level page to the components.
func SetPage(p apptypes.Page) tea.Cmd {
	return func() tea.Msg { return SetPageMsg{Page: p} }
}
