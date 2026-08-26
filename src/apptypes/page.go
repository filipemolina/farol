package apptypes

// Page is a top-level body surface: exactly one renders at a time, switched
// by the digit tabs (docs/DESIGN.md §5). It lives in apptypes rather than
// model so components that only display the current page (mainmenu) can
// receive it without importing the Bubble Tea half.
type Page int

const (
	PageActive Page = iota
	PageArchived
	PageSearch
)
