package scp

import (
	"strings"
)

type similarity float64
type accuracy float64

type fingerprint []byte

func (fp fingerprint) String() string {
	return string(fp)
}

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

func matchAccuracy(fp1, fp2 fingerprint) accuracy {
	if len(fp1) < len(fp2) {
		fp1, fp2 = fp2, fp1
	} else if len(fp1) == 0 && len(fp2) == 0 {
		return accuracy(1)
	}

	matchStart := -1
	matchEnd := -1
	fpIndex := 0
	for i := 0; i < len(fp2); i++ {
		foundOther := false
		for j := fpIndex; j < len(fp1); j++ {
			if fp2[i] == fp1[j] {
				foundOther = true
				if i == 0 {
					matchStart = j
				}
				matchEnd = j
				fpIndex = j + 1
				break
			}
		}
		if !foundOther {
			return accuracy(0)
		}
	}

	matchLength := (matchEnd - matchStart) + 1
	if matchLength < len(fp2) {
		return accuracy(0)
	}
	return accuracy(len(fp2)) / accuracy(matchLength)
}

func (fp fingerprint) Contains(other fingerprint) (bool, accuracy) {
	if len(fp) < len(other) {
		return false, accuracy(0)
	}

	accuracy := matchAccuracy(fp, other)
	return accuracy != 0, accuracy
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
