package message

import (
	"encoding/json"

	"github.com/wspowell/tabletop/game"
)

type State struct {
	*game.State
	MapData      json.RawMessage `json:"mapData,omitempty"`
}

func (self State) Type() string {
	return "state"
}
