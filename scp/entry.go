package scp

import (
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type distance int
type accuracy float64

type entry struct {
	s  string
	fp fingerprint
}

func newEntry(s string) entry {
	s = strings.ToUpper(strings.TrimSpace(s))
	return entry{s, extractFingerprint(s)}
}

func (e entry) CompareTo(o entry) (distance, accuracy) {
	matrix := levenshtein.MatrixForStrings([]rune(e.s), []rune(o.s), levenshtein.DefaultOptions)

	dist := levenshtein.DistanceForMatrix(matrix)
	ratio := levenshtein.RatioForMatrix(matrix)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}

	return distance(dist), accuracy(ratio)
}

func (e entry) EditTo(o entry) (distance, accuracy, AnnotatedMatch) {
	matrix := levenshtein.MatrixForStrings([]rune(e.s), []rune(o.s), levenshtein.DefaultOptions)

	dist := levenshtein.DistanceForMatrix(matrix)
	ratio := levenshtein.RatioForMatrix(matrix)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}
	script := levenshtein.EditScriptForMatrix(matrix, levenshtein.DefaultOptions)
	AnnotatedMatch := newAnnotatedMatch(e.s, o.s, script)

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
	OP MatchingOperation
	S  string
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
			lastPart.S += currentPart.S
		} else {
			if len(lastPart.S) > 0 {
				rawScript = append(rawScript, lastPart)
			}
			lastPart = currentPart
		}
	}

	if lastPart.OP == currentPart.OP && len(lastPart.S) > 0 {
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

		lastLen := len(lastPart.S)
		currentLen := len(currentPart.S)
		if lastLen > currentLen {
			result[len(result)-1] = Part{Substitute, lastPart.S[:currentLen]}
			result = append(result, Part{Insert, lastPart.S[currentLen:]})
			continue
		}
		if lastLen < currentLen {
			result[len(result)-1] = Part{Substitute, lastPart.S}
			result = append(result, Part{Delete, currentPart.S[lastLen:]})
			continue
		}
		result[len(result)-1] = Part{Substitute, lastPart.S}
	}

	return result
}

func (s AnnotatedMatch) String() string {
	var result string
	for _, e := range s {
		if e.OP != Delete {
			result += e.S
		}
	}
	return result
}

func (s AnnotatedMatch) LongestPart() int {
	result := 0
	for _, e := range s {
		if e.OP != NOP {
			continue
		}
		if result < len(e.S) {
			result = len(e.S)
		}
	}
	return result
}

type entrySet map[string]entry

func (set *entrySet) Add(entries ...entry) *entrySet {
	for _, e := range entries {
		(*set)[e.s] = e
	}
	return set
}

func (set entrySet) Entries() []entry {
	return set.Filter(func(e entry) bool { return true })
}

func (set entrySet) Filter(filter func(e entry) bool) []entry {
	result := make([]entry, 0, len(set))
	for _, e := range set {
		if filter(e) {
			result = append(result, e)
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
