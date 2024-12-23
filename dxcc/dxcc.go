package dxcc

import (
	"strings"
	"unicode/utf8"

	"github.com/ftl/hamradio/latlon"
)

// DefaultURL is the original URL of the cty.dat file: http://www.country-files.com/cty/cty.dat
const DefaultURL = "http://www.country-files.com/cty/cty.dat"

// DefaultLocalFilename is the default name for the file that is used to store the contents of cty.dat locally in the user's home directory.
const DefaultLocalFilename = ".config/hamradio/cty.dat"

// Prefixes contains all DXCC prefixes.
type Prefixes struct {
	items map[string][]Prefix
}

// Prefix contains the information for one specific DXCC prefix.
// The information from the cty.dat file is denormalized: each specific
// prefix carries all the information associated with its primary prefix.
type Prefix struct {
	Prefix           string
	Name             string
	CQZone           CQZone
	ITUZone          ITUZone
	Continent        string
	LatLon           latlon.LatLon
	TimeOffset       TimeOffset
	PrimaryPrefix    string
	NeedsExactMatch  bool
	NotARRLCompliant bool
}

// CQZone represents a CQ zone.
type CQZone int

// ITUZone represents an ITU zone.
type ITUZone int

// TimeOffset represents a time offset to UTC.
type TimeOffset float64

// DefaultPrefixes returns the default Prefixes instance. It optionally loads the latest update on demand.
func DefaultPrefixes(updateOnDemand bool) (*Prefixes, bool, error) {
	localFilename, err := LocalFilename()
	if err != nil {
		return nil, false, err
	}

	updated := false
	if updateOnDemand {
		updated, _ = Update(DefaultURL, localFilename)
	}

	result, err := LoadLocal(localFilename)
	return result, updated, err
}

// NewPrefixes creates a new instance of prefixes.
func NewPrefixes() *Prefixes {
	return &Prefixes{make(map[string][]Prefix)}
}

func (prefixes *Prefixes) Add(newPrefixes ...Prefix) {
	for _, prefix := range newPrefixes {
		key := strings.ToUpper(prefix.Prefix)
		ps, ok := prefixes.items[key]
		if !ok {
			ps = make([]Prefix, 0, 1)
		}
		prefixes.items[key] = append(ps, prefix)
	}
}

// Find returns the best matching prefixes for a given string.
// Since a prefix might be ambiguous, a slice of prefixes that match is returned.
func (prefixes Prefixes) Find(s string) ([]Prefix, bool) {
	normalString := strings.ToUpper(strings.TrimSpace(s))
	isExactMatch := true
	for len(normalString) > 0 {
		if ps, ok := prefixes.items[normalString]; ok {
			result := make([]Prefix, 0, len(ps))
			for _, prefix := range ps {
				if !prefix.NeedsExactMatch || isExactMatch {
					result = append(result, prefix)
				}
			}
			if len(result) > 0 {
				return result, true
			}
		}
		_, lastRuneSize := utf8.DecodeLastRuneInString(normalString)
		normalString = normalString[:len(normalString)-lastRuneSize]
		isExactMatch = false
	}

	return []Prefix{}, false
}
