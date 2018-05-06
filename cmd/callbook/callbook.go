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

	callbooks := loadCallbooks(config)
	infos := lookup(os.Args[1], callbooks)
	for name, info := range infos {
		printInfo(name, info)
		if useLocator {
			printDistanceAzimuth(info, locator)
		}
		fmt.Println()

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

func loadCallbooks(config cfg.Configuration) map[string]callbook.Callbook {
	params := []struct {
		name       string
		configPath string
		factory    callbook.Factory
	}{
		{"HamQTH.com", "callbook.hamqth", func(username, password string) callbook.Callbook {
			return callbook.NewHamQTH(username, password)
		}},
		{"QRZ.com", "callbook.qrz", func(username, password string) callbook.Callbook {
			return callbook.NewQRZ(username, password)
		}},
	}
	callbooks := make(map[string]callbook.Callbook)
	for _, param := range params {
		callbook, err := newCallbook(param.configPath, config, param.factory)
		if err == nil {
			callbooks[param.name] = callbook
		}
	}
	return callbooks
}

func newCallbook(configPath string, config cfg.Configuration, factory callbook.Factory) (callbook.Callbook, error) {
	useCallbook := config.Get(configPath, false) != false
	if !useCallbook {
		return nil, nil
	}

	username := config.Get(configPath+".username", "").(string)
	password := config.Get(configPath+".password", "").(string)
	if username == "" || password == "" {
		return nil, fmt.Errorf("cannot read username or password for %v", configPath)
	}
	return factory(username, password), nil
}

func lookup(callsign string, callbooks map[string]callbook.Callbook) map[string]callbook.Info {
	infos := make(map[string]callbook.Info)
	for name, callbook := range callbooks {
		info, err := callbook.Lookup(callsign)
		if err == nil {
			infos[name] = info
		}
	}
	return infos
}

func printInfo(title string, info callbook.Info) {
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", len(title)))
	fmt.Printf("Callsign %v\n", info.Callsign)
	fmt.Printf("Name: %s\n", info.Name)
	fmt.Printf("Address: %s\n", info.Address)
	fmt.Printf("QTH: %s\n", info.Country)
	fmt.Printf("Country: %s\n", info.Country)
	fmt.Printf("CQ: %d\n", info.CQZone)
	fmt.Printf("ITU: %d\n", info.ITUZone)
	fmt.Printf("Time Offset: UTC%+1.1f\n", info.TimeOffset)
	if !info.Locator.IsZero() {
		fmt.Printf("Locator: %v\n", info.Locator)
	}
	if info.LatLonValid {
		fmt.Printf("Lat/Lon: %v\n", info.LatLon)
	}
}

func printDistanceAzimuth(info callbook.Info, loc locator.Locator) {
	latLon1 := locator.ToLatLon(loc)
	var latLon2 latlon.LatLon
	if info.LatLonValid {
		latLon2 = info.LatLon
	} else if !info.Locator.IsZero() {
		latLon2 = locator.ToLatLon(info.Locator)
	} else {
		return
	}
	fmt.Printf("Distance: %v\n", latlon.Distance(latLon1, latLon2))
	fmt.Printf("Azimuth: %v\n", latlon.Azimuth(latLon1, latLon2))
}
