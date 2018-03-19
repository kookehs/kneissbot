package irc

import (
	"bytes"
	"log"

	"golang.org/x/net/websocket"
)

const (
	// Buffer size for channels
	ChannelSize = 100

	// MaxMessageSize is a fixed message length as specified by RFC1459
	MaxMessageSize = 512

	// Capabilities
	Commands   = "twitch.tv/commands"
	Membership = "twitch.tv/membership"
	Tags       = "twitch.tv/tags"
)

// Session contains variables required to interact with the IRC server
// over a websocket connection.
type Twitch struct {
	In        chan []byte
	Out       chan []byte
	Websocket *websocket.Conn
}

// NewTwitch creates and initializes a Twitch connected to the
// given URL.
func NewTwitch(origin, url string) (*Twitch, error) {
	ws, err := websocket.Dial(url, "", origin)

	if err != nil {
		return nil, err
	}

	// TODO: Buffered channels.
	return &Twitch{
		In:        make(chan []byte, ChannelSize),
		Out:       make(chan []byte, ChannelSize),
		Websocket: ws,
	}, nil
}

// Cap sends a request for the given capabilities.
func (t *Twitch) Cap(capabilities []string) {
	for _, cap := range capabilities {
		t.Out <- []byte("CAP REQ :" + cap)
	}
}

// Close forcefully shuts down the underlying websocket connection.
func (t *Twitch) Close() error {
	return t.Websocket.Close()
}

// Connect sends the given credentials to the IRC server for authentication.
// The blocking operation returns whether connecting to IRC server was successful
func (t *Twitch) Connect(nick, token string) bool {
	t.Out <- []byte("PASS oauth:" + token)
	t.Out <- []byte("NICK " + nick)

	// Wait until we receive end of MOTD
	for {
		in := <-t.In
		message := MakeMessage(string(in))

		switch message.Command {
		case "001":
		case "002":
		case "003":
		case "004":
		case "372":
		case "375":
		case "376":
			return true
		default:
			return false
		}
	}
}

// HandleCommands handles all incoming commands.
func (t *Twitch) HandleCommands() {
	for {
		in := <-t.In
		message := MakeMessage(string(in))

		switch message.Command {
		case "PING":
			t.Out <- []byte("PONG :tmi.twitch.tv")
		}
	}
}

// Join sends a request to join the given channel.
// The blocking operation returns whether joining the channel was successful
func (t *Twitch) Join(channel string) bool {
	t.Out <- []byte("JOIN #" + channel)

	// Wait until we receive end of names
	for {
		in := <-t.In
		message := MakeMessage(string(in))

		switch message.Command {
		case "JOIN":
		case "353":
		case "366":
			return true
		default:
			return false
		}
	}
}

// Part leaves the given channel.
func (t *Twitch) Part(channel string) {
	t.Out <- []byte("PART #" + channel)
}

// PrivMSG sends a private message to the given channel.
func (t *Twitch) PrivMSG(channel, message string) {
	t.Out <- []byte("PRIVMSG #" + channel + " :" + message)
}

// Read sends any incoming message from the IRC server to the In channel.
func (t *Twitch) Read() {
	for {
		buffer := make([]byte, MaxMessageSize)
		_, err := t.Websocket.Read(buffer)

		if err != nil {
			log.Println(err)
			continue
		}

		messages := bytes.Split(bytes.TrimRight(buffer, "\x00"), []byte{'\n'})

		for _, message := range messages {
			if len(message) > 0 {
				log.Println("in: " + string(message))
				t.In <- message
			}
		}
	}
}

// Start creates additional goroutines for reading and writing to the IRC server.
func (t *Twitch) Start() {
	go t.Read()
	go t.Write()
}

// Write sends any outgoing message from the Out channel to the IRC server.
func (t *Twitch) Write() {
	for {
		message := <-t.Out
		log.Println("out: " + string(message))

		if _, err := t.Websocket.Write(message); err != nil {
			log.Println(err)
		}
	}
}
