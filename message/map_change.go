package message

import "encoding/json"

type MapChange struct {
	MapName string          `json:"mapName"`
	MapData json.RawMessage `json:"mapData,omitempty"`
}

func (self MapChange) Type() string {
	return "mapChange"
}
