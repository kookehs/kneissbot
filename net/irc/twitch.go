package irc

import (
	"log"

	"golang.org/x/net/websocket"
)

const (
	// MaxMessageSize is a fixed message length as specified by RFC1459
	MaxMessageSize = 512
)

// Session contains variables required to interact with the IRC server
// over a websocket connection.
type Session struct {
	In        chan []byte
	Out       chan []byte
	Websocket *websocket.Conn
}

// NewSession creates and initializes a Session connected to the
// given URL.
func NewSession(origin, url string) (*Session, error) {
	ws, err := websocket.Dial(url, "", origin)

	if err != nil {
		return nil, err
	}

	// TODO: Buffered channels.
	return &Session{
		In:        make(chan []byte),
		Out:       make(chan []byte),
		Websocket: ws,
	}, nil
}

// Close forcefully shuts down the underlying websocket connection.
func (s *Session) Close() error {
	return s.Websocket.Close()
}

// Connect sends the given credentials to the IRC server for authentication.
func (s *Session) Connect(nick, token string) {
	s.Out <- []byte("PASS oauth:" + token)
	s.Out <- []byte("NICK " + nick)
}

// Read sends any incoming message from the IRC server to the In channel.
func (s *Session) Read() {
	for {
		message := make([]byte, MaxMessageSize)
		_, err := s.Websocket.Read(message)

		if err != nil {
			log.Println(err)
		}

		log.Println("in: " + string(message))
		s.In <- message
	}
}

// Start creates additional goroutines for reading and writing to the IRC server.
func (s *Session) Start() {
	go s.Read()
	go s.Write()
}

// Write sends any outgoing message from the Out channel to the IRC server.
func (s *Session) Write() {
	for {
		message := <-s.Out
		log.Println("out: " + string(message))

		if _, err := s.Websocket.Write(message); err != nil {
			log.Println(err)
		}
	}
}
