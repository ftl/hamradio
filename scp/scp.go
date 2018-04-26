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
	"sort"
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

	matchSet := make(map[string]match)
	allMatches := make([]match, 0)
	for _, b := range fp {
		entrySet, ok := database.items[b]
		if !ok {
			continue
		}
		matches := entrySet.FilterAndMap(func(e entry) interface{} {
			contains, accuracy := e.fp.Contains(fp)
			if !contains {
				return nil
			}
			return match{e, accuracy}
		})
		for _, value := range matches {
			m := value.(match)
			if _, ok := matchSet[m.s]; !ok {
				matchSet[m.s] = m
				allMatches = append(allMatches, m)
			}
		}
	}

	sort.Slice(allMatches, func(i, j int) bool {
		if allMatches[i].a == allMatches[j].a {
			return allMatches[i].s < allMatches[j].s
		}
		return allMatches[i].a > allMatches[j].a
	})

	result := make([]string, 0)
	for _, m := range allMatches {
		result = append(result, m.s)
	}
	return result, nil
}
