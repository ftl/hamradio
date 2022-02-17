package scp

import (
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type distance int
type accuracy float64

// Entry represents one entry in a Database.
type Entry struct {
	key         string
	fingerprint fingerprint
	fieldValues FieldValues
}

// FieldName defines the name of a field in an Entry.
type FieldName string

// FieldValues contains all fields of an Entry and their corresponding values.
type FieldValues map[FieldName]string

func newEntry(key string, fieldValues FieldValues) Entry {
	key = strings.ToUpper(strings.TrimSpace(key))
	return Entry{
		key:         key,
		fingerprint: extractFingerprint(key),
		fieldValues: fieldValues,
	}
}

func (e Entry) String() string {
	return e.key
}

// Key returns the key of this Entry.
func (e Entry) Key() string {
	return e.key
}

// Get the value of the field with the given name.
func (e Entry) Get(field FieldName) string {
	if e.fieldValues == nil {
		return ""
	}
	return e.fieldValues[field]
}

// GetValues returns the values of the fields with the given names as slice.
// The returned slice is of the same size as the number of field names. If a field
// is not populated, the corresponding slice entry is empty.
func (e Entry) GetValues(fields ...FieldName) []string {
	result := make([]string, len(fields))
	for i, field := range fields {
		result[i] = e.Get(field)
	}
	return result
}

// PopulatedFields returns a FieldSet that contains all populated fields of this Entry.
func (e Entry) PopulatedFields() FieldSet {
	result := make(FieldSet, 0, len(e.fieldValues))
	for fieldName := range e.fieldValues {
		result = append(result, fieldName)
	}
	return result
}

var levenshteinOptions = levenshtein.Options{
	InsCost: 1,
	DelCost: 100,
	SubCost: 2,
	Matches: levenshtein.IdenticalRunes,
}

// CompareTo compares this Entry's key with the key of the given Entry. It returns a measure
// of similarity in form of the editing distance and the matching accuracy.
func (e Entry) CompareTo(o Entry) (distance, accuracy) {
	d, a, _ := e.EditTo(o)
	return d, a
}

// EditTo provides the editing distance, matching accuracy, and the given Entry's key as MatchingAssembly
func (e Entry) EditTo(o Entry) (distance, accuracy, MatchingAssembly) {
	matrix := levenshtein.MatrixForStrings([]rune(e.key), []rune(o.key), levenshteinOptions)
	script := levenshtein.EditScriptForMatrix(matrix, levenshteinOptions)
	matchingAssembly := newMatchingAssembly(e.key, o.key, script)

	sourcelength := len(matrix) - 1
	targetlength := len(matrix[0]) - 1
	sum := sourcelength + targetlength

	dist := levenshtein.DistanceForMatrix(matrix)
	if matchingAssembly.ContainsFalseFriend() {
		dist--
	}

	var ratio float64
	if sum != 0 {
		ratio = float64(sum-dist) / float64(sum)
	}

	return distance(dist), accuracy(ratio), matchingAssembly
}

// MatchingOperation represents an editing operation that is applied to a key to transform it into another key.
type MatchingOperation int

const (
	// NOP = no edit required, the part matches exactly
	NOP MatchingOperation = iota
	// Insert this part
	Insert
	// Delete this part
	Delete
	// Substitute this part
	Substitute
	// FalseFriend is a subsitute that is close in CW to this part
	FalseFriend
)

// Part represents a part of a key with the corresponding editing operation.
type MatchingPart struct {
	OP    MatchingOperation
	Value string
}

// MatchingAssembly describes how a certain key matches to another key, using editing operations.
type MatchingAssembly []MatchingPart

func newMatchingAssembly(source, target string, script levenshtein.EditScript) MatchingAssembly {
	rawScript := make(MatchingAssembly, 0, len(script))

	lastPart := MatchingPart{NOP, ""}
	sourceIndex := 0
	targetIndex := 0
	var currentPart MatchingPart
	for _, lop := range script {
		switch lop {
		case levenshtein.Match:
			currentPart = MatchingPart{NOP, string(source[sourceIndex])}
			sourceIndex++
			targetIndex++
		case levenshtein.Ins:
			currentPart = MatchingPart{Insert, string(target[targetIndex])}
			targetIndex++
		case levenshtein.Del:
			currentPart = MatchingPart{Delete, string(source[sourceIndex])}
			sourceIndex++
		case levenshtein.Sub:
			currentPart = MatchingPart{Substitute, string(target[targetIndex])}
			if isFalseFriend(string(source[sourceIndex]), string(target[targetIndex])) {
				currentPart.OP = FalseFriend
			}
			sourceIndex++
			targetIndex++
		}

		if lastPart.OP == currentPart.OP {
			lastPart.Value += currentPart.Value
		} else {
			if len(lastPart.Value) > 0 {
				rawScript = append(rawScript, lastPart)
			}
			lastPart = currentPart
		}
	}

	if lastPart.OP == currentPart.OP && len(lastPart.Value) > 0 {
		rawScript = append(rawScript, lastPart)
	}

	if len(rawScript) == 0 {
		return nil
	}

	result := make(MatchingAssembly, 0, len(rawScript))
	result = append(result, rawScript[0])
	for i := 1; i < len(rawScript); i++ {
		resultIndex := len(result) - 1
		lastPart = result[resultIndex]
		currentPart = rawScript[i]
		if lastPart.OP != Insert || currentPart.OP != Delete {
			result = append(result, currentPart)
			continue
		}

		lastLen := len(lastPart.Value)
		currentLen := len(currentPart.Value)
		var headValue string
		var substitution, tail MatchingPart
		if lastLen > currentLen {
			headValue = currentPart.Value
			substitution = MatchingPart{Substitute, lastPart.Value[:currentLen]}
			tail = MatchingPart{Insert, lastPart.Value[currentLen:]}
		} else if lastLen < currentLen {
			headValue = currentPart.Value[:lastLen]
			substitution = MatchingPart{Substitute, lastPart.Value}
			tail = MatchingPart{Delete, currentPart.Value[lastLen:]}
		} else {
			headValue = currentPart.Value
			substitution = MatchingPart{Substitute, lastPart.Value}
		}
		if isFalseFriend(headValue, substitution.Value) {
			substitution.OP = FalseFriend
		}
		result[resultIndex] = substitution
		if tail.OP != NOP {
			result = append(result, tail)
		}
	}

	return result
}

func (m MatchingAssembly) String() string {
	var result string
	for _, e := range m {
		if e.OP != Delete {
			result += e.Value
		}
	}
	return result
}

// LongestPart returns the length of the longest matching part.
func (m MatchingAssembly) LongestPart() int {
	result := 0
	for _, e := range m {
		if e.OP != NOP {
			continue
		}
		if result < len(e.Value) {
			result = len(e.Value)
		}
	}
	return result
}

// ContainsFalseFriend indicates if this matching assembly contains a false friend.
func (m MatchingAssembly) ContainsFalseFriend() bool {
	for _, e := range m {
		if e.OP == FalseFriend {
			return true
		}
	}
	return false
}

var falseFriends = map[string][]string{
	"b": {"d", "6"},
	"d": {"n", "b"},
	"h": {"s", "5"},
	"j": {"1"},
	"s": {"h"},
	"v": {"4"},
	"1": {"j"},
	"4": {"v"},
	"5": {"h"},
	"6": {"b"},
}

func isFalseFriend(s, t string) bool {
	s, t = strings.ToLower(s), strings.ToLower(t)
	for _, friend := range falseFriends[s] {
		if friend == t {
			return true
		}
	}
	return false
}

// EntryParser defines an abstraction for parsing a single entry from a given line in a file.
type EntryParser interface {
	ParseEntry(string) (Entry, bool)
}

// EntryParserFunc wraps a matching function into the EntryParser interface
type EntryParserFunc func(string) (Entry, bool)

func (f EntryParserFunc) ParseEntry(line string) (Entry, bool) {
	return f(line)
}

type entrySet map[string]Entry

func (s *entrySet) Add(entries ...Entry) *entrySet {
	for _, entry := range entries {
		(*s)[entry.key] = entry
	}
	return s
}

func (s entrySet) Entries() []Entry {
	return s.Filter(func(Entry) bool { return true })
}

func (s entrySet) Filter(filter func(Entry) bool) []Entry {
	result := make([]Entry, 0, len(s))
	for _, entry := range s {
		if filter(entry) {
			result = append(result, entry)
		}
	}
	return result
}

func (set entrySet) FilterAndMap(filter func(e Entry) interface{}) []interface{} {
	result := make([]interface{}, 0, len(set))
	for _, e := range set {
		if value := filter(e); value != nil {
			result = append(result, value)
		}
	}
	return result
}

func (set entrySet) Do(f func(e Entry)) {
	for _, e := range set {
		f(e)
	}
}
