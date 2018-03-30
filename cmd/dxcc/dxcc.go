package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hamradio/latlon"
	"github.com/ftl/hamradio/locator"
)

func main() {
	localFilename, err := localFilename()
	if err != nil {
		log.Fatal(err)
	}
	updated, err := dxcc.Update(dxcc.DefaultURL, localFilename)
	if err != nil {
		log.Fatal(err)
	}
	if updated {
		fmt.Printf("updated local copy: %v\n", localFilename)
	}

	prefixes, err := dxcc.LoadLocal(localFilename)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("usage: %s <prefix> [<locator>]\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	foundPrefixes, _ := prefixes.Find(os.Args[1])
	loc, useLocator := parseLocator()
	for _, prefix := range foundPrefixes {
		printPrefix(prefix)
		if useLocator {
			printDistanceAzimuth(prefix, loc)
		}
		fmt.Println()
	}
}

func localFilename() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".dxcc/cty.dat"), nil
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

func printPrefix(prefix *dxcc.Prefix) {
	fmt.Printf("Prefix %s: %s (%s)\n", prefix.Prefix, prefix.Name, prefix.PrimaryPrefix)
	fmt.Printf("Continent: %s\n", prefix.Continent)
	fmt.Printf("CQ: %d\n", prefix.CQZone)
	fmt.Printf("ITU: %d\n", prefix.ITUZone)
	fmt.Printf("Location: %v\n", prefix.LatLon)
	fmt.Printf("Time Offset: UTC%+1.1f\n", prefix.TimeOffset)
	fmt.Printf("ARRL compliant: %t\n", !prefix.NotARRLCompliant)
	fmt.Printf("Exact Match: %t\n", prefix.NeedsExactMatch)
}

func printDistanceAzimuth(prefix *dxcc.Prefix, loc locator.Locator) {
	latLon := locator.ToLatLon(loc)
	fmt.Printf("Distance: %v\n", latlon.Distance(latLon, &prefix.LatLon))
	fmt.Printf("Azimuth: %v\n", latlon.Azimuth(latLon, &prefix.LatLon))
}
