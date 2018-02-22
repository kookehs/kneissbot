package irc

// Message is a RFC1459 message in the format specified by the protocol.
// <message> ::= ['@' <tags> <SPACE>] [':' <prefix> <SPACE> ] <command> <params> <crlf>
type Message struct {
	Command string
	Params  []string
	Prefix  Prefix
	Tags    Tags
}

// Prefix is an optional structure within a Message.
// Prefix contains information about the user who sent the message.
// <prefix> ::= <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
type Prefix struct {
	Host string
	Name string
	User string
}

// Tags are the IRCv3 message tags.
// <tags>	::= <tag> [';' <tag>]*
// <tag>	::= <key> ['=' <escaped value>]
type Tags map[string]string
