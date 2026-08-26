package mainmenu

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/filipemolina/farol/src/appstyles"
	"github.com/filipemolina/farol/src/apptypes"
	"github.com/filipemolina/farol/src/constants"
)

const versionGutter = 4

// tabLabel renders one page tab as "<digit> <label>" — the digit in the
// accent color advertises the key that jumps to it directly on the tab
// itself, the same look ../cais's mainmenu uses (docs/DESIGN.md §5). The tab
// set is closed — three pages — so this is three literals over one current-
// page value, not a registry; only the rendering is borrowed.
func tabLabel(digit, label string, fg color.Color, bold bool) string {
	d := lipgloss.NewStyle().
		Foreground(appstyles.Active.Accent).
		Background(appstyles.Active.BackgroundContent).
		Bold(true).
		Render(digit)

	l := lipgloss.NewStyle().
		Foreground(fg).
		Background(appstyles.Active.BackgroundContent).
		Bold(bold).
		Render(" " + label)

	return d + l
}

// viewModeIcon is the single glyph prefixing the view-mode indicator,
// matching ../pulso's mainmenu.sortIcon glyph (pulso/src/components/mainmenu/View.go).
const viewModeIcon = "⇅"

// View renders the header bar: the three page tabs on the left (1 Active,
// 2 Archived, 3 Search — keys.Global.PageActive/PageArchived/PageSearch), the
// task tree's view mode and version low-emphasis next to the wordmark on the
// right. No bottom border — the tier-2 background against the tier-3 panels
// below provides the section break, exactly like stack-stitcher.
func (m Model) View() tea.View {
	if m.terminalWidth <= 0 {
		return tea.NewView("")
	}

	barStyle := lipgloss.NewStyle().
		Background(appstyles.Active.BackgroundContent).
		Width(m.terminalWidth)

	// Cell styles carry only the spacing; tabLabel handles color/weight so the
	// digit is never competing with a foreground set on the enclosing style.
	cellStyle := lipgloss.NewStyle().
		Background(appstyles.Active.BackgroundContent).
		Padding(0, 2)
	// The selected cell has less left padding to compensate for the external ▌.
	activeCellStyle := cellStyle.Padding(0, 2, 0, 1)

	accentBar := lipgloss.NewStyle().
		Foreground(appstyles.Active.Accent).
		Background(appstyles.Active.BackgroundContent).
		Render("▌")

	tab := func(digit, label string, selected bool) string {
		if selected {
			cell := activeCellStyle.Render(tabLabel(digit, label, appstyles.Active.TextPrimary, true))
			return lipgloss.JoinHorizontal(lipgloss.Left, accentBar, cell)
		}
		return cellStyle.Render(tabLabel(digit, label, appstyles.Active.TextDim, false))
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Left,
		tab("1", "Active", m.page == apptypes.PageActive),
		tab("2", "Archived", m.page == apptypes.PageArchived),
		tab("3", "Search", m.page == apptypes.PageSearch),
	)

	wordmarkStyle := lipgloss.NewStyle().
		Foreground(appstyles.Active.Accent).
		Background(appstyles.Active.BackgroundContent).
		Bold(true).
		Padding(0, 2)

	versionStyle := lipgloss.NewStyle().
		Foreground(appstyles.Active.TextDim).
		Background(appstyles.Active.BackgroundContent)

	wordmark := wordmarkStyle.Render(constants.WORDMARK)

	// The right cluster is the view-mode indicator and the version, joined
	// by " . " only between pieces that are actually present — pulso's
	// "all . pulso-dusk . v0.3.0" look without a separator dangling off an
	// end. An unstamped build reports "unknown", which is identity noise
	// rather than information, so it renders as no version at all.
	//
	// The cluster is blank while a page other than Active is on screen:
	// Pending/Complete/All describes the task tree, which the Archive and
	// Search pages replace, and which page is on screen is carried by the
	// tabs themselves rather than a text label taking over this slot.
	modeText := ""
	if m.page == apptypes.PageActive && m.treeView != "" {
		modeText = viewModeIcon + " " + m.treeView
	}
	versionText := constants.Version()
	if versionText == "unknown" {
		versionText = ""
	}
	var right string
	switch {
	case modeText != "" && versionText != "":
		right = versionStyle.Render(modeText + " . " + versionText)
	case modeText != "":
		right = versionStyle.Render(modeText)
	case versionText != "":
		right = versionStyle.Render(versionText)
	}

	// Shed the cluster outward-in when it would crowd the wordmark: the
	// version first (static identity), then the mode (live state the user
	// checks — shedding the mode first blanked the tree-filter indicator at
	// everyday widths once the third tab tightened this budget). The tabs
	// and the wordmark never shed: the tabs are the primary navigation, the
	// wordmark the header's identity.
	fixedWidth := lipgloss.Width(tabs) + lipgloss.Width(wordmark) + versionGutter
	if fixedWidth+lipgloss.Width(right) > m.terminalWidth && modeText != "" {
		right = versionStyle.Render(modeText)
	}
	if fixedWidth+lipgloss.Width(right) > m.terminalWidth {
		right = ""
	}

	gapWidth := m.terminalWidth - fixedWidth - lipgloss.Width(right)
	if gapWidth < 0 {
		gapWidth = 0
	}

	gap := lipgloss.NewStyle().
		Background(appstyles.Active.BackgroundContent).
		Width(gapWidth).
		Render("")

	row := lipgloss.JoinHorizontal(lipgloss.Left, tabs, gap, right, wordmark)
	return tea.NewView(barStyle.Render(row))
}
