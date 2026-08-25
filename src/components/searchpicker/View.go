package searchpicker

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/filipemolina/farol/src/appstyles"
	"github.com/filipemolina/farol/src/components/chrome"
	"github.com/filipemolina/farol/src/keys"
)

func (m Model) View() tea.View {
	// Seal the input onto the modal surface every render: the bubbles
	// textinput default carries no foreground on focused text (and a
	// hardcoded white on the blurred one), which vanishes on a light theme's
	// modal (farol-day). Same per-render discipline as detailspanel's
	// inputs (docs/DESIGN.md §12).
	chrome.SealInput(&m.input, appstyles.Active.ModalBg, appstyles.Active.ModalBg)

	body := lipgloss.JoinVertical(lipgloss.Left,
		chrome.ModalTitle("Search all lists"),
		m.input.View(),
		"",
		m.renderResults(),
		m.renderError(),
		"",
		m.renderLegend(),
		"",
		m.renderHints(),
	)

	bg := appstyles.Active.ModalBg
	return tea.NewView(chrome.ModalSurface(bg, body))
}

// renderLegend explains the archived-list marker on the result rows. It is
// shown only when the current result set actually contains an archived list,
// so the picker does not burn a line advertising a state the user is not
// looking at (docs/DESIGN.md §12's "don't show chrome for empty state").
func (m Model) renderLegend() string {
	txt := ""
	for _, r := range m.results {
		if r.Archived {
			txt = "* archived list"
			break
		}
	}
	if txt == "" {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(appstyles.Active.TextDim).
		Background(appstyles.Active.ModalBg).
		Render(txt)
}

// renderResults renders the result rows on a fixed-height area: exactly
// m.visible lines every render, blank-padded below the matches (or the
// empty-state line). The modal's box therefore never changes size while the
// user types — a live search whose match count re-heights its own modal
// reads as a different surface on every keystroke.
func (m Model) renderResults() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.resultLines()...)
}

// resultLines returns exactly m.visible lines: the windowed, cursor-marked
// result rows when there are matches, else the empty-state line, each
// blank-padded to the fixed height. Pure so the height contract and the
// window math are table-testable.
func (m Model) resultLines() []string {
	blank := lipgloss.NewStyle().
		Background(appstyles.Active.ModalBg).
		Render(strings.Repeat(" ", m.rowWidth))
	if len(m.results) == 0 {
		dim := lipgloss.NewStyle().
			Foreground(appstyles.Active.TextDim).
			Background(appstyles.Active.ModalBg)
		var msg string
		if m.query == "" {
			msg = "Type to search across every list"
		} else {
			msg = "No matches"
		}
		lines := make([]string, 0, m.visible)
		lines = append(lines, dim.Render(chrome.Truncate(msg, m.rowWidth)))
		for len(lines) < m.visible {
			lines = append(lines, blank)
		}
		return lines
	}

	start := 0
	shown := m.results
	if len(m.results) > m.visible {
		start = max(0, min(m.cursor-(m.visible/2), len(m.results)-m.visible))
		shown = m.results[start : start+m.visible]
	}

	lines := make([]string, 0, m.visible)
	for i, r := range shown {
		selected := i+start == m.cursor
		lines = append(lines, renderResult(r, selected, m.rowWidth))
	}
	for len(lines) < m.visible {
		lines = append(lines, blank)
	}
	return lines
}

// renderResult renders one candidate row, built as "list › title" and styled
// like a selectable task row. The cursor row gets the selected card chrome —
// a ▌ accent bar column plus a visible lift (BackgroundElevated, the theme's
// raised tier) — while every other row sits unselected on ModalBg (the modal
// surface itself, so it recedes). Lifting to ModalBg would be invisible here:
// the modal surface *is* ModalBg, so the selected row must use a different
// theme tier to contrast. Truncated to the picker's fixed row width so no
// match can widen the modal.
func renderResult(r Result, selected bool, width int) string {
	// The bar column spends one cell; the row body gets width-1 so bar+body
	// == width and the selected row's lift spans the full row as a band.
	// The body is a fixed-width, unpadded box: Width pads a short label with
	// on-theme spaces (and lipgloss never wraps a pre-truncated single-line
	// string), so the band is full-width rather than a highlight behind the
	// text, and the box stays exactly one line per row.
	bodyWidth := max(0, width-1)
	rowBg := appstyles.Active.ModalBg
	if selected {
		rowBg = appstyles.Active.BackgroundElevated
	}
	titleStyle := lipgloss.NewStyle().
		Background(rowBg)
	if selected {
		titleStyle = titleStyle.Foreground(appstyles.Active.TextPrimary).Bold(true)
	} else {
		titleStyle = titleStyle.Foreground(appstyles.Active.TextMuted)
	}

	label := r.renderLabel(titleStyle, rowBg)
	label = chrome.Truncate(label, bodyWidth)

	wrapper := lipgloss.NewStyle().
		Width(max(0, bodyWidth)).
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
// list's name is dimmed (TextDim) and suffixed with "*" (see renderLegend) —
// the dim signals "this is a different, non-active kind of list", and the
// marker explains why — while an active list's name is styled exactly like
// the task title, so the two read as one label rather than the list receding
// from its task. An unresolvable list name (deleted between search and
// render) falls back to the bare title, treated as active. The task title's
// own style is passed in so the cursor row can be bold TextPrimary; the name
// never picks up the cursor's bold on an active list, keeping "the cursor is
// the one distinct row" signal intact alongside the bar.
func (r Result) renderLabel(titleStyle lipgloss.Style, rowBg color.Color) string {
	if r.ListName == "" {
		return titleStyle.Render(r.Title)
	}

	name := r.ListName
	if r.Archived {
		name += "*"
		return lipgloss.NewStyle().
			Foreground(appstyles.Active.TextDim).
			Background(rowBg).
			Render(name) + " › " + titleStyle.Render(r.Title)
	}
	return titleStyle.Render(name) + " › " + titleStyle.Render(r.Title)
}

func (m Model) renderError() string {
	if m.errMsg == "" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(appstyles.Active.StatusOverdue).Render(m.errMsg)
}

func (m Model) renderHints() string {
	return chrome.RenderKeyHints([]chrome.KeyHint{
		chrome.HintFor(keys.Overlay.Submit),
		{Key: "↑/↓", Desc: "navigate"},
		chrome.HintFor(keys.Overlay.Cancel),
	}, appstyles.Active.TextMuted)
}
