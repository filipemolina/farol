package searchpage

import (
	"github.com/filipemolina/farol/src/store"
	"github.com/sahilm/fuzzy"
)

// rank runs store.SearchTasks across all lists and orders the results the
// same way the CLI's rankSearch does (step 5): title matches first, ranked
// by fuzzy score, then candidates that matched only on notes in store
// order. Each result carries its list's name for context. A store error is
// returned so the page can surface it in the error tier (docs/DESIGN.md §12)
// rather than silently dropping the results.
func rank(s *store.Store, query string) ([]Result, error) {
	if query == "" {
		return nil, nil
	}

	candidates, err := s.SearchTasks(query, nil)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Prepend the notes fallback so a notes-only hit still shows, but title
	// matches always rank above it regardless of fuzzy score.
	lists := allLists(s)

	titles := make([]string, len(candidates))
	for i, c := range candidates {
		titles[i] = c.Title
	}

	matched := make([]bool, len(candidates))
	out := make([]Result, 0, len(candidates))
	for _, m := range fuzzy.Find(query, titles) {
		matched[m.Index] = true
		t := candidates[m.Index]
		out = append(out, Result{TaskID: t.ID, Title: t.Title, ListID: t.ListID, ListName: lists[t.ListID].Name, Archived: lists[t.ListID].Archived})
	}
	for i, c := range candidates {
		if matched[i] {
			continue
		}
		out = append(out, Result{TaskID: c.ID, Title: c.Title, ListID: c.ListID, ListName: lists[c.ListID].Name, Archived: lists[c.ListID].Archived})
	}
	return out, nil
}

// listContext is the subset of a list the result row needs: its display name
// and whether it is archived. A zero value means the list could not be
// resolved (deleted between search and render), which renders as a bare task
// title and is treated as active.
type listContext struct {
	Name     string
	Archived bool
}

// allLists builds the list-id -> context map once per search from
// store.ListAllLists (every list, archived included). The search runs across
// all lists regardless of archived state, so a result from an archived list
// must still carry its list's name — using store.ListLists here, which
// excludes archived lists, would render those results as " › Title" (a
// misleading "child of nothing" indent) and leave no way to tell the list
// apart. Archived is derived from the list's own ArchivedAt, so the page
// can mark it and route Enter to the Archive page.
func allLists(s *store.Store) map[string]listContext {
	lists, err := s.ListAllLists()
	if err != nil {
		return map[string]listContext{}
	}
	ctx := make(map[string]listContext, len(lists))
	for _, l := range lists {
		ctx[l.List.ID] = listContext{Name: l.List.Name, Archived: l.List.ArchivedAt != nil}
	}
	return ctx
}
