package message

import "github.com/wspowell/tabletop/game"

type PlayerHealth struct {
	PlayerHealth map[string]game.Health `json:"playerHealth"`
}

func (self PlayerHealth) Type() string {
	return "playerHealth"
}
