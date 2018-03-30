package latlon

import (
	"fmt"
	"math"
	"testing"
)

func TestNormalizeLat_ShouldTrimAtPlusMinus90(t *testing.T) {
	belowRange := normalizeLat(Latitude(-91.0))
	if belowRange < -90.0 {
		t.Errorf("normalization failed at lower bound: expected -90.0, got %v", belowRange)
	}
	aboveRange := normalizeLat(Latitude(91.0))
	if aboveRange > 90.0 {
		t.Errorf("normalization failed at upper bound: expected 90.0, got %v", aboveRange)
	}
}

func TestNormalizeLon_ShouldModuloPlusMinus180(t *testing.T) {
	belowRange := normalizeLon(-181.0)
	if belowRange != 179.0 {
		t.Errorf("normalization failed at lower bound: expected 179.0, got %v", belowRange)
	}
	aboveRange := normalizeLon(181.0)
	if aboveRange != -179.0 {
		t.Errorf("normalization failed at upper bound: expected -179.0, got %v", aboveRange)
	}
}

func TestNormalizeDegrees_ShouldModulo0To360(t *testing.T) {
	belowRange := normalizeDegrees(-1.0)
	if belowRange != 359.0 {
		t.Errorf("normalization failed at lower bound: expected 359.0, got %v", belowRange)
	}
	aboveRange := normalizeDegrees(361.0)
	if aboveRange != 1.0 {
		t.Errorf("normalization failed at lower bound: expected 1.0, got %v", belowRange)
	}
	inRange := normalizeDegrees(90.0)
	if inRange != 90.0 {
		t.Errorf("normalization failed in range: expected 90.0, got %v", inRange)
	}
}

func TestNewLatLon_ShouldNormalizeLatAndLon(t *testing.T) {
	latLon := NewLatLon(-91, 181)
	if latLon.Lat < -90.0 {
		t.Errorf("should normalize lattitude, but got %v", latLon.Lat)
	}
	if latLon.Lon != -179.0 {
		t.Errorf("should normalize longitude, but got %v", latLon.Lon)
	}
}

func TestLatToString(t *testing.T) {
	s := fmt.Sprintf("%v, %v", Latitude(23.4), Latitude(-23.4))
	if s != "23.40000N, 23.40000S" {
		t.Errorf("got %v", s)
	}
}

func TestLonToString(t *testing.T) {
	s := fmt.Sprintf("%v, %v", Longitude(23.4), Longitude(-23.4))
	if s != "23.40000E, 23.40000W" {
		t.Errorf("got %v", s)
	}
}

func TestLatLonToString(t *testing.T) {
	s := fmt.Sprintf("%v", NewLatLon(12.3, 45.6))
	if s != "(12.30000N, 45.60000E)" {
		t.Errorf("got %v", s)
	}
}

func TestParseLat(t *testing.T) {
	testCases := []struct {
		s        string
		expected float64
		valid    bool
	}{
		{"-23.4", -23.4, true},
		{"23.4S", -23.4, true},
		{"23", 23.0, true},
		{"-23", -23.0, true},
		{"23N", 23.0, true},
		{"23.4N", 23.4, true},
		{"23.4n", 23.4, true},
		{"23.4s", -23.4, true},
		{"", 0, false},
		{"ABCD", 0, false},
		{"2s3.4", 0, false},
		{"23.4X", 0, false},
	}
	for _, testCase := range testCases {
		l, err := ParseLat(testCase.s)
		if testCase.valid && l != Latitude(testCase.expected) {
			t.Errorf("parsing %v produced %v, %v, but expected %v", testCase.s, l, err, testCase.expected)
		} else if !testCase.valid && err == nil {
			t.Errorf("parsing %q should fail", testCase.s)
		}
	}
}

func TestParseLon(t *testing.T) {
	testCases := []struct {
		s        string
		expected float64
		valid    bool
	}{
		{"-23.4", -23.4, true},
		{"23.4W", -23.4, true},
		{"23", 23.0, true},
		{"-23", -23.0, true},
		{"23E", 23.0, true},
		{"23.4E", 23.4, true},
		{"23.4e", 23.4, true},
		{"23.4w", -23.4, true},
		{"", 0, false},
		{"ABCD", 0, false},
		{"2s3.4", 0, false},
		{"23.4X", 0, false},
	}
	for _, testCase := range testCases {
		l, err := ParseLon(testCase.s)
		if testCase.valid && l != Longitude(testCase.expected) {
			t.Errorf("parsing %v produced %v, %v, but expected %v", testCase.s, l, err, testCase.expected)
		} else if !testCase.valid && err == nil {
			t.Errorf("parsing %q should fail", testCase.s)
		}
	}
}

func TestDistance(t *testing.T) {
	testCases := []struct {
		value1, value2 *LatLon
		expected       int
	}{
		{NewLatLon(23.4, 23.4), NewLatLon(-23.4, -23.4), 7254},
		{NewLatLon(49.82, 10.74), NewLatLon(45.85, 40.12), 2221},
	}
	for _, testCase := range testCases {
		actual := Distance(testCase.value1, testCase.value2)
		if math.Abs(float64(actual-Km(testCase.expected))) > 1 {
			t.Errorf("expected %v, but got %v", testCase.expected, actual)
		}
	}
}

func TestAzimuth(t *testing.T) {
	testCases := []struct {
		value1, value2 *LatLon
		expected       int
	}{
		{NewLatLon(23.4, 23.4), NewLatLon(-23.4, -23.4), 227},
		{NewLatLon(49.82, 10.74), NewLatLon(45.85, 40.12), 90},
	}
	for _, testCase := range testCases {
		actual := Azimuth(testCase.value1, testCase.value2)
		if math.Abs(float64(actual-Degrees(testCase.expected))) > 1 {
			t.Errorf("expected %v, but got %v", testCase.expected, actual)
		}
	}
}
