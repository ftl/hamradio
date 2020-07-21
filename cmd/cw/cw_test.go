package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommandLine(t *testing.T) {
	tt := []struct {
		desc            string
		value           []string
		expectedHost    string
		expectedPort    int
		expectedCommand string
		expectedText    string
		invalid         bool
	}{
		{
			desc:  "empty",
			value: []string{},
		},
		{
			desc:            "tune, no host, no port",
			value:           []string{"tune"},
			expectedCommand: "tune",
		},
		{
			desc:            "tune before shortflag host, no port",
			value:           []string{"tune", "-h", "thehost"},
			expectedCommand: "tune",
			expectedHost:    "thehost",
		},
		{
			desc:            "tune after longflag host and shortflag port",
			value:           []string{"--host", "thehost", "-p", "123", "tune"},
			expectedCommand: "tune",
			expectedHost:    "thehost",
			expectedPort:    123,
		},
		{
			desc:            "send text to specific host, no port",
			value:           []string{"--host", "thehost", "send", "cq", "de", "dk0kd"},
			expectedCommand: "send",
			expectedHost:    "thehost",
			expectedText:    "cq de dk0kd",
		},
		{
			desc:    "missing host",
			value:   []string{"tune", "-h"},
			invalid: true,
		},
		{
			desc:    "missing port",
			value:   []string{"tune", "--port"},
			invalid: true,
		},
		{
			desc:    "invalid port",
			value:   []string{"--port", "theport", "tune"},
			invalid: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actualHost, actualPort, actualCommand, actualText, actualErr := parseCommandLine(tc.value)
			if !tc.invalid {
				assert.Equal(t, tc.expectedHost, actualHost)
				assert.Equal(t, tc.expectedPort, actualPort)
				assert.Equal(t, tc.expectedCommand, actualCommand)
				assert.Equal(t, tc.expectedText, actualText)
				assert.NoError(t, actualErr)
			} else {
				assert.Error(t, actualErr)
			}
		})
	}
}
