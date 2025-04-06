package game

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	cacheDir      = "." + string(filepath.Separator) + filepath.Join("data", "cache")
	cacheFilename = "." + string(filepath.Separator) + filepath.Join(cacheDir, "tokens.json")
)

type TokenData struct {
	TokenName string     `json:"tokenName"`
	Position  Coordinate `json:"position"`
	IsHome    bool       `json:"isHome"`
}

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type State struct {
	stateMutex sync.Mutex `json:"-"`

	// MapTokens stores location keyed on mapName -> tokenId
	MapTokens  map[string]map[string]TokenData `json:"mapTokens"`
	CurrentMap string                          `json:"currentMap"`
}

func LoadState() (*State, error) {
	state := &State{
		MapTokens:  map[string]map[string]TokenData{},
		stateMutex: sync.Mutex{},
	}

	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		log.Println(err)
		return nil, err
	}

	gameStateBytes, err := os.ReadFile(cacheFilename)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		err := fmt.Errorf("game state file could not be read: %s", err)
		log.Println(err)

		return nil, err
	}

	if len(gameStateBytes) == 0 {
		return state, nil
	}

	if err := json.Unmarshal(gameStateBytes, state); err != nil {
		err := fmt.Errorf("game state file could not be parsed: %s", err)
		log.Println(err)

		return nil, err
	}

	return state, nil
}

func (self *State) Save() error {
	self.stateMutex.Lock()
	defer self.stateMutex.Unlock()

	gameStateBytes, err := json.Marshal(self)
	if err != nil {
		log.Println(err)
		return err
	}

	file, err := os.OpenFile(cacheFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		log.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	// Create a buffered writer
	writer := bufio.NewWriter(file)

	// Write content to the file
	_, err = writer.Write(gameStateBytes)
	if err != nil {
		log.Println("Error writing to file:", err)
		return err
	}

	// Flush and close the writer
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing writer:", err)
		return err
	}

	log.Println("saved game state")

	return nil
}

func (self *State) SetTokenPosition(id string, tokenName string, mapName string, position TokenData) {
	self.stateMutex.Lock()
	if _, exists := self.MapTokens[mapName]; !exists {
		self.MapTokens[mapName] = map[string]TokenData{}
	}
	self.MapTokens[mapName][id] = TokenData{
		TokenName: tokenName,
		Position:  position.Position,
		IsHome:    position.IsHome,
	}
	self.stateMutex.Unlock()
}

func (self *State) DeleteToken(id string, mapName string) {
	self.stateMutex.Lock()
	if _, exists := self.MapTokens[mapName]; exists {
		delete(self.MapTokens[mapName], id)
	}
	self.stateMutex.Unlock()
}

func (self *State) SetCurrentMap(mapName string) {
	self.stateMutex.Lock()
	self.CurrentMap = mapName
	self.stateMutex.Unlock()
}
