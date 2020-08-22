/*
supercheck uses the Super Check Partial callsign database from http://www.supercheckpartial.com/
to find callsigns that are similar to a given string. The result is returned as a space separated
list of callsigns. The given string must be at least three characters long.

USAGE

	supercheck <string>

EXAMPLE

	> supercheck dneo

	DG9NEO DL1NEO KD0NEO

CONFIGURATION

	supercheck stores a MASTER.SCP file in ~/.config/hamradio. The file is automatically updated if
	there is a newer version available at http://www.supercheckpartial.com/MASTER.SCP.

*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ftl/hamradio/scp"
)

func main() {
	localFilename, err := scp.LocalFilename()
	if err != nil {
		log.Fatal(err)
	}
	updated, err := scp.Update(scp.DefaultURL, localFilename)
	if err != nil {
		fmt.Printf("update of local copy failed: %v\n", err)
	}
	if updated {
		fmt.Printf("updated local copy: %v\n", localFilename)
	}

	database, err := scp.LoadLocal(localFilename)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <text>\n", filepath.Base(os.Args[0]))
		os.Exit(0)
	}

	matches, err := database.Find(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Join(matches, " "))
}
