/*
dxcc prints information about a given DXCC prefix.
It can also calculate the distance and azimuth from an optionally given maidenhead locator
to that prefix.

If the given prefix is ambiguous (e.g. MZ is for the Shetlands and also for Scotland),
multiple datasets are returned.

USAGE

	dxcc <prefix> [locator]

EXAMPLE

	> dxcc mz em12af

	Prefix MZ: Shetland Islands (GM/s)
	Continent: EU
	CQ: 14
	ITU: 27
	Location: (60.50000N, 1.50000W)
	Time Offset: UTC+0.0
	ARRL compliant: false
	Exact Match: false
	Distance: 7264.9km
	Azimuth: 32.6°

	Prefix MZ: Scotland (GM)
	Continent: EU
	CQ: 14
	ITU: 27
	Location: (56.82000N, 4.18000W)
	Time Offset: UTC+0.0
	ARRL compliant: true
	Exact Match: false
	Distance: 7275.2km
	Azimuth: 36.9°

CONFIGURATION

	dxcc stores a cty.dat file in ~/.config/hamradio. The file is automatically updated if
	there is a newer version available at http://www.country-files.com/cty/cty.dat.
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

func main() {
	localFilename, err := dxcc.LocalFilename()
	if err != nil {
		log.Fatal(err)
	}
	updated, err := dxcc.Update(dxcc.DefaultURL, localFilename)
	if err != nil {
		fmt.Printf("update of local copy failed: %v\n", err)
	}
	if updated {
		fmt.Printf("updated local copy: %v\n", localFilename)
	}

	prefixes, err := dxcc.LoadLocal(localFilename)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("usage: %s <prefix> [locator]\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	config, err := cfg.LoadDefault()
	if err != nil {
		log.Fatalf("cannot load configuration file: %v", err)
	}

	foundPrefixes, _ := prefixes.Find(os.Args[1])
	loc, useLocator := parseLocator()
	if !useLocator {
		loc, useLocator = loadLocator(config)
	}
	for _, prefix := range foundPrefixes {
		printPrefix(prefix)
		if useLocator {
			printDistanceAzimuth(prefix, loc)
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

func loadLocator(config cfg.Configuration) (locator.Locator, bool) {
	value := config.Get(cfg.MyLocator, "").(string)

	loc, err := locator.Parse(value)
	if err != nil {
		fmt.Printf("cannot load locator: %v\n", err)
		return locator.Locator{}, false
	}
	return loc, true
}

func printPrefix(prefix dxcc.Prefix) {
	fmt.Printf("Prefix %s: %s (%s)\n", prefix.Prefix, prefix.Name, prefix.PrimaryPrefix)
	fmt.Printf("Continent: %s\n", prefix.Continent)
	fmt.Printf("CQ: %d\n", prefix.CQZone)
	fmt.Printf("ITU: %d\n", prefix.ITUZone)
	fmt.Printf("Location: %v\n", prefix.LatLon)
	fmt.Printf("Time Offset: UTC%+1.1f\n", prefix.TimeOffset)
	fmt.Printf("ARRL compliant: %t\n", !prefix.NotARRLCompliant)
	fmt.Printf("Exact Match: %t\n", prefix.NeedsExactMatch)
}

func printDistanceAzimuth(prefix dxcc.Prefix, loc locator.Locator) {
	latLon := locator.ToLatLon(loc)
	fmt.Printf("Distance: %v\n", latlon.Distance(latLon, prefix.LatLon))
	fmt.Printf("Azimuth: %v\n", latlon.Azimuth(latLon, prefix.LatLon))
}
