package scp

import (
	"strings"
	"testing"
)

const testSCP = `# This is a comment
2E0AOZ
2E0BNI
2E0BPP`

func TestDatabase_Read(t *testing.T) {
	database, err := Read(strings.NewReader(testSCP))
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedLengths := []struct {
		b      byte
		length int
	}{
		{'0', 3},
		{'2', 3},
		{'A', 1},
		{'B', 2},
		{'E', 3},
		{'I', 1},
		{'N', 1},
		{'O', 1},
		{'P', 1},
		{'Z', 1},
	}
	if len(database.items) != len(expectedLengths) {
		t.Errorf("expected %d bigrams, but got %d", len(expectedLengths), len(database.items))
	}
	for _, expectedLength := range expectedLengths {
		if len(database.items[expectedLength.b]) != expectedLength.length {
			t.Errorf("expected %d entries for %q, but got %d",
				expectedLength.length,
				string(expectedLength.b),
				len(database.items[expectedLength.b]))
		}
	}
}

func TestDatabase_Find(t *testing.T) {
	database, err := Read(strings.NewReader(testSCP))
	if err != nil {
		t.Errorf("%v", err)
	}

	testCases := []struct {
		value    string
		expected []string
	}{
		{"", []string{}},
		{"2", []string{}},
		{"2E", []string{}},
		{"E0B", []string{"2E0BNI", "2E0BPP"}},
		{"E0BN", []string{"2E0BNI"}},
		{"2E0BNIX", []string{}},
	}

	for _, testCase := range testCases {
		actual, err := database.Find(testCase.value)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(actual) != len(testCase.expected) {
			t.Errorf("%s: expected %d entries, but got %d, %v", testCase.value, len(testCase.expected), len(actual), actual)
		}
	}
}
