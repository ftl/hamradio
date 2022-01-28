package scp

import (
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type distance int
type accuracy float64

type entry struct {
	key         string
	fingerprint fingerprint
}

func newEntry(key string) entry {
	key = strings.ToUpper(strings.TrimSpace(key))
	return entry{key, extractFingerprint(key)}
}

func (e entry) CompareTo(o entry) (distance, accuracy) {
	matrix := levenshtein.MatrixForStrings([]rune(e.key), []rune(o.key), levenshtein.DefaultOptions)

	dist := levenshtein.DistanceForMatrix(matrix)
	ratio := levenshtein.RatioForMatrix(matrix)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}

	return distance(dist), accuracy(ratio)
}

func (e entry) EditTo(o entry) (distance, accuracy, AnnotatedMatch) {
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

type MatchingOperation int

const (
	NOP MatchingOperation = iota
	Insert
	Delete
	Substitute
)

type Part struct {
	OP    MatchingOperation
	Value string
}

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

type entrySet map[string]entry

func (s *entrySet) Add(entries ...entry) *entrySet {
	for _, entry := range entries {
		(*s)[entry.key] = entry
	}
	return s
}

func (s entrySet) Entries() []entry {
	return s.Filter(func(entry) bool { return true })
}

func (s entrySet) Filter(filter func(entry) bool) []entry {
	result := make([]entry, 0, len(s))
	for _, entry := range s {
		if filter(entry) {
			result = append(result, entry)
		}
	}
	return result
}

func (set entrySet) FilterAndMap(filter func(e entry) interface{}) []interface{} {
	result := make([]interface{}, 0, len(set))
	for _, e := range set {
		if value := filter(e); value != nil {
			result = append(result, value)
		}
	}
	return result
}

func (set entrySet) Do(f func(e entry)) {
	for _, e := range set {
		f(e)
	}
}
