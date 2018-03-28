package core

import (
	"bytes"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/kookehs/kneissbot/net/irc"
	"github.com/kookehs/kneissbot/net/irc/twitch"
	"github.com/kookehs/watchmen/primitives"
)

// TODO: Blacklist of users (never moderator)
// TODO: Whitelist of users (always moderator?)
// TODO: Handle making people moderators / not moderators

const (
	// BalanceFormat defines the search pattern for retrieving users' balances.
	BalanceFormat = `!balance ((?:\w+ ?)*)`
	// CommandFormat defines the search pattern for a command.
	CommandFormat = `!(\w+)`
	// SendFormat defines the search pattern for sending funds to another user.
	SendFormat = `!send (\w+) (\d+)`
	// VoteFormat defines the search pattern for voting for delegates.
	VoteFormat = `!vote ((?:[+|-]\w+ ?)+)`
)

var (
	// BalanceRegExp is the regular expression used to find balance(s) for given user(s).
	BalanceRegExp = regexp.MustCompile(BalanceFormat)
	// CommandRegExp is the regular expression used to find commands.
	CommandRegExp = regexp.MustCompile(CommandFormat)
	// SendRegExp is the regular expression used to find the receiver.
	SendRegExp = regexp.MustCompile(SendFormat)
	// VoteRegExp is the regular expression used to find the delegates elected.
	VoteRegExp = regexp.MustCompile(VoteFormat)

	// BotCommands is a mapping of strings to functions related to the bot.
	BotCommands = make(map[string]func(*Bot, irc.Message))
	// TwitchCommands is a mapping of strings to functions related to IRC.
	TwitchCommands = make(map[string]func(*Bot, irc.Message))
)

func init() {
	// Set up mapping of commands to functions.
	BotCommands["balance"] = Balance
	BotCommands["register"] = Register
	BotCommands["send"] = Send
	BotCommands["vote"] = Vote
	TwitchCommands["CLEARCHAT"] = ClearChat
	TwitchCommands[irc.RPL_ENDOFMOTD] = EndOfMOTD
	TwitchCommands[irc.RPL_ENDOFNAMES] = EndOfNames
	TwitchCommands["PING"] = Ping
	TwitchCommands["PRIVMSG"] = PrivMSG

	// Twitch has a 500 character limit, not including line endings, not a 512 byte limit.
	irc.MaxMessageSize = 512 * utf8.UTFMax
}

// Bot contains logic realted to both the API and IRC.
type Bot struct {
	Event      chan string
	Management *Management
	Session    *irc.Session
}

// NewBot returns a pointer to an initialized Bot struct.
func NewBot(username string) (*Bot, error) {
	session, err := irc.NewSession(twitch.Origin, twitch.IRC)

	if err != nil {
		return nil, err
	}

	return &Bot{
		Event:      make(chan string),
		Management: NewManagement(username),
		Session:    session,
	}, nil
}

// Balance returns the user's balance or a set of users.
func Balance(bot *Bot, message irc.Message) {
	username := message.Prefix.User
	users := make([]string, 0)

	for _, param := range message.Params {
		matches := BalanceRegExp.FindStringSubmatch(param)

		if matches == nil || len(matches) < 1 {
			continue
		}

		users = strings.Fields(matches[1])
		break
	}

	if len(users) == 0 {
		users = append(users, username)
	}

	// Respond with a whisper to the user who requested the balance(s).
	buffer := bytes.NewBufferString("/w ")
	buffer.WriteString(username)
	buffer.WriteByte(' ')

	for _, user := range users {
		iban := bot.Management.Ledger.Users[user]

		if block := bot.Management.Ledger.LatestBlock(iban); block != nil {
			balance := block.Balance()
			buffer.WriteString(user)
			buffer.WriteByte('(')
			buffer.WriteString(balance.Text('f', -1))
			buffer.WriteString(") ")
		}
	}

	bot.Session.Write(buffer.String())
}

// ClearChat is the handler for the CLEARCHAT command sent from IRC.
func ClearChat(bot *Bot, message irc.Message) {
	if strings.Compare(message.Tags["ban-duration"], "") == 0 {
		bot.Management.Bans++
	} else {
		bot.Management.Timeouts++
	}
}

// EndOfMOTD is the handler for the ENDOFMOTD command sent from IRC.
func EndOfMOTD(bot *Bot, message irc.Message) {
	bot.Event <- message.Command
}

// EndOfNames is the handler for the ENDOFNAMES command sent from IRC.
func EndOfNames(bot *Bot, message irc.Message) {
	bot.Event <- message.Command
}

// Ping is the handler for the PING command sent from IRC.
func Ping(bot *Bot, message irc.Message) {
	bot.Session.Write("PONG :tmi.twitch.tv")
}

// PrivMSG is the handler for the PRIVMSG command sent from IRC.
func PrivMSG(bot *Bot, message irc.Message) {
	bot.Management.Messages++
	bot.ParseCommand(message)
}

// Register creates an account for the given user in the ledger.
func Register(bot *Bot, message irc.Message) {
	username := message.Prefix.User

	if _, ok := bot.Management.Ledger.Users[username]; !ok {
		bot.Management.Ledger.OpenAccount(bot.Management.Node, username)
	}
}

// Send sends tokens from one user to another.
func Send(bot *Bot, message irc.Message) {
	username := message.Prefix.User

	for _, param := range message.Params {
		matches := SendRegExp.FindStringSubmatch(param)

		if matches == nil || len(matches) < 2 {
			continue
		}

		receiver := matches[1]
		amount, err := strconv.Atoi(matches[2])

		if err != nil {
			log.Println(err)
			return
		}

		funds := primitives.NewAmount(float64(amount))
		src := bot.Management.Ledger.Users[username]
		dst := bot.Management.Ledger.Users[receiver]
		_, err = bot.Management.Ledger.Transfer(funds, dst, src, bot.Management.Node)

		if err != nil {
			log.Println(err)
			return
		}

		break
	}
}

// Vote handles a user's choice to add or remove delegates.
func Vote(bot *Bot, message irc.Message) {
	username := message.Prefix.User

	for _, param := range message.Params {
		matches := VoteRegExp.FindStringSubmatch(param)

		if matches == nil || len(matches) < 1 {
			continue
		}

		delegates := strings.Fields(matches[1])
		iban := bot.Management.Ledger.Users[username]
		account := bot.Management.Ledger.Accounts[iban.String()]

		if err := bot.Management.DPoS.Elect(account, delegates, bot.Management.Ledger, bot.Management.Node); err != nil {
			log.Println(err)
			continue
		}

		break
	}
}

// Cap sends a request for the given capabilities.
// Default capabilities are used if none are given.
func (b *Bot) Cap(capabilities []string) {
	caps := []string{
		twitch.CommandsCapability,
		twitch.MembershipCapability,
		twitch.TagsCapability,
	}

	if capabilities != nil && len(capabilities) > 0 {
		caps = capabilities
	}

	for _, cap := range caps {
		b.Session.Write("CAP REQ :" + cap)
	}
}

// Close shuts down channels and the underlying websocket connection.
func (b *Bot) Close() error {
	close(b.Event)

	if err := b.Session.Close(); err != nil {
		return err
	}

	return nil
}

// Connect sends the given credentials to the IRC server for authentication.
// The blocking operation returns whether connecting to IRC server was successful
func (b *Bot) Connect(nick, token string) bool {
	b.Session.Write("PASS oauth:" + token)
	b.Session.Write("NICK " + nick)

	// Wait until we receive end of MOTD.
	event := <-b.Event

	switch event {
	case irc.RPL_ENDOFMOTD:
		return true
	}

	return false
}

// In handles all incoming messages.
func (b *Bot) In(input []byte) {
	message := irc.MakeMessage(string(input))

	if callback, ok := TwitchCommands[message.Command]; ok {
		callback(b, message)
	}
}

// Join sends a request to join the given channel.
// The blocking operation returns whether joining the channel was successful
func (b *Bot) Join(channel string) bool {
	b.Session.Write("JOIN #" + channel)

	// Wait until we receive end of names.
	event := <-b.Event

	switch event {
	case irc.RPL_ENDOFNAMES:
		return true
	}

	return false
}

// ParseCommand parses commands related to the bot.
func (b *Bot) ParseCommand(message irc.Message) {
	for _, param := range message.Params {
		matches := CommandRegExp.FindStringSubmatch(param)

		if matches == nil || len(matches) < 1 {
			continue
		}

		command := matches[1]

		if callback, ok := BotCommands[command]; ok {
			callback(b, message)
		}

		break
	}
}

// Part leaves the given channel.
func (b *Bot) Part(channel string) {
	b.Session.Write("PART #" + channel)
}

// PrivMSG sends a private message to the given channel.
func (b *Bot) PrivMSG(channel, message string) {
	b.Session.Write("PRIVMSG #" + channel + " :" + message)
}

// Start creates additional goroutines for reading from the IRC server.
func (b *Bot) Start() {
	go b.Management.Update()
	go b.Session.Listen(b)
}
