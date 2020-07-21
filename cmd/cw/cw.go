/*
cw uses a cwdaemon server running locally on port 6789 to output CW.

USAGE

	cw <command> [parameters]

Valid commands are:
	send <text>
	speed <wpm>
	tune <duration>

EXAMPLES

	Send "hello world":
	> cw send hello world

	Set the speed to 15 WpM:
	> cw speed 15

	Key down for 5 seconds:
	> cw tune 5
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

type commandFunc func(*cwclient.Client, string)

func main() {
	commands := map[string]commandFunc{
		"send":  send,
		"speed": speed,
		"tune":  tune,
	}

	host, port, commandName, text, err := parseCommandLine(os.Args[1:])
	if err != nil {
		log.Fatal(err)
		printUsage()
	}

	client, err := cwclient.New(host, port)
	if err != nil {
		log.Fatalf("cannot create a client for cwdaemon: %v", err)
	}

	err = client.Connect()
	if err != nil {
		log.Fatalf("cannot connect to cwdaemon: %v", err)
	}
	defer client.Disconnect()

	if commandName == "" {
		printUsage()
	}

	if command, ok := commands[commandName]; ok {
		command(client, text)
	} else {
		printUsage()
	}
}

func parseCommandLine(args []string) (host string, port int, command string, text string, err error) {
	lastIndex := len(args) - 1
	text = ""
	for i := 0; i < len(args); i++ {
		arg := strings.ToLower(args[i])
		switch arg {
		case "-h", "--host":
			if i >= lastIndex {
				err = fmt.Errorf("missing actual hostname after host flag")
				return
			}
			i += 1
			host = args[i]
		case "-p", "--port":
			if i >= lastIndex {
				err = fmt.Errorf("missing actual hostname after host flag")
				return
			}
			i += 1
			port, err = strconv.Atoi(args[i])
			if err != nil {
				err = fmt.Errorf("invalid port: %w", err)
				return
			}
		default:
			if command == "" {
				command = arg
			} else if text == "" {
				text = arg
			} else {
				text += " " + arg
			}
		}
	}
	return
}

func printUsage() {
	fmt.Printf("usage: %s [flags] <command> [parameters]\n", filepath.Base(os.Args[0]))
	fmt.Print(`

valid commands:
	send <text>          send the given text
	speed <wpm>          set the given speed in WpM for the next transmissions
	tune  <duration>     key down for the given duration in seconds for tuning

flags:
	-h, --host [host]    use the given host as target instead of the default host localhost
	-p, --port [port]    use the given port as target instead of the default port 6789

`)
	os.Exit(0)
}

func send(client *cwclient.Client, text string) {
	if text == "" {
		printUsage()
	}

	client.Send(text)
	client.Wait()
}

func speed(client *cwclient.Client, text string) {
	if text == "" {
		printUsage()
	}

	speed, err := strconv.Atoi(text)
	if err != nil {
		log.Fatalf("%v is not a valid speed value, it must be a number between 5 and 60", os.Args[2])
	}

	client.Speed(speed)
}

func tune(client *cwclient.Client, text string) {
	if text == "" {
		printUsage()
	}

	duration, err := strconv.Atoi(text)
	if err != nil {
		log.Fatalf("%v is not a valid duration value, it must be a number between 0 and 10", os.Args[2])
	}

	client.Tune(duration)
}
