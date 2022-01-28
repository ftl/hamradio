package scp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func TestEntrySet(t *testing.T) {
	e1_1 := newEntry("one")
	e1_2 := newEntry("one")
	e2 := newEntry("two")
	e3 := newEntry("abc")

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

	entries = set.Filter(func(e entry) bool {
		return e.key == "ABC"
	})
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "ABC", entries[0].key)
}

func TestNewAnnotatedMatch(t *testing.T) {
	tt := []struct {
		input    string
		entry    string
		expected AnnotatedMatch
	}{
		{"abcd", "abcd", AnnotatedMatch{Part{NOP, "abcd"}}},
		{"abc", "abcd", AnnotatedMatch{Part{NOP, "abc"}, Part{Insert, "d"}}},
		{"abcd", "abc", AnnotatedMatch{Part{NOP, "abc"}, Part{Delete, "d"}}},
		{"efgd", "abcd", AnnotatedMatch{Part{Substitute, "abc"}, Part{NOP, "d"}}},
		{"efghd", "abcd", AnnotatedMatch{Part{Substitute, "abc"}, Part{Delete, "h"}, Part{NOP, "d"}}},
		{"aefgd", "abcd", AnnotatedMatch{Part{NOP, "a"}, Part{Substitute, "bc"}, Part{Delete, "g"}, Part{NOP, "d"}}},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s -> %s", tc.input, tc.entry), func(t *testing.T) {
			matrix := levenshtein.MatrixForStrings([]rune(tc.input), []rune(tc.entry), levenshtein.DefaultOptions)
			script := levenshtein.EditScriptForMatrix(matrix, levenshtein.DefaultOptions)

			actual := newAnnotatedMatch(tc.input, tc.entry, script)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.entry, actual.String())
		})
	}
}
