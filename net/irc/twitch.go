package irc

import (
	"log"

	"golang.org/x/net/websocket"
)

const (
	MaxMessageSize = 512
)

type Session struct {
	Websocket *websocket.Conn
	In        chan<- []byte
	Out       <-chan []byte
}

func NewSession(o, u string) (*Session, error) {
	ws, err := websocket.Dial(u, "", o)

	if err != nil {
		return nil, err
	}

	return &Session{
		Websocket: ws,
	}, nil
}

func (s *Session) Close() error {
	return s.Websocket.Close()
}

func (s *Session) Read() {
	for {
		message := make([]byte, MaxMessageSize)
		_, err := s.Websocket.Read(message)

		if err != nil {
			log.Println(err)
		}

		s.In <- message
	}
}

func (s *Session) Write() {
	for {
		select {
		case message := <-s.Out:
			if _, err := s.Websocket.Write(message); err != nil {
				log.Println(err)
			}
		}
	}
}
