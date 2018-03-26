package core

import (
	"strings"
	"unicode/utf8"

	"github.com/kookehs/kneissbot/net/irc"
)

const (
	// URLs
	TwitchOrigin = "http://twitch.tv"
	TwitchIRC    = "ws://irc-ws.chat.twitch.tv:80"

	// Capabilities
	CommandsCapability   = "twitch.tv/commands"
	MembershipCapability = "twitch.tv/membership"
	TagsCapability       = "twitch.tv/tags"
)

var (
	Commands            = make(map[string]func(irc.Message, *Twitch))
	DefaultCapabilities = []string{CommandsCapability, MembershipCapability, TagsCapability}
)

func init() {
	// Set up mapping of commands to functions.
	Commands["CLEARCHAT"] = ClearChat
	Commands[irc.RPL_ENDOFMOTD] = EndOfMOTD
	Commands[irc.RPL_ENDOFNAMES] = EndOfNames
	Commands["PING"] = Ping
	Commands["PRIVMSG"] = PrivMSG

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

// ClearChat is the handler for the CLEARCHAT command sent from IRC.
func ClearChat(message irc.Message, twitch *Twitch) {
	if strings.Compare(message.Tags["ban-duration"], "") == 0 {
		twitch.Management.Bans++
	} else {
		twitch.Management.Timeouts++
	}
}

// EndOfMOTD is the handler for the ENDOFMOTD command sent from IRC.
func EndOfMOTD(message irc.Message, twitch *Twitch) {
	twitch.Event <- message.Command
}

// EndOfNames is the handler for the ENDOFNAMES command sent from IRC.
func EndOfNames(message irc.Message, twitch *Twitch) {
	twitch.Event <- message.Command
}

// Ping is the handler for the PING command sent from IRC.
func Ping(message irc.Message, twitch *Twitch) {
	twitch.Session.Write("PONG :tmi.twitch.tv")
}

// PrivMSG is the handler for the PRIVMSG command sent from IRC.
func PrivMSG(message irc.Message, twitch *Twitch) {
	// TODO: Handle bot commands
	username := message.Prefix.User

	// TODO: Remove this temporary code.
	if _, ok := twitch.Management.Ledger.Users[username]; !ok {
		twitch.Management.Ledger.OpenAccount(twitch.Management.Node, username)
	}

	twitch.Management.Messages++
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
	message := irc.MakeMessage(string(input))

	if callback, ok := Commands[message.Command]; ok {
		callback(message, t)
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
