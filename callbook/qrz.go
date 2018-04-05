package callbook

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

// QRZ represents a connection to qrz.com with a certain user account.
// For more information about the API see https://www.qrz.com/page/current_spec.html.
type QRZ struct {
	Username  string
	password  string
	sessionID string
	url       string
}

// NewQRZ creates a new QRZ instance with the given username and password.
func NewQRZ(username, password string) *QRZ {
	return &QRZ{Username: username, password: password, url: "https://xmldata.qrz.com/xml/current/"}
}

// Lookup looks up information about the given callsign.
func (qrz *QRZ) Lookup(callsign string) (*Info, error) {
	var response *qrzResponse
	var err error
	for retryCount := 0; retryCount < 2; retryCount++ {
		err = qrz.login()
		if err != nil {
			return &Info{}, err
		}

		response, err = qrz.get(map[string]string{
			"s":        qrz.sessionID,
			"callsign": callsign,
			"agent":    "go-hamradio-callbook",
		})
		if err != nil && err.Error() == "Session Timeout" {
			qrz.sessionID = ""
			continue
		} else if err != nil {
			return &Info{}, err
		} else {
			break
		}
	}
	if err != nil {
		return &Info{}, err
	}

	info, err := qrzCallsignToInfo(response.Callsign)
	if err != nil {
		return &Info{}, err
	}
	return info, nil
}

func (qrz *QRZ) login() error {
	if qrz.sessionID != "" {
		return nil
	}
	response, err := qrz.get(map[string]string{
		"username": qrz.Username,
		"password": qrz.password,
	})
	if err != nil {
		return err
	}

	if response.Session == nil || response.Session.SessionID == "" {
		return fmt.Errorf("failed to get a session ID from qrz.com")
	}
	qrz.sessionID = response.Session.SessionID

	return nil
}

func (qrz *QRZ) get(params map[string]string) (*qrzResponse, error) {
	request, err := http.NewRequest("GET", qrz.url, nil)
	if err != nil {
		return new(qrzResponse), err
	}

	query := request.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	request.URL.RawQuery = query.Encode()

	response, err := httpClient.Do(request)
	if err != nil {
		return new(qrzResponse), err
	}
	defer response.Body.Close()

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(response.Body)
	if err != nil {
		return new(qrzResponse), err
	}

	result := new(qrzResponse)
	err = xml.Unmarshal(buffer.Bytes(), result)
	if err != nil {
		return new(qrzResponse), err
	}

	if result.Session != nil && result.Session.Error != "" {
		return new(qrzResponse), fmt.Errorf("%v", strings.TrimSpace(result.Session.Error))
	}

	return result, nil
}

type qrzResponse struct {
	XMLName  xml.Name     `xml:"http://xmldata.qrz.com QRZDatabase"`
	Version  string       `xml:"version,attr"`
	Session  *qrzSession  `xml:"Session"`
	Callsign *qrzCallsign `xml:"Callsign"`
}

type qrzSession struct {
	XMLName                xml.Name     `xml:"Session"`
	SessionID              string       `xml:"Key"`
	Count                  int          `xml:"Count"`
	SubscriptionExpiration string       `xml:"SubExp"`
	Timestamp              qrzTimestamp `xml:"GMTime"`
	Message                string       `xml:"Message"`
	Error                  string       `xml:"Error"`
}

type qrzCallsign struct {
	XMLName               xml.Name `xml:"Callsign"`
	Callsign              string   `xml:"call"`
	Aliases               string   `xml:"aliases"`
	FirstName             string   `xml:"fname"`
	LastName              string   `xml:"name"`
	Address1              string   `xml:"addr1"`
	Address2              string   `xml:"addr2"`
	ZIP                   string   `xml:"zip"`
	Country               string   `xml:"country"`
	DXCCCountryCode       string   `xml:"ccode"`
	DXCCCountryName       string   `xml:"land"`
	ITUZone               string   `xml:"ituzone"`
	CQZone                string   `xml:"cqzone"`
	Grid                  string   `xml:"grid"`
	Latitude              string   `xml:"lat"`
	Longitude             string   `xml:"lon"`
	LocationSource        string   `xml:"geoloc"`
	TimeZone              string   `xml:"TimeZone"`
	UTCOffset             string   `xml:"GMTOffset"`
	USState               string   `xml:"state"`
	USCounty              string   `xml:"county"`
	FIPS                  string   `xml:"fips"`
	IOTA                  string   `xml:"iota"`
	LicenseEffectiveDate  string   `xml:"efdate"`
	LicenseExpirationDate string   `xml:"expdate"`
	PreviousCallsign      string   `xml:"p_call"`
	LicenseClass          string   `xml:"class"`
	LicenseTypeCodes      string   `xml:"codes"`
	QSLManager            string   `xml:"qslmgr"`
	Email                 string   `xml:"email"`
	Website               string   `xml:"url"`
	QRZPageViews          string   `xml:"u_views"`
	BioLength             string   `xml:"bio"`
	BioLastUpdate         string   `xml:"biodate"`
	ImageURL              string   `xml:"image"`
	ImageInfo             string   `xml:"imageinfo"`
	Serial                string   `xml:"serial"`
	LastUpdate            string   `xml:"moddate"`
	MSA                   string   `xml:"MSA"`
	AreaCode              string   `xml:"AreaCode"`
	DaylightSavingTime    string   `xml:"DST"`
	Lotw                  string   `xml:"lotw"`
	Eqsl                  string   `xml:"eqsl"`
	PaperQSL              string   `xml:"mqsl"`
	DateOfBirth           string   `xml:"born"`
	ManagingUser          string   `xml:"user"`
}

type qrzTimestamp time.Time

// QRZTimeFormat describes the time format used by qrz.com
const QRZTimeFormat = "Mon Jan  2 15:04:05 2006"

func (t *qrzTimestamp) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	err := d.DecodeElement(&value, &start)
	if err != nil {
		return err
	}

	parsedTime, err := time.Parse(QRZTimeFormat, value)
	if err != nil {
		return err
	}

	*t = qrzTimestamp(parsedTime)
	return nil
}

func (t qrzTimestamp) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	value := time.Time(t).Format(QRZTimeFormat)
	e.EncodeElement(value, start)
	return nil
}

func qrzCallsignToInfo(q *qrzCallsign) (*Info, error) {
	var result Info
	var err error
	result.Callsign, err = callsign.Parse(q.Callsign)
	if err != nil {
		return nil, err
	}
	result.Name = join([]string{q.FirstName, q.LastName}, " ")
	result.QTH = join([]string{q.Address1, q.Address2}, ", ")
	result.Country = q.Country
	result.Locator, _ = locator.Parse(q.Grid)
	result.LatLon, _ = latlon.ParseLatLon(q.Latitude, q.Longitude)
	result.CQZone, _ = dxcc.ParseCQZone(q.CQZone)
	result.ITUZone, _ = dxcc.ParseITUZone(q.ITUZone)
	result.TimeOffset, _ = dxcc.ParseTimeOffset(q.UTCOffset)

	return &result, nil
}

func join(values []string, separator string) string {
	var buffer bytes.Buffer
	separatorBefore := ""
	for _, value := range values {
		normalString := strings.TrimSpace(value)
		if normalString == "" {
			continue
		}
		buffer.WriteString(separatorBefore)
		buffer.WriteString(normalString)
		separatorBefore = separator
	}
	return buffer.String()
}
