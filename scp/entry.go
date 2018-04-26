package scp

import (
	"sort"
)

type entry struct {
	s  string
	fp fingerprint
}

func newEntry(s string) entry {
	return entry{s, extractFingerprint(s)}
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
	sort.Slice(result, func(i, j int) bool {
		return result[i].s < result[j].s
	})
	return result
}
