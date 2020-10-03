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
	"sync"
	"time"
)

// Client is a client for the cwdaemon server application.
type Client struct {
	localAddr      *net.UDPAddr
	remoteAddr     *net.UDPAddr
	connection     *net.UDPConn
	connectionLock *sync.Mutex
	receiveBuffer  []byte
	sendBuffer     chan interface{}
	disconnected   chan struct{}
	sendQueue      *sendQueue
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
// If the hostname is empty, localhost will be used. If the port is 0, the default port 6789 will be used.
func New(hostname string, port int) (*Client, error) {
	client := Client{
		connectionLock: new(sync.Mutex),
		receiveBuffer:  make([]byte, 32),
		sendBuffer:     make(chan interface{}),
		sendQueue:      newSendQueue(),
	}

	if port == 0 {
		port = 6789
	}

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

	client.disconnected = make(chan struct{})
	close(client.disconnected)

	return &client, nil
}

// NewDefault returns a Client for a cwdaemon server running on localhost:6789.
func NewDefault() (*Client, error) {
	return New("", 0)
}

// Connect sets up a connection between the client and the server.
func (client *Client) Connect() error {
	client.connectionLock.Lock()
	defer client.connectionLock.Unlock()

	if client.IsConnected() {
		return nil
	}

	connection, err := net.DialUDP("udp", client.localAddr, client.remoteAddr)
	if err != nil {
		return err
	}
	client.connection = connection
	client.disconnected = make(chan struct{})

	go client.communicate()

	return nil
}

func (client *Client) communicate() {
	for {
		select {
		case _ = <-client.disconnected:
			client.sendQueue.Reset()
			return
		case m := <-client.sendBuffer:
			var err error
			switch message := m.(type) {
			case string:
				err = client.send(message)
			case syncText:
				token := client.sendQueue.Enqueue()
				err = client.send(fmt.Sprintf("\x1Bh%d", token), message.text)
			default:
				panic(fmt.Errorf("unknown send message type: %T", m))
			}
			if err != nil {
				log.Printf("Error sending %v: %T %v", m, err, err)
				client.sendQueue.Reset()
				// client.handleConnectionError()
			}
		default:
			message, err := client.receive()
			if err != nil {
				log.Printf("Error receiving: %T %v", err, err)
				// client.handleConnectionError()
				continue
			}
			if message == "" {
				continue
			}
			switch {
			case strings.HasPrefix(message, "h"):
				var token int
				_, err := fmt.Sscanf(message, "h%d", &token)
				if err != nil {
					log.Printf("Error parsing sequence number %v: %v", message, err)
				}
				client.sendQueue.Finish(token)
			case message == "break":
				finished, total := client.sendQueue.Current()
				log.Printf("CW output aborted at %d/%d", finished, total)
				client.sendQueue.Reset()
			}
		}
	}
}

type syncText struct {
	text string
}

func (client *Client) send(messages ...string) error {
	for _, message := range messages {
		buf := []byte(message)
		_, err := client.connection.Write(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) receive() (string, error) {
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

func (client *Client) handleConnectionError() {
	select {
	case <-client.disconnected:
	default:
		close(client.disconnected)
	}
}

// Disconnect closes the connection between the client and the server.
func (client *Client) Disconnect() {
	select {
	case <-client.disconnected:
		return
	default:
	}
	client.connectionLock.Lock()
	defer client.connectionLock.Unlock()

	client.connection.Close()
	select {
	case <-client.disconnected:
	default:
		close(client.disconnected)
	}
}

// IsConnected indicates if the client has an active connection to the server.
func (client *Client) IsConnected() bool {
	select {
	case <-client.disconnected:
		return false
	default:
		return true
	}
}

// IsIdle returns true if there are no texts waiting on the server for output as CW.
func (client *Client) IsIdle() bool {
	return client.sendQueue.Idle()
}

// Wait waits for all pending text to be output as CW.
func (client *Client) Wait() {
	for !client.IsIdle() && client.IsConnected() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (client *Client) command(format string, values ...interface{}) {
	command := fmt.Sprintf("\x1B%s", fmt.Sprintf(format, values...))
	if !client.IsConnected() {
		log.Printf("Cannot send command %q, the client is not connected. Reconnect and try again.", command)
		return
	}
	client.sendBuffer <- command
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
	if !client.IsConnected() {
		log.Printf("Cannot send %q, the client is not connected. Reconnect and try again.", text)
	}
	if strings.HasPrefix(text, "\x1B") {
		log.Panicf("Cannot send escape sequence %s, use the dedicated methods for that.", text[1:])
	}

	client.sendBuffer <- syncText{text}
}

type sendQueue struct {
	queued, finished int
	lock             *sync.RWMutex
}

func newSendQueue() *sendQueue {
	return &sendQueue{
		lock: new(sync.RWMutex),
	}
}

func (q *sendQueue) Enqueue() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queued++
	return q.queued
}

func (q *sendQueue) Finish(token int) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if token > q.queued {
		return
	}
	q.finished = token
}

func (q *sendQueue) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queued = 0
	q.finished = 0
}

func (q *sendQueue) Current() (int, int) {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.finished, q.queued
}

func (q *sendQueue) Idle() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.queued == q.finished
}
