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
	QTH        string
	Country    string
	Locator    locator.Locator
	LatLon     *latlon.LatLon
	CQZone     dxcc.CQZone
	ITUZone    dxcc.ITUZone
	TimeOffset dxcc.TimeOffset
}

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}
