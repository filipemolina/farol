// Package searchpicker is the cross-list search modal opened by the global
// binding F. A text input at the top live-searches every list; the results
// below show each match's list context ("list › title"). Enter jumps to the
// selected task (switching the active list if needed), esc cancels.
package searchpicker

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/filipemolina/farol/src/store"
)

// modalChrome is how many rows the picker's chrome (border, title, input,
// hint lines) needs beyond its result list. It is the same number
// chrome.modalListChrome reserves internally; the picker sizes its own list
// so it re-derives it rather than importing the chrome helper.
const modalChrome = 10

// rowWidthChrome is the columns a result row yields to the modal's own
// chrome (border 2 + padding 4) plus a margin so the centered surface never
// touches the terminal edges. minRowWidth keeps the reserved area honest on
// terminals too narrow for the subtraction to leave anything.
const (
	rowWidthChrome = 10
	minRowWidth    = 24
)

// Result is one candidate: the task plus the name of the list it lives in,
// enough to render "<list> › <title>" and to jump to it. Archived marks a
// result whose list is archived, so the picker can label it distinctly and
// route Enter to the Archive page instead of trying to make an archived list
// the active one.
type Result struct {
	TaskID   string
	Title    string
	ListID   string
	ListName string
	Archived bool
}

// Model is the cross-list search picker: a focused text input, a live result
// list, and a cursor over it.
type Model struct {
	input    textinput.Model
	query    string
	results  []Result
	cursor   int
	errMsg   string
	store    *store.Store
	visible  int // result rows the modal reserves (fixed for its lifetime)
	rowWidth int // display columns every result row is truncated/padded to
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

// New builds the picker. termHeight sizes the result area and termWidth caps
// each row, both fixed at open so live filtering re-fills the list in place
// instead of resizing the modal on every keystroke. The input starts focused
// so the user can type immediately.
func New(s *store.Store, termWidth, termHeight int) tea.Model {
	input := textinput.New()
	input.Focus()
	input.Placeholder = "search all lists"

	rowWidth := max(minRowWidth, termWidth-rowWidthChrome)
	// The input sits directly above the result list and spans the same fixed
	// width, or a narrow default width truncates the placeholder mid-word
	// (the reported lone "s"). It is never resized on a keystroke — only the
	// content re-fills — so the box stays put, the same "no live resize"
	// contract as the result rows.
	input.SetWidth(max(0, rowWidth))

	return Model{
		input:    input,
		store:    s,
		visible:  max(3, termHeight-modalChrome),
		rowWidth: rowWidth,
	}
}

// runSearch re-runs the store search for the input's current value and
// replaces the results, keeping the cursor in range.
func (m *Model) runSearch() {
	m.query = m.input.Value()
	m.results = rank(m.store, m.query)
	m.errMsg = ""
	m.clampCursor()
}

// clampCursor keeps the cursor within the result list (or at -1 when there
// is nothing to pick).
func (m *Model) clampCursor() {
	if len(m.results) == 0 {
		m.cursor = -1
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
}
