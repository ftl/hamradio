package callsign

import (
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		callsign string
		expected Callsign
		valid    bool
	}{
		{"K/DL1ABC/9/P", Callsign{Prefix: "K", BaseCall: "DL1ABC", Suffix: "9", WorkingCondition: "P"}, true},
		{"DL1ABC/9/P", Callsign{Prefix: "", BaseCall: "DL1ABC", Suffix: "9", WorkingCondition: "P"}, true},
		{"DL1ABC/9", Callsign{Prefix: "", BaseCall: "DL1ABC", Suffix: "9", WorkingCondition: ""}, true},
		{"DL1ABC/P", Callsign{Prefix: "", BaseCall: "DL1ABC", Suffix: "", WorkingCondition: "P"}, true},
		{"K/DL1ABC/P", Callsign{Prefix: "K", BaseCall: "DL1ABC", Suffix: "", WorkingCondition: "P"}, true},
		{"K/DL1ABC", Callsign{Prefix: "K", BaseCall: "DL1ABC", Suffix: "", WorkingCondition: ""}, true},
		{"K/DL1ABC/9", Callsign{Prefix: "K", BaseCall: "DL1ABC", Suffix: "9", WorkingCondition: ""}, true},
		{"DL1ABC", Callsign{Prefix: "", BaseCall: "DL1ABC", Suffix: "", WorkingCondition: ""}, true},
		{"3DA0RU", Callsign{Prefix: "", BaseCall: "3DA0RU", Suffix: "", WorkingCondition: ""}, true},
		{"", Callsign{}, false},
		{"DLABC", Callsign{}, false},
	}

	for _, testCase := range testCases {
		actual, err := Parse(testCase.callsign)
		if testCase.valid {
			if err != nil {
				t.Errorf("parsing failed: %v", err)
				continue
			}
			if actual != testCase.expected {
				t.Errorf("expected %v, but got %v", testCase.expected, actual)
			}
		} else {
			if err == nil {
				t.Errorf("%v should raise an error", testCase.callsign)
			}
		}
	}
}
func TestCallsignToString(t *testing.T) {
	testCases := []struct {
		callsign Callsign
		expected string
	}{
		{Callsign{Prefix: "PRE", BaseCall: "BASE", Suffix: "SUF", WorkingCondition: "WC"}, "PRE/BASE/SUF/wc"},
		{Callsign{Prefix: "", BaseCall: "BASE", Suffix: "SUF", WorkingCondition: "WC"}, "BASE/SUF/wc"},
		{Callsign{Prefix: "", BaseCall: "BASE", Suffix: "SUF", WorkingCondition: ""}, "BASE/SUF"},
		{Callsign{Prefix: "", BaseCall: "BASE", Suffix: "", WorkingCondition: "WC"}, "BASE/wc"},
		{Callsign{Prefix: "PRE", BaseCall: "BASE", Suffix: "", WorkingCondition: "WC"}, "PRE/BASE/wc"},
		{Callsign{Prefix: "PRE", BaseCall: "BASE", Suffix: "", WorkingCondition: ""}, "PRE/BASE"},
		{Callsign{Prefix: "PRE", BaseCall: "BASE", Suffix: "SUF", WorkingCondition: ""}, "PRE/BASE/SUF"},
		{Callsign{Prefix: "", BaseCall: "BASE", Suffix: "", WorkingCondition: ""}, "BASE"},
	}

	for _, testCase := range testCases {
		s := testCase.callsign.String()
		if s != testCase.expected {
			t.Errorf("expected %q, but got %q", testCase.expected, s)
		}
	}
}

func TestFindAll(t *testing.T) {
	testCases := []struct {
		s     string
		count int
	}{
		{"", 0},
		{"DL1ABC", 1},
		{"DL1ABCDL1ABC", 1},
		{"DL1ABC DL1ABC", 2},
	}

	for _, testCase := range testCases {
		found := FindAll(testCase.s)
		if len(found) != testCase.count {
			t.Errorf("expected %d, but got %v", testCase.count, found)
		}
	}
}
