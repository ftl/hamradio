package callbook

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

// HamQTH represents a connection to hamqth.com with a certain user account.
// For more information about the API see https://www.hamqth.com/developers.php.
type HamQTH struct {
	Username  string
	password  string
	sessionID string
	url       string
}

// NewHamQTH creates a new HamQTH instance with the given username and password.
func NewHamQTH(username, password string) *HamQTH {
	return &HamQTH{Username: username, password: password, url: "https://www.hamqth.com/xml.php"}
}

// Lookup looks up information about the given callsign.
func (hamqth *HamQTH) Lookup(callsign string) (Info, error) {
	var response hamqthResponse
	var err error
	for retryCount := 0; retryCount < 2; retryCount++ {
		err = hamqth.login()
		if err != nil {
			return Info{}, err
		}

		response, err = hamqth.get(map[string]string{
			"id":       hamqth.sessionID,
			"callsign": callsign,
			"prg":      "go-hamradio-callbook",
		})
		if err != nil && err.Error() == "Session does not exist or expired" {
			hamqth.sessionID = ""
			continue
		} else if err != nil {
			return Info{}, err
		} else {
			break
		}
	}
	if err != nil {
		return Info{}, err
	}

	info, err := hamqthSearchToInfo(response.Search)
	if err != nil {
		return Info{}, err
	}
	return info, nil
}

func (hamqth *HamQTH) login() error {
	if hamqth.sessionID != "" {
		return nil
	}
	response, err := hamqth.get(map[string]string{
		"u": hamqth.Username,
		"p": hamqth.password,
	})
	if err != nil {
		return err
	}

	if response.Session == nil || response.Session.SessionID == "" {
		return fmt.Errorf("failed to get a session ID from hamqth.com")
	}
	hamqth.sessionID = response.Session.SessionID

	return nil
}

func (hamqth HamQTH) get(params map[string]string) (hamqthResponse, error) {
	request, err := http.NewRequest("GET", hamqth.url, nil)
	if err != nil {
		return hamqthResponse{}, err
	}

	query := request.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	request.URL.RawQuery = query.Encode()

	response, err := httpClient.Do(request)
	if err != nil {
		return hamqthResponse{}, err
	}
	defer response.Body.Close()

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(response.Body)
	if err != nil {
		return hamqthResponse{}, err
	}

	result := hamqthResponse{}
	err = xml.Unmarshal(buffer.Bytes(), &result)
	if err != nil {
		return hamqthResponse{}, err
	}

	if result.Session != nil && result.Session.Error != "" {
		return hamqthResponse{}, fmt.Errorf("%v", strings.TrimSpace(result.Session.Error))
	}

	return result, nil
}

type hamqthResponse struct {
	XMLName xml.Name       `xml:"https://www.hamqth.com HamQTH"`
	Version string         `xml:"version,attr"`
	Session *hamqthSession `xml:"session"`
	Search  *hamqthSearch  `xml:"search"`
}

type hamqthSession struct {
	XMLName   xml.Name `xml:"session"`
	SessionID string   `xml:"session_id"`
	Error     string   `xml:"error"`
}

type hamqthSearch struct {
	XMLName          xml.Name `xml:"search"`
	Callsign         string   `xml:"callsign"`
	Nick             string   `xml:"nick"`
	QTH              string   `xml:"qth"`
	Country          string   `xml:"country"`
	DXCCCountryCode  string   `xml:"adif"`
	ITUZone          string   `xml:"itu"`
	CQZone           string   `xml:"cq"`
	Grid             string   `xml:"grid"`
	Latitude         string   `xml:"latitude"`
	Longitude        string   `xml:"longitude"`
	Continent        string   `xml:"continent"`
	UTCOffset        string   `xml:"utc_offset"`
	AdrName          string   `xml:"adr_name"`
	AdrStreet1       string   `xml:"adr_street1"`
	AdrStreet2       string   `xml:"adr_street2"`
	AdrStreet3       string   `xml:"adr_street3"`
	AdrCity          string   `xml:"adr_city"`
	AdrZIP           string   `xml:"adr_zip"`
	AdrADIFCountryID string   `xml:"adr_adif"`
	District         string   `xml:"district"`
	USState          string   `xml:"us_state"`
	USCounty         string   `xml:"us_county"`
	Oblast           string   `xml:"oblast"`
	DOK              string   `xml:"dok"`
	IOTA             string   `xml:"iota"`
	QSLVia           string   `xml:"qsl_via"`
	Lotw             string   `xml:"lotw"`
	Eqsl             string   `xml:"eqsl"`
	QSLBuro          string   `xml:"qsl"`
	QSLDirect        string   `xml:"qsldirect"`
	Email            string   `xml:"email"`
	Jabber           string   `xml:"jabber"`
	ICQ              string   `xml:"icq"`
	MSN              string   `xml:"msn"`
	Skype            string   `xml:"skype"`
	BirthYear        string   `xml:"birth_year"`
	LicenseYear      string   `xml:"lic_year"`
	Picture          string   `xml:"picture"`
	Facebook         string   `xml:"facebook"`
	Twitter          string   `xml:"twitter"`
	GooglePlus       string   `xml:"gplus"`
	Youtube          string   `xml:"youtube"`
	LinkedIn         string   `xml:"linkedin"`
	Flickr           string   `xml:"flicker"`
	Vimeo            string   `xml:"vimeo"`
}

func hamqthSearchToInfo(h *hamqthSearch) (Info, error) {
	var result Info
	var err error
	result.Callsign, err = callsign.Parse(h.Callsign)
	if err != nil {
		return Info{}, err
	}
	result.Name = h.Nick
	result.Address = join(", ", h.AdrName, h.AdrStreet1, h.AdrStreet2, h.AdrStreet3, h.AdrZIP, h.AdrCity)
	result.QTH = h.QTH
	result.Country = h.Country
	result.Locator, _ = locator.Parse(h.Grid)
	result.LatLon, err = latlon.ParseLatLon(h.Latitude, h.Longitude)
	result.LatLonValid = err == nil
	result.CQZone, _ = dxcc.ParseCQZone(h.CQZone)
	result.ITUZone, _ = dxcc.ParseITUZone(h.ITUZone)
	result.TimeOffset, _ = dxcc.ParseTimeOffset(h.UTCOffset)

	return result, nil
}
