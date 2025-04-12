package main

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/wspowell/tabletop/account"
	"github.com/wspowell/tabletop/message"
)

type Session struct {
	Connection        *websocket.Conn
	LastKeepAliveTime time.Time
	User              account.User
}

func NewSession(user account.User, connection *websocket.Conn) *Session {
	return &Session{
		User:              user,
		Connection:        connection,
		LastKeepAliveTime: time.Now(),
	}
}

func (self *Session) SendBytes(payloadBytes []byte) error {
	return self.Connection.WriteMessage(websocket.TextMessage, payloadBytes)
}

func (self *Session) SendMessage(payload message.Payload) error {
	payloadBytes, err := message.Marshal(payload)
	if err != nil {
		return err
	}
	return self.Connection.WriteMessage(websocket.TextMessage, payloadBytes)
}
