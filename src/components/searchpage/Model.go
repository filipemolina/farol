// Package searchpage is the cross-list Search page: a full-body surface that
// replaces Tasks and Lists the way the Archive page does (docs/DESIGN.md §5 —
// the picker promoted from a modal to a page). A text input at the top
// live-searches every list; the results below show each match's list context
// ("list › title"). Enter jumps to the selected task (switching the active
// list when it lives on an active list, or opening the Archive page revealed
// on an archived one), esc closes the page onto Active, and the digits leave
// for their pages. The query, results, and cursor persist while the app runs.
package searchpage

import (
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/filipemolina/farol/src/appstyles"
	"github.com/filipemolina/farol/src/cmds"
	"github.com/filipemolina/farol/src/components/chrome"
	"github.com/filipemolina/farol/src/constants"
	"github.com/filipemolina/farol/src/keys"
	"github.com/filipemolina/farol/src/store"
	"github.com/sahilm/fuzzy"
)

// focusedZoneID is the zone id this component answers to. Like Details and the
// Archive page, the Search page is focused only while it is visible, entered
// by AppModel's explicit open transition — it is never in the tab/shift+tab
// cycle.
const focusedZoneID = constants.COMPONENT_SEARCH_PAGE

// Result is one candidate: the task plus the name of the list it lives in,
// enough to render "<list> › <title>" and to jump to it. Archived marks a
// result whose list is archived, so the page can label it distinctly and
// route Enter to the Archive page instead of trying to make an archived list
// the active one.
type Result struct {
	TaskID   string
	Title    string
	ListID   string
	ListName string
	Archived bool
}

// Span is a rune-index range [Start, End) inside a title that the matcher
// consumed. Rendering highlights these in Accent (docs/DESIGN.md §12).
type Span struct {
	Start int
	End   int
}

// Model is the cross-list Search page: a focused text input, a live result
// list, and a cursor over it.
type Model struct {
	input   textinput.Model
	query   string
	results []Result
	cursor  int
	errMsg  string
	store   *store.Store
	focused bool
	body    cmds.SetBodyLayoutMsg
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

// New builds the Search page. It takes no terminal dimensions: the component is
// constructed once and kept in AppModel's component set (like the Archive page
// and the task/lists panels), and is sized exclusively through the
// SetBodyLayoutMsg broadcast every body surface takes — so a resize while it is
// open reflows it the same way every other body surface reflows, and nothing
// here hand-subtracts layout constants (docs/DESIGN.md §5).
func New(s *store.Store) tea.Model {
	input := textinput.New()
	input.Focus()
	input.Placeholder = "search all lists"

	return Model{store: s, input: input}
}

// Update runs the message through update and returns the result. The page is a
// body surface, not a modal: it receives the same layout and focus broadcasts
// as the other body components, and owns the keyboard only while AppModel has
// switched the page to it.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cmds.SetBodyLayoutMsg:
		m.body = msg
		// Size the input from the broadcast: the page is constructed once
		// with no dimensions, and a zero-width textinput renders a single
		// rune of the placeholder ("s"). The modal sized its input at New
		// from terminal dims the page no longer receives; the input takes
		// the same body width the result rows render at.
		m.input.SetWidth(max(0, chrome.PanelBodyWidth(msg.TerminalWidth)))
		return m, nil

	case cmds.SetFocusMsg:
		m.focused = int(msg) == focusedZoneID
		return m, nil

	// AppModel issues this on open and on every poll tick while the page is
	// open, so a non-empty query's results never go stale behind the cursor
	// (docs/DESIGN.md §5).
	case cmds.RefreshSearchMsg:
		m.runSearch()
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey matches the page's bindings (declared once in src/keys, like every
// other surface) and otherwise feeds the keystroke to the text input so it
// re-runs the live search. j/k are printable query characters here — only the
// arrow keys navigate, so the matcher can consume them (docs/DESIGN.md §5).
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.SearchPage.Cancel):
		// esc closes the page onto Active: search is a visit, not a
		// destination, so remembering which page it was opened from buys
		// nothing over the two digits that already jump anywhere
		// (docs/DESIGN.md §5).
		return m, cmds.CloseSearchPage(nil)

	case key.Matches(msg, keys.Global.PageActive):
		// 1 is a second way off the page, landing on Active (the same
		// "the digit always jumps there" contract the Archive page's own 1
		// uses). Plain switch, no ladder.
		return m, cmds.CloseSearchPage(nil)

	case key.Matches(msg, keys.Global.PageArchived):
		// 2 leaves for the Archived page (loads its lists the same way the
		// global 2 does from Active).
		return m, cmds.OpenArchivePage()

	case key.Matches(msg, keys.Global.PageSearch):
		// 3 is this page's own tab: already home, so this consumes the
		// keystroke as an idempotent no-op — the digits are never query
		// characters anywhere in the app (docs/DESIGN.md §5). F is NOT
		// matched here: it falls through to the input as a printable, so
		// titles like "Farol v0.4" stay searchable.
		return m, nil

	case key.Matches(msg, keys.SearchPage.Submit):
		if len(m.results) == 0 || m.cursor < 0 {
			return m, nil
		}
		r := m.results[m.cursor]
		// Close first, then hand off as the follow-up, so the page is gone
		// before AppModel switches lists or opens the Archive page. An
		// archived list cannot become the active list, so a result from one
		// goes to the Archive page (revealed on that list and task) rather
		// than the JumpToTask that moves the active list.
		var follow tea.Cmd
		if r.Archived {
			follow = cmds.OpenArchivedTask(r.TaskID, r.ListID)
		} else {
			follow = cmds.JumpToTask(r.TaskID, r.ListID)
		}
		return m, cmds.CloseSearchPage(follow)

	case key.Matches(msg, keys.SearchPage.Navigate):
		dir := 1
		if msg.String() == "up" {
			dir = -1
		}
		if len(m.results) > 0 {
			m.cursor += dir
			m.clampCursor()
		}
		return m, nil
	}

	// Default: a printable keystroke edits the query and re-runs the live
	// search.
	m.errMsg = ""
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.runSearch()
	return m, cmd
}

// runSearch re-runs the store search for the input's current value and
// replaces the results, keeping the cursor in range. A store error is kept
// for the error line rather than silently dropping the results.
func (m *Model) runSearch() {
	m.query = m.input.Value()
	results, err := rank(m.store, m.query)
	if err != nil {
		m.errMsg = err.Error()
		m.results = nil
		m.clampCursor()
		return
	}
	m.errMsg = ""
	m.results = results
	m.clampCursor()
}

// clampCursor keeps the cursor within the result list (or at -1 when there is
// nothing to pick).
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

// matchSpans returns the title spans the fuzzy matcher consumed for query
// against title. A title that fuzzy-matches (the subsequence the user typed,
// possibly non-contiguous) yields those matched positions merged into
// contiguous spans; a title that did not match the query at all (a notes-only
// hit) yields nil, so rendering highlights nothing and honestly shows the
// title did not match. Pure so ranking and rendering cannot disagree about
// what matched (docs/DESIGN.md §12); table-tested like any rule-heavy
// function.
//
// sahilm/fuzzy reports MatchedIndexes as BYTE offsets, while spans index
// runes — renderTitleWithSpans slices []rune — and on a multibyte title the
// two diverge: an unconverted byte offset lands on the wrong character or
// slices past the end of the rune slice entirely. Each matched byte offset is
// therefore mapped to the rune that contains it before any span math.
func matchSpans(query, title string) []Span {
	if query == "" || title == "" {
		return nil
	}
	matches := fuzzy.Find(query, []string{title})
	if len(matches) == 0 {
		return nil
	}
	idxs := matches[0].MatchedIndexes
	if len(idxs) == 0 {
		return nil
	}

	byteToRune := make(map[int]int, len(title))
	bi, ri := 0, 0
	for bi < len(title) {
		_, size := utf8.DecodeRuneInString(title[bi:])
		for k := 0; k < size; k++ {
			byteToRune[bi+k] = ri
		}
		bi += size
		ri++
	}

	var spans []Span
	start, prev := -1, -1
	for _, b := range idxs {
		r, ok := byteToRune[b]
		if !ok {
			continue
		}
		switch {
		case start == -1:
			start, prev = r, r
		case r == prev:
			// Two matched bytes inside one rune collapse to one position.
		case r == prev+1:
			prev = r
		default:
			spans = append(spans, Span{Start: start, End: prev + 1})
			start, prev = r, r
		}
	}
	if start != -1 {
		spans = append(spans, Span{Start: start, End: prev + 1})
	}
	return spans
}

// renderTitleWithSpans renders title with the matched spans drawn in Accent
// over the base style, the rest in base. Pure given its inputs.
func renderTitleWithSpans(title string, spans []Span, base lipgloss.Style) string {
	if len(spans) == 0 {
		return base.Render(title)
	}
	runes := []rune(title)
	var b strings.Builder
	i := 0
	for _, s := range spans {
		if s.Start > i {
			b.WriteString(base.Render(string(runes[i:s.Start])))
		}
		b.WriteString(base.Foreground(appstyles.Active.Accent).Render(string(runes[s.Start:s.End])))
		i = s.End
	}
	if i < len(runes) {
		b.WriteString(base.Render(string(runes[i:])))
	}
	return b.String()
}
