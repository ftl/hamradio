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

func TestFingerprint_Similar(t *testing.T) {
	testCases := []struct {
		value1, value2 string
		expected       similarity
	}{
		{"", "", 1},
		{"abc", "abc", 1},
		{"abc", "def", 0},
		{"def", "abc", 0},
		{"abc", "bcd", 0.5},
	}
	for _, testCase := range testCases {
		fp1 := extractFingerprint(testCase.value1)
		fp2 := extractFingerprint(testCase.value2)
		actual := fp1.Similar(fp2)
		if actual != testCase.expected {
			t.Errorf("%v, %v expected %v, but got %v", fp1, fp2, testCase.expected, actual)
		}
	}
}

func TestFingerprint_Contains(t *testing.T) {
	testCases := []struct {
		value1, value2 string
		expected       bool
	}{
		{"", "", true},
		{"abc", "abc", true},
		{"abcd", "abc", true},
		{"abc", "abcd", false},
		{"abc", "def", false},
		{"abc", "bcd", false},
	}
	for _, testCase := range testCases {
		fp1 := extractFingerprint(testCase.value1)
		fp2 := extractFingerprint(testCase.value2)
		actual := fp1.Contains(fp2)
		if actual != testCase.expected {
			t.Errorf("%v, %v expected %t, but got %t", fp1, fp2, testCase.expected, actual)
		}
	}
}

func TestNewFingerprint(t *testing.T) {
	testCases := []struct {
		value    []byte
		expected fingerprint
	}{
		{
			[]byte{},
			fingerprint{},
		},
		{
			[]byte{'a', 'b', 'c'},
			fingerprint{'a', 'b', 'c'},
		},
		{
			[]byte{'b', 'c', 'a', 'b'},
			fingerprint{'b', 'c', 'a', 'b'},
		},
		{
			[]byte{'b', 'c', 'a', 'b', 'b', 'c'},
			fingerprint{'b', 'c', 'a', 'b', 'c'},
		},
	}

	for _, testCase := range testCases {
		actual := newFingerprint(testCase.value...)
		if !testCase.expected.Equal(actual) {
			t.Errorf("expected %v, but got %v", testCase.expected, actual)
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
	}
	for _, testCase := range testCases {
		actual := extractFingerprint(testCase.value)
		if !testCase.expected.Equal(actual) {
			t.Errorf("expected %v but got %v", testCase.expected, actual)
		}
	}
}
