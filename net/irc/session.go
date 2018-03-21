package irc

import (
	"bytes"
	"io"
	"log"

	"golang.org/x/net/websocket"
)

const (
	// IRC numeric reply definitions
	RPL_WELCOME    = "001"
	RPL_YOURHOST   = "002"
	RPL_CREATED    = "003"
	RPL_MYINFO     = "004"
	RPL_NAMREPLY   = "353"
	RPL_ENDOFNAMES = "366"
	RPL_MOTD       = "372"
	RPL_MOTDSTART  = "375"
	RPL_ENDOFMOTD  = "376"
)

var (
	// MaxMessageSize is a fixed message length in bytes as specified by RFC1459
	MaxMessageSize = 512
)

// Session contains variables required to interact with the IRC server
// over a websocket connection.
type Session struct {
	Websocket *websocket.Conn
}

// NewSession creates and initializes a Session connected to the given URL.
func NewSession(origin, url string) (*Session, error) {
	ws, err := websocket.Dial(url, "", origin)

	if err != nil {
		return nil, err
	}

	return &Session{
		Websocket: ws,
	}, nil
}

// Close forcefully shuts down the underlying websocket connection.
func (s *Session) Close() error {
	return s.Websocket.Close()
}

// Listen sends any incoming message from the IRC server to the handler.
// Listen should be run as a goroutine. Handler functions are ran as goroutines.
func (s *Session) Listen(handler Handler) {
	for {
		buffer := make([]byte, MaxMessageSize)
		_, err := s.Websocket.Read(buffer)

		if err != nil {
			log.Println(err)

			if err == io.EOF {
				// TODO: Reconnect to IRC.
				panic("[IRC]: Lost connection to IRC server")
			}

			continue
		}

		messages := bytes.Split(bytes.TrimRight(buffer, "\x00"), []byte{'\n'})

		for _, message := range messages {
			if len(message) > 0 {
				log.Println("[IRC]: " + string(message))
				go handler.In(message)
			}
		}
	}
}

// Write sends an outgoing message to the IRC server.
func (s *Session) Write(message string) {
	out := []byte(message)
	log.Println("[IRC]: " + message)

	if _, err := s.Websocket.Write(out); err != nil {
		log.Println(err)
	}
}
