package scp

import (
	"io"
	"strings"
)

// DefaultFieldSet defines the default field set that is used if no !!Order!! directive is present in the call history file.
var DefaultFieldSet = NewFieldSet("Call", "Name", "Loc1", "Loc2", "Sect", "State", "CK", "BirthDate", "Exch1", "Misc", "UserText", "LastUpdateNote")

const (
	FieldCall     FieldName = "Call"
	FieldUserName FieldName = "Name"
	FieldUserText FieldName = "UserText"
	FieldIgnore   FieldName = ""
)

// ReadCallHistory creates a new Database and fills it from the call history that is read with the given reader.
func ReadCallHistory(r io.Reader) (*Database, error) {
	parser := NewCallHistoryParser()
	result, err := Read(r, parser)
	result.fieldSet = parser.fieldSet
	return result, err
}

// CallHistoryParser is used to parse the entries in a call history file to fill the database.
type CallHistoryParser struct {
	fieldSet FieldSet
}

// NewCallHistoryParser creates a new CallHistoryParser that uses the DefaultFieldSet.
func NewCallHistoryParser() *CallHistoryParser {
	return &CallHistoryParser{
		fieldSet: DefaultFieldSet,
	}
}

// ParseEntry parses the given line and returns the corresponding entry.
// If the line contains other information (for example a comment or a directive),
// this method returns false in the second return value.
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

// FieldSet defines a set of fields used in a call history file.
type FieldSet []FieldName

// NewFieldSet creates a new FieldSet with the given field names.
func NewFieldSet(fieldNames ...string) FieldSet {
	result := make(FieldSet, len(fieldNames))
	for i, rawName := range fieldNames {
		result[i] = FieldName(strings.TrimSpace(rawName))
	}
	return result
}

// IndexOf returns the index of the field with the given name in this FieldSet, or -1 if the field is not in this FieldSet.
func (s FieldSet) IndexOf(field FieldName) int {
	field = FieldName(strings.TrimSpace(string(field)))
	for i, name := range s {
		if name == field {
			return i
		}
	}
	return -1
}

// CallIndex returns the index of the Call field.
func (s FieldSet) CallIndex() int {
	return s.IndexOf(FieldCall)
}

// Get returns the field name at the given index.
func (s FieldSet) Get(index int) FieldName {
	if index < 0 || index >= len(s) {
		return FieldIgnore
	}
	return s[index]
}

// UsableNames returns a slice of usable field names (excluding Call and empty field names).
func (s FieldSet) UsableNames() []FieldName {
	result := make([]FieldName, 0, len(s))
	for _, fieldName := range s {
		if fieldName != FieldIgnore && fieldName != FieldCall {
			result = append(result, fieldName)
		}
	}
	return result
}
