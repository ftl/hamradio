package scp

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_Read(t *testing.T) {
	const testSCP = `# This is a comment
	2E0AOZ
	2E0BNI
	2E0BPP
	N1MM`

	database, err := Read(strings.NewReader(testSCP), SCPFormat)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedLengths := []struct {
		b      byte
		length int
	}{
		{'0', 3},
		{'1', 1},
		{'2', 3},
		{'A', 1},
		{'B', 2},
		{'E', 3},
		{'I', 1},
		{'N', 2},
		{'M', 1},
		{'O', 1},
		{'P', 1},
		{'Z', 1},
	}
	if len(database.items) != len(expectedLengths) {
		t.Errorf("expected %d buckets, but got %d", len(expectedLengths), len(database.items))
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
	file, err := os.Open("testdata/MASTER.SCP")
	require.NoError(t, err)
	defer file.Close()
	database, err := Read(file, SCPFormat)
	require.NoError(t, err)

	tt := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"D", []string{}},
		{"DB", []string{}},
		{"DBB", []string{"DJ8BB", "DK9BB"}},
		{"DAB", []string{"DK1AB", "DL1ABC", "DL2ABC"}},
		{"D1AB", []string{"DK1AB", "DL1ABC"}},
		{"DLAB", []string{"DL1ABC", "DL2ABC", "DK1AB"}},
		{"DABC", []string{"DL1ABC", "DL2ABC"}},
		{"DK1ABC", []string{"DL1ABC", "DL2ABC"}}}
	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := database.FindStrings(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
