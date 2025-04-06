package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wspowell/tabletop/account"
	"github.com/wspowell/tabletop/game"
	"github.com/wspowell/tabletop/message"
)

func serveWebSocket(ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	const serverAddress = ":3000"

	gameState, err := game.LoadState()
	if err != nil {
		log.Fatal(err)
		return
	}

	defer gameState.Save()

	webSocketHandler := &webSocketHandler{
		upgrader: websocket.Upgrader{
			HandshakeTimeout:  5 * time.Second,
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			WriteBufferPool:   nil,
			Subprotocols:      nil,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				log.Printf("origin: %s", origin)
				return true // origin == "http://0.0.0.0:8080"
			},
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				log.Printf("websocket error: %d, %s", status, reason)
			},
		},
		sessionsMutex: sync.RWMutex{},
		sessions:      map[string]*Session{},
		state:         gameState,
	}

	http.Handle("/websocket", webSocketHandler)

	log.Print("Starting server...")

	log.Fatal(http.ListenAndServe(serverAddress, nil))
}

type webSocketHandler struct {
	upgrader websocket.Upgrader

	sessionsMutex sync.RWMutex
	sessions      map[string]*Session

	state *game.State
}

func (self *webSocketHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	self.upgradeRequestToWebSocket(responseWriter, request)
}

func (self *webSocketHandler) upgradeRequestToWebSocket(responseWriter http.ResponseWriter, request *http.Request) {
	connection, err := self.upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		log.Printf("error when upgrading connection to websocket: %s", err)
		return
	}

	log.Printf("upgraded request to websocket")

	// connection.SetPingHandler(func(appData string) error {
	// 	log.Printf("received ping: %s", appData)
	// 	return nil
	// })

	self.handleWebSocket(connection)
}

func (self *webSocketHandler) handleWebSocket(connection *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %s", err)
		}
	}()

	defer func() {
		log.Println("closing connection")
		connection.Close()
	}()

	log.Printf("waiting for login")

	session, err := self.waitForLogin(connection)
	if err != nil {
		return
	}

	defer func() {
		log.Printf("logging out: %+v", session)
		self.onLogout(session)
	}()

	for {
		if err := self.handleNextMessage(session); err != nil {
			return
		}
	}
}

func (self *webSocketHandler) onLogout(session *Session) {
	_ = session.SendMessage(message.Logout{
		Username: session.User.Username,
	})

	_ = self.broadcastToAllSessions(message.UserOffline{
		Username: session.User.Username,
	})

	_ = self.state.Save()
}

func (self *webSocketHandler) messageSession(username string, payload message.Payload) error {
	self.sessionsMutex.RLock()
	defer self.sessionsMutex.RUnlock()

	session, exists := self.sessions[username]
	if !exists {
		return ErrUserNotConnected
	}

	return session.SendMessage(payload)
}

func (self *webSocketHandler) broadcastToAllSessions(payload message.Payload) error {
	payloadBytes, err := message.Marshal(payload)
	if err != nil {
		return err
	}

	self.sessionsMutex.RLock()
	defer self.sessionsMutex.RUnlock()

	for _, session := range self.sessions {
		_ = session.SendBytes(payloadBytes)
	}

	return nil
}

func (self *webSocketHandler) broadcastToOtherSessions(session *Session, payload message.Payload) error {
	payloadBytes, err := message.Marshal(payload)
	if err != nil {
		return err
	}

	self.sessionsMutex.RLock()
	defer self.sessionsMutex.RUnlock()

	for _, otherSession := range self.sessions {
		if session.User.Username == otherSession.User.Username {
			continue
		}

		_ = otherSession.SendBytes(payloadBytes)
	}

	return nil
}

var (
	ErrLoginFailure     = errors.New("login failure")
	ErrUnknownMessage   = errors.New("unknown message")
	ErrUserNotConnected = errors.New("user not connected")
)

func (self *webSocketHandler) waitForLogin(connection *websocket.Conn) (*Session, error) {
	for {
		_, data, err := connection.ReadMessage()
		if err != nil {
			log.Printf("error when reading message from client: %s", err)
			return nil, err
		}

		// Detect the payload based on the type.
		typedData, err := message.UnmarshalType(data)
		if err != nil {
			log.Println(err)
			continue
		}

		if typedData.Type != "login" {
			err := fmt.Errorf("%w: must login before sending messages", ErrLoginFailure)
			log.Println(err)

			payloadBytes, _ := message.Marshal(message.Error{
				TypeOfFailedMessage: "login",
				ErrorMessage:        "invalid login message",
			})
			_ = connection.WriteMessage(websocket.TextMessage, payloadBytes)
			continue
		}

		var payload message.Login
		if err := message.Unmarshal(typedData, &payload); err != nil {
			log.Println(err)

			payloadBytes, _ := message.Marshal(message.Error{
				TypeOfFailedMessage: "login",
				ErrorMessage:        "invalid login message",
			})
			_ = connection.WriteMessage(websocket.TextMessage, payloadBytes)
			continue
		}

		if payload.Secret == "" {
			err := fmt.Errorf("%w: secret is blank", ErrLoginFailure)
			log.Println(err)

			payloadBytes, _ := message.Marshal(message.Error{
				TypeOfFailedMessage: "login",
				ErrorMessage:        "invalid username/secret",
			})
			_ = connection.WriteMessage(websocket.TextMessage, payloadBytes)
			continue
		}

		user, err := account.LoadUser(payload.Username)
		if err != nil {
			payloadBytes, _ := message.Marshal(message.Error{
				TypeOfFailedMessage: "login",
				ErrorMessage:        "internal server failure",
			})
			_ = connection.WriteMessage(websocket.TextMessage, payloadBytes)
			return nil, err
		}

		if !user.Authenticate(payload.Secret) {
			payloadBytes, _ := message.Marshal(message.Error{
				TypeOfFailedMessage: "login",
				ErrorMessage:        "invalid username/secret",
			})
			_ = connection.WriteMessage(websocket.TextMessage, payloadBytes)
			continue
		}

		session := NewSession(user, connection)

		self.sessionsMutex.Lock()
		self.sessions[payload.Username] = session
		self.sessionsMutex.Unlock()

		session.SendMessage(message.LoginSuccess{
			User: user,
		})

		// Send all sessions this session username to trigger it as online for everyone.
		self.broadcastToAllSessions(message.UserOnline{
			User: user,
		})

		// Send the session the other usernames so that it sees them as online.
		self.sessionsMutex.RLock()
		defer self.sessionsMutex.RUnlock()
		for _, otherSession := range self.sessions {
			if otherSession.User.Username == payload.Username {
				continue
			}

			session.SendMessage(message.UserOnline{
				User: otherSession.User,
			})
		}

		// Get list of maps and send them to the user.
		// Only do this for the Story Teller.
		// if session.User.IsStoryTeller {
		sendMapList(session)
		sendTokenBank(session)
		// }

		// Push current state to all users so they also know where to put your token.
		self.broadcastToAllSessions(message.State{
			State:   self.state,
			MapData: getMapData(self.state.CurrentMap),
		})

		log.Printf("session logged in: %s", session.User.Username)

		return session, nil
	}
}

func sendMapList(session *Session) {
	dirFiles, err := os.ReadDir("." + string(filepath.Separator) + filepath.Join("data", "maps"))
	if err != nil {
		err := fmt.Errorf("%w: failed reading map list, %s", ErrLoginFailure, err)
		log.Println(err)
	}

	mapNames := []string{}
	for index := range dirFiles {
		file := dirFiles[index]

		if strings.HasSuffix(file.Name(), ".png") {
			mapName, _, _ := strings.Cut(file.Name(), ".png")
			mapNames = append(mapNames, mapName)
		}
	}

	session.SendMessage(message.MapList{
		MapNames: mapNames,
	})
}

func sendTokenBank(session *Session) {
	dirFiles, err := os.ReadDir("." + string(filepath.Separator) + filepath.Join("data", "npcs"))
	if err != nil {
		err := fmt.Errorf("%w: failed reading token bank, %s", ErrLoginFailure, err)
		log.Println(err)
	}

	tokenNames := []string{}
	for index := range dirFiles {
		file := dirFiles[index]

		if strings.HasSuffix(file.Name(), ".png") {
			tokenName, _, _ := strings.Cut(file.Name(), ".png")
			tokenNames = append(tokenNames, tokenName)
		}
	}

	session.SendMessage(message.TokenBank{
		TokenNames: tokenNames,
	})
}

func (self *webSocketHandler) handleNextMessage(session *Session) error {
	_, data, err := session.Connection.ReadMessage()
	if err != nil {
		log.Printf("error when reading message from client: %s", err)
		return err
	}

	log.Println("received message", string(data))

	// Detect the payload based on the type.
	typedData, err := message.UnmarshalType(data)
	if err != nil {
		log.Println(err)
		return err
	}

	switch typedData.Type {
	case "login":
		err := fmt.Errorf("%w: already logged in", ErrLoginFailure)

		_ = session.SendMessage(message.Error{
			TypeOfFailedMessage: typedData.Type,
			ErrorMessage:        "already logged in",
		})

		return err
	case "tokenPosition":
		var payload message.TokenPosition
		if err := message.Unmarshal(typedData, &payload); err != nil {
			log.Println(err)
			_ = session.SendMessage(message.Error{
				TypeOfFailedMessage: typedData.Type,
				ErrorMessage:        "failed reading message",
			})

			return err
		}

		self.onTokenPosition(session, payload)
	case "deleteToken":
		var payload message.DeleteToken
		if err := message.Unmarshal(typedData, &payload); err != nil {
			log.Println(err)
			_ = session.SendMessage(message.Error{
				TypeOfFailedMessage: typedData.Type,
				ErrorMessage:        "failed reading message",
			})

			return err
		}

		self.onDeleteToken(session, payload)
	case "mapSize":
		var payload message.MapSize
		if err := message.Unmarshal(typedData, &payload); err != nil {
			log.Println(err)
			_ = session.SendMessage(message.Error{
				TypeOfFailedMessage: typedData.Type,
				ErrorMessage:        "failed reading message",
			})

			return err
		}

		self.onMapSize(payload)
	case "mapChange":
		var payload message.MapChange
		if err := message.Unmarshal(typedData, &payload); err != nil {
			log.Println(err)
			_ = session.SendMessage(message.Error{
				TypeOfFailedMessage: typedData.Type,
				ErrorMessage:        "failed reading message",
			})

			return err
		}

		self.onMapChange(payload)
	default:
		err := fmt.Errorf("%w: %s", ErrUnknownMessage, data)
		log.Println(err)
		_ = session.SendMessage(message.Error{
			TypeOfFailedMessage: typedData.Type,
			ErrorMessage:        "unknown message type",
		})
		return err
	}

	return nil
}

func (self *webSocketHandler) onTokenPosition(session *Session, payload message.TokenPosition) {
	// Update the game state with the location of the token
	self.state.SetTokenPosition(payload.Id, payload.TokenName, payload.MapName, game.TokenData{
		TokenName: payload.TokenName,
		Position:  payload.Position,
		IsHome:    payload.IsHome,
	})

	log.Println(payload)
	_ = self.broadcastToOtherSessions(session, payload)
}

func (self *webSocketHandler) onDeleteToken(session *Session, payload message.DeleteToken) {
	// Update the game state with the location of the token
	self.state.DeleteToken(payload.Id, payload.MapName)

	log.Println(payload)
	_ = self.broadcastToOtherSessions(session, payload)
}

func (self *webSocketHandler) onMapSize(payload message.MapSize) {
	log.Println(payload)
	_ = self.broadcastToAllSessions(payload)
}

func (self *webSocketHandler) onMapChange(payload message.MapChange) {
	// Update the game state with the current map.
	self.state.SetCurrentMap(payload.MapName)

	mapDataBytes := getMapData(payload.MapName)

	if len(mapDataBytes) != 0 {
		payload.MapData = mapDataBytes
	}

	log.Println(payload)
	_ = self.broadcastToAllSessions(payload)
}

func getMapData(mapName string) []byte {
	mapDataBytes, err := os.ReadFile("." + string(filepath.Separator) + filepath.Join("data", "maps", mapName+".json"))
	if err != nil {
		err := fmt.Errorf("map data file could not be read: %s", err)
		log.Println(err)
		mapDataBytes = nil
	}

	return mapDataBytes
}
