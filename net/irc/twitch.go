package irc

import (
	"log"

	"golang.org/x/net/websocket"
)

const (
	MaxMessageSize = 512
)

type Session struct {
	In        chan []byte
	Out       chan []byte
	Websocket *websocket.Conn
}

func NewSession(o, u string) (*Session, error) {
	ws, err := websocket.Dial(u, "", o)

	if err != nil {
		return nil, err
	}

	return &Session{
		In:        make(chan []byte),
		Out:       make(chan []byte),
		Websocket: ws,
	}, nil
}

func (s *Session) Close() error {
	return s.Websocket.Close()
}

func (s *Session) Connect(nick, token string) {
	s.Out <- []byte("PASS oauth:" + token)
	s.Out <- []byte("NICK " + nick)
}

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

func (s *Session) Start() {
	go s.Read()
	go s.Write()
}

func (s *Session) Write() {
	for {
		select {
		case message := <-s.Out:
			log.Println("out: " + string(message))

			if _, err := s.Websocket.Write(message); err != nil {
				log.Println(err)
			}
		}
	}
}
