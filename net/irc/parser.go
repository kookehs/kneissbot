package irc

import (
	"regexp"
	"strings"
)

// Regular Expressions for searching
const (
	CommandFormat = `(?:(?P<Command>[a-zA-Z]+)|(?P<Command>\d{3}))`
	ParamFormat   = `(?P<Params>\s(?::(?P<Trailing>.*)|(?P<Middle>[^:]\w+)))*`
	PrefixFormat  = `:(?P<Prefix>(?P<Name>\S+?)(?:!(?P<User>\S+?))?(?:@(?P<Host>\S+?))?\s)?`
	TagFormat     = `(?:@(?P<Tags>\S+)\s)?`
)

// MessageRegExp is used to search for components of an IRC message through capture groups.
var MessageRegExp = regexp.MustCompile(TagFormat + PrefixFormat + CommandFormat + ParamFormat)

// Message is a RFC1459 message in the format specified by the protocol with support for IRCv3 tags.
// <message> ::= ['@' <tags> <SPACE>] [':' <prefix> <SPACE> ] <command> <params> <crlf>
type Message struct {
	Command string
	Params  []string
	Prefix  Prefix
	Tags    map[string]string
}

// MakeMessage creates a Message from the given string.
func MakeMessage(input string) Message {
	result := MessageRegExp.FindStringSubmatch(input)
	names := MessageRegExp.SubexpNames()
	message := Message{
		Params: make([]string, 0),
		Prefix: Prefix{},
		Tags:   make(map[string]string),
	}

	for i, v := range result {
		switch names[i] {
		// Commands
		case "Command":
			if len(message.Command) == 0 {
				message.Command = v
			}
		// Params
		case "Middle":
			fallthrough
		case "Trailing":
			if len(v) != 0 {
				message.Params = append(message.Params, v)
			}
		// Prefix
		case "Host":
			message.Prefix.Host = v
		case "Name":
			message.Prefix.Name = v
		case "User":
			message.Prefix.User = v
		// Tags
		case "Tags":
			if len(v) != 0 {
				message.Tags = MakeTags(v)
			}
		}
	}

	return message
}

// Prefix is an optional structure within a Message.
// Prefix contains information about the user who sent the message.
// <prefix> ::= <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
type Prefix struct {
	Host string
	Name string
	User string
}

// MakeTags creates a Tags from the given string.
// Tags are the IRCv3 message tags.
// <tags>	::= <tag> [';' <tag>]*
// <tag>	::= <key> ['=' <escaped value>]
func MakeTags(input string) map[string]string {
	tags := make(map[string]string)
	pairs := strings.Split(input, ";")

	for _, v := range pairs {
		pair := strings.SplitN(v, "=", 2)

		if len(pair) == 1 {
			tags[pair[0]] = ""
			continue
		}

		tags[pair[0]] = pair[1]
	}

	return tags
}
