package message

type MapSize struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (self MapSize) Type() string {
	return "mapSize"
}
