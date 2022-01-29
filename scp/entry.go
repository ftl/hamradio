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

// CompareTo compares this Entry's key with the key of the given Entry. It returns a measure
// of similarity in form of the editing distance and the matching accuracy.
func (e Entry) CompareTo(o Entry) (distance, accuracy) {
	matrix := levenshtein.MatrixForStrings([]rune(e.key), []rune(o.key), levenshtein.DefaultOptions)

	dist := levenshtein.DistanceForMatrix(matrix)
	ratio := levenshtein.RatioForMatrix(matrix)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}

	return distance(dist), accuracy(ratio)
}

// EditTo provides the editing distance, matching accuracy, and the given Entry's key as AnnotatedMatch
func (e Entry) EditTo(o Entry) (distance, accuracy, AnnotatedMatch) {
	matrix := levenshtein.MatrixForStrings([]rune(e.key), []rune(o.key), levenshtein.DefaultOptions)

	dist := levenshtein.DistanceForMatrix(matrix)
	ratio := levenshtein.RatioForMatrix(matrix)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}
	script := levenshtein.EditScriptForMatrix(matrix, levenshtein.DefaultOptions)
	AnnotatedMatch := newAnnotatedMatch(e.key, o.key, script)

	return distance(dist), accuracy(ratio), AnnotatedMatch
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
)

// Part represents a part of a key with the corresponding editing operation.
type Part struct {
	OP    MatchingOperation
	Value string
}

// AnnotatedMatch describes how a certain key matches to another key, using editing operations.
type AnnotatedMatch []Part

func newAnnotatedMatch(source, target string, script levenshtein.EditScript) AnnotatedMatch {
	rawScript := make(AnnotatedMatch, 0, len(script))

	lastPart := Part{NOP, ""}
	sourceIndex := 0
	targetIndex := 0
	var currentPart Part
	for _, lop := range script {
		switch lop {
		case levenshtein.Match:
			currentPart = Part{NOP, string(source[sourceIndex])}
			sourceIndex++
			targetIndex++
		case levenshtein.Ins:
			currentPart = Part{Insert, string(target[targetIndex])}
			targetIndex++
		case levenshtein.Del:
			currentPart = Part{Delete, string(source[sourceIndex])}
			sourceIndex++
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

	result := make(AnnotatedMatch, 0, len(rawScript))
	result = append(result, rawScript[0])
	for i := 1; i < len(rawScript); i++ {
		lastPart = result[len(result)-1]
		currentPart = rawScript[i]
		if lastPart.OP != Insert || currentPart.OP != Delete {
			result = append(result, currentPart)
			continue
		}

		lastLen := len(lastPart.Value)
		currentLen := len(currentPart.Value)
		if lastLen > currentLen {
			result[len(result)-1] = Part{Substitute, lastPart.Value[:currentLen]}
			result = append(result, Part{Insert, lastPart.Value[currentLen:]})
			continue
		}
		if lastLen < currentLen {
			result[len(result)-1] = Part{Substitute, lastPart.Value}
			result = append(result, Part{Delete, currentPart.Value[lastLen:]})
			continue
		}
		result[len(result)-1] = Part{Substitute, lastPart.Value}
	}

	return result
}

func (m AnnotatedMatch) String() string {
	var result string
	for _, e := range m {
		if e.OP != Delete {
			result += e.Value
		}
	}
	return result
}

// LongestPart returns the length of the longest matching part.
func (m AnnotatedMatch) LongestPart() int {
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
