package dxcc

import (
	"testing"
)

func TestPrefixes_add(t *testing.T) {
	prefixes := NewPrefixes()
	prefix := Prefix{Prefix: "P", Name: "N"}

	prefixes.Add(prefix)

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
	prefixes.Add(
		Prefix{Prefix: "P1"},
		Prefix{Prefix: "P2"},
		Prefix{Prefix: "P2"},
		Prefix{Prefix: "P3AB", NeedsExactMatch: true},
		Prefix{Prefix: "P4"},
		Prefix{Prefix: "P4A", NeedsExactMatch: true},
		Prefix{Prefix: "P5"},
		Prefix{Prefix: "P5B", NotARRLCompliant: true},
	)

	testCases := []struct {
		value, prefix string
		count         int
		arrlCompliant bool
	}{
		{"P1", "P1", 1, false},
		{"P1A", "P1", 1, false},
		{"P2", "P2", 2, false},
		{"P3AB", "P3AB", 1, false},
		{"P3A", "", 0, false},
		{"P3ABC", "", 0, false},
		{"P4A", "P4A", 1, false},
		{"P4AB", "P4", 1, false},
		{"P5BA", "P5B", 1, false},
		{"P5BC", "P5", 1, true},
	}
	for _, testCase := range testCases {
		foundPrefixes, _ := prefixes.find(testCase.value, testCase.arrlCompliant)
		if len(foundPrefixes) != testCase.count {
			t.Errorf("%q: expected %d, but found %d: %v", testCase.value, testCase.count, len(foundPrefixes), foundPrefixes)
			continue
		}
		if testCase.count > 0 && foundPrefixes[0].Prefix != testCase.prefix {
			t.Errorf("%q: expected %s, but found %s", testCase.value, testCase.prefix, foundPrefixes[0].Prefix)
		}
	}
}
