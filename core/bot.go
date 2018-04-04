package core

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kookehs/kneissbot/net/api/twitch"
	"github.com/kookehs/kneissbot/net/irc"
	"github.com/kookehs/kneissbot/net/server"
	"github.com/kookehs/watchmen/primitives"
)

// TODO: Blacklist of users (never moderator)
// TODO: Whitelist of users (always moderator)
// TODO: Provide uptime and share statistics of delegates

const (
	// BalanceFormat defines the search pattern for retrieving users' balances.
	BalanceFormat = `!balance ((?:\w+ ?)*)`
	// CommandFormat defines the search pattern for a command.
	CommandFormat = `!(\w+)`
	// ModeratorFormat defines the search pattern for a list of moderators.
	ModeratorFormat = `The moderators of this channel are: ((?:\w+(?:, )?)+)`
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
	// ModeratorRegExp is the regular expression used to find moderators.
	ModeratorRegExp = regexp.MustCompile(ModeratorFormat)
	// SendRegExp is the regular expression used to find the receiver.
	SendRegExp = regexp.MustCompile(SendFormat)
	// VoteRegExp is the regular expression used to find the delegates elected.
	VoteRegExp = regexp.MustCompile(VoteFormat)

	// BotCommands is a mapping of strings to functions related to the bot.
	BotCommands = make(map[string]func(*Bot, irc.Message))
	// TwitchCommands is a mapping of strings to functions related to IRC.
	TwitchCommands = make(map[string]func(*Bot, irc.Message))

	// UpdateInterval is the time in seconds for an update to trigger.
	UpdateInterval time.Duration = 60
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
	TwitchCommands["NOTICE"] = Notice
	TwitchCommands["PING"] = Ping
	TwitchCommands["PRIVMSG"] = PrivMSG

	// Twitch has a 500 character limit, not including line endings, not a 512 byte limit.
	irc.MaxMessageSize = 512 * utf8.UTFMax
}

// Bot contains logic realted to both the API and IRC.
type Bot struct {
	API        *twitch.API
	Config     *Config
	Event      chan irc.Message
	Management *Management
	Session    *irc.Session
	Timer      *time.Timer
}

// NewBot returns a pointer to an initialized Bot struct.
func NewBot() (*Bot, error) {
	bot := new(Bot)
	session, err := irc.NewSession(twitch.Origin, twitch.IRC)

	if err != nil {
		return nil, err
	}

	bot.Config = NewConfig()
	home, err := Home()

	if err != nil {
		return nil, err
	}

	path := home + "/.kneissbot"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModeDir)

		if err != nil {
			log.Println(err)
		}
	}

	bot.Config.Files["config"] = path + "/config.bin"
	bot.Config.Files["ledger"] = path + "/ledger.bin"
	bot.Config.Files["ma"] = path + "/ma.bin"

	bot.Event = make(chan irc.Message)
	bot.Management = NewManagement(bot, bot.Config.Twitch.Username)
	bot.Session = session
	bot.Timer = time.NewTimer(UpdateInterval * time.Second)

	if _, err := os.Stat(bot.Config.Files["config"]); os.IsNotExist(err) {
		output := make(chan string)
		defer close(output)
		twitchAuth := server.NewTwitchAuth(output)
		go twitchAuth.ListenAndServe()

		if err := twitchAuth.Authenticate(); err != nil {
			panic(err)
		}

		defer twitchAuth.Close()
		bot.Config.Twitch.AccessToken = <-output
	} else {
		bot.Deserialize()
		log.Println(bot.Config.Twitch.AccessToken)
	}

	bot.API = twitch.NewAPI(bot.Config.Twitch.AccessToken)
	response, err := bot.API.ValidToken()

	if err != nil || !response.Token.Valid {
		if !response.Token.Valid {
			return nil, errors.New("Invalid token")
		}

		return nil, err
	}

	bot.Config.Twitch.Username = response.Token.Username
	return bot, nil
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

	bot.PrivMSG(buffer.String())
}

// ClearChat is the handler for the CLEARCHAT command sent from IRC.
func ClearChat(bot *Bot, message irc.Message) {
	if strings.Compare(message.Tags["ban-duration"], "") == 0 {
		bot.Management.Bans++
	} else {
		bot.Management.Timeouts++
	}
}

// Deserialize retrieves state information of the bot from disk.
func (b *Bot) Deserialize() {
	configPath := b.Config.Files["config"]
	configFile, err := os.OpenFile(configPath, os.O_RDONLY, 0666)

	if err == nil {
		defer configFile.Close()
		reader := bufio.NewReader(configFile)
		b.Config.Deserialize(reader)
	}

	ledgerPath := b.Config.Files["ledger"]
	ledgerFile, err := os.OpenFile(ledgerPath, os.O_RDONLY, 0666)

	if err == nil {
		defer ledgerFile.Close()
		reader := bufio.NewReader(ledgerFile)
		b.Management.Ledger.Deserialize(reader)
	}

	maPath := b.Config.Files["ma"]
	maFile, err := os.OpenFile(maPath, os.O_RDONLY, 0666)

	if err == nil {
		defer maFile.Close()
		reader := bufio.NewReader(maFile)
		b.Management.MovingAverage.Deserialize(reader)
	}
}

// EndOfMOTD is the handler for the ENDOFMOTD command sent from IRC.
func EndOfMOTD(bot *Bot, message irc.Message) {
	bot.Event <- message
}

// EndOfNames is the handler for the ENDOFNAMES command sent from IRC.
func EndOfNames(bot *Bot, message irc.Message) {
	bot.Event <- message
}

// Notice is the handler for the NOTICE command sent from IRC.
func Notice(bot *Bot, message irc.Message) {
	bot.Event <- message
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

// Amend makes changes to the current set of moderators.
func (b *Bot) Amend(moderators []string) {
	// Retrieve current list of moderators.
	b.PrivMSG("/mods")
	message := <-b.Event
	id, ok := message.Tags["msg-id"]

	if !ok || strings.Compare(id, "room_mods") != 0 {
		return
	}

	// A mapping of users to mod or unmod.
	amendment := make(map[string]bool)

	for _, moderator := range moderators {
		amendment[moderator] = true
	}

	for _, param := range message.Params {
		matches := ModeratorRegExp.FindStringSubmatch(param)

		if matches == nil || len(matches) < 1 {
			continue
		}

		mods := strings.Split(matches[1], ", ")

		for _, mod := range mods {
			if _, exist := amendment[mod]; !exist {
				amendment[mod] = false
			}
		}

		break
	}

	for moderator, operator := range amendment {
		if operator {
			b.PrivMSG("/mod " + moderator)
		} else {
			b.PrivMSG("/unmod " + moderator)
		}
	}
}

// Available returns whether or not the given user is online.
func (b *Bot) Available(username string) bool {
	resp, err := b.API.GetChatters(b.Config.Twitch.Username)

	if err != nil {
		log.Println(err)
		return false
	}

	for _, admin := range resp.Chatters.Admins {
		if strings.Compare(username, admin) == 0 {
			return true
		}
	}

	for _, mod := range resp.Chatters.GlobalMods {
		if strings.Compare(username, mod) == 0 {
			return true
		}
	}

	for _, mod := range resp.Chatters.Moderators {
		if strings.Compare(username, mod) == 0 {
			return true
		}
	}

	for _, staff := range resp.Chatters.Staff {
		if strings.Compare(username, staff) == 0 {
			return true
		}
	}

	for _, viewer := range resp.Chatters.Viewers {
		if strings.Compare(username, viewer) == 0 {
			return true
		}
	}

	return false
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
func (b *Bot) Connect() bool {
	b.Session.Write("PASS oauth:" + b.Config.Twitch.AccessToken)
	b.Session.Write("NICK " + b.Config.Twitch.Username)

	// Wait until we receive end of MOTD.
	message := <-b.Event

	switch message.Command {
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
	message := <-b.Event

	switch message.Command {
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

// PrivMSG sends a private message.
func (b *Bot) PrivMSG(message string) {
	b.Session.Write("PRIVMSG #" + b.Config.Twitch.Username + " :" + message)
}

// Serialize stores state information of the bot to disk in byte data.
func (b *Bot) Serialize() {
	// TODO: Encrypt data.
	configPath := b.Config.Files["config"]
	configFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	defer configFile.Close()
	writer := bufio.NewWriter(configFile)
	b.Config.Serialize(writer)
	writer.Flush()

	ledgerPath := b.Config.Files["ledger"]
	ledgerFile, err := os.OpenFile(ledgerPath, os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	defer ledgerFile.Close()
	writer = bufio.NewWriter(ledgerFile)
	b.Management.Ledger.Serialize(writer)
	writer.Flush()

	maPath := b.Config.Files["ma"]
	maFile, err := os.OpenFile(maPath, os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	defer maFile.Close()
	writer = bufio.NewWriter(maFile)
	b.Management.MovingAverage.Serialize(writer)
	writer.Flush()
}

// Start creates additional goroutines for reading from the IRC server.
func (b *Bot) Start() {
	go b.Update()
	go b.Session.Listen(b)
}

// Update calls nested update functions and applies changes to moderators.
func (b *Bot) Update() {
	for {
		<-b.Timer.C
		b.Management.Update()
		moderators := make([]string, 0)

		for _, moderator := range b.Management.DPoS.Round.Forgers {
			account := moderator.Account
			username := b.Management.Ledger.Username(account.IBAN)
			moderators = append(moderators, username)
		}

		b.Amend(moderators)
		b.Serialize()
		b.Timer.Reset(UpdateInterval * time.Second)
	}
}
