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
	"sync"
)

// DefaultURL is the original URL of the MASTER.SCP file: http://www.supercheckpartial.com/MASTER.SCP
const DefaultURL = "http://www.supercheckpartial.com/MASTER.SCP"

// DefaultLocalFilename is the default name for the file that is used to store the contents of MASTER.SCP locally in the user's home directory.
const DefaultLocalFilename = ".config/hamradio/MASTER.SCP"

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
	if len(s) < 3 {
		return []string{}, nil
	}
	fp := extractFingerprint(s)

	matches := make(chan match, 100)
	merged := make(chan []match)
	waiter := &sync.WaitGroup{}

	byteMap := make(map[byte]bool)
	for _, b := range fp {
		if byteMap[b] {
			continue
		}
		byteMap[b] = true
		entrySet, ok := database.items[b]
		if !ok {
			continue
		}

		waiter.Add(1)
		go findMatches(matches, entrySet, fp, waiter)
	}
	go collectMatches(merged, matches)

	waiter.Wait()
	close(matches)
	allMatches := <-merged
	close(merged)

	result := make([]string, 0)
	for _, m := range allMatches {
		result = append(result, m.s)
	}
	return result, nil
}

func findMatches(matches chan<- match, entries entrySet, fp fingerprint, waiter *sync.WaitGroup) {
	defer waiter.Done()

	entries.Do(func(e entry) {
		contains, accuracy := e.fp.Contains(fp)
		if contains {
			matches <- match{e, accuracy}
		}
	})
}

func collectMatches(result chan<- []match, matches <-chan match) {
	allMatches := make([]match, 0)
	matchSet := make(map[string]match)
	for match := range matches {
		if _, ok := matchSet[match.s]; !ok {
			matchSet[match.s] = match
			allMatches = append(allMatches, match)
		}
	}
	sort.Slice(allMatches, func(i, j int) bool {
		if allMatches[i].a == allMatches[j].a {
			return allMatches[i].s < allMatches[j].s
		}
		return allMatches[i].a > allMatches[j].a
	})
	result <- allMatches
}
