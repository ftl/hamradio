/*
locator calculates latitude and longitude for a given maidenhead locator.
It can also calculate the distance and azimuth to an optionally given second maidenhead locator.

USAGE

	locator <locator1> [locator2]

EXAMPLE

	> locator EM12af KO94bx

	Locator EM12af = (32.22917N, 97.95833W)
	Locator KO94bx = (54.97917N, 38.12500E)
	Distance: 9452.2km
	Azimuth: 23.6°
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ftl/hamradio/locator"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("usage: %s <locator1> [locator2]\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	locator1, err := locator.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Locator %v = %v\n", locator1, locator.ToLatLon(locator1))
	if len(os.Args) == 2 {
		os.Exit(0)
	}

	locator2, err := locator.Parse(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Locator %v = %v\n", locator2, locator.ToLatLon(locator2))
	fmt.Printf("Distance: %v\nAzimuth: %v\n",
		locator.Distance(locator1, locator2),
		locator.Azimuth(locator1, locator2))
}
