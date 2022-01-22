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
		return e.s == "ABC"
	})
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "ABC", entries[0].s)
}

func TestNewEditScript(t *testing.T) {
	tt := []struct {
		input    string
		entry    string
		expected EditScript
	}{
		{"abcd", "abcd", EditScript{Edit{NOP, "abcd"}}},
		{"abc", "abcd", EditScript{Edit{NOP, "abc"}, Edit{Delete, "d"}}},
		{"abcd", "abc", EditScript{Edit{NOP, "abc"}, Edit{Insert, "d"}}},
		{"efgd", "abcd", EditScript{Edit{Substitute, "efg"}, Edit{NOP, "d"}}},
		{"efghd", "abcd", EditScript{Edit{Substitute, "efg"}, Edit{Insert, "h"}, Edit{NOP, "d"}}},
		{"aefgd", "abcd", EditScript{Edit{NOP, "a"}, Edit{Substitute, "ef"}, Edit{Insert, "g"}, Edit{NOP, "d"}}},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s -> %s", tc.entry, tc.input), func(t *testing.T) {
			matrix := levenshtein.MatrixForStrings([]rune(tc.entry), []rune(tc.input), levenshtein.DefaultOptions)
			script := levenshtein.EditScriptForMatrix(matrix, levenshtein.DefaultOptions)

			actual := newEditScript(tc.entry, tc.input, script)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
