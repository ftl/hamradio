// Package latlon implements handling of geodetic coordinates as latitude and longitude.
package latlon

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Latitude represents a latitudinal value in degrees.
type Latitude float64

// Longitude represents a longitudinal value in degrees.
type Longitude float64

// LatLon contains a pair of coordinates consisting of latitude and longitude.
type LatLon struct {
	Lat Latitude
	Lon Longitude
}

// Km represents kilometers.
type Km float64

// Degrees represents an angle in degrees.
type Degrees float64

func (lat Latitude) String() string {
	value := float64(lat)
	var direction string
	if value >= 0 {
		direction = "N"
	} else {
		direction = "S"
	}
	return fmt.Sprintf("%3.5f%s", math.Abs(value), direction)
}

// ParseLat parses a string into a latitudinal value.
func ParseLat(s string) (Latitude, error) {
	value, err := parseCoordinate(s, "NSns", 'S')
	if err != nil {
		return Latitude(0), fmt.Errorf("Cannot parse latitude: %v", err)
	}
	return normalizeLat(Latitude(value)), nil
}

func parseCoordinate(s string, directionIndicators string, negativeDirection rune) (float64, error) {
	valuePart := s
	directionFactor := 1.0
	if lastRune, size := utf8.DecodeLastRuneInString(strings.ToUpper(s)); strings.Contains(directionIndicators, string(lastRune)) {
		i := len(s) - size
		valuePart = s[:i]
		if lastRune == negativeDirection {
			directionFactor = -1.0
		}
	}

	value, err := strconv.ParseFloat(valuePart, 64)
	if err != nil {
		return 0, err
	}

	value *= directionFactor
	return value, nil
}

func normalizeLat(lat Latitude) Latitude {
	return Latitude(math.Max(-90.0, math.Min(float64(lat), 90.0)))
}

func (lon Longitude) String() string {
	value := float64(lon)
	var direction string
	if value >= 0 {
		direction = "E"
	} else {
		direction = "W"
	}
	return fmt.Sprintf("%.5f%s", math.Abs(value), direction)
}

// ParseLon parses a string into a longitudinal value.
func ParseLon(s string) (Longitude, error) {
	value, err := parseCoordinate(s, "EWew", 'W')
	if err != nil {
		return Longitude(0), fmt.Errorf("Cannot parse longitude: %v", err)
	}
	return normalizeLon(Longitude(value)), nil
}

func normalizeLon(lon Longitude) Longitude {
	if lon < -180 {
		return Longitude(lon + 360.0)
	}
	if lon >= 180 {
		return Longitude(lon - 360.0)
	}
	return lon
}

// ParseLatLon parses a latitude and longitude from two strings and returns the parsed data as LatLon.
func ParseLatLon(latString, lonString string) (*LatLon, error) {
	lat, err := ParseLat(strings.TrimSpace(latString))
	if err != nil {
		return &LatLon{}, err
	}
	lon, err := ParseLon(strings.TrimSpace(lonString))
	if err != nil {
		return &LatLon{}, err
	}

	return NewLatLon(lat, lon), nil
}

func (latLon LatLon) String() string {
	return fmt.Sprintf("(%v, %v)", latLon.Lat, latLon.Lon)
}

func (d Km) String() string {
	return fmt.Sprintf("%.1fkm", float64(d))
}

func (a Degrees) String() string {
	return fmt.Sprintf("%.1f°", float64(a))
}

func normalizeDegrees(d Degrees) Degrees {
	if d < 0 {
		return d + 360
	}
	if d >= 360 {
		return d - 360
	}
	return d
}

// NewLatLon creates a new pair of coordinates. It normalizes the given latitude
// and longitude.
func NewLatLon(lat Latitude, lon Longitude) *LatLon {
	return &LatLon{normalizeLat(lat), normalizeLon(lon)}
}

// Distance calculates the great circle distance between two coordinates
// in kilometers using the haversine formula.
// See http://www.movable-type.co.uk/scripts/latlong.html for more details.
func Distance(latLon1, latLon2 *LatLon) Km {
	const radius = 6371.0 // km
	φ1 := radians(float64(latLon1.Lat))
	φ2 := radians(float64(latLon2.Lat))
	Δφ := radians(float64(latLon2.Lat - latLon1.Lat))
	Δλ := radians(float64(latLon2.Lon - latLon1.Lon))
	a := math.Pow(math.Sin(Δφ/2), 2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Pow(math.Sin(Δλ/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return Km(radius * c)
}

// Azimuth calculates the azimuth between two coordinates in degrees.
// See http://www.movable-type.co.uk/scripts/latlong.html for more details.
func Azimuth(latLon1, latLon2 *LatLon) Degrees {
	φ1 := radians(float64(latLon1.Lat))
	φ2 := radians(float64(latLon2.Lat))
	Δλ := radians(float64(latLon2.Lon - latLon1.Lon))
	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) -
		math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	azimuth := degrees(math.Atan2(y, x))
	return normalizeDegrees(Degrees(azimuth))
}

func radians(degrees float64) float64 {
	return (degrees * math.Pi) / 180.0
}

func degrees(radians float64) float64 {
	return (radians * 180.0) / math.Pi
}
