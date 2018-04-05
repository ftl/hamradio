package callbook

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestQRZ_get(t *testing.T) {
	testServer := httptest.NewServer(serveValidQRZRequest)
	defer testServer.Close()

	qrz := NewQRZ("the_username", "the_password")
	qrz.url = testServer.URL

	response, err := qrz.get(map[string]string{})
	if err.Error() != "Username/password incorrect" {
		t.Errorf("connection error: %v", err)
	} else if err == nil {
		t.Errorf("request without parameters should raise an error: %v", response)
	}

	response, err = qrz.get(map[string]string{
		"username": qrz.Username,
		"password": qrz.password,
	})
	if err != nil {
		t.Errorf("failed to send valid request with username and password: %v", err)
	}
	if response.Session == nil || response.Session.SessionID == "" {
		t.Errorf("failed to parse session id: %v", response)
	}

	response, err = qrz.get(map[string]string{
		"s":        "the_session",
		"callsign": "the_callsign",
	})
	if err != nil {
		t.Errorf("failed to send valid request with session ID and callsign: %v", err)
	}
	if response.Callsign == nil {
		t.Errorf("failed to parse search: %v", response)
	}
}

func TestQRZ_login(t *testing.T) {
	var requestCount int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response qrzResponse
		timestamp := qrzTimestamp(time.Now())
		if requestCount == 0 {
			response.Session = &qrzSession{Error: "Username/password incorrect ", Timestamp: timestamp}
		} else {
			response.Session = &qrzSession{SessionID: "123", Timestamp: timestamp}
		}
		body, _ := xml.Marshal(response)
		w.Write(body)
		requestCount++
	}))
	defer testServer.Close()

	qrz := NewQRZ("the_username", "the_password")
	qrz.url = testServer.URL

	err := qrz.login()
	if err.Error() != "Username/password incorrect" {
		t.Errorf("connection error: %v", err)
	} else if err == nil {
		t.Errorf("login should faile on first attempt")
	}

	err = qrz.login()
	if err != nil {
		t.Errorf("login failed on second attempt: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("failed to request session ID")
	}
	if qrz.sessionID != "123" {
		t.Errorf("failed to set received session ID: %v", qrz)
	}

	err = qrz.login()
	if err != nil {
		t.Errorf("login failed on third attempt: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("qrz should not request new session ID if it already has one: %d", requestCount)
	}
}

func TestJoin(t *testing.T) {
	testCases := []struct {
		value    []string
		expected string
	}{
		{[]string{"head", "middle", "tail"}, "head, middle, tail"},
		{[]string{"", "middle", "tail"}, "middle, tail"},
		{[]string{"head", "", "tail"}, "head, tail"},
		{[]string{"head", "middle", ""}, "head, middle"},
		{[]string{"head", "", ""}, "head"},
		{[]string{"", "middle", ""}, "middle"},
		{[]string{"", "", "tail"}, "tail"},
		{[]string{"", "", ""}, ""},
	}
	for _, testCase := range testCases {
		actual := join(testCase.value, ", ")
		if actual != testCase.expected {
			t.Errorf("expected %q, but got %q", testCase.expected, actual)
		}
	}
}

func TestQRZ_Lookup(t *testing.T) {
	testServer := httptest.NewServer(serveValidQRZRequest)
	defer testServer.Close()

	qrz := NewQRZ("the_username", "the_password")
	qrz.url = testServer.URL

	qrz.sessionID = "timeout"
	info, err := qrz.Lookup("dl1abc")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if info.Callsign.String() != "DL1ABC" {
		t.Errorf("failed to receive callsign: %q", info.Callsign)
	}
	if info.Name != "the_firstname the_lastname" {
		t.Errorf("failed to receive name: %q", info.Name)
	}
	if info.QTH != "the_street, the_city" {
		t.Errorf("failed to receive QTH: %q", info.QTH)
	}
	if info.Locator.String() != "EM42lm" {
		t.Errorf("failed to receive locator: %v", info.Locator)
	}
}

var serveValidQRZRequest = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	sessionID := r.URL.Query().Get("s")
	callsign := r.URL.Query().Get("callsign")
	timestamp := qrzTimestamp(time.Now())

	var response qrzResponse
	if username != "" && password != "" {
		response.Session = &qrzSession{SessionID: "123", Timestamp: timestamp}
	} else if sessionID == "timeout" {
		response.Session = &qrzSession{Error: "Session Timeout", Timestamp: timestamp}
	} else if sessionID != "" && callsign != "" {
		response.Session = &qrzSession{SessionID: "123", Timestamp: timestamp}
		response.Callsign = &qrzCallsign{
			Callsign:  strings.ToUpper(callsign),
			FirstName: "the_firstname",
			LastName:  "the_lastname",
			Address1:  "the_street",
			Address2:  "the_city",
			Grid:      "EM42lm",
		}
	} else {
		response.Session = &qrzSession{Error: "Username/password incorrect ", Timestamp: timestamp}
	}

	body, _ := xml.Marshal(response)
	w.Write(body)
})
