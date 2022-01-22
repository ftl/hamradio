package scp

import (
	"strings"
)

type fingerprint []byte

func (fp fingerprint) String() string {
	return string(fp)
}

func (fp fingerprint) Equal(other fingerprint) bool {
	return string(fp) == string(other)
}

func extractFingerprint(s string) fingerprint {
	bytes := make([]byte, 0)
	for _, b := range []byte(strings.ToUpper(s)) {
		if !isCallsignChar(b) {
			continue
		}
		bytes = append(bytes, b)
	}
	return fingerprint(bytes)
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
