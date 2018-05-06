// Package locator implements the handling of maidenhead locators (https://en.wikipedia.org/wiki/Maidenhead_Locator_System)
package locator

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/ftl/hamradio/latlon"
)

// Locator represents a maidenhead locator with up to four pairs of characters.
// This is equivalent to an accuracy of 500m.
type Locator [8]byte

var parseLocatorExpression = regexp.MustCompile("([A-R]{2})(?:([0-9]{2})(?:([A-X]{2})([0-9]{2})?)?)?")

// Parse parses a maidenhead locator from a string.
func Parse(s string) (Locator, error) {
	normalString := strings.ToUpper(strings.TrimSpace(s))
	if !parseLocatorExpression.MatchString(normalString) {
		return Locator{}, fmt.Errorf("%q is not a valid maidenhead locator", s)
	}

	matches := parseLocatorExpression.FindAllStringSubmatch(normalString, 1)
	if len(matches[0][0]) != len(normalString) {
		return Locator{}, fmt.Errorf("%q is not a valid maidenhead locator", s)
	}

	normalizedString := fmt.Sprintf("%s%s%s%s",
		strings.ToUpper(matches[0][1]),
		matches[0][2],
		strings.ToLower(matches[0][3]),
		matches[0][4],
	)

	locator := Locator{}
	copy(locator[:], []byte(normalizedString))

	return locator, nil
}

func (l Locator) String() string {
	return strings.TrimRight(strings.TrimSpace(string(l[:])), "\x00")
}

// IsZero returns true if this locator is zero.
func (l Locator) IsZero() bool {
	return l[0] == 0
}

// ToLatLon converts a maidenhead locator into a pair of latitude and longitude coordinates.
// The coordinates represent the center of the square with the given precision of the locator.
func ToLatLon(locator Locator) latlon.LatLon {
	lonPrecision := 0
	lon := latlon.Longitude(locator[0]-'A') * 20.0
	if locator[2] > 0 {
		lon += latlon.Longitude(locator[2]-'0') * 2.0
		lonPrecision++
	} else {
		lon += 10.0
	}
	if lonPrecision == 1 {
		if locator[4] > 0 {
			lon += latlon.Longitude(locator[4]-'a') * 5.0 / 60
			lonPrecision++
		} else {
			lon += 0.5 * 2.0
		}
	}
	if lonPrecision == 2 {
		if locator[6] > 0 {
			lon += latlon.Longitude(locator[6]-'0') * 0.1 * 5.0 / 60
		} else {
			lon += 0.5 * 5.0 / 60
		}
	}
	lon -= 180

	latPrecision := 0
	lat := latlon.Latitude(locator[1]-'A') * 10
	if locator[3] > 0 {
		lat += latlon.Latitude(locator[3] - '0')
		latPrecision++
	} else {
		lat += 5
	}
	if latPrecision == 1 {
		if locator[5] > 0 {
			lat += latlon.Latitude(locator[5]-'a') * 2.5 / 60
			latPrecision++
		} else {
			lat += 0.5
		}
	}
	if latPrecision == 2 {
		if locator[7] > 0 {
			lat += latlon.Latitude(locator[7]-'0') * 0.1 * 2.5 / 60
			latPrecision++
		} else {
			lat += 0.5 * 2.5 / 60
		}
	}
	lat -= 90

	return latlon.NewLatLon(lat, lon)
}

// LatLonToLocator converts latitude and longitude into a maidenhead locator of the given length.
// The length must be 2, 4, 6, or 8. The returned locator describes a square of the desired precision
// that contains the given coordinates.
func LatLonToLocator(latLon latlon.LatLon, length int) Locator {
	if length%2 == 1 || length < 2 || length > 8 {
		panic("The length of a maidenhead locator must be 2, 4, 6, or 8!")
	}
	lon := float64(latLon.Lon) + 180.0
	lonRemainder := lon - math.Floor(lon)
	lonSubsquare := lonRemainder / 5.0 * 60.0
	lat := float64(latLon.Lat) + 90.0
	latRemainder := lat - math.Floor(lat)
	latSubsquare := latRemainder / 2.5 * 60.0

	locator := Locator{}
	locator[0] = byte('A' + int(lon/20))
	locator[1] = byte('A' + int(lat/10))
	if length > 2 {
		locator[2] = byte('0' + int(lon/2)%10)
		locator[3] = byte('0' + int(lat)%10)
	}
	if length > 4 {
		locator[4] = byte('a' + int(lonSubsquare))
		locator[5] = byte('a' + int(latSubsquare))
	}
	if length > 6 {
		locator[6] = byte('0' + int((lonSubsquare-math.Floor(lonSubsquare))/0.1))
		locator[7] = byte('0' + int((latSubsquare-math.Floor(latSubsquare))/0.1))
	}
	return locator
}

// Distance calculates the great circle distance between two maidenhead locators
// in kilometers using the haversine formula.
// See http://www.movable-type.co.uk/scripts/latlong.html for more details.
func Distance(locator1, locator2 Locator) latlon.Km {
	return latlon.Distance(ToLatLon(locator1), ToLatLon(locator2))
}

// Azimuth calculates the azimuth between two maidenhead locators in degrees.
// See http://www.movable-type.co.uk/scripts/latlong.html for more details.
func Azimuth(locator1, locator2 Locator) latlon.Degrees {
	return latlon.Azimuth(ToLatLon(locator1), ToLatLon(locator2))
}
