/*
Package scp provides access to the Super Check Partial (http://www.supercheckpartial.com)
database stored in the SCP format. The package also provides functions to download,
store and update a MASTER.SCP file. The default remote location for the MASTER.SCP file
is http://www.supercheckpartial.com/MASTER.SCP.

File Format Description

1. The file is in plain text format (ASCII).
2. Each line contains one callsign.
3. Lines that begin with # are comments that can be ignored.
*/
package scp

import (
	"bufio"
	"io"
	"strings"
)

// DefaultURL is the original URL of the MASTER.SCP file: http://www.supercheckpartial.com/MASTER.SCP
const DefaultURL = "http://www.supercheckpartial.com/MASTER.SCP"

// Database represents the SCP database.
type Database struct {
	items map[byte]entrySet
}

// Read the database from a reader.
func Read(r io.Reader) (*Database, error) {
	database := &Database{make(map[byte]entrySet)}
	lines := bufio.NewScanner(r)
	for lines.Scan() {
		line := lines.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		entry := newEntry(line)
		for _, b := range entry.fp {
			es, ok := database.items[b]
			if !ok {
				es = entrySet{}
			}
			es.Add(entry)
			database.items[b] = es
		}
	}

	return database, nil
}

// Find all strings in database that partially match the given string
func (database Database) Find(s string) ([]string, error) {
	fp := extractFingerprint(s)
	if len(fp) < 3 {
		return []string{}, nil
	}

	resultEntries := entrySet{}
	for _, b := range fp {
		entrySet, ok := database.items[b]
		if !ok {
			continue
		}
		entries := entrySet.Filter(func(e entry) bool {
			return e.fp.Contains(fp)
		})
		resultEntries.Add(entries...)
	}

	result := make([]string, 0)
	for _, entry := range resultEntries.Entries() {
		result = append(result, entry.s)
	}
	return result, nil
}
