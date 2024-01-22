// Package callsign implements a representation and handling of callsigns.
package callsign

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Callsign represents a callsign consisting of a base call, a prefix, a suffix and the working condition:
// Prefix/BaseCall/Suffix/WorkingCondition: K/DL1ABC/9/p
// Prefix, suffix and working condition are optional and may be empty.
type Callsign struct {
	Prefix           string
	BaseCall         string
	Suffix           string
	WorkingCondition string
}

var parseCallsignExpression = regexp.MustCompile(`\b(?:([A-Z0-9]+)/)?((?:[A-Z]|[A-Z][A-Z]|[0-9][A-Z]|[0-9][A-Z][A-Z])[0-9][A-Z0-9]*[A-Z])(?:/([A-Z0-9]+))?(?:/(P|A|M|MM|AM))?\b`)
var callsignWorkingConditions = map[string]bool{
	"P":  true,
	"A":  true,
	"M":  true,
	"MM": true,
	"AM": true,
}

// MustParse parses a callsign from a string. The function panics if the parsing fails.
func MustParse(s string) Callsign {
	result, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return result
}

// Parse parses a callsign from a string.
func Parse(s string) (Callsign, error) {
	normalString := strings.ToUpper(strings.TrimSpace(s))
	if !parseCallsignExpression.MatchString(normalString) {
		return Callsign{}, fmt.Errorf("%q is not a valid callsign", s)
	}

	matches := parseCallsignExpression.FindAllStringSubmatch(normalString, 1)
	if len(matches[0][0]) != len(normalString) {
		return Callsign{}, fmt.Errorf("%q is not a valid callsign", s)
	}

	callsign := Callsign{
		Prefix:           matches[0][1],
		BaseCall:         matches[0][2],
		Suffix:           matches[0][3],
		WorkingCondition: matches[0][4],
	}

	if _, ok := callsignWorkingConditions[callsign.Suffix]; ok && callsign.WorkingCondition == "" {
		callsign.Suffix, callsign.WorkingCondition = callsign.WorkingCondition, callsign.Suffix
	}

	return callsign, nil
}

func (callsign Callsign) String() string {
	var buffer bytes.Buffer

	if callsign.Prefix != "" {
		buffer.WriteString(callsign.Prefix)
		buffer.WriteString("/")
	}
	buffer.WriteString(callsign.BaseCall)
	if callsign.Suffix != "" {
		buffer.WriteString("/")
		buffer.WriteString(callsign.Suffix)
	}
	if callsign.WorkingCondition != "" {
		buffer.WriteString("/")
		buffer.WriteString(strings.ToLower(callsign.WorkingCondition))
	}
	return buffer.String()
}

// FindAll returns all callsigns that are contained in the given string.
func FindAll(s string) []Callsign {
	normalString := strings.ToUpper(strings.TrimSpace(s))
	matches := parseCallsignExpression.FindAllStringSubmatch(normalString, -1)

	callsigns := make([]Callsign, 0, len(matches))
	for _, match := range matches {
		callsign, err := Parse(match[0])
		if err == nil {
			callsigns = append(callsigns, callsign)
		}
	}
	return callsigns
}
