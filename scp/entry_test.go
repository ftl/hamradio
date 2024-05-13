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
		{"aady", "aaney", MatchingAssembly{MatchingPart{NOP, "aa"}, MatchingPart{Substitute, "n"}, MatchingPart{Insert, "e"}, MatchingPart{NOP, "y"}}},
		{"aaney", "aady", MatchingAssembly{MatchingPart{NOP, "aa"}, MatchingPart{FalseFriend, "d"}, MatchingPart{Delete, "e"}, MatchingPart{NOP, "y"}}},
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
	input := newEntry("dl4m", nil)
	entry1 := newEntry("dl4w", nil)
	entry2 := newEntry("dl4k", nil)
	entry3 := newEntry("dl4fm", nil)
	d1, a1, m1 := input.EditTo(entry1)
	d2, a2, m2 := input.EditTo(entry2)
	d3, a3, m3 := input.EditTo(entry3)

	assert.Equal(t, distance(1), d1, "first distance")
	assert.Equal(t, accuracy(0.875), a1, "first accuracy")
	assert.Equal(t, MatchingAssembly{MatchingPart{NOP, "DL4"}, MatchingPart{FalseFriend, "W"}}, m1, "first matching assembly")
	assert.True(t, m1.ContainsFalseFriend(), "first entry contains false friend")

	assert.Less(t, d1, d2, "distance order 1")
	assert.Greater(t, a1, a2, "accuracy order 1")
	assert.Equal(t, MatchingAssembly{MatchingPart{NOP, "DL4"}, MatchingPart{Substitute, "K"}}, m2, "second matching assembly")
	assert.False(t, m2.ContainsFalseFriend(), "second entry contains no false friend")

	assert.Equal(t, d2, d3, "distance order 2")
	assert.NotEqual(t, a2, a3, "accuracy order 2")
	assert.Equal(t, MatchingAssembly{MatchingPart{NOP, "DL4"}, MatchingPart{Insert, "F"}, MatchingPart{NOP, "M"}}, m3, "third matching assembly")
	assert.False(t, m3.ContainsFalseFriend(), "third entry contains no false friend")

	match1 := Match{entry1, d1, a1, m1}
	match2 := Match{entry2, d2, a2, m2}
	match3 := Match{entry3, d3, a3, m3}
	assert.True(t, match1.LessThan(match2), "match order 1")
	assert.True(t, match1.LessThan(match3), "match order 2")
}
