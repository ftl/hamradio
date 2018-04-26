package scp

type entry struct {
	s  string
	fp fingerprint
}

type match struct {
	entry
	a accuracy
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
