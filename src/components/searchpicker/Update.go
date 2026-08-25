package searchpicker

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/filipemolina/farol/src/cmds"
	"github.com/filipemolina/farol/src/keys"
)

// navBinding is the up/down arrow pair the picker uses to move the result
// cursor. It deliberately excludes j/k — those are printable characters the
// text input needs, so only the arrows navigate.
var navBinding = key.NewBinding(key.WithKeys("up", "down"))

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Overlay.Cancel):
			return m, cmds.CloseModal(nil)

		case key.Matches(msg, keys.Overlay.Submit):
			if len(m.results) == 0 || m.cursor < 0 {
				return m, nil
			}
			r := m.results[m.cursor]
			// Close first, then hand off as the follow-up, so the modal is
			// gone before AppModel switches lists or opens the Archive page.
			// An archived list cannot become the active list, so a result
			// from one goes to the Archive page (revealed on that list and
			// task) rather than the JumpToTask that moves the active list.
			var follow tea.Cmd
			if r.Archived {
				follow = cmds.OpenArchivedTask(r.TaskID, r.ListID)
			} else {
				follow = cmds.JumpToTask(r.TaskID, r.ListID)
			}
			return m, cmds.CloseModal(follow)

		case key.Matches(msg, navBinding):
			dir := 1
			if msg.String() == "up" {
				dir = -1
			}
			if len(m.results) > 0 {
				m.cursor += dir
				m.clampCursor()
			}
			return m, nil

		default:
			m.errMsg = ""
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.runSearch()
			return m, cmd
		}
	}

	// Non-key messages (cursor blink) go straight to the input.
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
