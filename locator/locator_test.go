package locator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ftl/hamradio/latlon"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		value, expected string
		valid           bool
	}{
		{"   fn   ", "FN      ", true},
		{"Fn20", "FN20    ", true},
		{"Fn20aX", "FN20ax  ", true},
		{"Fn20Ax12", "FN20ax12", true},
		{"", "", false},
		{"ZZ", "", false},
		{"Fn20Ax12bD34", "", false},
		{"20Ar12bD34", "", false},
	}
	for _, testCase := range testCases {
		l, err := Parse(testCase.value)
		if testCase.valid {
			if err != nil {
				t.Errorf("parsing failed: %q", err)
			}

			s := strings.Replace(string(l[:]), "\x00", " ", -1)
			if s != testCase.expected {
				t.Errorf("got %q", l)
			}
		} else {
			if err == nil {
				t.Errorf("%q should raise an error", testCase.value)
			}
		}
	}
}

func TestToString(t *testing.T) {
	testCases := []struct {
		value, expected string
	}{
		{"fn", "FN"},
		{"Fn20", "FN20"},
		{"Fn20aX", "FN20ax"},
		{"Fn20Ax12", "FN20ax12"},
	}
	for _, testCase := range testCases {
		l, _ := Parse(testCase.value)
		actual := fmt.Sprintf("%v", l)
		if actual != testCase.expected {
			t.Errorf("expected %q, but got %q", testCase.expected, actual)
		}
	}
}

func TestToLatLon(t *testing.T) {
	testCases := []struct {
		value, expected string
	}{
		{"fn", "(45.00000N, 70.00000W)"},
		{"fn20", "(40.50000N, 75.00000W)"},
		{"fn20ab", "(40.06250N, 75.95833W)"},
		{"fn20ab36", "(40.06667N, 75.97500W)"},
	}
	for _, testCase := range testCases {
		l, _ := Parse(testCase.value)
		latLon := ToLatLon(l).String()
		if latLon != testCase.expected {
			t.Errorf("%v: expected %q, but got %q", testCase.value, testCase.expected, latLon)
		}
	}
}

func TestLatLonToLocator(t *testing.T) {
	testCases := []struct {
		value    *latlon.LatLon
		expected string
	}{
		{latlon.NewLatLon(45, -70), "FN"},
		{latlon.NewLatLon(40.5, -75), "FN20"},
		{latlon.NewLatLon(40.06250, -75.95833), "FN20ab"},
		{latlon.NewLatLon(40.06667, -75.97500), "FN20ab36"},
	}
	for _, testCase := range testCases {
		l := LatLonToLocator(testCase.value, len(testCase.expected))
		if l.String() != testCase.expected {
			t.Errorf("%v: expected %q, but got %q", testCase.value, testCase.expected, l)
		}
	}
}
