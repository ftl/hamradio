package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ftl/hamradio/callbook"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

func main() {
	if len(os.Args) < 4 || len(os.Args) > 5 {
		fmt.Printf("usage: %s <callsign> <username> <password> [locator]\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}
	locator, useLocator := parseLocator()

	hamqth := callbook.NewHamQTH(os.Args[2], os.Args[3])

	info, err := hamqth.Lookup(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	printInfo(info)
	if useLocator {
		printDistanceAzimuth(info, locator)
	}
}

func parseLocator() (locator.Locator, bool) {
	if len(os.Args) != 5 {
		return locator.Locator{}, false
	}

	loc, err := locator.Parse(os.Args[4])
	if err != nil {
		fmt.Printf("cannot parse locator: %v\n", err)
		return locator.Locator{}, false
	}
	return loc, true
}
func printInfo(info *callbook.Info) {
	fmt.Printf("Callsign %v\n", info.Callsign)
	fmt.Printf("Name: %s\n", info.Name)
	fmt.Printf("QTH: %s\n", info.Country)
	fmt.Printf("Country: %s\n", info.Country)
	fmt.Printf("CQ: %d\n", info.CQZone)
	fmt.Printf("ITU: %d\n", info.ITUZone)
	fmt.Printf("Time Offset: UTC%+1.1f\n", info.TimeOffset)
	if !info.Locator.IsZero() {
		fmt.Printf("Locator: %v\n", info.Locator)
	}
	if info.LatLon != nil {
		fmt.Printf("Lat/Lon: %v\n", info.LatLon)
	}
}

func printDistanceAzimuth(info *callbook.Info, loc locator.Locator) {
	latLon1 := locator.ToLatLon(loc)
	var latLon2 *latlon.LatLon
	if info.LatLon != nil {
		latLon2 = info.LatLon
	} else if !info.Locator.IsZero() {
		latLon2 = locator.ToLatLon(info.Locator)
	} else {
		return
	}
	fmt.Printf("Distance: %v\n", latlon.Distance(latLon1, latLon2))
	fmt.Printf("Azimuth: %v\n", latlon.Azimuth(latLon1, latLon2))
}
