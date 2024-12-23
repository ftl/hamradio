package dxcc

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/latlon"
)

// Read parses a set of DXCC entires from a reader
func Read(in io.Reader) (*Prefixes, error) {
	allPrefixes := NewPrefixes()
	for {
		prefixes, err := readDXCCEntry(in)
		if err == io.EOF {
			break
		} else if err != nil {
			return NewPrefixes(), err
		}
		allPrefixes.Add(prefixes...)
	}
	return allPrefixes, nil
}

func readDXCCEntry(in io.Reader) ([]Prefix, error) {
	lines := bufio.NewReader(in)
	line, err := lines.ReadString('\n')
	if err != nil {
		return []Prefix{}, err
	}
	header, err := parseHeaderLine(line)

	allPrefixes := make([]Prefix, 0, 10)
	lastLine := false
	for !lastLine {
		line, err = lines.ReadString('\n')
		if err != nil {
			return []Prefix{}, err
		}
		var prefixes []Prefix
		prefixes, lastLine, err = parsePrefixesLine(line, header)
		if err != nil {
			return []Prefix{}, err
		}
		allPrefixes = append(allPrefixes, prefixes...)
	}

	return allPrefixes, nil
}

type dxccHeader struct {
	Name             string
	CQZone           CQZone
	ITUZone          ITUZone
	Continent        string
	LatLon           latlon.LatLon
	TimeOffset       TimeOffset
	PrimaryPrefix    string
	NotARRLCompliant bool
}

func parseHeaderLine(line string) (dxccHeader, error) {
	fields := strings.Split(line, ":")
	if len(fields) != 9 {
		return dxccHeader{}, fmt.Errorf("The DXCC header line must have 8 fields, separated by ':' and it must end with a ':'")
	}

	var err error
	header := dxccHeader{}
	header.Name = strings.TrimSpace(fields[0])
	header.CQZone, err = ParseCQZone(fields[1])
	if err != nil {
		return dxccHeader{}, err
	}
	header.ITUZone, err = ParseITUZone(fields[2])
	if err != nil {
		return dxccHeader{}, err
	}
	header.Continent = strings.TrimSpace(fields[3])
	header.LatLon, err = parseLatLon(fields[4], fields[5])
	if err != nil {
		return dxccHeader{}, err
	}
	header.TimeOffset, err = ParseTimeOffset(fields[6])
	if err != nil {
		return dxccHeader{}, err
	}
	header.PrimaryPrefix = strings.TrimSpace(fields[7])
	if strings.HasPrefix(header.PrimaryPrefix, "*") {
		header.PrimaryPrefix = header.PrimaryPrefix[1:]
		header.NotARRLCompliant = true
	}
	return header, nil
}

// ParseCQZone parses the CQ zone information from a string.
func ParseCQZone(s string) (CQZone, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(s), 10, 8)
	if err != nil {
		return CQZone(0), fmt.Errorf("cannot parse CQ zone: %v", err)
	}
	return CQZone(value), nil
}

// ParseITUZone parses the ITU zone information from a string.
func ParseITUZone(s string) (ITUZone, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(s), 10, 8)
	if err != nil {
		return ITUZone(0), fmt.Errorf("cannot parse ITU zone: %v", err)
	}
	return ITUZone(value), err
}

func parseLatLon(latString, lonString string) (latlon.LatLon, error) {
	lat, err := latlon.ParseLat(strings.TrimSpace(latString))
	if err != nil {
		return latlon.LatLon{}, err
	}
	lon, err := latlon.ParseLon(strings.TrimSpace(lonString))
	if err != nil {
		return latlon.LatLon{}, err
	}

	return latlon.NewLatLon(lat, lon*-1), nil
}

// ParseTimeOffset parses the time offset information from a string.
func ParseTimeOffset(s string) (TimeOffset, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return TimeOffset(0), fmt.Errorf("cannot parse TimeOffset: %v", err)
	}
	return TimeOffset(value), err
}

func parsePrefixesLine(line string, header dxccHeader) (prefixes []Prefix, lastLine bool, err error) {
	normalLine := strings.TrimSpace(line)
	lastLine = strings.HasSuffix(normalLine, ";")
	lastIndex := len(normalLine)
	if lastLine || strings.HasSuffix(normalLine, ",") {
		lastIndex--
	}
	values := strings.Split(strings.TrimSpace(normalLine[:lastIndex]), ",")
	prefixes = make([]Prefix, 0, len(values))
	for _, value := range values {
		prefix, prefixError := parsePrefix(value, header)
		if prefixError != nil {
			err = prefixError
			prefixes = make([]Prefix, 0, len(values))
			return
		}
		prefixes = append(prefixes, prefix)
	}
	return
}

func parsePrefix(s string, header dxccHeader) (Prefix, error) {
	prefix := Prefix{
		Name:             header.Name,
		CQZone:           header.CQZone,
		ITUZone:          header.ITUZone,
		Continent:        header.Continent,
		LatLon:           header.LatLon,
		TimeOffset:       header.TimeOffset,
		PrimaryPrefix:    header.PrimaryPrefix,
		NotARRLCompliant: header.NotARRLCompliant,
	}

	startIndex := 0
	if strings.HasPrefix(s, "=") {
		prefix.NeedsExactMatch = true
		startIndex = 1
	}
	endIndex := len(s)
	if i := strings.IndexAny(s, "([<{~"); i > -1 {
		endIndex = i
	}
	prefix.Prefix = s[startIndex:endIndex]

	if err := overrideCQZone(&prefix, s); err != nil {
		return Prefix{}, err
	}
	if err := overrideITUZone(&prefix, s); err != nil {
		return Prefix{}, err
	}
	if err := overrideLatLon(&prefix, s); err != nil {
		return Prefix{}, err
	}
	overrideContinent(&prefix, s)
	if err := overrideTimeOffset(&prefix, s); err != nil {
		return Prefix{}, err
	}

	return prefix, nil
}

var overrideCQZoneExpression = regexp.MustCompile("\\(([0-9]+)\\)")

func overrideCQZone(prefix *Prefix, s string) (err error) {
	matches := overrideCQZoneExpression.FindStringSubmatch(s)
	if matches != nil {
		prefix.CQZone, err = ParseCQZone(matches[1])
	}
	return
}

var overrideITUZoneExpression = regexp.MustCompile("\\[([0-9]+)\\]")

func overrideITUZone(prefix *Prefix, s string) (err error) {
	matches := overrideITUZoneExpression.FindStringSubmatch(s)
	if matches != nil {
		prefix.ITUZone, err = ParseITUZone(matches[1])
	}
	return
}

var overrideLatLonExpression = regexp.MustCompile("<(-?[0-9]+(?:\\.[0-9]+)?)/(-?[0-9]+(?:\\.[0-9]+)?)>")

func overrideLatLon(prefix *Prefix, s string) (err error) {
	matches := overrideLatLonExpression.FindStringSubmatch(s)
	if matches != nil {
		prefix.LatLon, err = parseLatLon(matches[1], matches[2])
	}
	return
}

var overrideContinentExpression = regexp.MustCompile("\\{(EU|AF|AS|NA|SA|OC)\\}")

func overrideContinent(prefix *Prefix, s string) {
	matches := overrideContinentExpression.FindStringSubmatch(s)
	if matches != nil {
		prefix.Continent = matches[1]
	}
	return
}

var overrideTimeOffsetExpression = regexp.MustCompile("~(-?[0-9]+(?:\\.[0-9]+)?)~")

func overrideTimeOffset(prefix *Prefix, s string) (err error) {
	matches := overrideTimeOffsetExpression.FindStringSubmatch(s)
	if matches != nil {
		prefix.TimeOffset, err = ParseTimeOffset(matches[1])
	}
	return
}
