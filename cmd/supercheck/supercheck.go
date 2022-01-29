/*
supercheck uses the Super Check Partial callsign database from http://www.supercheckpartial.com/
to find callsigns that are similar to a given string. The result is returned as a space separated
list of callsigns. The given string must be at least three characters long.

USAGE

	supercheck <string>

EXAMPLE

	> supercheck dneo

	DL1NEO KD0NEO N0EO NE6O NE8O NE9O

CONFIGURATION

	supercheck stores a MASTER.SCP file in ~/.config/hamradio. The file is automatically updated if
	there is a newer version available at http://www.supercheckpartial.com/MASTER.SCP.

*/
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"

	"github.com/ftl/hamradio/scp"
)

var options struct {
	CallHistoryFilename string   `short:"c" long:"callhistory" description:"use this call history file instead of MASTER.SCP"`
	Fields              []string `short:"f" long:"field" description:"when using a call history file, show the values of these fields next to the callsigns"`
	Lines               bool     `short:"l" long:"lines" description:"output each matching callsign in a separate line"`
	Reverse             bool     `short:"r" long:"reverse" description:"print the callsigns in reverse order (best match last)"`
	Args                struct {
		Input []string `positional-arg-name:"input" required:"1"`
	} `positional-args:"yes"`
}

func main() {
	_, err := flags.Parse(&options)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil && !flags.WroteHelp(err) {
		log.Fatal(err)
	}

	var database *scp.Database
	if options.CallHistoryFilename != "" {
		database, err = loadCallhistory(options.CallHistoryFilename)
	} else {
		database, err = loadMasterScp()
	}
	if err != nil {
		log.Fatal(err)
	}

	entries, err := database.FindEntries(options.Args.Input[0])
	if err != nil {
		log.Fatal(err)
	}

	fieldSet := make(scp.FieldSet, len(options.Fields))
	for i, field := range options.Fields {
		fieldSet[i] = scp.FieldName(field)
	}

	matches := make([]string, len(entries))
	for i, entry := range entries {
		match := entry.Key()
		if options.CallHistoryFilename != "" {
			match = fmt.Sprintf("%s,%s", match, strings.Join(entry.GetValues(fieldSet...), ","))
		}
		var j int
		if options.Reverse {
			j = len(matches) - 1 - i
		} else {
			j = i
		}
		matches[j] = match
	}

	separator := " "
	if options.Lines {
		separator = "\n"
	}
	fmt.Println(strings.Join(matches, separator))
}

func loadMasterScp() (*scp.Database, error) {
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

	return scp.LoadLocal(localFilename)
}

func loadCallhistory(filename string) (*scp.Database, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return scp.ReadCallHistory(file)
}
