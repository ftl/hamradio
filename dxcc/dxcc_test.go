package dxcc

import (
	"testing"
)

func TestPrefixes_add(t *testing.T) {
	prefixes := NewPrefixes()
	prefix := Prefix{Prefix: "P", Name: "N"}

	prefixes.add(prefix)

	if len(prefixes.items) != 1 {
		t.Errorf("failed to add prefix")
		t.FailNow()
	}
	if prefixes.items[prefix.Prefix][0] != prefix {
		t.Errorf("expected %v, but got %v", prefix, prefixes.items[prefix.Prefix][0])
	}
}

func TestPrefixes_Find(t *testing.T) {
	prefixes := NewPrefixes()
	prefixes.add(
		Prefix{Prefix: "P1"},
		Prefix{Prefix: "P2"},
		Prefix{Prefix: "P2"},
		Prefix{Prefix: "P3AB", NeedsExactMatch: true},
		Prefix{Prefix: "P4"},
		Prefix{Prefix: "P4A", NeedsExactMatch: true},
	)

	testCases := []struct {
		value, prefix string
		count         int
	}{
		{"P1", "P1", 1},
		{"P1A", "P1", 1},
		{"P2", "P2", 2},
		{"P3AB", "P3AB", 1},
		{"P3A", "", 0},
		{"P3ABC", "", 0},
		{"P4A", "P4A", 1},
		{"P4AB", "P4", 1},
	}
	for _, testCase := range testCases {
		foundPrefixes, _ := prefixes.Find(testCase.value)
		if len(foundPrefixes) != testCase.count {
			t.Errorf("%q: expected %d, but found %d: %v", testCase.value, testCase.count, len(foundPrefixes), foundPrefixes)
			continue
		}
		if testCase.count > 0 && foundPrefixes[0].Prefix != testCase.prefix {
			t.Errorf("%q: expected %s, but found %s", testCase.value, testCase.prefix, foundPrefixes[0].Prefix)
		}
	}
}
