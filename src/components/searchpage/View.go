package searchpage

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/filipemolina/farol/src/appstyles"
	"github.com/filipemolina/farol/src/components/chrome"
)

// View renders the page filling the whole body — the terminal width, not just
// a Tasks-sized column, since the Search page replaces the Tasks/Lists split
// entirely the way the Archive page does (docs/DESIGN.md §5). The keybinding
// bar stays live around it (unlike the Archive page, whose footer blanks), so
// this page renders no private hint line of its own — the footer advertises
// navigate / open / back from the page's bindings in src/keys.
func (m Model) View() tea.View {
	width := m.body.TerminalWidth
	height := m.body.Height
	bg := chrome.PanelBg(m.focused)
	bodyW := max(1, chrome.PanelBodyWidth(width))
	bodyH := max(1, chrome.PanelBodyHeight(height))

	// Seal the input onto the page surface every render: the bubbles
	// textinput default carries no foreground on focused text (and a
	// hardcoded white on the blurred one), which vanishes on a light theme's
	// panel. Same per-render discipline as detailspanel's inputs
	// (docs/DESIGN.md §12).
	chrome.SealInput(&m.input, bg, bg)

	// The input owns one line; the legend (when an archived result is present)
	// owns another. Reserve both so the results window never collides with
	// them.
	hasArchived := containsArchived(m.results)
	reserved := 1 // input line
	if hasArchived {
		reserved++
	}
	contentH := max(0, bodyH-reserved-1) // -1 for the blank spacer under the input

	var content string
	switch {
	case m.errMsg != "":
		content = chrome.EmptyStateCard(m.errMsg, bodyW, contentH, bg)
	case len(m.results) == 0:
		msg := "Type to search across every list"
		if m.query != "" {
			msg = "No matches"
		}
		content = chrome.EmptyStateCard(msg, bodyW, contentH, bg)
	default:
		content = m.renderResults(bodyW, contentH, bg)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, m.input.View(), "", content)
	if hasArchived {
		legend := lipgloss.NewStyle().
			Foreground(appstyles.Active.TextDim).
			Background(bg).
			Render("* archived list")
		body = lipgloss.JoinVertical(lipgloss.Left, body, legend)
	}

	title := "Search"
	return tea.NewView(chrome.PanelFrameWithRightTitle(title, "", m.focused, width, height, body))
}

// containsArchived reports whether any result in the set lives on an archived
// list, which is what gates the "* archived list" legend (docs/DESIGN.md §5).
func containsArchived(results []Result) bool {
	for _, r := range results {
		if r.Archived {
			return true
		}
	}
	return false
}

// renderResults renders the result rows on a single column, windowed around
// the cursor. The layout derives its width from the body broadcast (not the
// terminal), and leaves the door open for a second read-only preview column
// the way the Archive page splits its two — so nothing here assumes the
// results own the full width (docs/DESIGN.md §5).
func (m Model) renderResults(width, height int, bg color.Color) string {
	start := 0
	shown := m.results
	if len(m.results) > height {
		start = max(0, min(m.cursor-height/2, len(m.results)-height))
		shown = m.results[start : start+height]
	}

	rows := make([]string, 0, len(shown))
	for i, r := range shown {
		selected := i+start == m.cursor
		rows = append(rows, renderResult(r, selected, m.focused, width, m.query))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	// Fill the column to height on the page surface so a short result set
	// never bleeds the background.
	return chrome.PanelBodyWithFooter(width, height, bg, content, "")
}

// renderResult renders one candidate row, built as "list › title" and styled
// like a selectable task row. The cursor row gets the selected-task treatment
// outright — the ▌ accent bar plus the lift (chrome.BarColumn and the shared
// bar-and-body row split, so bar + body sum to the row width and the lift
// reads as a full-width band) exactly as the task tree's selected row does
// (docs/DESIGN.md §12): the modal-era detour through BackgroundElevated
// existed only because the modal's own ModalBg surface made the standard
// signal invisible there. Within a row, the characters the matcher consumed in
// the title draw in Accent (matchSpans); a notes-only hit highlights nothing.
// An active list's name is styled exactly like its task title — the two read
// as one label — and only an archived list's name is drawn in TextDim
// (suffixed with "*"). The cursor row keeps TextPrimary bold on the title,
// never on the list name, so the bold cue, the accent bar, and the match
// highlight all point at the one selected row.
func renderResult(r Result, selected, focused bool, width int, query string) string {
	rowBg := chrome.ListRowBg(selected, focused)

	titleStyle := lipgloss.NewStyle().Background(rowBg)
	if selected {
		titleStyle = titleStyle.Foreground(appstyles.Active.TextPrimary).Bold(true)
	} else {
		titleStyle = titleStyle.Foreground(appstyles.Active.TextMuted)
	}

	label := r.renderLabel(titleStyle, rowBg, query)
	label = chrome.Truncate(label, max(0, width-1))

	wrapper := lipgloss.NewStyle().
		Width(max(0, width-1)).
		Background(rowBg).
		Render(label)

	barFg := rowBg // invisible on an unselected row
	if selected {
		barFg = appstyles.Active.Accent
	}
	return appstyles.FillBackground(rowBg,
		lipgloss.JoinHorizontal(lipgloss.Left, chrome.BarColumn(barFg, rowBg, wrapper), wrapper))
}

// renderLabel styles one result's "list › title" string. Only an ARCHIVED
// list's name is dimmed (TextDim) and suffixed with "*" (see the legend) — the
// dim signals "this is a different, non-active kind of list" — while an active
// list's name is styled exactly like the task title, so the two read as one
// label rather than the list receding from its task. The title carries the
// match highlight (Accent on the matched subsequence) computed by matchSpans.
// An unresolvable list name (deleted between search and render) falls back to
// the bare title, treated as active. Every fragment the row emits — name,
// separator, title — sets the row background explicitly, so the band stays
// unbroken across the whole label.
func (r Result) renderLabel(titleStyle lipgloss.Style, rowBg color.Color, query string) string {
	titleRendered := renderTitleWithSpans(r.Title, matchSpans(query, r.Title), titleStyle)
	if r.ListName == "" {
		return titleRendered
	}

	// The "›" is chrome between the label's two halves rather than part of
	// either, so it carries the row background explicitly and draws in TextDim
	// on every row. Left bare it inherited the terminal's own default colors
	// instead of the theme's, punching a default-background gap through the
	// selected row's full-width band and vanishing outright wherever those
	// defaults happened to land on the row beneath it. TextDim clears the 2.2
	// glyph floor on both row surfaces in every theme (Contrast_test.go).
	sep := lipgloss.NewStyle().
		Foreground(appstyles.Active.TextDim).
		Background(rowBg).
		Render(" › ")

	if r.Archived {
		name := lipgloss.NewStyle().
			Foreground(appstyles.Active.TextDim).
			Background(rowBg).
			Render(r.ListName + "*")
		return name + sep + titleRendered
	}
	// Active list name styled exactly like the task title (the cursor row's
	// bold carries through titleStyle, so it stays bold when selected).
	name := titleStyle.Render(r.ListName)
	return name + sep + titleRendered
}
