package scp

import (
	"testing"
)

func TestEntrySet(t *testing.T) {
	e1_1 := newEntry("one")
	e1_2 := newEntry("one")
	e2 := newEntry("two")
	e3 := newEntry("abc")

	set := entrySet{}
	entries := set.Entries()
	if len(entries) != 0 {
		t.Errorf("should be empty after creation: %d %v", len(entries), entries)
	}

	set.Add(e1_1)
	entries = set.Entries()
	if len(entries) != 1 {
		t.Errorf("should contain 1 entry, but got %d %v", len(entries), entries)
	}

	set.Add(e1_2)
	entries = set.Entries()
	if len(entries) != 1 {
		t.Errorf("should contain 1 entry, but got %d %v", len(entries), entries)
	}

	set.Add(e2)
	entries = set.Entries()
	if len(entries) != 2 {
		t.Errorf("should contain 2 entries, but got %d %v", len(entries), entries)
	}

	set.Add(e3)
	entries = set.Entries()
	if len(entries) != 3 {
		t.Errorf("should contain 3 entries, but got %d %v", len(entries), entries)
	}

	entries = set.Filter(func(e entry) bool {
		return e.s == "abc"
	})
	if len(entries) != 1 {
		t.Errorf("should contain 1 entry, but got %d %v", len(entries), entries)
	}
	if entries[0].s != "abc" {
		t.Errorf("filtering abc failed: %v", entries)
	}
}
