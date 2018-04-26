// Package callbook allows to retrieve information about a call from various online sources.
// Supported sources: qrz.com, hamqth.com
package callbook

import (
	"net/http"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

// Info contains the information from a callbook service about a callsign.
type Info struct {
	Callsign   *callsign.Callsign
	Name       string
	Address    string
	QTH        string
	Country    string
	Locator    locator.Locator
	LatLon     *latlon.LatLon
	CQZone     dxcc.CQZone
	ITUZone    dxcc.ITUZone
	TimeOffset dxcc.TimeOffset
}

// Callbook defines the Lookup functionality in a callbook.
type Callbook interface {
	Lookup(callsign string) (Info, error)
}

// Factory is a function that creates a new callbook instance from username and password
type Factory func(username, password string) Callbook

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}
