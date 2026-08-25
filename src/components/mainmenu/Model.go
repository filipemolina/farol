package mainmenu

import (
	tea "charm.land/bubbletea/v2"
	"github.com/filipemolina/farol/src/cmds"
)

// Model is the top menu bar. It is not focusable and handles no keys — page
// switching is a global digit key (keys.Global.PageActive/PageArchived)
// handled in AppModel, not a click target here — it renders two page tabs
// (1 Active, 2 Archived), the wordmark, and the version (when it fits) on a
// tier-2 strip.
type Model struct {
	terminalWidth int
	// treeView is the task tree's current Pending/Complete/All view mode,
	// low-emphasis next to the version (mirrors ../pulso's mainmenu, which
	// renders resultslist's TableStateMsg the same way). Defaults to "all",
	// the tree's own default, so the header agrees with the tree before any
	// SetTaskTreeViewMsg has arrived.
	treeView string
	// archiveOpen tracks the Archive page (docs/DESIGN.md §5): it selects
	// which of the two tabs is highlighted, and blanks the tree's view-mode
	// slot while true (that mode describes a surface the Archive page has
	// replaced). mainmenu learns this from the same Open/CloseArchivePageMsg
	// AppModel itself reacts to, via the ordinary component fan-out — there
	// is no separate broadcast for it.
	archiveOpen bool
	// searchOpen tracks the Search page (docs/DESIGN.md §5). Search has no tab
	// of its own — it is a visit opened from either page and left by esc or a
	// digit — so while it is open neither tab is highlighted. mainmenu learns
	// this from the same Open/CloseSearchPageMsg AppModel reacts to.
	searchOpen bool
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// New builds the header component.
func New() tea.Model {
	return Model{treeView: "all"}
}

// Update tracks the terminal width from the broadcast layout, and the task
// tree's view mode so the header stays in step with it.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cmds.SetBodyLayoutMsg:
		m.terminalWidth = msg.TerminalWidth
	case cmds.SetTaskTreeViewMsg:
		m.treeView = msg.View
	case cmds.OpenArchivePageMsg:
		m.archiveOpen = true
	case cmds.CloseArchivePageMsg:
		m.archiveOpen = false
	case cmds.OpenSearchPageMsg:
		m.searchOpen = true
	case cmds.CloseSearchPageMsg:
		m.searchOpen = false
	}
	return m, nil
}
