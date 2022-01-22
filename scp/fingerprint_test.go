package scp

import (
	"testing"
)

func TestFingerprint_Equal(t *testing.T) {
	testCases := []struct {
		value1, value2 fingerprint
		expected       bool
	}{
		{fingerprint{}, fingerprint{}, true},
		{fingerprint{'a', 'b', 'c'}, fingerprint{'a', 'b', 'c'}, true},
		{fingerprint{'a', 'a', 'b'}, fingerprint{'a', 'b', 'c'}, false},
		{fingerprint{'a', 'b', 'c'}, fingerprint{'a', 'b', 'c', 'd'}, false},
	}
	for _, testCase := range testCases {
		if testCase.value1.Equal(testCase.value2) != testCase.expected {
			t.Errorf("%v == %v failed", testCase.value1, testCase.value2)
		}
	}
}

func TestIsCallsignChar(t *testing.T) {
	testCases := []struct {
		value    byte
		expected bool
	}{
		{'a', false},
		{'B', true},
		{'3', true},
		{' ', false},
		{'/', false},
	}
	for _, testCase := range testCases {
		if isCallsignChar(testCase.value) != testCase.expected {
			t.Errorf("%v failed", testCase.value)
		}
	}
}

func TestExtractFingerprint(t *testing.T) {
	testCases := []struct {
		value    string
		expected fingerprint
	}{
		{"abc", fingerprint{'A', 'B', 'C'}},
		{"abcabc", fingerprint{'A', 'B', 'C', 'A', 'B', 'C'}},
		{"F/DL1ABC/p", fingerprint{'F', 'D', 'L', '1', 'A', 'B', 'C', 'P'}},
		{"EA7/DL1ABC/p", fingerprint{'E', 'A', '7', 'D', 'L', '1', 'A', 'B', 'C', 'P'}},
		{"nmm", fingerprint{'N', 'M', 'M'}},
	}
	for _, testCase := range testCases {
		actual := extractFingerprint(testCase.value)
		if !testCase.expected.Equal(actual) {
			t.Errorf("expected %v but got %v", testCase.expected, actual)
		}
	}
}
