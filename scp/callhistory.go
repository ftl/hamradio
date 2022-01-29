package scp

import (
	"io"
	"strings"
)

var DefaultFieldSet = NewFieldSet("Call", "Name", "Loc1", "Loc2", "Sect", "State", "CK", "BirthDate", "Exch1", "Misc", "UserText", "LastUpdateNote")

const (
	FieldCall   FieldName = "Call"
	FieldIgnore FieldName = ""
)

func ReadCallHistory(r io.Reader) (*Database, error) {
	return Read(r, NewCallHistoryParser())
}

type CallHistoryParser struct {
	fieldSet FieldSet
}

func NewCallHistoryParser() *CallHistoryParser {
	return &CallHistoryParser{
		fieldSet: DefaultFieldSet,
	}
}
func (p *CallHistoryParser) ParseEntry(line string) (Entry, bool) {
	switch {
	case strings.HasPrefix(line, "#"):
		return Entry{}, false
	case strings.HasPrefix(line, "!!Order!!,"):
		p.handleFieldSetDirective(line[10:])
		return Entry{}, false
	case strings.HasPrefix(line, "!!"): // any other directives are currently ignored
		return Entry{}, false
	default:
		return p.parseEntry(line)
	}
}

func (p *CallHistoryParser) handleFieldSetDirective(line string) {
	fieldNames := strings.Split(line, ";")
	if len(fieldNames) <= 1 {
		fieldNames = strings.Split(line, ",")
	}

	p.fieldSet = NewFieldSet(fieldNames...)
}

func (p *CallHistoryParser) parseEntry(line string) (Entry, bool) {
	values := strings.Split(line, ";")
	if len(values) <= 1 {
		values = strings.Split(line, ",")
	}
	callIndex := p.fieldSet.CallIndex()
	if callIndex < 0 || callIndex >= len(values) {
		return Entry{}, false
	}
	key := strings.TrimSpace(values[callIndex])
	fieldValues := make(FieldValues)
	for i, value := range values {
		fieldName := p.fieldSet.Get(i)
		if fieldName == FieldCall {
			continue
		}
		if fieldName == FieldIgnore {
			continue
		}
		fieldValues[fieldName] = strings.TrimSpace(value)
	}
	return newEntry(key, fieldValues), true
}

type FieldSet []FieldName

func NewFieldSet(fieldNames ...string) FieldSet {
	result := make(FieldSet, len(fieldNames))
	for i, rawName := range fieldNames {
		result[i] = FieldName(strings.TrimSpace(rawName))
	}
	return result
}

func (s FieldSet) IndexOf(field FieldName) int {
	field = FieldName(strings.TrimSpace(string(field)))
	for i, name := range s {
		if name == field {
			return i
		}
	}
	return -1
}

func (s FieldSet) CallIndex() int {
	return s.IndexOf(FieldCall)
}

func (s FieldSet) Get(index int) FieldName {
	if index < 0 || index >= len(s) {
		return FieldIgnore
	}
	return s[index]
}
