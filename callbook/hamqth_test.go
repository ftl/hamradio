package callbook

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHamQTH_get(t *testing.T) {
	testServer := httptest.NewServer(serveValidRequest)
	defer testServer.Close()

	hamQTH := NewHamQTH("the_username", "the_password")
	hamQTH.url = testServer.URL

	info, err := hamQTH.get(map[string]string{})
	if err.Error() != "Username or password missing" {
		t.Errorf("connection error: %v", err)
	} else if err == nil {
		t.Errorf("request without parameters should raise an error: %v", info)
	}

	info, err = hamQTH.get(map[string]string{
		"u": hamQTH.Username,
		"p": hamQTH.password,
	})
	if err != nil {
		t.Errorf("failed to send valid request with username and password: %v", err)
	}
	if info.Session == nil || info.Session.SessionID == "" {
		t.Errorf("failed to parse session id: %v", info)
	}

	info, err = hamQTH.get(map[string]string{
		"id":       "the_session",
		"callsign": "the_callsign",
	})
	if err != nil {
		t.Errorf("failed to send valid request with session ID and callsign: %v", err)
	}
	if info.Search == nil {
		t.Errorf("failed to parse search: %v", info)
	}
}

func TestHamQTH_login(t *testing.T) {
	var requestCount int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response hamqthResponse
		if requestCount == 0 {
			response.Session = &hamqthSession{Error: "Wrong user name or password"}
		} else {
			response.Session = &hamqthSession{SessionID: "123"}
		}
		body, _ := xml.Marshal(response)
		w.Write(body)
		requestCount++
	}))
	defer testServer.Close()

	hamQTH := NewHamQTH("the_username", "the_password")
	hamQTH.url = testServer.URL

	err := hamQTH.login()
	if err.Error() != "Wrong user name or password" {
		t.Errorf("connection error: %v", err)
	} else if err == nil {
		t.Errorf("login should faile on first attempt")
	}

	err = hamQTH.login()
	if err != nil {
		t.Errorf("login failed on second attempt: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("failed to request session ID")
	}
	if hamQTH.sessionID != "123" {
		t.Errorf("failed to set received session ID: %v", hamQTH)
	}

	err = hamQTH.login()
	if err != nil {
		t.Errorf("login failed on third attempt: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("hamQTH should not request new session ID if it already has one: %d", requestCount)
	}
}

func TestHamQTH_Lookup(t *testing.T) {
	testServer := httptest.NewServer(serveValidRequest)
	defer testServer.Close()

	hamQTH := NewHamQTH("the_username", "the_password")
	hamQTH.url = testServer.URL

	hamQTH.sessionID = "timeout"
	info, err := hamQTH.Lookup("dl1abc")
	if err != nil {
		t.Error(err)
	}
	if info.Callsign.String() != "DL1ABC" {
		t.Errorf("failed to receive callsign: %q", info.Callsign)
	}
	if info.Name != "the_nick" {
		t.Errorf("failed to receive name: %q", info.Name)
	}
	if info.QTH != "the_qth" {
		t.Errorf("failed to receive QTH: %q", info.QTH)
	}
	if info.Locator.String() != "EM42lm" {
		t.Errorf("failed to receive locator: %v", info.Locator)
	}
}

var serveValidRequest = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("u")
	password := r.URL.Query().Get("p")
	sessionID := r.URL.Query().Get("id")
	callsign := r.URL.Query().Get("callsign")

	var response hamqthResponse
	if username != "" && password != "" {
		response.Session = &hamqthSession{SessionID: "123"}
	} else if sessionID == "timeout" {
		response.Session = &hamqthSession{Error: "Session does not exist or expired"}
	} else if sessionID != "" && callsign != "" {
		response.Search = &hamqthSearch{
			Callsign: strings.ToUpper(callsign),
			Nick:     "the_nick",
			QTH:      "the_qth",
			Grid:     "EM42lm",
		}
	} else {
		response.Session = &hamqthSession{Error: "Username or password missing"}
	}

	body, _ := xml.Marshal(response)
	w.Write(body)
})
