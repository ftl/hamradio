package dxcc

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/ftl/hamradio/latlon"
)

const (
	dxccHeaderLine        = "Sov Mil Order of Malta:   15:  28:  EU:   41.90:   -12.43:    -1.0:  1A:"
	dxccNonARRLHeaderLine = "Sov Mil Order of Malta:   15:  28:  EU:   41.90:   -12.43:    -1.0:  *1A:"
	dxccPrefixLine        = "    1A,=3D2C,3H0(23)[42],B2A[33],=KC4AAA(39),=KC4AAC[73],CH2(2),DL;"
)

func TestParseHeaderLine(t *testing.T) {
	header, err := parseHeaderLine(dxccHeaderLine)
	if err != nil {
		t.Errorf("parsing failed: %q", err)
		t.FailNow()
	}

	expectedHeader := dxccHeader{
		Name:             "Sov Mil Order of Malta",
		CQZone:           CQZone(15),
		ITUZone:          ITUZone(28),
		Continent:        "EU",
		LatLon:           latlon.LatLon{Lat: 41.9, Lon: 12.43},
		TimeOffset:       TimeOffset(-1.0),
		PrimaryPrefix:    "1A",
		NotARRLCompliant: false,
	}
	if *header != expectedHeader {
		t.Errorf("expected %v, got %v", expectedHeader, header)
	}
}

func TestParseNonARRLHeaderLine(t *testing.T) {
	header, err := parseHeaderLine(dxccNonARRLHeaderLine)
	if err != nil {
		t.Errorf("parsing failed: %q", err)
		t.FailNow()
	}

	if header.PrimaryPrefix != "1A" {
		t.Errorf("expected 1A, got %v", header.PrimaryPrefix)
	}
	if !header.NotARRLCompliant {
		t.Errorf("failed to set NotARRLCompliant")
	}
}

func TestParsePrefix(t *testing.T) {
	testCases := []struct {
		value    string
		expected Prefix
	}{
		{"1A", Prefix{"1A", "Sov Mil Order of Malta", CQZone(15), ITUZone(28), "EU", latlon.LatLon{Lat: 41.9, Lon: 12.43}, TimeOffset(-1), "1A", false, false}},
		{"=3D2C<12.3/45.6>", Prefix{"3D2C", "Sov Mil Order of Malta", CQZone(15), ITUZone(28), "EU", latlon.LatLon{Lat: 12.3, Lon: -45.6}, TimeOffset(-1), "1A", true, false}},
		{"3H0(23)[42]", Prefix{"3H0", "Sov Mil Order of Malta", CQZone(23), ITUZone(42), "EU", latlon.LatLon{Lat: 41.9, Lon: 12.43}, TimeOffset(-1), "1A", false, false}},
		{"B2A[33]", Prefix{"B2A", "Sov Mil Order of Malta", CQZone(15), ITUZone(33), "EU", latlon.LatLon{Lat: 41.9, Lon: 12.43}, TimeOffset(-1), "1A", false, false}},
		{"=KC4AAA(39){SA}", Prefix{"KC4AAA", "Sov Mil Order of Malta", CQZone(39), ITUZone(28), "SA", latlon.LatLon{Lat: 41.9, Lon: 12.43}, TimeOffset(-1), "1A", true, false}},
		{"=KC4AAC[73]{SA}", Prefix{"KC4AAC", "Sov Mil Order of Malta", CQZone(15), ITUZone(73), "SA", latlon.LatLon{Lat: 41.9, Lon: 12.43}, TimeOffset(-1), "1A", true, false}},
		{"CH2(2)~-2.5~", Prefix{"CH2", "Sov Mil Order of Malta", CQZone(2), ITUZone(28), "EU", latlon.LatLon{Lat: 41.9, Lon: 12.43}, TimeOffset(-2.5), "1A", false, false}},
	}

	header, _ := parseHeaderLine(dxccHeaderLine)
	for _, testCase := range testCases {
		actual, err := parsePrefix(testCase.value, header)
		if err != nil {
			t.Errorf("parsing of %q failed: %v", testCase.value, err)
			continue
		}
		if *actual != testCase.expected {
			t.Errorf("expected %v, but got %v", testCase.expected, actual)
		}
	}
}

func TestParsePrefixesLine_LastLine(t *testing.T) {
	header, _ := parseHeaderLine(dxccHeaderLine)

	prefixes, lastLine, err := parsePrefixesLine(dxccPrefixLine, header)
	if err != nil {
		t.Errorf("parsing failed: %q", err)
		t.FailNow()
	}

	if !lastLine {
		t.Errorf("should be the last line")
	}

	if len(prefixes) != 8 {
		t.Errorf("expected 8 entries, but got %d", len(prefixes))
	}
}

func TestParsePrefixesLine_IntermediateLines(t *testing.T) {
	header, _ := parseHeaderLine(dxccHeaderLine)

	line := dxccPrefixLine[:len(dxccPrefixLine)-1]
	prefixes, lastLine, _ := parsePrefixesLine(line, header)
	if lastLine {
		t.Errorf("should not be the last line")
	}

	if prefixes[7].Prefix != "DL" {
		t.Errorf("expected DL, but got %s", prefixes[7].Prefix)
	}

	line = line + ","
	prefixes, lastLine, _ = parsePrefixesLine(line, header)
	if lastLine {
		t.Errorf("should not be the last line")
	}
}

func TestParseDXCCEntry(t *testing.T) {
	entry := dxccHeaderLine + "\n" + dxccPrefixLine + "\n"
	in := strings.NewReader(entry)

	infos, err := readDXCCEntry(in)
	if err != nil {
		t.Errorf("parsing failed: %q", err)
		t.FailNow()
	}

	if len(infos) != 8 {
		t.Errorf("expected 8, but got %d", len(infos))
	}
}

func TestRead(t *testing.T) {
	file, err := os.Open("./testdata/cty.dat")
	if err != nil {
		t.Errorf("open failed: %v", err)
		t.FailNow()
	}
	defer file.Close()
	in := bufio.NewReader(file)

	prefixes, err := Read(in)
	if err != nil {
		t.Errorf("parsing failed: %v", err)
	}
	if len(prefixes.items) != 5517 {
		t.Errorf("expected 5517, but got %d prefixes in file", len(prefixes.items))
	}
}
