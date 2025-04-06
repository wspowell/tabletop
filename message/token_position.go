package message

import "github.com/wspowell/tabletop/game"

type TokenPosition struct {
	Id        string          `json:"id"`
	TokenName string          `json:"tokenName"`
	MapName   string          `json:"mapName"`
	Position  game.Coordinate `json:"position"`
	IsHome    bool            `json:"isHome"`
}

func (self TokenPosition) Type() string {
	return "tokenPosition"
}

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}
