/*
cw uses a cwdaemon server running locally on port 6789 to output CW.

USAGE

	cw <command> [parameters]

Valid commands are:
	send <text>
	speed <wpm>

EXAMPLES

	Send "hello world":
	> cw send hello world

	Set the speed to 15 WpM:
	> cw speed 15
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/cwclient"
)

type commandFunc func(*cwclient.Client)

func main() {
	commands := map[string]commandFunc{
		"send":  send,
		"speed": speed,
	}

	client, err := cwclient.NewDefault()
	if err != nil {
		log.Fatalf("cannot create a client for cwdaemon: %v", err)
	}

	err = client.Connect()
	if err != nil {
		log.Fatalf("cannot connect to cwdaemon: %v", err)
	}

	if len(os.Args) == 1 {
		printUsage()
	}

	if command, ok := commands[os.Args[1]]; ok {
		command(client)
	} else {
		printUsage()
	}

	defer client.Disconnect()
}

func printUsage() {
	fmt.Printf("usage: %s <command> [parameters]\n", filepath.Base(os.Args[0]))
	fmt.Printf("valid commands:\n\tsend <text>\n\tspeed <wpm>\n")
	os.Exit(0)
}

func send(client *cwclient.Client) {
	if len(os.Args) < 3 {
		printUsage()
	}

	message := strings.Join(os.Args[2:], " ")
	client.Send(message)
	client.Wait()
}

func speed(client *cwclient.Client) {
	if len(os.Args) != 3 {
		printUsage()
	}

	speed, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("%v is not a valid speed value, it must be a number between 5 and 60", os.Args[3])
	}

	client.Speed(speed)
}
