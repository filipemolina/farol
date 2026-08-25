package searchpage

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/filipemolina/farol/src/store"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// rank orders title matches first (by fuzzy score), then notes-only hits —
// the same decision the CLI's rankSearch makes. Each result carries its
// list's name so the page can render "<list> › <title>" without a second
// lookup.
func TestRankTitlesFirstThenNotes(t *testing.T) {
	s := testStore(t)
	lid, err := s.CreateList("Errands", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	tid, err := s.CreateTask(lid, "Buy milk", nil, "")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	// A notes-only hit: its title does not match, its notes do.
	if _, err := s.CreateTask(lid, "Groceries", nil, "remember milk at the store"); err != nil {
		t.Fatalf("create notes task: %v", err)
	}

	results, err := rank(s, "milk")
	if err != nil {
		t.Fatalf("rank: %v", err)
	}
	// "Buy milk" is a title match; "Groceries" only matches via notes.
	if len(results) != 2 {
		t.Fatalf("rank returned %d results, want 2", len(results))
	}
	if results[0].TaskID != tid {
		t.Errorf("results[0] = %q, want the title match first", results[0].TaskID)
	}

	for _, r := range results {
		if r.ListName != "Errands" {
			t.Errorf("result %q list name = %q, want Errands (context for the row)", r.Title, r.ListName)
		}
	}
}

// An empty query searches nothing: the page shows its guidance empty state,
// not every task in the store.
func TestRankEmptyQueryReturnsNoResults(t *testing.T) {
	s := testStore(t)
	results, err := rank(s, "")
	if err != nil {
		t.Fatalf("rank: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("rank(\"\") returned %d results, want 0", len(results))
	}
}

// The page's jump needs the task's list id so AppModel can switch to it;
// rank must carry that through.
func TestRankCarriesListID(t *testing.T) {
	s := testStore(t)
	lid, err := s.CreateList("List", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	tid, _ := s.CreateTask(lid, "unicorn", nil, "")

	results, err := rank(s, "unicorn")
	if err != nil {
		t.Fatalf("rank: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].ListID != lid || results[0].TaskID != tid {
		t.Errorf("carried ListID=%q/TaskID=%q, want %q/%q", results[0].ListID, results[0].TaskID, lid, tid)
	}
}

// Regression: a result from an ARCHIVED list must still carry its list name
// and be flagged Archived. Before the fix rank resolved names via
// store.ListLists, which excludes archived lists, so an archived-list result
// rendered as the anonymous, misleading " › Title". allLists now reads
// ListAllLists, so the Archived flag (which the page uses to mark the row and
// route Enter to the Archive page) can be derived.
func TestRankNamesArchivedListAndFlagsIt(t *testing.T) {
	s := testStore(t)
	lid, err := s.CreateList("Farol v0.4", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	if _, err := s.CreateTask(lid, "Tag a list as collaborative", nil, ""); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := s.ArchiveList(lid); err != nil {
		t.Fatalf("archive list: %v", err)
	}

	results, err := rank(s, "collaborative")
	if err != nil {
		t.Fatalf("rank: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("rank returned %d results, want 1", len(results))
	}
	r := results[0]
	if !r.Archived {
		t.Error("result.Archived = false, want true (the list is archived)")
	}
	if r.ListName == "" {
		t.Error("result.ListName is empty; archived-list results must carry their list name")
	}
	// The archived list's name carries its archive-date suffix (store.ArchiveList),
	// so it must contain the original name.
	if want := "Farol v0.4"; !strings.Contains(r.ListName, want) {
		t.Errorf("result.ListName = %q, want it to contain %q", r.ListName, want)
	}
}
