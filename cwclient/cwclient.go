/*
Package cwclient provides a client library for the cwdaemon (https://github.com/acerion/cwdaemon) server application.

The client communicates with the cwdaemon server via UDP using the proprietary protocol defined in the cwdaemon documentation.

To run the cwdaemon locally for testing use the following command line: "cwdaemon -yi -xs -n -d null"
This will start the cwdaemon as foreground process listening on port 6789, its output is written to stdout and stderr.
To kill the process, hit Ctrl+C.
*/
package cwclient

import (
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"time"
)

// Client is a client for the cwdaemon server application.
type Client struct {
	localAddr     *net.UDPAddr
	remoteAddr    *net.UDPAddr
	connection    *net.UDPConn
	receiveBuffer []byte
	sendBuffer    chan string
	disconnected  chan struct{}
	inCount       int
	outCount      int
}

// Soundsystem supported by the cwdaemon
type Soundsystem string

// The Soundsystems supported by the cwdaemon
const (
	PCSpeaker  Soundsystem = "c"
	OSS        Soundsystem = "o"
	ALSA       Soundsystem = "a"
	PulseAudio Soundsystem = "p"
	None       Soundsystem = "n"
	Soundcard  Soundsystem = "s"
)

// SSBSource supported by cwdaemon
type SSBSource int

// The SSB sources supported by the cwdaemon
const (
	SSBFromMicrophone = 0
	SSBFromSoundcard  = 1
)

// New creates a new Client for a cwdaemon server running on the given hostname and port.
// If the hostname is empty, localhost will be used.
func New(hostname string, port int) (*Client, error) {
	client := Client{}

	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}
	client.localAddr = localAddr
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return nil, err
	}
	client.remoteAddr = remoteAddr

	client.receiveBuffer = make([]byte, 32)
	client.sendBuffer = make(chan string)
	client.disconnected = make(chan struct{})

	return &client, nil
}

// NewDefault returns a Client for a cwdaemon server running on localhost:6789.
func NewDefault() (*Client, error) {
	return New("", 6789)
}

// Connect sets up a connection between the client and the server.
func (client *Client) Connect() error {
	if client.connection != nil {
		return nil
	}

	connection, err := net.DialUDP("udp", client.localAddr, client.remoteAddr)
	if err != nil {
		return err
	}
	client.connection = connection
	go client.communicate()

	return nil
}

func (client *Client) communicate() {
	for {
		select {
		case _ = <-client.disconnected:
			return
		case message := <-client.sendBuffer:
			err := client.send(message)
			if err != nil {
				log.Printf("error sending %v: %v", message, err)
			}
		default:
			message, err := client.receive()
			if err != nil {
				log.Printf("error receiving: %v", err)
			}
			if message == "" {
				continue
			}
			switch {
			case strings.HasPrefix(message, "h"):
				_, err := fmt.Sscanf(message, "h%d", &client.inCount)
				if err != nil {
					log.Printf("error parsing sequence number %v: %v", message, err)
				}
			case message == "break":
				log.Printf("CW output aborted at %d/%d", client.inCount, client.outCount)
				client.inCount = 0
				client.outCount = 0
			}
		}
	}
}

func (client *Client) send(s string) error {
	if client.connection == nil {
		return fmt.Errorf("client not connected")
	}

	buf := []byte(s)
	_, err := client.connection.Write(buf)
	return err
}

func (client *Client) receive() (string, error) {
	if client.connection == nil {
		return "", fmt.Errorf("client not connected")
	}

	client.connection.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, err := client.connection.Read(client.receiveBuffer)
	if err, ok := err.(net.Error); ok {
		if err.Timeout() {
			return "", nil
		}
		return "", err
	} else if err != nil {
		return "", err
	}

	message := client.receiveBuffer[0:n]
	return strings.TrimSpace(string(message)), nil
}

// Disconnect closes the connection between the client and the server.
func (client *Client) Disconnect() {
	if client.connection == nil {
		return
	}
	defer client.connection.Close()
	client.disconnected <- struct{}{}
	client.connection = nil
}

// IsConnected indicates if the client has an active connection to the server.
func (client *Client) IsConnected() bool {
	return client.connection != nil
}

// IsIdle returns true if there are no texts waiting on the server for output as CW.
func (client *Client) IsIdle() bool {
	return client.inCount == client.outCount
}

// Wait waits for all pending text to be output as CW.
func (client *Client) Wait() {
	for !client.IsIdle() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (client *Client) command(format string, values ...interface{}) {
	client.sendBuffer <- fmt.Sprintf("\x1B%s", fmt.Sprintf(format, values...))
}

// Reset resets the server to the default values:
// speed = 24 WpM
// tone = 800 Hz
// sound = on
// wordmode = off
// weight = 0
// UDP port = 6789
// PTT delay = 0 (off)
// device = parport0
// sound device = console buzzer
func (client *Client) Reset() {
	client.command("0")
}

// Speed sets the speed to the given speed in WpM [5..60]
func (client *Client) Speed(speed int) {
	normalizedSpeed := int(math.Max(5, math.Min(float64(speed), 60)))
	client.command("2%d", normalizedSpeed)
}

// Tone sets the generated sidetone to the given frequency in Hz  [300..1000]
func (client *Client) Tone(tone int) {
	normalizedTone := int(math.Max(300, math.Min(float64(tone), 1000)))
	client.command("3%d", normalizedTone)
}

// ToneOff turns of the generated sidetone
func (client *Client) ToneOff() {
	client.command("30")
}

// Abort aborts the output of CW and discards all pending texts.
func (client *Client) Abort() {
	client.command("4")
}

// Wordmode sets the cwdaemon server into the word mode
func (client *Client) Wordmode() {
	client.command("6")
}

// Weight sets the weighting between dit and dah [-50..50].
func (client *Client) Weight(weight int) {
	normalizedWeight := int(math.Max(-50, math.Min(float64(weight), 50)))
	client.command("7%d", normalizedWeight)
}

// Device sets the device for CW output, default is parport0.
func (client *Client) Device(device string) {
	client.command("8%s", device)
}

// PTT enables or disables the PTT keying.
func (client *Client) PTT(on bool) {
	var onAsInt int
	if on {
		onAsInt = 1
	}
	client.command("a%d", onAsInt)
}

// SSBSource sets the source for the SSB signal either to microphone or soundcard.
func (client *Client) SSBSource(source SSBSource) {
	client.command("b%d", source)
}

// Tune tunes for the given duration in seconds [0..10].
func (client *Client) Tune(seconds int) {
	normalizedSeconds := int(math.Max(0, math.Min(float64(seconds), 10)))
	client.command("c%d", normalizedSeconds)
}

// PTTDelay sets the PTT delay to the given duration in milliseconds [0..50].
func (client *Client) PTTDelay(milliseconds int) {
	normalizedMilliseconds := int(math.Max(0, math.Min(float64(milliseconds), 50)))
	client.command("d%d", normalizedMilliseconds)
}

// BandIndex outputs the given band index on the pins 2 (lsb), 7, 8, 9 (msb) of the parport.
func (client *Client) BandIndex(bandIndex int) {
	normalizedBandIndex := int(math.Max(0, math.Min(float64(bandIndex), 64)))
	client.command("e%d", normalizedBandIndex)
}

// Soundsystem instructs the cwdaemon to use the given soundsystem
func (client *Client) Soundsystem(soundsystem Soundsystem) {
	client.command("f%s", soundsystem)
}

// Volume sets the volume of the generated sidetone [0..100].
func (client *Client) Volume(volume int) {
	normalizedVolume := int(math.Max(0, math.Min(float64(volume), 100)))
	client.command("g%d", normalizedVolume)
}

// Send sends the given text to the server to be output it as CW.
func (client *Client) Send(text string) {
	if strings.HasPrefix(text, "\x1B") {
		log.Panicf("cannot send escape sequence %s, use the dedicated methods for that", text[1:])
	}

	client.outCount++
	client.command("h%d", client.outCount)
	client.sendBuffer <- text
}
