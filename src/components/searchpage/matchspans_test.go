package searchpage

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// Transcribed from docs/DESIGN.md §12 ("The Search page's result rows play by
// the page rules"): a fuzzy title match yields the matched subsequence's
// spans — possibly non-contiguous; a title that did not match yields nil, so
// a notes-only hit highlights nothing; empty inputs yield nil. The multibyte
// case is the regression for sahilm/fuzzy reporting BYTE offsets while spans
// index runes.
func TestMatchSpans(t *testing.T) {
	cases := []struct {
		name  string
		query string
		title string
		want  []Span
	}{
		{"empty query", "", "Buy milk", nil},
		{"empty title", "milk", "", nil},
		{"no title match (notes-only hit)", "zzzz", "Buy milk", nil},
		{"contiguous substring", "milk", "Buy milk", []Span{{4, 8}}},
		{
			"non-contiguous subsequence",
			"bmk", "Buy milk",
			[]Span{{0, 1}, {4, 5}, {7, 8}},
		},
		{
			"gap then run",
			"mlk", "Buy milk",
			[]Span{{4, 5}, {6, 8}},
		},
		{"case-insensitive", "MILK", "Buy milk", []Span{{4, 8}}},
		{
			"multibyte title stays rune-aligned",
			"fe", "Café ☕ latte",
			[]Span{{2, 3}, {11, 12}},
		},
		{
			"accented character inside a contiguous match",
			"café", "Café latte",
			[]Span{{0, 4}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchSpans(tc.query, tc.title)
			if len(got) != len(tc.want) {
				t.Fatalf("matchSpans(%q, %q) = %v, want %v", tc.query, tc.title, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("span[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// renderTitleWithSpans must place every highlighted rune inside the title and
// leave the rest to the base style — a span past the rune length would panic
// here rather than render.
func TestRenderTitleWithSpansStaysInBounds(t *testing.T) {
	title := "Café ☕ latte"
	spans := matchSpans("fe", title)
	base := lipgloss.NewStyle()
	out := renderTitleWithSpans(title, spans, base)
	if out == "" {
		t.Fatal("rendered nothing")
	}
	// A nil-span title renders through the base style untouched.
	if plain := renderTitleWithSpans(title, nil, base); plain == "" {
		t.Error("nil spans rendered nothing")
	}
}
