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

func (e entry) EditTo(o entry) (distance, accuracy, EditScript) {
	matrix := levenshtein.MatrixForStrings([]rune(e.s), []rune(o.s), levenshtein.DefaultOptions)

	dist := levenshtein.DistanceForMatrix(matrix)
	ratio := levenshtein.RatioForMatrix(matrix)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}
	script := levenshtein.EditScriptForMatrix(matrix, levenshtein.DefaultOptions)
	editScript := newEditScript(e.s, o.s, script)

	return distance(dist), accuracy(ratio), editScript
}

type EditOperation int

const (
	NOP EditOperation = iota
	Insert
	Delete
	Substitute
)

type Edit struct {
	OP EditOperation
	S  string
}

type EditScript []Edit

func newEditScript(source, target string, script levenshtein.EditScript) EditScript {
	rawScript := make(EditScript, 0, len(script))

	lastEdit := Edit{NOP, ""}
	sourceIndex := 0
	targetIndex := 0
	var currentEdit Edit
	for _, lop := range script {
		switch lop {
		case levenshtein.Match:
			currentEdit = Edit{NOP, string(source[sourceIndex])}
			sourceIndex++
			targetIndex++
		case levenshtein.Ins:
			currentEdit = Edit{Insert, string(target[targetIndex])}
			targetIndex++
		case levenshtein.Del:
			currentEdit = Edit{Delete, string(source[sourceIndex])}
			sourceIndex++
		}

		if lastEdit.OP == currentEdit.OP {
			lastEdit.S += currentEdit.S
		} else {
			if len(lastEdit.S) > 0 {
				rawScript = append(rawScript, lastEdit)
			}
			lastEdit = currentEdit
		}
	}

	if lastEdit.OP == currentEdit.OP && len(lastEdit.S) > 0 {
		rawScript = append(rawScript, lastEdit)
	}

	if len(rawScript) == 0 {
		return nil
	}

	result := make(EditScript, 0, len(rawScript))
	result = append(result, rawScript[0])
	for i := 1; i < len(rawScript); i++ {
		lastEdit = result[len(result)-1]
		currentEdit = rawScript[i]
		if lastEdit.OP != Insert || currentEdit.OP != Delete {
			result = append(result, currentEdit)
			continue
		}

		lastLen := len(lastEdit.S)
		currentLen := len(currentEdit.S)
		if lastLen > currentLen {
			result[len(result)-1] = Edit{Substitute, lastEdit.S[:currentLen]}
			result = append(result, Edit{Insert, lastEdit.S[currentLen:]})
			continue
		}
		if lastLen < currentLen {
			result[len(result)-1] = Edit{Substitute, lastEdit.S}
			result = append(result, Edit{Delete, currentEdit.S[currentLen:]})
			continue
		}
		result[len(result)-1] = Edit{Substitute, lastEdit.S}
	}

	return result
}

func (s EditScript) String() string {
	var result string
	for _, e := range s {
		result += e.S
	}
	return result
}

func (s EditScript) LongestMatch() int {
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
