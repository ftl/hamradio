package scp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func TestEntrySet(t *testing.T) {
	e1_1 := newEntry("one", nil)
	e1_2 := newEntry("one", nil)
	e2 := newEntry("two", nil)
	e3 := newEntry("abc", nil)

	set := entrySet{}
	entries := set.Entries()
	assert.Equal(t, 0, len(entries), "should be empty after creation")

	set.Add(e1_1)
	entries = set.Entries()
	assert.Equal(t, 1, len(entries))

	set.Add(e1_2)
	entries = set.Entries()
	assert.Equal(t, 1, len(entries))

	set.Add(e2)
	entries = set.Entries()
	assert.Equal(t, 2, len(entries))

	set.Add(e3)
	entries = set.Entries()
	assert.Equal(t, 3, len(entries))

	entries = set.Filter(func(e Entry) bool {
		return e.key == "ABC"
	})
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "ABC", entries[0].key)
}

func TestNewAnnotatedMatch(t *testing.T) {
	tt := []struct {
		input    string
		entry    string
		expected MatchingAssembly
	}{
		{"abcd", "abcd", MatchingAssembly{MatchingPart{NOP, "abcd"}}},
		{"abc", "abcd", MatchingAssembly{MatchingPart{NOP, "abc"}, MatchingPart{Insert, "d"}}},
		{"abcd", "abc", MatchingAssembly{MatchingPart{NOP, "abc"}, MatchingPart{Delete, "d"}}},
		{"efgd", "abcd", MatchingAssembly{MatchingPart{Substitute, "abc"}, MatchingPart{NOP, "d"}}},
		{"efghd", "abcd", MatchingAssembly{MatchingPart{Substitute, "abc"}, MatchingPart{Delete, "h"}, MatchingPart{NOP, "d"}}},
		{"aefgd", "abcd", MatchingAssembly{MatchingPart{NOP, "a"}, MatchingPart{Substitute, "bc"}, MatchingPart{Delete, "g"}, MatchingPart{NOP, "d"}}},
		{"aady", "aaney", MatchingAssembly{MatchingPart{NOP, "aa"}, MatchingPart{FalseFriend, "n"}, MatchingPart{Insert, "e"}, MatchingPart{NOP, "y"}}},
		{"aaney", "aady", MatchingAssembly{MatchingPart{NOP, "aa"}, MatchingPart{Substitute, "d"}, MatchingPart{Delete, "e"}, MatchingPart{NOP, "y"}}},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s -> %s", tc.input, tc.entry), func(t *testing.T) {
			matrix := levenshtein.MatrixForStrings([]rune(tc.input), []rune(tc.entry), levenshtein.DefaultOptions)
			script := levenshtein.EditScriptForMatrix(matrix, levenshtein.DefaultOptions)

			actual := newMatchingAssembly(tc.input, tc.entry, script)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.entry, actual.String())
		})
	}
}

func TestPreferFalseFriends(t *testing.T) {
	input := newEntry("dl3dy", nil)
	entry := newEntry("dl3ney", nil)
	d, a, m := input.EditTo(entry)

	assert.Equal(t, distance(2), d)
	assert.Equal(t, accuracy(0.8181818181818182), a)
	assert.Equal(t, MatchingAssembly{MatchingPart{NOP, "DL3"}, MatchingPart{FalseFriend, "N"}, MatchingPart{Insert, "E"}, MatchingPart{NOP, "Y"}}, m)
	assert.True(t, m.ContainsFalseFriend())
}
