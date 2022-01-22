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

func (e entry) DistanceTo(o entry) (distance, accuracy) {
	source := []rune(e.s)
	target := []rune(o.s)
	distanceOptions := levenshtein.DefaultOptions
	ratioOptions := levenshtein.DefaultOptions

	dist := levenshtein.DistanceForStrings(source, target, distanceOptions)
	ratio := levenshtein.RatioForStrings(source, target, ratioOptions)
	if ratio > 1 {
		ratio = 1.0 / ratio
	}

	return distance(dist), accuracy(ratio)
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
