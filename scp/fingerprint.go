package scp

import (
	"strings"
)

type similarity float64

type fingerprint []byte

func (fp fingerprint) Equal(other fingerprint) bool {
	return string(fp) == string(other)
}

func (fp fingerprint) Similar(other fingerprint) similarity {
	union := make(map[byte]bool, len(fp)+len(other))
	for _, b := range fp {
		union[b] = false
	}
	for _, b := range other {
		if _, ok := union[b]; ok {
			union[b] = true
		} else {
			union[b] = false
		}
	}
	all := 0
	same := 0
	for _, value := range union {
		all++
		if value {
			same++
		}
	}
	if all == 0 {
		return 1
	}
	return similarity(same) / similarity(all)
}

func (fp fingerprint) Contains(other fingerprint) bool {
	if len(fp) < len(other) {
		return false
	}

	fpIndex := 0
	for i := 0; i < len(other); i++ {
		foundOther := false
		for j := fpIndex; j < len(fp); j++ {
			if other[i] == fp[j] {
				foundOther = true
				fpIndex = j
			}
		}
		if !foundOther {
			return false
		}
	}

	return true
}

func extractFingerprint(s string) fingerprint {
	bytes := make([]byte, len(s))
	for _, b := range []byte(strings.ToUpper(s)) {
		if !isCallsignChar(b) {
			continue
		}
		bytes = append(bytes, b)
	}
	return newFingerprint(bytes...)
}

func newFingerprint(bytes ...byte) fingerprint {
	normalized := fingerprint{}
	var last byte
	for _, b := range bytes {
		if last == b {
			continue
		}
		normalized = append(normalized, b)
		last = b
	}
	return normalized
}

func isCallsignChar(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z':
		return true
	case b >= '0' && b <= '9':
		return true
	}
	return false
}
