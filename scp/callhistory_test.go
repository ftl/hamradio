package scp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEntry(t *testing.T) {
	tt := []struct {
		desc     string
		lines    []string
		expected []Entry
	}{
		{
			desc:     "single comment",
			lines:    []string{"# a single comment"},
			expected: []Entry{{}},
		},
		{
			desc:  "entry with default field set",
			lines: []string{"Call,Name,Loc1,Loc2,Sect,State,CK,BirthDate,Exch1,Misc,UserText,LastUpdateNote"},
			expected: []Entry{
				newEntry("Call", FieldValues{
					"Name":           "Name",
					"Loc1":           "Loc1",
					"Loc2":           "Loc2",
					"Sect":           "Sect",
					"State":          "State",
					"CK":             "CK",
					"BirthDate":      "BirthDate",
					"Exch1":          "Exch1",
					"Misc":           "Misc",
					"UserText":       "UserText",
					"LastUpdateNote": "LastUpdateNote",
				}),
			},
		},
		{
			desc:  "an order directive and a matching line",
			lines: []string{"!!Order!!,Call,Name,CK,Sect", "DL1ABC, Klaus ,43,B01"},
			expected: []Entry{
				{},
				newEntry("DL1ABC", FieldValues{
					"Name": "Klaus",
					"CK":   "43",
					"Sect": "B01",
				}),
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			parser := NewCallHistoryParser()
			actualEntries := make([]Entry, 0, len(tc.expected))
			for _, line := range tc.lines {
				actualEntry, _ := parser.ParseEntry(line)
				actualEntries = append(actualEntries, actualEntry)
			}

			assert.Equal(t, tc.expected, actualEntries)
		})
	}
}

func TestLoadCallHistoryFromFile(t *testing.T) {
	file, err := os.Open("testdata/DefaultFieldSet.callhistory")
	require.NoError(t, err)
	defer file.Close()
	database, err := ReadCallHistory(file)
	require.NoError(t, err)

	actual, err := database.FindStrings("dl3ney")

	assert.NoError(t, err)
	assert.Equal(t, []string{"DL3NEY"}, actual)
}
