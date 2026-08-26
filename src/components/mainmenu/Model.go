package mainmenu

import (
	tea "charm.land/bubbletea/v2"

	"github.com/filipemolina/farol/src/apptypes"
	"github.com/filipemolina/farol/src/cmds"
)

// Model is the top menu bar. It is not focusable and handles no keys — page
// switching is a global digit key (keys.Global.PageActive/PageArchived/
// PageSearch) handled in AppModel, not a click target here — it renders the
// three page tabs (1 Active, 2 Archived, 3 Search), the wordmark, and the
// version (when it fits) on a tier-2 strip.
type Model struct {
	terminalWidth int
	// treeView is the task tree's current Pending/Complete/All view mode,
	// low-emphasis next to the version (mirrors ../pulso's mainmenu, which
	// renders resultslist's TableStateMsg the same way). Defaults to "all",
	// the tree's own default, so the header agrees with the tree before any
	// SetTaskTreeViewMsg has arrived.
	treeView string
	// page is the current top-level surface, broadcast by AppModel on every
	// transition (cmds.SetPageMsg). One value, not a pair of open flags: the
	// tabs are mutually exclusive on screen, and two independent booleans
	// desynced the highlight whenever a transition skipped the other page's
	// close message — 2 -> 3 lit both tabs, and the stale Archive flag kept
	// the tree's view-mode slot blanked after landing back on Active.
	page apptypes.Page
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// New builds the header component. The zero Page is PageActive, so the first
// frame highlights the Active tab before any broadcast arrives.
func New() tea.Model {
	return Model{treeView: "all"}
}

// Update tracks the terminal width from the broadcast layout, the task
// tree's view mode so the header stays in step with it, and the current page
// so the tabs always agree with AppModel's enum.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cmds.SetBodyLayoutMsg:
		m.terminalWidth = msg.TerminalWidth
	case cmds.SetTaskTreeViewMsg:
		m.treeView = msg.View
	case cmds.SetPageMsg:
		m.page = msg.Page
	}
	return m, nil
}
