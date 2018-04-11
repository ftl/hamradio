/*
callbook retrieves information about a given callsign from hamqth.com and qrz.com and prints this information.
It can also calculate the distance and azimuth from an optionally given maidenhead locator
to the retrieved location of the callsign.

USAGE

	callbook <callsign> [locator]

EXAMPLE

	> callbook aa7bq em12af

	HamQTH.com
	==========
	Callsign AA7BQ
	Name: Fred
	QTH: United States
	Country: United States
	CQ: 3
	ITU: 6
	Time Offset: UTC+7.0
	Locator: DM43
	Lat/Lon: (33.66000N, 111.87000W)
	Distance: 1306.9km
	Azimuth: 280.7°

	QRZ.com
	=======
	Callsign AA7BQ
	Name: FRED L LLOYD
	QTH: United States
	Country: United States
	CQ: 3
	ITU: 6
	Time Offset: UTC-7.0
	Locator: DM43bq
	Lat/Lon: (33.69826N, 111.89150W)
	Distance: 1309.1km
	Azimuth: 280.9°

CONFIGURATION

	callbook expects the hamradio configuration file (~/.config/hamradio/conf.json) to contain
	the credentials for hamqth.com and qrz.com. If it can't find the credentials for one
	of these sites, it will not try to query the respective site.

	The expected JSON structure for the credentials is as follows:

	{
		"callbook": {
			"hamqth": {
				"username": "your hamqth.com username",
				"password": "your hamqth.com password"
			},
			"qrz": {
				"username": "your qrz.com username",
				"password": "your qrz.com password"
			},
		}
	}
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ftl/hamradio/callbook"
	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("usage: %s <callsign> [locator]\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}
	locator, useLocator := parseLocator()

	config, err := cfg.LoadDefault()
	if err != nil {
		log.Fatalf("cannot load configuration file: %v", err)
	}

	hamQTHInfo, err := lookupHamQTH(os.Args[1], config)
	if err != nil {
		log.Fatal(err)
	}
	qrzInfo, err := lookupQRZ(os.Args[1], config)
	if err != nil {
		log.Fatal(err)
	}
	if hamQTHInfo != nil {
		printInfo("HamQTH.com", hamQTHInfo)
		if useLocator {
			printDistanceAzimuth(hamQTHInfo, locator)
		}
	}
	if qrzInfo != nil {
		printInfo("QRZ.com", qrzInfo)
		if useLocator {
			printDistanceAzimuth(qrzInfo, locator)
		}
	}
}

func parseLocator() (locator.Locator, bool) {
	if len(os.Args) != 3 {
		return locator.Locator{}, false
	}

	loc, err := locator.Parse(os.Args[2])
	if err != nil {
		fmt.Printf("cannot parse locator: %v\n", err)
		return locator.Locator{}, false
	}
	return loc, true
}

func lookupHamQTH(callsign string, config cfg.Configuration) (*callbook.Info, error) {
	useHamQTH := config.Get("callbook.hamqth", false) != false
	if !useHamQTH {
		return nil, nil
	}

	username := config.Get("callbook.hamqth.username", "").(string)
	password := config.Get("callbook.hamqth.password", "").(string)
	if username == "" || password == "" {
		return nil, fmt.Errorf("cannot read username or password for hamqth.com")
	}
	hamqth := callbook.NewHamQTH(username, password)

	return hamqth.Lookup(os.Args[1])
}

func lookupQRZ(callsign string, config cfg.Configuration) (*callbook.Info, error) {
	useQRZ := config.Get("callbook.qrz", false) != false
	if !useQRZ {
		return nil, nil
	}

	username := config.Get("callbook.qrz.username", "").(string)
	password := config.Get("callbook.qrz.password", "").(string)
	if username == "" || password == "" {
		return nil, fmt.Errorf("cannot read username or password for qrz.com")
	}
	qrz := callbook.NewQRZ(username, password)

	return qrz.Lookup(os.Args[1])
}

func printInfo(title string, info *callbook.Info) {
	if info == nil {
		return
	}

	fmt.Println(title)
	fmt.Println(strings.Repeat("=", len(title)))
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
