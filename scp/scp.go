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

var SCPFormat = EntryParserFunc(func(line string) (Entry, bool) {
	if strings.HasPrefix(line, "#") {
		return Entry{}, false
	}
	return newEntry(line, nil), true
})

type Match struct {
	Entry
	distance distance
	accuracy accuracy
	Assembly MatchingAssembly
}

// LessThan returns true if this match is less than the other based on the default ordering for matches (the better the lesser).
func (m Match) LessThan(o Match) bool {
	mLongestPart := m.Assembly.LongestPart()
	oLongestPart := o.Assembly.LongestPart()
	if mLongestPart != oLongestPart {
		return mLongestPart > oLongestPart
	}
	if m.accuracy != o.accuracy {
		return m.accuracy > o.accuracy
	}
	if m.distance != o.distance {
		return m.distance < o.distance
	}
	if len(m.key) != len(o.key) {
		return len(m.key) < len(o.key)
	}
	return m.key < o.key
}

// Read the database from a reader using the SCP format.
func ReadSCP(r io.Reader) (*Database, error) {
	return Read(r, SCPFormat)
}

// Read the database from a reader unsing the given entry parser.
func Read(r io.Reader, parser EntryParser) (*Database, error) {
	database := &Database{make(map[byte]entrySet)}
	lines := bufio.NewScanner(r)
	for lines.Scan() {
		line := strings.TrimSpace(lines.Text())
		if len(line) == 0 {
			continue
		}
		entry, ok := parser.ParseEntry(line)
		if !ok {
			continue
		}
		for _, b := range entry.fingerprint {
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

// FindStrings returns all strings in database that partially match the given string
func (database Database) FindStrings(s string) ([]string, error) {
	allMatches, err := database.Find(s)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(allMatches))
	for i, m := range allMatches {
		result[i] = m.key
	}

	return result, nil
}

// Find returns all entries in database that are similar to the given string.
func (database Database) Find(s string) ([]Match, error) {
	if len(s) < 3 {
		return nil, nil
	}
	source := newEntry(s, nil)

	matches := make(chan Match, 100)
	merged := make(chan []Match)
	waiter := &sync.WaitGroup{}

	byteMap := make(map[byte]bool)
	for _, b := range source.fingerprint {
		if byteMap[b] {
			continue
		}
		byteMap[b] = true
		entrySet, ok := database.items[b]
		if !ok {
			continue
		}

		waiter.Add(1)
		go findMatches(matches, source, entrySet, waiter)
	}
	go collectMatches(merged, matches)

	waiter.Wait()
	close(matches)
	result := <-merged
	close(merged)
	return result, nil
}

func findMatches(matches chan<- Match, input Entry, entries entrySet, waiter *sync.WaitGroup) {
	defer waiter.Done()
	const accuracyThreshold = 0.65

	entries.Do(func(e Entry) {
		distance, accuracy, assembly := input.EditTo(e)
		if accuracy >= accuracyThreshold {
			matches <- Match{e, distance, accuracy, assembly}
		}
	})
}

func collectMatches(result chan<- []Match, matches <-chan Match) {
	allMatches := make([]Match, 0)
	matchSet := make(map[string]Match)
	for match := range matches {
		if _, ok := matchSet[match.key]; !ok {
			matchSet[match.key] = match
			allMatches = append(allMatches, match)
		}
	}
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].LessThan(allMatches[j])
	})
	result <- allMatches
}
