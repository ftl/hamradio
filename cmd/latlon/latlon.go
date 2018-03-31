/*
latlon calculates the distance and azimuth between two coordinates given as pairs of latitude and longitude.

USAGE

	latlon <latitude1> <longitude1> <latitude2> <longitude2>

EXAMPLE

	> latlon 12.3 45.6 78.9 10.1

	Distance: 7646.4km, Azimuth: 353.1°
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ftl/hamradio/latlon"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Printf("usage: %s <lat1> <lon1> <lat2> <lon2>\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	lat1, err := latlon.ParseLat(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	lon1, err := latlon.ParseLon(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	lat2, err := latlon.ParseLat(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	lon2, err := latlon.ParseLon(os.Args[4])
	if err != nil {
		log.Fatal(err)
	}

	latLon1 := latlon.LatLon{Lat: lat1, Lon: lon1}
	latLon2 := latlon.LatLon{Lat: lat2, Lon: lon2}

	distance := latlon.Distance(&latLon1, &latLon2)
	azimuth := latlon.Azimuth(&latLon1, &latLon2)

	fmt.Printf("Distance: %v, Azimuth: %v\n", distance, azimuth)
}
