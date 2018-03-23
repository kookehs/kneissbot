package core

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/kookehs/kneissbot/net/irc"
)

const (
	// URLs
	TwitchOrigin = "http://twitch.tv"
	TwitchIRC    = "ws://irc-ws.chat.twitch.tv:80"

	// Capabilities
	Commands   = "twitch.tv/commands"
	Membership = "twitch.tv/membership"
	Tags       = "twitch.tv/tags"
)

var (
	DefaultCapabilities = []string{Commands, Membership, Tags}
)

func init() {
	// Twitch has a 500 character limit, not including line endings, not a 512 byte limit.
	irc.MaxMessageSize = 512 * utf8.UTFMax
}

// Twitch contains logic realted to both the API and IRC.
type Twitch struct {
	Event      chan string
	Management *Management
	Session    *irc.Session
}

// NewTwitch returns a pointer to an initialized Twitch struct.
func NewTwitch() (*Twitch, error) {
	session, err := irc.NewSession(TwitchOrigin, TwitchIRC)

	if err != nil {
		return nil, err
	}

	return &Twitch{
		Event:      make(chan string),
		Management: NewManagement(),
		Session:    session,
	}, nil
}

// Cap sends a request for the given capabilities.
// Default capabilities are used if none are given.
func (t *Twitch) Cap(capabilities []string) {
	caps := DefaultCapabilities

	if capabilities != nil && len(capabilities) > 0 {
		caps = capabilities
	}

	for _, cap := range caps {
		t.Session.Write("CAP REQ :" + cap)
	}
}

// Close shuts down channels and the underlying websocket connection.
func (t *Twitch) Close() error {
	close(t.Event)

	if err := t.Session.Close(); err != nil {
		return err
	}

	return nil
}

// Connect sends the given credentials to the IRC server for authentication.
// The blocking operation returns whether connecting to IRC server was successful
func (t *Twitch) Connect(nick, token string) bool {
	t.Session.Write("PASS oauth:" + token)
	t.Session.Write("NICK " + nick)

	// Wait until we receive end of MOTD.
	event := <-t.Event

	switch event {
	case irc.RPL_ENDOFMOTD:
		return true
	}

	return false
}

// In handles all incoming messages.
func (t *Twitch) In(input []byte) {
	// TODO: Create a map[string]func for handling commands.
	message := irc.MakeMessage(string(input))

	switch message.Command {
	case irc.RPL_CREATED:
	case irc.RPL_ENDOFMOTD:
		t.Event <- message.Command
	case irc.RPL_ENDOFNAMES:
		t.Event <- message.Command
	case irc.RPL_MOTD:
	case irc.RPL_MOTDSTART:
	case irc.RPL_MYINFO:
	case irc.RPL_NAMREPLY:
	case irc.RPL_WELCOME:
	case irc.RPL_YOURHOST:
	case "CAP":
	case "CLEARCHAT":
		if strings.Compare(message.Tags["ban-duration"], "") == 0 {
			t.Management.Bans++
		} else {
			t.Management.Timeouts++
		}
	case "JOIN":
	case "MODE":
	case "PART":
	case "PING":
		t.Session.Write("PONG :tmi.twitch.tv")
	case "PRIVMSG":
		t.Management.Messages++
	case "USERNOTICE":
	default:
		log.Printf("Unknown command: %v", message.Command)
	}
}

// Join sends a request to join the given channel.
// The blocking operation returns whether joining the channel was successful
func (t *Twitch) Join(channel string) bool {
	t.Session.Write("JOIN #" + channel)

	// Wait until we receive end of names.
	event := <-t.Event

	switch event {
	case irc.RPL_ENDOFNAMES:
		return true
	}

	return false
}

// Part leaves the given channel.
func (t *Twitch) Part(channel string) {
	t.Session.Write("PART #" + channel)
}

// PrivMSG sends a private message to the given channel.
func (t *Twitch) PrivMSG(channel, message string) {
	t.Session.Write("PRIVMSG #" + channel + " :" + message)
}

// Start creates additional goroutines for reading from the IRC server.
func (t *Twitch) Start() {
	go t.Management.Update()
	go t.Session.Listen(t)
}
